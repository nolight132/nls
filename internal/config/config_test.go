package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/nolight132/nls/internal/listing"
)

func TestDefaults(t *testing.T) {
	d := Defaults()
	if d.Icons {
		t.Fatal("icons should default off")
	}
	if d.DirSize.Timing != TimingBalanced {
		t.Fatalf("default timing = %q, want balanced", d.DirSize.Timing)
	}
	if d.DirSize.DefaultDepth != 0 {
		t.Fatalf("default depth = %d, want 0", d.DirSize.DefaultDepth)
	}
}

func TestResolveAppliesDefaultsForMissingFields(t *testing.T) {
	resolved, err := (&Config{}).Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Icons {
		t.Fatal("empty config should keep icons off")
	}
	if resolved.DirSize.Timing != TimingBalanced {
		t.Fatalf("empty config timing = %q, want balanced", resolved.DirSize.Timing)
	}
}

func TestResolvePreservesExplicitValues(t *testing.T) {
	c := Config{
		Icons: true,
		DirSize: DirSizeConfig{
			DefaultDepth: 3,
			Timing:       TimingRelaxed,
		},
	}
	resolved, err := c.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if !resolved.Icons {
		t.Fatal("icons should stay on")
	}
	if resolved.DirSize.DefaultDepth != 3 {
		t.Fatalf("depth = %d, want 3", resolved.DirSize.DefaultDepth)
	}
	if resolved.DirSize.Timing != TimingRelaxed {
		t.Fatalf("timing = %q, want relaxed", resolved.DirSize.Timing)
	}
}

func TestResolveRejectsUnknownTiming(t *testing.T) {
	if _, err := (&Config{DirSize: DirSizeConfig{Timing: "turbo"}}).Resolve(); err == nil {
		t.Fatal("expected error for unknown timing preset")
	}
}

func TestResolveRejectsUnknownColumn(t *testing.T) {
	if _, err := (&Config{DefaultColumns: []ColumnEntry{"bogus"}}).Resolve(); err == nil {
		t.Fatal("expected error for unknown column name")
	}
}

func TestResolveAcceptsAllKnownColumns(t *testing.T) {
	all := []ColumnEntry{
		ColumnId, ColumnName, ColumnType, ColumnSize,
		ColumnModified, ColumnAccessed, ColumnChanged,
		ColumnPermissions, ColumnLinks, ColumnOwner, ColumnGroup,
		ColumnInode, ColumnBlocks,
	}
	resolved, err := (&Config{DefaultColumns: all}).Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved.DefaultColumns) != len(all) {
		t.Fatalf("resolved columns = %d, want %d", len(resolved.DefaultColumns), len(all))
	}
}

func TestLimitsForStrict(t *testing.T) {
	c := Config{DirSize: DirSizeConfig{Timing: TimingStrict}}
	limits := c.Limits()
	if limits.WalkDuration >= 50*time.Millisecond {
		t.Fatalf("strict walk budget = %v, should be under balanced 50ms", limits.WalkDuration)
	}
	if limits.MaxDirsPerListing >= 6 {
		t.Fatalf("strict dirs cap = %d, should be under balanced 6", limits.MaxDirsPerListing)
	}
}

func TestLimitsForRelaxed(t *testing.T) {
	c := Config{DirSize: DirSizeConfig{Timing: TimingRelaxed}}
	limits := c.Limits()
	if limits.WalkDuration <= 50*time.Millisecond {
		t.Fatalf("relaxed walk budget = %v, should exceed balanced 50ms", limits.WalkDuration)
	}
	if limits.MaxWalkEntries <= 400 {
		t.Fatalf("relaxed entry cap = %d, should exceed balanced 400", limits.MaxWalkEntries)
	}
}

func TestLimitsForBalancedMatchesDefaults(t *testing.T) {
	c := Config{DirSize: DirSizeConfig{Timing: TimingBalanced}}
	limits := c.Limits()
	def := listing.DefaultBoundedLimits()
	if limits.WalkDuration != def.WalkDuration ||
		limits.ListingDuration != def.ListingDuration ||
		limits.MaxWalkEntries != def.MaxWalkEntries ||
		limits.MaxDirsPerListing != def.MaxDirsPerListing {
		t.Fatalf("balanced limits = %+v, want %+v", limits, def)
	}
}

