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
	Name          string
	Kind          Kind
	Mode          fs.FileMode
	Size          int64
	SizeApprox    bool
	Modified      time.Time
	Accessed      time.Time
	Changed       time.Time
	Permissions   string
	LinkTarget    string
	LinkTargetDir bool
	Inode         uint64
	Blocks        int64
	Links         uint64
	Owner         string
	Group         string
	GitState      GitState
}

func formatPermissions(mode fs.FileMode) string {
	p := mode & fs.ModePerm
	var b strings.Builder
	b.Grow(10)

	b.WriteByte(fileTypeChar(mode))

	writePermTriplet(&b, p, 0o400, 0o200, 0o100, mode&fs.ModeSetuid != 0, 's')
	writePermTriplet(&b, p, 0o040, 0o020, 0o010, mode&fs.ModeSetgid != 0, 's')
	writePermTriplet(&b, p, 0o004, 0o002, 0o001, mode&fs.ModeSticky != 0, 't')

	return b.String()
}

func fileTypeChar(mode fs.FileMode) byte {
	switch {
	case mode&fs.ModeSymlink != 0:
		return 'l'
	case mode&fs.ModeDir != 0:
		return 'd'
	case mode&fs.ModeCharDevice != 0:
		return 'c'
	case mode&fs.ModeDevice != 0:
		return 'b'
	case mode&fs.ModeNamedPipe != 0:
		return 'p'
	case mode&fs.ModeSocket != 0:
		return 's'
	default:
		return '-'
	}
}

func writePermTriplet(b *strings.Builder, mode, r, w, x fs.FileMode, special bool, specialChar byte) {
	if mode&r != 0 {
		b.WriteByte('r')
	} else {
		b.WriteByte('-')
	}
	if mode&w != 0 {
		b.WriteByte('w')
	} else {
		b.WriteByte('-')
	}
	switch {
	case special && mode&x != 0:
		b.WriteByte(specialChar)
	case special:
		b.WriteByte(specialChar &^ 0x20)
	case mode&x != 0:
		b.WriteByte('x')
	default:
		b.WriteByte('-')
	}
}

// ClassifySuffix returns ls -F / -p style suffix.
func ClassifySuffix(kind Kind, mode fs.FileMode, classify, dirSlash bool) string {
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
	}
	switch {
	case mode&fs.ModeSocket != 0:
		return "="
	case mode&fs.ModeNamedPipe != 0:
		return "|"
	case kind == KindExecutable:
		return "*"
	default:
		return ""
	}
}

// DisplayName formats a name for output.
func DisplayName(e Entry, classify, dirSlash, quote, showLinkTarget bool) string {
	name := e.Name
	if showLinkTarget && e.Kind == KindSymlink && e.LinkTarget != "" {
		target := e.LinkTarget
		if e.LinkTargetDir && classify {
			target += "/"
		}
		name = fmt.Sprintf("%s -> %s", e.Name, target)
		if quote {
			return fmt.Sprintf("%q", name)
		}
		return name
	}
	name += ClassifySuffix(e.Kind, e.Mode, classify, dirSlash)
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
	entry, _ := entryFromInfo(target, name, info, ListOptions{})
	return entry
}
