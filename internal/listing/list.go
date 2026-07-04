package listing

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	EstimateDepthMax = iota - 2
	EstimateDepthBounded
)

// ListOptions control directory reads.
type ListOptions struct {
	All           bool
	AlmostAll     bool
	IgnoreBackups bool
	Dereference   bool
	Directory     bool
	Recursive     bool
	ResolveAbs    bool
	LongListing   bool
	ShowInode     bool
	ShowBlocks    bool
	Classify      bool
	DirSlash      bool
	QuoteNames    bool
	Commas        bool
	EstimateSizes bool
	EstimateDepth int
	Precise       bool
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
func List(paths []string, opts ListOptions) ([]Block, error) {
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

func listRecursive(dir string, opts ListOptions) ([]Block, error) {
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

func readDirAt(dir string, opts ListOptions) ([]Entry, error) {
	if opts.Sort.Field == SortByNone {
		return readDirAtUnsorted(dir, opts)
	}

	entries, err := readDirEntries(dir, opts.Sort.Field != SortByNone)
	if err != nil {
		return nil, err
	}

	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if !includeName(e.Name(), opts) {
			continue
		}
		entry, err := classify(dir, e, opts)
		if err != nil {
			return nil, err
		}
		out = append(out, entry)
	}

	if opts.All {
		out = appendDotEntries(dir, out)
	}

	if opts.EstimateSizes {
		estimateDirectorySizes(dir, out, opts.EstimateDepth, opts.Precise)
	}
	sortEntries(out, opts.Sort)
	return out, nil
}

func readDirAtUnsorted(dir string, opts ListOptions) ([]Entry, error) {
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

	if opts.EstimateSizes {
		estimateDirectorySizes(dir, out, opts.EstimateDepth, opts.Precise)
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

func entryFromInfo(fullPath, name string, info os.FileInfo, opts ListOptions) (Entry, error) {
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

func classify(dir string, e fs.DirEntry, opts ListOptions) (Entry, error) {
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
func ReadDir(dir string, opts ListOptions) ([]Entry, error) {
	blocks, err := List([]string{dir}, opts)
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, nil
	}
	return blocks[0].Entries, nil
}

// ResolvePath cleans a path; Abs is only used for interactive display.
func ResolvePath(raw string, resolveAbs bool) string {
	if resolveAbs {
		if abs, err := filepath.Abs(raw); err == nil {
			return abs
		}
	}
	return filepath.Clean(raw)
}
