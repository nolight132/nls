package listing

import (
	"io/fs"
	"os"
	"path/filepath"
)

// NeedsFullMetadata reports whether listing must stat every entry.
func NeedsFullMetadata(opts Options) bool {
	if opts.EstimateDirSizes || opts.Recursive || opts.Directory || opts.Dereference {
		return true
	}
	if opts.All || opts.AlmostAll {
		return true
	}
	if opts.LongListing || opts.ShowInode || opts.ShowBlocks {
		return true
	}
	s := opts.Sort
	if s.DirsFirst {
		return true
	}
	if s.Field == SortByTime || s.Field == SortBySize || s.Field == SortByExtension {
		return true
	}
	return opts.Classify || opts.DirSlash
}

func classifyFast(dir string, e fs.DirEntry, opts Options) (Entry, error) {
	if needsEntryStat(opts) {
		return classify(dir, e, opts)
	}

	entry := Entry{Name: e.Name()}
	if opts.Classify || opts.DirSlash {
		entry.Kind = kindFromType(e.Type())
	}
	return entry, nil
}

func needsEntryStat(opts Options) bool {
	return opts.LongListing || opts.ShowInode || opts.ShowBlocks ||
		opts.Sort.Field == SortByTime || opts.Sort.Field == SortBySize
}

func kindFromType(mode fs.FileMode) Kind {
	switch {
	case mode&fs.ModeSymlink != 0:
		return KindSymlink
	case mode.IsDir():
		return KindDirectory
	case mode&0o111 != 0:
		return KindExecutable
	default:
		return KindFile
	}
}

func appendDotEntriesFast(dir string, entries []Entry, fullMeta bool) []Entry {
	if fullMeta {
		return appendDotEntries(dir, entries)
	}
	return append([]Entry{{Name: ".", Kind: KindDirectory}, {Name: "..", Kind: KindDirectory}}, entries...)
}

// FastListNames reads directory names with minimal work (native ls speed).
func FastListNames(dir string, opts Options) ([]string, error) {
	info, err := os.Lstat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{filepath.Clean(dir)}, nil
	}

	entries, err := readDirEntries(dir, opts.Sort.Field != SortByNone)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if includeName(e.Name(), opts) {
			names = append(names, e.Name())
		}
	}

	if opts.All {
		names = append([]string{".", ".."}, names...)
	}

	if opts.Sort.Field != SortByNone {
		sortNames(names, opts.Sort)
	}
	return names, nil
}

func sortNames(names []string, sort SortOptions) {
	nameCmp := newNameComparer()
	for i := 1; i < len(names); i++ {
		j := i
		for j > 0 && compareNamesWithComparer(names[j-1], names[j], sort, nameCmp) > 0 {
			names[j-1], names[j] = names[j], names[j-1]
			j--
		}
	}
}

func compareNames(a, b string, sort SortOptions) int {
	return compareNamesWithComparer(a, b, sort, newNameComparer())
}

func compareNamesWithComparer(a, b string, sort SortOptions, names nameComparer) int {
	cmp := names.compare(a, b)
	if sort.Reverse {
		cmp = -cmp
	}
	return cmp
}

// CanFastList reports whether the ultra-light name-only path is valid.
func CanFastList(opts Options) bool {
	if NeedsFullMetadata(opts) {
		return false
	}
	if opts.Recursive || opts.Directory {
		return false
	}
	if opts.Classify || opts.DirSlash || opts.QuoteNames {
		return false
	}
	if opts.LongListing || opts.Commas || opts.ShowInode || opts.ShowBlocks {
		return false
	}
	if opts.Sort.Field != SortByName && opts.Sort.Field != SortByNone {
		return false
	}
	if opts.Sort.DirsFirst {
		return false
	}
	return true
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
