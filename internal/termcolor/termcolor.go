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
	var styler *lscolor.Styler
	if enabled {
		styler = lscolor.New()
	}
	return &Style{enabled: enabled, styler: styler}
}

func (s *Style) sprint(c *color.Color, value string) string {
	if !s.enabled {
		return value
	}
	c.EnableColor()
	return c.Sprint(value)
}

// Heading colors the heading text.
func (s *Style) Heading(value string) string {
	return s.sprint(color.New(color.FgBlue), value)
}

func (s *Style) Header(value string) string {
	return s.sprint(color.New(color.FgBlue, color.Bold), value)
}

// Index colors the index column
func (s *Style) Index(value string) string {
	return s.sprint(color.New(color.FgBlue, color.Bold), value)
}

// Name colors a filename using LS_COLORS-compatible rules.
func (s *Style) Name(name string, kind listing.Kind) string {
	if !s.enabled || s.styler == nil {
		return name
	}
	return s.styler.Colorize(name, kind)
}

// NameGit colors a filename by its git state, falling back to the
// LS_COLORS rules for clean entries and entries outside a repo.
func (s *Style) NameGit(name string, kind listing.Kind, state listing.GitState) string {
	switch state {
	case listing.GitStateUntracked:
		return s.sprint(color.New(color.FgHiGreen), name)
	case listing.GitStateModified:
		return s.sprint(color.New(color.FgYellow), name)
	case listing.GitStateIgnored:
		return s.sprint(color.New(color.FgHiBlack), name)
	default:
		return s.Name(name, kind)
	}
}

// Size colors the size column
func (s *Style) Size(value string) string {
	return s.sprint(color.New(color.FgCyan), value)
}

// Modified colors the modified column
func (s *Style) Modified(value string) string {
	if !s.enabled || value == "-" {
		return value
	}
	return s.sprint(color.New(color.FgMagenta), value)
}

// Error colors error text red.
func (s *Style) Error(msg string) string {
	return s.sprint(color.New(color.FgRed), msg)
}

// Empty colors empty table messages.
func (s *Style) Empty(msg string) string {
	return s.sprint(color.New(color.FgHiBlack), msg)
}
