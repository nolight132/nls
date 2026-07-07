package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

type IconsConfig struct {
	Enabled      bool `toml:"enabled"`
	SpecialIcons bool `toml:"special_icons"`
}

// DirSizeConfig controls default interactive directory size estimation.
type DirSizeConfig struct {
	Enabled      bool   `toml:"enabled"`
	DefaultDepth int    `toml:"default_depth"`
	Timing       string `toml:"timing"`
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

	// Git status.
	ColumnGitStatus ColumnEntry = "git"
)

// Config is the nls user configuration loaded from XDG paths.
type Config struct {
	// Icons enables Nerd Font icons by default. --no-icons still
	// overrides this.
	Icons IconsConfig `toml:"icons"`
	// DirSize holds defaults for bounded directory size estimation.
	DirSize DirSizeConfig `toml:"dir_size"`
	// Layout holds defaults for the layout of the listing.
	DefaultColumns []ColumnEntry `toml:"default_columns"`
}

// Defaults returns the configuration used when no file is present.
func Defaults() Config {
	return Config{
		Icons: IconsConfig{
			Enabled:      true,
			SpecialIcons: true,
		},
		DirSize: DirSizeConfig{
			Enabled:      true,
			DefaultDepth: 0,
			Timing:       "balanced",
		},
		DefaultColumns: []ColumnEntry{
			ColumnId,
			ColumnName,
			ColumnSize,
			ColumnModified,
		},
	}
}

// User is the process-wide user configuration loaded at startup.
var User = Defaults()

// Resolve validates the result after defaults have been applied.
func (c Config) Resolve() (Config, error) {
	for _, col := range c.DefaultColumns {
		if !isValidColumn(col) {
			return c, fmt.Errorf("unknown column %q in default_columns", col)
		}
	}
	return c, nil
}

func isValidColumn(c ColumnEntry) bool {
	switch c {
	case ColumnId, ColumnName, ColumnType, ColumnSize,
		ColumnModified, ColumnAccessed, ColumnChanged,
		ColumnPermissions, ColumnLinks, ColumnOwner, ColumnGroup,
		ColumnInode, ColumnBlocks, ColumnGitStatus:
		return true
	}
	return false
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
	raw := Defaults()
	md, err := toml.Decode(string(data), &raw)
	if err != nil {
		return Defaults(), fmt.Errorf("parse config %s: %w", path, err)
	}
	if keys := md.Undecoded(); len(keys) > 0 {
		return Defaults(), fmt.Errorf("parse config %s: unknown key %q", path, keys[0].String())
	}
	return raw.Resolve()
}

// LoadUser loads the XDG user config and stores it for process-wide use.
func LoadUser() (Config, error) {
	cfg, err := Load()
	if err != nil {
		cfg = Defaults()
	}
	User = cfg
	return cfg, err
}
