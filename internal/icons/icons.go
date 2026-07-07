package icons

import (
	"github.com/nolight132/nls/internal/listing"
)

// Set controls which glyphs are used for entry types.
type Set int

const (
	SetNone Set = iota
	SetNerdBasic
	SetNerd
)

// Resolve picks an icon set from flags and config.
// Precedence: noIcons flag > configEnabled > config default (on).
func Resolve(noIcons bool, configEnabled bool, specialIcons bool) Set {
	if noIcons {
		return SetNone
	}
	if !configEnabled {
		return SetNone
	}
	if !specialIcons {
		return SetNerdBasic
	}
	return SetNerd
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
