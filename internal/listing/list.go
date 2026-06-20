package listing

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Options control directory reads.
type Options struct {
	All              bool
	AlmostAll        bool
	IgnoreBackups    bool
	Dereference      bool
	Directory        bool
	Recursive        bool
	EstimateDirSizes bool
	FastPath         bool
	ResolveAbs       bool
	LongListing      bool
	ShowInode        bool
	ShowBlocks       bool
	Classify         bool
	DirSlash         bool
	QuoteNames       bool
	Commas           bool
	Sort             SortOptions
}

// List resolves paths into output blocks.
func List(paths []string, opts Options) ([]Block, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var blocks []Block
	for _, raw := range paths {
		path := ResolvePath(raw, opts.ResolveAbs)

		info, err := statPath(path, opts.Dereference)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", raw, err)
		}

		if !info.IsDir() {
			entry, err := entryFromInfo(path, filepath.Base(path), info, opts)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, Block{Entries: []Entry{entry}})
			continue
		}

		if opts.Directory {
			entry, err := entryFromInfo(path, filepath.Base(path), info, opts)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, Block{Entries: []Entry{entry}})
			continue
		}

		if opts.Recursive {
			rec, err := listRecursive(path, opts)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", raw, err)
			}
			blocks = append(blocks, rec...)
			continue
		}

		entries, err := readDirAt(path, opts)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", raw, err)
		}
		header := ""
		if len(paths) > 1 {
			header = path
		}
		blocks = append(blocks, Block{Header: header, Entries: entries})
	}

	return blocks, nil
}

func listRecursive(dir string, opts Options) ([]Block, error) {
	entries, err := readDirAt(dir, opts)
	if err != nil {
		return nil, err
	}

	blocks := []Block{{Header: dir + string(os.PathSeparator), Entries: entries}}
	for _, e := range entries {
		if e.Kind != KindDirectory {
			continue
		}
		if e.Name == "." || e.Name == ".." {
			continue
		}
		sub, err := listRecursive(filepath.Join(dir, e.Name), opts)
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

func readDirAt(dir string, opts Options) ([]Entry, error) {
	entries, err := os.ReadDir(dir)
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

	if opts.EstimateDirSizes {
		estimateDirectorySizes(dir, out)
	}
	sortEntries(out, opts.Sort)
	return out, nil
}

func appendDotEntries(dir string, entries []Entry) []Entry {
	dot := BlockDotEntry(dir, ".")
	dotdot := BlockDotEntry(filepath.Dir(dir), "..")
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
	}

	mode := info.Mode()
	switch {
	case mode&os.ModeSymlink != 0 && !opts.Dereference:
		entry.Kind = KindSymlink
		target, err := os.Readlink(fullPath)
		if err == nil {
			entry.LinkTarget = target
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

func blocksOf(info os.FileInfo) int64 {
	const blockSize = 512
	if info.Size() == 0 {
		return 0
	}
	return (info.Size() + blockSize - 1) / blockSize
}
