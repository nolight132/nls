package listing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEstimateDirectorySizes(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "a.txt"), make([]byte, 1000), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), make([]byte, 500), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), make([]byte, 100), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err := ReadDir(dir, Options{EstimateDirSizes: true})
	if err != nil {
		t.Fatal(err)
	}

	var docs *Entry
	for i := range entries {
		if entries[i].Name == "docs" {
			docs = &entries[i]
			break
		}
	}
	if docs == nil {
		t.Fatal("docs dir not found")
	}
	if docs.Size != 1500 {
		t.Fatalf("docs size = %d, want 1500", docs.Size)
	}
}
