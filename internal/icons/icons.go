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
	SetEmoji
	SetNerd
)

// Resolve picks an icon set from flags and environment.
// Icons are off by default to match Nushell ls.
func Resolve(noIcons bool) Set {
	if noIcons {
		return SetNone
	}
	switch strings.ToLower(os.Getenv("NLS_ICONS")) {
	case "1", "true", "yes", "on", "nerd":
		if nerdFontAvailable() {
			return SetNerd
		}
		return SetEmoji
	case "emoji":
		return SetEmoji
	default:
		return SetNone
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
func For(kind listing.Kind, set Set) string {
	switch set {
	case SetNone:
		return ""
	case SetNerd:
		return nerdIcon(kind)
	default:
		return emojiIcon(kind)
	}
}

func nerdIcon(kind listing.Kind) string {
	switch kind {
	case listing.KindDirectory:
		return "\uf07b "
	case listing.KindSymlink:
		return "\uf0c1 "
	case listing.KindExecutable:
		return "\uf013 "
	default:
		return "\uf15b "
	}
}

func emojiIcon(kind listing.Kind) string {
	switch kind {
	case listing.KindDirectory:
		return "📁 "
	case listing.KindSymlink:
		return "🔗 "
	case listing.KindExecutable:
		return "⚙️ "
	default:
		return "📄 "
	}
}
