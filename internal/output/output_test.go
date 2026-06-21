package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nolight132/nls/internal/icons"
	"github.com/nolight132/nls/internal/listing"
)

func blocks(entries ...listing.Entry) []listing.Block {
	return []listing.Block{{Entries: entries}}
}

func TestRenderPlain(t *testing.T) {
	entries := []listing.Entry{{Name: "a"}, {Name: "b"}}
	var buf bytes.Buffer
	if err := Render(&buf, blocks(entries...), Options{IsTTY: false, Plain: PlainOne}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "a\nb\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestRenderPlainLong(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	entries := []listing.Entry{{
		Name:        "file.txt",
		Kind:        listing.KindFile,
		Size:        10,
		Modified:    now,
		Permissions: "-rw-r--r--",
	}}
	var buf bytes.Buffer
	if err := Render(&buf, blocks(entries...), Options{
		IsTTY: false,
		Plain: PlainLong,
		Human: true,
		Now:   now,
	}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "-rw-r--r--") || !strings.Contains(out, "file.txt") {
		t.Fatalf("unexpected long output: %q", out)
	}
}

func TestRenderJSON(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	entries := []listing.Entry{{
		Name:        "file.txt",
		Kind:        listing.KindFile,
		Size:        10,
		Modified:    now.Add(-time.Hour),
		Permissions: "-rw-r--r--",
	}}

	var buf bytes.Buffer
	if err := Render(&buf, blocks(entries...), Options{JSON: true, Human: true, Now: now}); err != nil {
		t.Fatal(err)
	}

	var rows []JSONRow
	if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if rows[0].Name != "file.txt" || rows[0].Type != "file" {
		t.Fatalf("unexpected row: %#v", rows[0])
	}
	if rows[0].SizeHuman != "10 B" {
		t.Fatalf("size human: %q", rows[0].SizeHuman)
	}
}

func TestRenderTable(t *testing.T) {
	entries := []listing.Entry{{
		Name:        "docs",
		Kind:        listing.KindDirectory,
		Permissions: "drwxr-xr-x",
	}}

	var buf bytes.Buffer
	if err := Render(&buf, blocks(entries...), Options{
		UseTable: true,
		IsTTY:    true,
		Color:    false,
		IconSet:  icons.SetNone,
		Now:      time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	for _, col := range []string{"╭", "name", "type", "size", "modified", "docs", "dir", "│ 0 │"} {
		if !strings.Contains(out, col) {
			t.Fatalf("missing %q in %q", col, out)
		}
	}
}

func TestRenderClassify(t *testing.T) {
	entries := []listing.Entry{{Name: "bin", Kind: listing.KindDirectory}}
	var buf bytes.Buffer
	if err := Render(&buf, blocks(entries...), Options{
		IsTTY:    false,
		Plain:    PlainOne,
		Classify: true,
	}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "bin/\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestFormatPathErrorLikeLs(t *testing.T) {
	temp := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}

	_, err = os.Lstat("missing")
	if err == nil {
		t.Fatal("expected missing file error")
	}
	got := formatError(err, false)
	wantPrefix := "nls: missing: "
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("got %q, want prefix %q", got, wantPrefix)
	}
	if got == wantPrefix {
		t.Fatalf("got %q, want OS error message", got)
	}
}

func TestFormatPathErrorSuggestsOnlyForTTY(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alpha"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	err := &os.PathError{Op: "lstat", Path: filepath.Join(dir, "aplha"), Err: os.ErrNotExist}

	tty := formatError(err, true)
	if !strings.Contains(tty, "Did you mean '") || !strings.Contains(tty, filepath.Join(dir, "alpha")) {
		t.Fatalf("missing suggestion in %q", tty)
	}

	nonTTY := formatError(err, false)
	if strings.Contains(nonTTY, "Did you mean") {
		t.Fatalf("suggestion leaked into non-TTY output: %q", nonTTY)
	}
}

func TestFormatPathErrorKeepsRelativeNestedPath(t *testing.T) {
	dir := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatal(err)
		}
	}()
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nested", "target"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	abs := filepath.Join(dir, "nested", "targte")
	err = fmt.Errorf("nested/targte: %w", &os.PathError{Op: "lstat", Path: abs, Err: os.ErrNotExist})
	got := formatError(err, true)
	if !strings.Contains(got, "nls: nested/targte") {
		t.Fatalf("did not preserve relative path: %q", got)
	}
	if !strings.Contains(got, "Did you mean 'nested/target'?") {
		t.Fatalf("did not suggest relative path: %q", got)
	}
}
