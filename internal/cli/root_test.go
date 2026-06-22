package cli

import (
	"bytes"
	"strings"
	"testing"

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
	opts := buildListOptions(&Config{JSON: true}, false)
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
	}, false)
	if opts.EstimateDepth != listing.EstimateDepthMax {
		t.Fatalf("estimate depth = %d, want max", opts.EstimateDepth)
	}
}

func TestEstimateDepthExplicitZeroIsBounded(t *testing.T) {
	opts := buildListOptions(&Config{EstimateSet: true}, false)
	if opts.EstimateDepth != listing.EstimateDepthBounded {
		t.Fatalf("estimate depth = %d, want bounded", opts.EstimateDepth)
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
