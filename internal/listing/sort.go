package listing

import (
	"os"
	"slices"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
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
	names := newNameComparer()
	slices.SortStableFunc(entries, func(a, b Entry) int {
		return compare(a, b, sort, names)
	})
}

func compare(a, b Entry, sort SortOptions, names nameComparer) int {
	if sort.DirsFirst {
		switch {
		case sortGroupDir(a) && !sortGroupDir(b):
			return -1
		case sortGroupDir(b) && !sortGroupDir(a):
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
		case names.compare(ea, eb) < 0:
			cmp = -1
		case names.compare(ea, eb) > 0:
			cmp = 1
		default:
			if names.compare(a.Name, b.Name) < 0 {
				cmp = -1
			} else if names.compare(a.Name, b.Name) > 0 {
				cmp = 1
			}
		}
	default:
		if names.compare(a.Name, b.Name) < 0 {
			cmp = -1
		} else if names.compare(a.Name, b.Name) > 0 {
			cmp = 1
		}
	}
	if cmp == 0 && sort.Field != SortByName {
		cmp = names.compare(a.Name, b.Name)
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

func sortGroupDir(e Entry) bool {
	return e.Kind == KindDirectory || e.LinkTargetDir
}

type nameComparer struct {
	collator *collate.Collator
}

func newNameComparer() nameComparer {
	tag, ok := collationLanguage()
	if !ok {
		return nameComparer{}
	}
	return nameComparer{collator: collate.New(tag)}
}

func (c nameComparer) compare(a, b string) int {
	if c.collator != nil {
		ak, bk := localeSortKey(a), localeSortKey(b)
		if cmp := c.collator.CompareString(ak, bk); cmp != 0 {
			return cmp
		}
	}
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func localeSortKey(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func collationLanguage() (language.Tag, bool) {
	locale := collationLocale()
	base := localeBase(locale)
	if base == "" || base == "C" || base == "POSIX" {
		return language.Und, false
	}
	tag, err := language.Parse(strings.ReplaceAll(base, "_", "-"))
	if err != nil {
		return language.Und, true
	}
	return tag, true
}

func collationLocale() string {
	for _, key := range []string{"LC_ALL", "LC_COLLATE", "LANG"} {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return "C"
}

func localeBase(locale string) string {
	if i := strings.IndexAny(locale, ".@"); i >= 0 {
		return locale[:i]
	}
	return locale
}
