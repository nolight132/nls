package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/nolight132/nls/internal/listing"
)

// TimingPreset names the trade-off curve for bounded directory size estimates.
// Concrete budgets are derived in code; the TOML file only ever stores the name.
type TimingPreset string

const (
	// TimingStrict caps walks aggressively so interactive listings stay snappy
	// on huge trees. Sizes are more likely to be marked approximate.
	TimingStrict TimingPreset = "strict"
	// TimingBalanced is the default curve: enough headroom for typical repos
	// without making large directories feel sluggish.
	TimingBalanced TimingPreset = "balanced"
	// TimingRelaxed allows much longer walks before truncating, useful when
	// accurate directory sizes matter more than latency.
	TimingRelaxed TimingPreset = "relaxed"
)

// presetLimits returns the concrete budgets for a preset.
func presetLimits(p TimingPreset) (listing.Limits, error) {
	switch p {
	case TimingStrict:
		return listing.Limits{
			WalkDuration:      25 * time.Millisecond,
			ListingDuration:   60 * time.Millisecond,
			MaxWalkEntries:    200,
			MaxDirsPerListing: 4,
		}, nil
	case TimingBalanced:
		return listing.DefaultBoundedLimits(), nil
	case TimingRelaxed:
		return listing.Limits{
			WalkDuration:      200 * time.Millisecond,
			ListingDuration:   500 * time.Millisecond,
			MaxWalkEntries:    2000,
			MaxDirsPerListing: 12,
		}, nil
	default:
		return listing.Limits{}, fmt.Errorf("unknown timing preset %q (want strict, balanced, or relaxed)", p)
	}
}

// DirSizeConfig controls bounded directory size estimation defaults, used
// when --estimate-depth is not passed on a TTY.
type DirSizeConfig struct {
	// DefaultDepth caps how deep the bounded walk goes per directory.
	// 0 means unlimited depth, bounded only by the timing preset.
	DefaultDepth int `toml:"default_depth"`
	// Timing selects the budget preset. Empty is treated as "balanced".
	Timing TimingPreset `toml:"timing"`
}

// ColumnEntry names a single column in the listing layout.
type ColumnEntry string

const (
	// Base table columns.
	ColumnId   ColumnEntry = "id"
	ColumnName ColumnEntry = "name"
	ColumnType ColumnEntry = "type"
	ColumnSize ColumnEntry = "size"

	// Time fields. modified is the default; accessed/changed apply with
	// --access-time/-u and --ctime/-c.
	ColumnModified ColumnEntry = "modified"
	ColumnAccessed ColumnEntry = "accessed"
	ColumnChanged  ColumnEntry = "changed"

	// Long listing (-l) metadata.
	ColumnPermissions ColumnEntry = "permissions"
	ColumnLinks       ColumnEntry = "links"
	ColumnOwner       ColumnEntry = "owner"
	ColumnGroup       ColumnEntry = "group"

	// Optional counts, shown with -i and -s.
	ColumnInode  ColumnEntry = "inode"
	ColumnBlocks ColumnEntry = "blocks"
)

// Config is the nls user configuration loaded from XDG paths.
type Config struct {
	// Icons enables Nerd Font icons by default. --no-icons and NLS_ICONS
	// still override this.
	Icons bool `toml:"icons"`
	// DirSize holds defaults for bounded directory size estimation.
	DirSize DirSizeConfig `toml:"dir_size"`
	// Layout holds defaults for the layout of the listing.
	DefaultColumns []ColumnEntry `toml:"default_columns"`
}

// Defaults returns the configuration used when no file is present.
func Defaults() Config {
	return Config{
		Icons: false,
		DirSize: DirSizeConfig{
			DefaultDepth: 0,
			Timing:       TimingBalanced,
		},
		DefaultColumns: []ColumnEntry{
			ColumnId,
			ColumnName,
			ColumnType,
			ColumnSize,
			ColumnModified,
		},
	}
}

// Resolve applies defaults for missing fields and validates the result.
func (c Config) Resolve() (Config, error) {
	resolved := Defaults()
	if c.Icons {
		resolved.Icons = true
	}
	if c.DirSize.DefaultDepth > 0 {
		resolved.DirSize.DefaultDepth = c.DirSize.DefaultDepth
	}
	if c.DirSize.Timing != "" {
		resolved.DirSize.Timing = c.DirSize.Timing
	}
	if _, err := presetLimits(resolved.DirSize.Timing); err != nil {
		return resolved, err
	}
	if len(c.DefaultColumns) > 0 {
		for _, col := range c.DefaultColumns {
			if !isValidColumn(col) {
				return resolved, fmt.Errorf("unknown column %q in default_columns", col)
			}
		}
		resolved.DefaultColumns = c.DefaultColumns
	}
	return resolved, nil
}

func isValidColumn(c ColumnEntry) bool {
	switch c {
	case ColumnId, ColumnName, ColumnType, ColumnSize,
		ColumnModified, ColumnAccessed, ColumnChanged,
		ColumnPermissions, ColumnLinks, ColumnOwner, ColumnGroup,
		ColumnInode, ColumnBlocks:
		return true
	}
	return false
}

// Limits returns the concrete budgets for the configured timing preset.
// Falls back to balanced on any error.
func (c Config) Limits() listing.Limits {
	limits, err := presetLimits(c.DirSize.Timing)
	if err != nil {
		limits, _ = presetLimits(TimingBalanced)
	}
	limits.MaxDepth = c.DirSize.DefaultDepth
	return limits
}

// Dir returns the nls config directory.
//
// On Unix-like systems this honors the XDG Base Directory specification:
// $XDG_CONFIG_HOME/nls, falling back to ~/.config/nls.
//
// On Windows this uses the conventional %APPDATA% location: %APPDATA%\nls,
// falling back to %USERPROFILE%\AppData\Roaming\nls when APPDATA is unset.
func Dir() (string, error) {
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			if !filepath.IsAbs(appdata) {
				return "", fmt.Errorf("APPDATA must be absolute, got %q", appdata)
			}
			return filepath.Join(appdata, "nls"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "AppData", "Roaming", "nls"), nil
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		if !filepath.IsAbs(xdg) {
			return "", fmt.Errorf("XDG_CONFIG_HOME must be absolute, got %q", xdg)
		}
		return filepath.Join(xdg, "nls"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "nls"), nil
}

// Path returns the resolved config file path under the XDG hierarchy.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load reads the TOML config from the XDG path. Missing files yield Defaults.
// Errors loading the directory (e.g. bad XDG_CONFIG_HOME) are returned; a
// missing config.toml is not an error.
func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Defaults(), err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Defaults(), nil
		}
		return Defaults(), fmt.Errorf("read config %s: %w", path, err)
	}
	var raw Config
	md, err := toml.Decode(string(data), &raw)
	if err != nil {
		return Defaults(), fmt.Errorf("parse config %s: %w", path, err)
	}
	if keys := md.Undecoded(); len(keys) > 0 {
		return Defaults(), fmt.Errorf("parse config %s: unknown key %q", path, keys[0].String())
	}
	return raw.Resolve()
}

// NormalizeTiming trims and lowercases a preset name for tolerant matching.
func NormalizeTiming(s string) TimingPreset {
	return TimingPreset(strings.ToLower(strings.TrimSpace(s)))
}
