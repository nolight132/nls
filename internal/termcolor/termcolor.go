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

var (
	headingColor   = color.New(color.FgBlue)
	headerColor    = color.New(color.FgBlue, color.Bold)
	untrackedColor = color.New(color.FgHiGreen)
	ignoredColor   = color.New(color.FgHiBlack)
	dirtyColor     = color.New(color.FgYellow)
	sizeColor      = color.New(color.FgCyan)
	modifiedColor  = color.New(color.FgMagenta)
	errorColor     = color.New(color.FgRed)
	emptyColor     = color.New(color.FgHiBlack)
)

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
	return s.sprint(headingColor, value)
}

func (s *Style) Header(value string) string {
	return s.sprint(headerColor, value)
}

// Index colors the index column
func (s *Style) Index(value string) string {
	return s.sprint(headerColor, value)
}

// Name colors a filename using LS_COLORS-compatible rules.
func (s *Style) Name(name string, kind listing.Kind) string {
	if !s.enabled {
		return name
	}
	return s.styler.Colorize(name, kind)
}

// NameGit colors a filename by its git state, falling back to the
// LS_COLORS rules for clean entries and entries outside a repo.
func (s *Style) NameGit(name string, kind listing.Kind, state listing.GitState) string {
	clean := listing.GitState{Staging: listing.StatusUnmodified, Worktree: listing.StatusUnmodified}
	switch {
	case state.Staging == listing.StatusUntracked && state.Worktree == listing.StatusUntracked:
		return s.sprint(untrackedColor, name)
	case state.Staging == listing.StatusIgnored || state.Worktree == listing.StatusIgnored:
		return s.sprint(ignoredColor, name)
	case state != clean && state != (listing.GitState{}):
		return s.sprint(dirtyColor, name)
	default:
		return s.Name(name, kind)
	}
}

// Size colors the size column
func (s *Style) Size(value string) string {
	return s.sprint(sizeColor, value)
}

// Modified colors the modified column
func (s *Style) Modified(value string) string {
	if !s.enabled || value == "-" {
		return value
	}
	return s.sprint(modifiedColor, value)
}

// Error colors error text red.
func (s *Style) Error(msg string) string {
	return s.sprint(errorColor, msg)
}

// Empty colors empty table messages.
func (s *Style) Empty(msg string) string {
	return s.sprint(emptyColor, msg)
}
