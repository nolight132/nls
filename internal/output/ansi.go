package output

import (
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

func visibleWidth(s string) int {
	return runewidth.StringWidth(stripANSI(s))
}

func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] == '\x1b' {
			i = skipEscape(s, i)
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// truncateANSI shortens s to at most max display cells, appending an
// ellipsis. Escape sequences are preserved (including trailing resets)
// so truncation never leaks color state into later cells.
func truncateANSI(s string, max int) string {
	if max <= 0 || visibleWidth(s) <= max {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	width := 0
	target := max - 1
	for i := 0; i < len(s); {
		if s[i] == '\x1b' {
			j := skipEscape(s, i)
			b.WriteString(s[i:j])
			i = j
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		w := runewidth.RuneWidth(r)
		if width+w > target {
			i += size
			continue
		}
		b.WriteString(s[i : i+size])
		width += w
		i += size
	}
	b.WriteString("…")
	return b.String()
}

func skipEscape(s string, i int) int {
	if i+1 >= len(s) {
		return i + 1
	}
	switch s[i+1] {
	case '[':
		j := i + 2
		for j < len(s) {
			c := s[j]
			if c >= 0x40 && c <= 0x7E {
				return j + 1
			}
			j++
		}
		return j
	case ']':
		j := i + 2
		for j < len(s) {
			if s[j] == '\x07' {
				return j + 1
			}
			if s[j] == '\x1b' && j+1 < len(s) && s[j+1] == '\\' {
				return j + 2
			}
			j++
		}
		return j
	default:
		return i + 2
	}
}
