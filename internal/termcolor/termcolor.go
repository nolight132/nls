package termcolor

import (
	"github.com/fatih/color"
	"github.com/nolight132/nls/internal/listing"
	"github.com/nolight132/nls/internal/lscolor"
)

// Style applies ANSI colors.
type Style struct {
	enabled bool
	styler  *lscolor.Styler
}

// New returns a color style helper.
func New(enabled bool) *Style {
	if !enabled {
		color.NoColor = true
	} else {
		color.NoColor = false
	}
	var styler *lscolor.Styler
	if enabled {
		styler = lscolor.New()
	}
	return &Style{enabled: enabled, styler: styler}
}

// Heading colors the heading text.
func (s *Style) Heading(value string) string {
	if !s.enabled {
		return value
	}
	return color.New(color.FgGreen).Sprint(value)
}

func (s *Style) Header(value string) string {
	if !s.enabled {
		return value
	}
	return color.New(color.FgGreen, color.Bold).Sprint(value)
}

// Index colors the index column
func (s *Style) Index(value string) string {
	if !s.enabled {
		return value
	}
	return color.New(color.FgGreen, color.Bold).Sprint(value)
}

// Name colors a filename using LS_COLORS-compatible rules.
func (s *Style) Name(name string, kind listing.Kind) string {
	if !s.enabled || s.styler == nil {
		return name
	}
	return s.styler.Colorize(name, kind)
}

// Size colors the size column
func (s *Style) Size(value string) string {
	if !s.enabled {
		return value
	}
	return color.New(color.FgCyan).Sprint(value)
}

// Modified colors the modified column
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

// Empty colors empty table messages.
func (s *Style) Empty(msg string) string {
	if !s.enabled {
		return msg
	}
	return color.New(color.FgHiBlack).Sprint(msg)
}
