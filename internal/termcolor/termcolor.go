package termcolor

import (
	"github.com/fatih/color"
	"github.com/nolight132/nls/internal/listing"
	"github.com/nolight132/nls/internal/lscolor"
)

// Style applies Nushell-like ANSI colors.
type Style struct {
	enabled bool
	ls      *lscolor.Styler
}

// New returns a color style helper.
func New(enabled bool) *Style {
	if !enabled {
		color.NoColor = true
	} else {
		color.NoColor = false
	}
	var ls *lscolor.Styler
	if enabled {
		ls = lscolor.New()
	}
	return &Style{enabled: enabled, ls: ls}
}

// Name colors a filename using LS_COLORS-compatible rules.
func (s *Style) Name(name string, kind listing.Kind) string {
	if !s.enabled || s.ls == nil {
		return name
	}
	return s.ls.Colorize(name, kind)
}

// Size colors the size column (Nushell filesize: cyan).
func (s *Style) Size(value string) string {
	if !s.enabled {
		return value
	}
	return color.New(color.FgCyan).Sprint(value)
}

// Modified colors the modified column (Nushell datetime: purple).
func (s *Style) Modified(value string) string {
	if !s.enabled || value == "-" {
		return value
	}
	return color.New(color.FgMagenta).Sprint(value)
}

// Error colors error text red.
func (s *Style) Error(msg string) string {
	if !s.enabled {
		return msg
	}
	return color.New(color.FgRed).Sprint(msg)
}
