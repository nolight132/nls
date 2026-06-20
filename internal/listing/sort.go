package listing

import (
	"time"
)

// SortField controls entry ordering.
type SortField int

const (
	SortByName SortField = iota
	SortByTime
	SortBySize
	SortByExtension
	SortByNone
)

// TimeField selects which timestamp -t/-u/-c use.
type TimeField int

const (
	TimeModified TimeField = iota
	TimeAccessed
	TimeChanged
)

// SortOptions configure ordering.
type SortOptions struct {
	Field     SortField
	TimeField TimeField
	Reverse   bool
	DirsFirst bool
}

func sortEntries(entries []Entry, sort SortOptions) {
	if sort.Field == SortByNone {
		return
	}

	for i := 1; i < len(entries); i++ {
		j := i
		for j > 0 && compare(entries[j-1], entries[j], sort) > 0 {
			entries[j-1], entries[j] = entries[j], entries[j-1]
			j--
		}
	}
}

func compare(a, b Entry, sort SortOptions) int {
	if sort.DirsFirst {
		switch {
		case a.Kind == KindDirectory && b.Kind != KindDirectory:
			if sort.Reverse {
				return 1
			}
			return -1
		case b.Kind == KindDirectory && a.Kind != KindDirectory:
			if sort.Reverse {
				return -1
			}
			return 1
		}
	}

	var cmp int
	switch sort.Field {
	case SortByTime:
		at, bt := entryTime(a, sort.TimeField), entryTime(b, sort.TimeField)
		switch {
		case at.After(bt):
			cmp = -1
		case at.Before(bt):
			cmp = 1
		}
	case SortBySize:
		switch {
		case a.Size > b.Size:
			cmp = -1
		case a.Size < b.Size:
			cmp = 1
		}
	case SortByExtension:
		ea, eb := extensionKey(a.Name), extensionKey(b.Name)
		switch {
		case ea < eb:
			cmp = -1
		case ea > eb:
			cmp = 1
		default:
			if a.Name < b.Name {
				cmp = -1
			} else if a.Name > b.Name {
				cmp = 1
			}
		}
	default:
		if a.Name < b.Name {
			cmp = -1
		} else if a.Name > b.Name {
			cmp = 1
		}
	}

	if sort.Reverse {
		cmp = -cmp
	}
	return cmp
}

func entryTime(e Entry, field TimeField) time.Time {
	switch field {
	case TimeAccessed:
		if !e.Accessed.IsZero() {
			return e.Accessed
		}
	case TimeChanged:
		if !e.Changed.IsZero() {
			return e.Changed
		}
	}
	return e.Modified
}

func extensionKey(name string) string {
	if i := lastDot(name); i > 0 {
		return name[i:]
	}
	return ""
}

func lastDot(name string) int {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return i
		}
	}
	return -1
}

// EntryTime returns the timestamp used for sorting/display field selection.
func EntryTime(e Entry, field TimeField) time.Time {
	return entryTime(e, field)
}
