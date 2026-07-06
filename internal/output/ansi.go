package output

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

func visibleWidth(s string) int {
	return runewidth.StringWidth(stripANSI(s))
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
