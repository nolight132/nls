package cli

import (
	"bytes"
	"strings"
	"testing"
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