func TestLimitsCarriesDefaultDepth(t *testing.T) {
	c := Config{DirSize: DirSizeConfig{DefaultDepth: 5, Timing: TimingBalanced}}
	if got := c.Limits().MaxDepth; got != 5 {
		t.Fatalf("MaxDepth = %d, want 5", got)
	}
}

func TestLimitsFallsBackOnBadPreset(t *testing.T) {
	c := Config{DirSize: DirSizeConfig{Timing: "nope"}}
	limits := c.Limits()
	def := listing.DefaultBoundedLimits()
	if limits.WalkDuration != def.WalkDuration {
		t.Fatalf("bad preset should fall back to balanced walk budget, got %v", limits.WalkDuration)
	}
}

func TestDirHonorsXDGConfigHome(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG_CONFIG_HOME is a Unix convention; Windows uses APPDATA")
	}
	t.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if dir != filepath.Join("/custom/xdg", "nls") {
		t.Fatalf("dir = %q, want /custom/xdg/nls", dir)
	}
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join("/custom/xdg", "nls", "config.toml") {
		t.Fatalf("path = %q, want /custom/xdg/nls/config.toml", path)
	}
}

func TestDirHonorsAppDataOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("APPDATA lookup only applies on Windows")
	}
	t.Setenv("APPDATA", `C:\Users\test\AppData\Roaming`)
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(`C:\Users\test\AppData\Roaming`, "nls")
	if dir != want {
		t.Fatalf("dir = %q, want %q", dir, want)
	}
}

func TestDirFallsBackToHomeConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows fallback differs; covered by TestDirFallsBackToUserProfileOnWindows")
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".config", "nls")
	if dir != want {
		t.Fatalf("dir = %q, want %q", dir, want)
	}
}

func TestDirFallsBackToUserProfileOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("USERPROFILE fallback only applies on Windows")
	}
	t.Setenv("APPDATA", "")
	profile := t.TempDir()
	t.Setenv("USERPROFILE", profile)
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(profile, "AppData", "Roaming", "nls")
	if dir != want {
		t.Fatalf("dir = %q, want %q", dir, want)
	}
}

func TestDirRejectsRelativeXDG(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG_CONFIG_HOME is a Unix convention")
	}
	t.Setenv("XDG_CONFIG_HOME", "relative/path")
	if _, err := Dir(); err == nil {
		t.Fatal("expected error for relative XDG_CONFIG_HOME")
	}
}

func TestDirRejectsRelativeAppDataOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("APPDATA validation only applies on Windows")
	}
	t.Setenv("APPDATA", "relative/path")
	if _, err := Dir(); err == nil {
		t.Fatal("expected error for relative APPDATA")
	}
}

func setConfigDirEnv(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", dir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", dir)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	setConfigDirEnv(t, t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := Defaults()
	if cfg.Icons != want.Icons ||
		cfg.DirSize != want.DirSize ||
		!columnsEqual(cfg.DefaultColumns, want.DefaultColumns) {
		t.Fatalf("missing file: got %+v, want defaults %+v", cfg, want)
	}
}

func columnsEqual(a, b []ColumnEntry) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestLoadReadsAndResolves(t *testing.T) {
	root := t.TempDir()
	setConfigDirEnv(t, root)
	dir := filepath.Join(root, "nls")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	contents := `
icons = true
[dir_size]
default_depth = 2
timing = "relaxed"
`
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Icons {
		t.Fatal("icons should be on")
	}
	if cfg.DirSize.DefaultDepth != 2 {
		t.Fatalf("depth = %d, want 2", cfg.DirSize.DefaultDepth)
	}
	if cfg.DirSize.Timing != TimingRelaxed {
		t.Fatalf("timing = %q, want relaxed", cfg.DirSize.Timing)
	}
	if got := cfg.Limits().MaxDepth; got != 2 {
		t.Fatalf("limits MaxDepth = %d, want 2", got)
	}
}

func TestLoadParseErrorIsReturned(t *testing.T) {
	root := t.TempDir()
	setConfigDirEnv(t, root)
	dir := filepath.Join(root, "nls")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("icons = = bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestNormalizeTiming(t *testing.T) {
	if got := NormalizeTiming("  Relaxed "); got != TimingRelaxed {
		t.Fatalf("normalize = %q, want relaxed", got)
	}
}
