package output

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nolight132/nls/internal/icons"
	"github.com/nolight132/nls/internal/listing"
)

func TestRenderFastUsesTableOnTTY(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := RenderFast(
		&buf,
		[]string{dir},
		listing.ListOptions{Sort: listing.SortOptions{Field: listing.SortByName}},
		RenderOptions{
			UseTable: true,
			IsTTY:    true,
			Color:    false,
			IconSet:  icons.SetNone,
			Now:      time.Now(),
			Columns:  []string{"id", "name", "type", "size", "modified"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "╭") {
		t.Fatalf("expected table output, got %q", buf.String())
	}
}

func TestRenderFastUsesCompatPathWhenNotTable(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := RenderFast(
		&buf,
		[]string{dir},
		listing.ListOptions{Sort: listing.SortOptions{Field: listing.SortByName}},
		RenderOptions{Plain: PlainOne},
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "╭") {
		t.Fatalf("expected plain compat output, got %q", buf.String())
	}
	if buf.String() != "alpha.txt\n" {
		t.Fatalf("got %q", buf.String())
	}
}
