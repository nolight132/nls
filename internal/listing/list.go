package listing

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const (
	EstimateDepthMax = iota - 2
	EstimateDepthOff
	EstimateDepthBounded
)

// Limits holds the concrete budgets for bounded directory size estimation.
// They are only consulted when EstimateDepth == EstimateDepthBounded.
type Limits struct {
	// WalkDuration caps a single directory walk.
	WalkDuration time.Duration
	// ListingDuration caps the total estimate work for one listing.
	ListingDuration time.Duration
	// MaxWalkEntries caps the number of entries walked per directory.
	MaxWalkEntries int
	// MaxDirsPerListing caps how many directories get estimated at all.
	MaxDirsPerListing int
	// MaxDepth caps walk depth; 0 means unlimited within the time budget.
	MaxDepth int
}

// DefaultBoundedLimits returns the balanced budgets used when no user config
// is present. These match the historical constants.
func DefaultBoundedLimits() Limits {
	return Limits{
		WalkDuration:      maxDirWalkDuration,
		ListingDuration:   maxListingEstimate,
		MaxWalkEntries:    maxDirWalkEntries,
		MaxDirsPerListing: maxDirsPerListingDefault,
	}
}

// SafetyLimits returns generous caps applied to --estimate-depth max so that
// full-walk mode cannot hang on huge filesystems like /. No time limits —
// only an entry-count cap high enough to fully scan a typical home directory.
func SafetyLimits() Limits {
	return Limits{
		WalkDuration:      0,
		ListingDuration:   0,
		MaxWalkEntries:    200000,
		MaxDirsPerListing: 50,
		MaxDepth:          0,
	}
}

// Options control directory reads.
type Options struct {
	All           bool
	AlmostAll     bool
	IgnoreBackups bool
	Dereference   bool
	Directory     bool
	Recursive     bool
	FastPath      bool
	ResolveAbs    bool
	LongListing   bool
	ShowInode     bool
	ShowBlocks    bool
	Classify      bool
	DirSlash      bool
	QuoteNames    bool
	Commas        bool
	EstimateDepth int
	// BoundedLimits applies when EstimateDepth == EstimateDepthBounded.
	// A zero value falls back to DefaultBoundedLimits().
	BoundedLimits Limits
	Sort          SortOptions
}

type operand struct {
	raw     string
	path    string
	display string
	info    os.FileInfo
	entry   Entry
}

// List resolves paths into output blocks.
func List(paths []string, opts Options) ([]Block, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	operands := make([]operand, 0, len(paths))
	for _, raw := range paths {
		path := ResolvePath(raw, opts.ResolveAbs)
		displayPath := filepath.Clean(raw)

		info, err := statPath(path, opts.Dereference)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", raw, err)
		}
		entry, err := entryFromInfo(path, displayPath, info, opts)
		if err != nil {
			return nil, err
		}
		operands = append(operands, operand{raw: raw, path: path, display: displayPath, info: info, entry: entry})
	}

	if len(operands) == 1 {
		op := operands[0]
		if !op.info.IsDir() || opts.Directory {
			return []Block{{Entries: []Entry{op.entry}}}, nil
		}
		if opts.Recursive {
			rec, err := listRecursive(op.path, opts)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op.raw, err)
			}
			return rec, nil
		}
		entries, err := readDirAt(op.path, opts)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op.raw, err)
		}
		return []Block{{Entries: entries, Directory: true}}, nil
	}

	var files []Entry
	var dirs []operand
	for _, op := range operands {
		if !op.info.IsDir() || opts.Directory {
			files = append(files, op.entry)
			continue
		}
		dirs = append(dirs, op)
	}
	sortEntries(files, opts.Sort)
	sortOperands(dirs, opts.Sort)

	var blocks []Block
	if len(files) > 0 {
		blocks = append(blocks, Block{Entries: files})
	}
	for _, op := range dirs {
		if opts.Recursive {
			rec, err := listRecursive(op.path, opts)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op.raw, err)
			}
			blocks = append(blocks, rec...)
			continue
		}

		entries, err := readDirAt(op.path, opts)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op.raw, err)
		}
		blocks = append(blocks, Block{Header: op.path, Entries: entries, Directory: true})
	}

	return blocks, nil
}

func sortOperands(operands []operand, sort SortOptions) {
	if sort.Field == SortByNone {
		return
	}
	names := newNameComparer()
	for i := 1; i < len(operands); i++ {
		j := i
		for j > 0 && compare(operands[j-1].entry, operands[j].entry, sort, names) > 0 {
			operands[j-1], operands[j] = operands[j], operands[j-1]
			j--
		}
	}
}

func listRecursive(dir string, opts Options) ([]Block, error) {
	entries, err := readDirAt(dir, opts)
	if err != nil {
		return nil, err
	}

	blocks := []Block{{Header: dir, Entries: entries, Directory: true}}
	for _, e := range entries {
		if e.Kind != KindDirectory {
			continue
		}
		if e.Name == "." || e.Name == ".." {
			continue
		}
		sub, err := listRecursive(childPath(dir, e.Name), opts)
		if err != nil {
			if os.IsPermission(err) {
				continue
			}
			return nil, err
		}
		blocks = append(blocks, sub...)
	}
	return blocks, nil
}

