package listing

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"
)

// Kind classifies directory entries for display and sorting.
type Kind int

const (
	KindDirectory Kind = iota
	KindSymlink
	KindFile
	KindExecutable
)

// Entry is a single directory listing row.
type Entry struct {
	Name        string
	Kind        Kind
	Size        int64
	SizeApprox  bool
	Modified    time.Time
	Accessed    time.Time
	Changed     time.Time
	Permissions string
	LinkTarget  string
	Inode       uint64
	Blocks      int64
}

func formatPermissions(mode fs.FileMode) string {
	p := mode & fs.ModePerm
	var b strings.Builder
	b.Grow(10)

	switch {
	case mode&fs.ModeSymlink != 0:
		b.WriteByte('l')
	case mode&fs.ModeDir != 0:
		b.WriteByte('d')
	default:
		b.WriteByte('-')
	}

	writePermTriplet(&b, p, 0o400, 0o200, 0o100, 'r', 'w', 'x')
	writePermTriplet(&b, p, 0o040, 0o020, 0o010, 'r', 'w', 'x')
	writePermTriplet(&b, p, 0o004, 0o002, 0o001, 'r', 'w', 'x')

	return b.String()
}

func writePermTriplet(b *strings.Builder, mode, r, w, x fs.FileMode, rc, wc, xc byte) {
	if mode&r != 0 {
		b.WriteByte(rc)
	} else {
		b.WriteByte('-')
	}
	if mode&w != 0 {
		b.WriteByte(wc)
	} else {
		b.WriteByte('-')
	}
	if mode&x != 0 {
		b.WriteByte(xc)
	} else {
		b.WriteByte('-')
	}
}

// ClassifySuffix returns ls -F / -p style suffix.
func ClassifySuffix(kind Kind, classify, dirSlash bool) string {
	if dirSlash && kind == KindDirectory {
		return "/"
	}
	if !classify {
		return ""
	}
	switch kind {
	case KindDirectory:
		return "/"
	case KindSymlink:
		return "@"
	case KindExecutable:
		return "*"
	default:
		return ""
	}
}

// DisplayName formats a name for output with optional link target.
func DisplayName(e Entry, classify, dirSlash, quote bool) string {
	name := e.Name
	if e.Kind == KindSymlink && e.LinkTarget != "" {
		name = fmt.Sprintf("%s -> %s", e.Name, e.LinkTarget)
	}
	name += ClassifySuffix(e.Kind, classify, dirSlash)
	if quote {
		return fmt.Sprintf("%q", name)
	}
	return name
}

// BlockDotEntry builds a . or .. row.
func BlockDotEntry(target, name string) Entry {
	info, err := os.Stat(target)
	if err != nil {
		return Entry{Name: name, Kind: KindDirectory}
	}
	entry, _ := entryFromInfo(target, name, info, Options{})
	return entry
}
