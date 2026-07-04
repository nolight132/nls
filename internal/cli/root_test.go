package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nolight132/nls/internal/config"
	"github.com/nolight132/nls/internal/listing"
)

func TestVersionFlag(t *testing.T) {
	cmd := Root()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.Contains(got, "nls version ") {
		t.Fatalf("version output = %q", got)
	}
}

func TestLoadUserConfigFallsBackOnInvalidConfig(t *testing.T) {
	root := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", root)
	} else {
		t.Setenv("XDG_CONFIG_HOME", root)
	}
	dir := filepath.Join(root, "nls")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("icons = = bad"), 0o644); err != nil {
		t.Fatal(err)
	}

	var errOut bytes.Buffer
	got := loadUserConfig(&errOut)
	if !columnsEqual(got.DefaultColumns, config.Defaults().DefaultColumns) {
		t.Fatalf("columns = %v, want defaults", got.DefaultColumns)
	}
	if !columnsEqual(config.User.DefaultColumns, config.Defaults().DefaultColumns) {
		t.Fatalf("global columns = %v, want defaults", config.User.DefaultColumns)
	}
	if !strings.Contains(errOut.String(), "using defaults") {
		t.Fatalf("warning = %q, want using defaults", errOut.String())
	}
}

func TestUseTableAllowsListingFlagsOnTTY(t *testing.T) {
	cfg := &Config{Long: true, Recursive: true, SortTime: true, Inode: true}
	if !useTable(cfg, true) {
		t.Fatal("listing flags should keep table output on a TTY")
	}
}

func TestUseTableRejectsAlternateOutputShapes(t *testing.T) {
	for _, cfg := range []*Config{{One: true}, {Commas: true}, {JSON: true}} {
		if useTable(cfg, true) {
			t.Fatalf("%+v should not use table output", cfg)
		}
	}
	if useTable(&Config{}, false) {
		t.Fatal("non-TTY output should not use table output")
	}
}

func TestJSONResolvesAbsAndStaysRich(t *testing.T) {
	setUserForTest(t, config.Defaults())
	opts := buildListOptions(&Config{JSON: true}, false)
	if !opts.ResolveAbs {
		t.Fatal("JSON output should resolve absolute paths")
	}
}

func TestInteractiveUsesBoundedEstimateWhenConfigEnabled(t *testing.T) {
	setUserForTest(t, config.Defaults())
	opts := buildListOptions(&Config{}, true)
	if opts.EstimateDepth != listing.EstimateDepthBounded {
		t.Fatalf("estimate depth = %d, want bounded", opts.EstimateDepth)
	}
	if !opts.EstimateSizes {
		t.Fatal("interactive default should enable size estimation when config enables dir_size")
	}
}

func TestInteractiveUsesConfigDefaultDepthInListing(t *testing.T) {
	cfg := config.Defaults()
	cfg.DirSize.DefaultDepth = 2
	setUserForTest(t, cfg)
	opts := buildListOptions(&Config{}, true)
	if opts.EstimateDepth != listing.EstimateDepthBounded {
		t.Fatalf("estimate depth = %d, want bounded", opts.EstimateDepth)
	}
	if !opts.EstimateSizes {
		t.Fatal("interactive default should enable size estimation")
	}
}

func TestPreciseEnablesExactUnlimitedEstimates(t *testing.T) {
	setUserForTest(t, config.Defaults())
	opts := buildListOptions(&Config{Precise: true}, false)
	if !opts.EstimateSizes {
		t.Fatal("precise should enable size estimation")
	}
	if opts.EstimateDepth != listing.EstimateDepthMax {
		t.Fatalf("estimate depth = %d, want max", opts.EstimateDepth)
	}
	if !opts.Precise {
		t.Fatal("precise flag should pass through to listing")
	}
}

func TestInteractiveSkipsEstimateWhenConfigDisabled(t *testing.T) {
	cfg := config.Defaults()
	cfg.DirSize.Enabled = false
	setUserForTest(t, cfg)
	opts := buildListOptions(&Config{}, true)
	if opts.EstimateSizes {
		t.Fatal("dir_size.enabled=false should disable default interactive estimation")
	}
}

func TestBuildColumnsDefaults(t *testing.T) {
	setUserForTest(t, config.Defaults())
	cols := buildColumns(&Config{})
	want := []string{"id", "name", "size", "modified"}
	if len(cols) != len(want) {
		t.Fatalf("cols = %v, want %v", cols, want)
	}
	for i, w := range want {
		if cols[i] != w {
			t.Fatalf("cols[%d] = %q, want %q", i, cols[i], w)
		}
	}
}

func TestBuildColumnsConfigOverridesOrder(t *testing.T) {
	userCfg := config.Config{
		DefaultColumns: []config.ColumnEntry{
			config.ColumnName,
			config.ColumnId,
			config.ColumnSize,
		},
	}
	setUserForTest(t, userCfg)
	cols := buildColumns(&Config{})
	if cols[0] != "name" || cols[1] != "id" || cols[2] != "size" {
		t.Fatalf("cols = %v, want [name id size]", cols)
	}
}

func TestBuildColumnsFlagsAppendIfMissing(t *testing.T) {
	userCfg := config.Config{
		DefaultColumns: []config.ColumnEntry{config.ColumnName, config.ColumnSize},
	}
	setUserForTest(t, userCfg)
	cols := buildColumns(&Config{Inode: true, Blocks: true, Long: true})
	want := []string{"name", "size", "inode", "blocks", "permissions"}
	if len(cols) != len(want) {
		t.Fatalf("cols = %v, want %v", cols, want)
	}
	for i, w := range want {
		if cols[i] != w {
			t.Fatalf("cols[%d] = %q, want %q", i, cols[i], w)
		}
	}
}

func TestBuildColumnsFlagsDontDuplicateConfigColumns(t *testing.T) {
	userCfg := config.Config{
		DefaultColumns: []config.ColumnEntry{
			config.ColumnName, config.ColumnInode, config.ColumnPermissions,
		},
	}
	setUserForTest(t, userCfg)
	cols := buildColumns(&Config{Inode: true, Long: true})
	count := 0
	for _, c := range cols {
		if c == "inode" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("inode appeared %d times in %v, want 1", count, cols)
	}
}

func columnsEqual(a, b []config.ColumnEntry) bool {
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

func setUserForTest(t *testing.T, cfg config.Config) {
	t.Helper()
	prev := config.User
	config.User = cfg
	t.Cleanup(func() { config.User = prev })
}