func childPath(dir, name string) string {
	if dir == "." {
		return "." + string(os.PathSeparator) + name
	}
	return filepath.Join(dir, name)
}

func readDirAt(dir string, opts Options) ([]Entry, error) {
	if opts.Sort.Field == SortByNone {
		return readDirAtUnsorted(dir, opts)
	}

	entries, err := readDirEntries(dir, opts.Sort.Field != SortByNone)
	if err != nil {
		return nil, err
	}

	out := make([]Entry, 0, len(entries))
	fullMeta := !opts.FastPath
	for _, e := range entries {
		if !includeName(e.Name(), opts) {
			continue
		}
		var entry Entry
		var err error
		if opts.FastPath {
			entry, err = classifyFast(dir, e, opts)
		} else {
			entry, err = classify(dir, e, opts)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, entry)
	}

	if opts.All {
		out = appendDotEntriesFast(dir, out, fullMeta)
	}

	if opts.EstimateDepth != EstimateDepthOff {
		estimateDirectorySizes(dir, out, opts.EstimateDepth, opts.BoundedLimits)
	}
	sortEntries(out, opts.Sort)
	return out, nil
}

func readDirAtUnsorted(dir string, opts Options) ([]Entry, error) {
	names, err := readDirNamesUnsorted(dir)
	if err != nil {
		return nil, err
	}

	out := make([]Entry, 0, len(names))
	for _, name := range names {
		if !includeName(name, opts) {
			continue
		}
		full := childPath(dir, name)
		info, err := statPath(full, opts.Dereference)
		if err != nil {
			return nil, fmt.Errorf("stat %q: %w", name, err)
		}
		entry, err := entryFromInfo(full, name, info, opts)
		if err != nil {
			return nil, err
		}
		out = append(out, entry)
	}

	if opts.EstimateDepth != EstimateDepthOff {
		estimateDirectorySizes(dir, out, opts.EstimateDepth, opts.BoundedLimits)
	}
	return out, nil
}

func readDirEntries(dir string, sorted bool) ([]fs.DirEntry, error) {
	if sorted {
		return os.ReadDir(dir)
	}
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.ReadDir(-1)
}

func appendDotEntries(dir string, entries []Entry) []Entry {
	dot := BlockDotEntry(dir, ".")
	dotdot := BlockDotEntry(childPath(dir, ".."), "..")
	return append([]Entry{dot, dotdot}, entries...)
}

func statPath(path string, dereference bool) (os.FileInfo, error) {
	if dereference {
		return os.Stat(path)
	}
	return os.Lstat(path)
}

func entryFromInfo(fullPath, name string, info os.FileInfo, opts Options) (Entry, error) {
	accessed, changed := fileTimes(info)
	entry := Entry{
		Name:        name,
		Size:        info.Size(),
		Modified:    info.ModTime(),
		Accessed:    accessed,
		Changed:     changed,
		Permissions: formatPermissions(info.Mode()),
		Inode:       inodeOf(info),
		Blocks:      blocksOf(info),
		Links:       linksOf(info),
	}
	entry.Owner, entry.Group = ownerGroupOf(info)

	mode := info.Mode()
	switch {
	case mode&os.ModeSymlink != 0 && !opts.Dereference:
		entry.Kind = KindSymlink
		target, err := os.Readlink(fullPath)
		if err == nil {
			entry.LinkTarget = target
		}
		if targetInfo, err := os.Stat(fullPath); err == nil && targetInfo.IsDir() {
			entry.LinkTargetDir = true
		}
	case info.IsDir():
		entry.Kind = KindDirectory
	case mode&0o111 != 0:
		entry.Kind = KindExecutable
	default:
		entry.Kind = KindFile
	}

	return entry, nil
}

func classify(dir string, e fs.DirEntry, opts Options) (Entry, error) {
	full := filepath.Join(dir, e.Name())
	info, err := entryInfo(full, e, opts.Dereference)
	if err != nil {
		return Entry{}, fmt.Errorf("stat %q: %w", e.Name(), err)
	}
	entry, err := entryFromInfo(full, e.Name(), info, opts)
	if err != nil {
		return entry, err
	}
	if opts.Dereference && e.Type()&fs.ModeSymlink != 0 && info.IsDir() {
		entry.Kind = KindDirectory
	}
	return entry, nil
}

func entryInfo(path string, e fs.DirEntry, dereference bool) (os.FileInfo, error) {
	if dereference {
		return os.Stat(path)
	}
	return e.Info()
}

// ReadDir lists one directory. Prefer List for full flag support.
func ReadDir(dir string, opts Options) ([]Entry, error) {
	blocks, err := List([]string{dir}, opts)
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, nil
	}
	return blocks[0].Entries, nil
}
