package output

import (
	"strings"
	"unicode/utf8"
)

func visibleWidth(s string) int {
	return utf8.RuneCountInString(stripANSI(s))
}

func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inEscape := false
	for i := 0; i < len(s); i++ {
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		if s[i] == '\x1b' {
			inEscape = true
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
