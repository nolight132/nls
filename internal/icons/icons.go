package icons

import (
	"os"
	"strings"

	"github.com/nolight132/nls/internal/listing"
)

// Set controls which glyphs are used for entry types.
type Set int

const (
	SetNone Set = iota
	SetNerdBasic
	SetNerd
)

// Resolve picks an icon set from flags, config, and environment.
// Precedence: noIcons flag > NLS_ICONS env > configEnabled > default off.
// Icons are off by default to match Nushell ls.
func Resolve(noIcons bool, configEnabled bool, specialIcons bool) Set {
	if noIcons {
		return SetNone
	}
	if !configEnabled && !envIconsOn() {
		return SetNone
	}
	if !nerdFontAvailable() {
		return SetNone
	}
	if !specialIcons {
		return SetNerdBasic
	}
	return SetNerd
}

func envIconsOn() bool {
	switch strings.ToLower(os.Getenv("NLS_ICONS")) {
	case "1", "true", "yes", "on", "nerd":
		return true
	default:
		return false
	}
}

func nerdFontAvailable() bool {
	for _, key := range []string{"NERD_FONT", "NLS_NERD_FONT"} {
		switch strings.ToLower(os.Getenv(key)) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	}

	term := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(term, "alacritty") ||
		strings.Contains(term, "kitty") ||
		strings.Contains(term, "wezterm") ||
		strings.Contains(term, "ghostty") {
		return true
	}

	for _, key := range []string{"FONT", "FONTFACE", "FONT_FAMILY"} {
		if strings.Contains(strings.ToLower(os.Getenv(key)), "nerd") {
			return true
		}
	}

	return true
}

// For returns the icon for an entry kind.
func For(entry listing.Entry, set Set) string {
	if set != SetNerd {
		if set == SetNerdBasic {
			return basicIcon(entry.Kind)
		}
		return ""
	}
	icon := MatchIcon(entry.Name)
	if icon != "" {
		return icon
	}
	return basicIcon(entry.Kind)
}
