package cli

import (
	"bytes"
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

func TestJSONDisablesFastPath(t *testing.T) {
	opts := buildListOptions(&Config{JSON: true}, false, config.Defaults())
	if opts.FastPath {
		t.Fatal("JSON output needs full metadata")
	}
	if !opts.ResolveAbs {
		t.Fatal("JSON output should resolve absolute paths")
	}
}

func TestEstimateDepthMaxFlag(t *testing.T) {
	opts := buildListOptions(&Config{
		EstimateDepth: listing.EstimateDepthMax,
		EstimateSet:   true,
	}, false, config.Defaults())
	if opts.EstimateDepth != listing.EstimateDepthMax {
		t.Fatalf("estimate depth = %d, want max", opts.EstimateDepth)
	}
}

func TestEstimateDepthExplicitZeroIsBounded(t *testing.T) {
	opts := buildListOptions(&Config{EstimateSet: true}, false, config.Defaults())
	if opts.EstimateDepth != listing.EstimateDepthBounded {
		t.Fatalf("estimate depth = %d, want bounded", opts.EstimateDepth)
	}
}

func TestBuildListOptionsAppliesConfigLimits(t *testing.T) {
	userCfg := config.Config{
		Icons: true,
		DirSize: config.DirSizeConfig{
			DefaultDepth: 4,
			Timing:       config.TimingRelaxed,
		},
	}
	opts := buildListOptions(&Config{EstimateSet: true}, true, userCfg)
	if opts.EstimateDepth != listing.EstimateDepthBounded {
		t.Fatalf("estimate depth = %d, want bounded", opts.EstimateDepth)
	}
	limits := userCfg.Limits()
	if opts.BoundedLimits.MaxDepth != 4 {
		t.Fatalf("bounded MaxDepth = %d, want 4", opts.BoundedLimits.MaxDepth)
	}
	if opts.BoundedLimits.WalkDuration != limits.WalkDuration {
		t.Fatalf("bounded walk budget = %v, want %v", opts.BoundedLimits.WalkDuration, limits.WalkDuration)
	}
	if opts.BoundedLimits.MaxDirsPerListing != limits.MaxDirsPerListing {
		t.Fatalf("bounded dirs cap = %d, want %d", opts.BoundedLimits.MaxDirsPerListing, limits.MaxDirsPerListing)
	}
}

func TestBuildListOptionsUsesDefaultsWhenBounded(t *testing.T) {
	opts := buildListOptions(&Config{EstimateSet: true}, true, config.Defaults())
	if opts.BoundedLimits == (listing.Limits{}) {
		t.Fatal("bounded limits should not be zero when config defaults are applied")
	}
	def := listing.DefaultBoundedLimits()
	if opts.BoundedLimits.WalkDuration != def.WalkDuration {
		t.Fatalf("default walk budget = %v, want %v", opts.BoundedLimits.WalkDuration, def.WalkDuration)
	}
}

func TestBuildColumnsDefaults(t *testing.T) {
	cols := buildColumns(&Config{}, config.Defaults())
	want := []string{"id", "name", "type", "size", "modified"}
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
	cols := buildColumns(&Config{}, userCfg)
	if cols[0] != "name" || cols[1] != "id" || cols[2] != "size" {
		t.Fatalf("cols = %v, want [name id size]", cols)
	}
}

func TestBuildColumnsFlagsAppendIfMissing(t *testing.T) {
	userCfg := config.Config{
		DefaultColumns: []config.ColumnEntry{config.ColumnName, config.ColumnSize},
	}
	cols := buildColumns(&Config{Inode: true, Blocks: true, Long: true}, userCfg)
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
	cols := buildColumns(&Config{Inode: true, Long: true}, userCfg)
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

func TestEstimateDepthFlagSet(t *testing.T) {
	var depth int
	var set bool
	flag := &estimateDepthFlag{value: &depth, set: &set}

	if err := flag.Set("max"); err != nil {
		t.Fatal(err)
	}
	if depth != listing.EstimateDepthMax || !set {
		t.Fatalf("max: depth=%d set=%v", depth, set)
	}
	if flag.String() != "max" {
		t.Fatalf("String() = %q", flag.String())
	}

	if err := flag.Set("3"); err != nil {
		t.Fatal(err)
	}
	if depth != 3 {
		t.Fatalf("depth = %d, want 3", depth)
	}

	if flag.Set("-1") == nil {
		t.Fatal("expected error for negative depth")
	}
	if flag.Set("nope") == nil {
		t.Fatal("expected error for invalid depth")
	}
}
