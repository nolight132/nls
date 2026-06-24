package listing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDirSortsAlphabetically(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"zebra.txt", "alpha", "mango"} {
		path := filepath.Join(dir, name)
		if name == "alpha" || name == "mango" {
			if err := os.Mkdir(path, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := ReadDir(dir, ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("len = %d, want 3", len(entries))
	}
	if entries[0].Name != "alpha" || entries[1].Name != "mango" || entries[2].Name != "zebra.txt" {
		t.Fatalf("unexpected order: %#v", entries)
	}
}

func TestReadDirHidden(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "visible"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	without, err := ReadDir(dir, ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(without) != 1 || without[0].Name != "visible" {
		t.Fatalf("without all: %#v", without)
	}

	with, err := ReadDir(dir, ListOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(with) != 4 {
		t.Fatalf("with all: got %d entries, want 4 (. .. .hidden visible): %#v", len(with), with)
	}
}

func TestRecursiveAllSkipsDotEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	blocks, err := List([]string{dir}, ListOptions{All: true, Recursive: true, Sort: SortOptions{Field: SortByName}})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want root and sub only: %#v", len(blocks), blocks)
	}
}

func TestFormatPermissions(t *testing.T) {
	got := formatPermissions(0o755)
	if got != "-rwxr-xr-x" {
		t.Fatalf("got %q", got)
	}

	dirMode := os.ModeDir | 0o755
	got = formatPermissions(dirMode)
	if got != "drwxr-xr-x" {
		t.Fatalf("got %q", got)
	}
}
