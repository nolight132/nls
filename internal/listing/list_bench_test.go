package listing

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func setupBenchDir(b *testing.B, files, symlinks int) string {
	b.Helper()
	dir := b.TempDir()
	exts := []string{".go", ".txt", ".log", ".png", ".rs", ".md"}
	for i := range files {
		name := fmt.Sprintf("file-%05d%s", i, exts[i%len(exts)])
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			b.Fatal(err)
		}
	}
	for i := range symlinks {
		name := fmt.Sprintf("link-%05d", i)
		if err := os.Symlink("file-00000.go", filepath.Join(dir, name)); err != nil {
			b.Fatal(err)
		}
	}
	return dir
}

func benchList(b *testing.B, dir string, opts ListOptions) {
	b.Helper()
	for b.Loop() {
		blocks, errs := List([]string{dir}, opts)
		if len(errs) > 0 {
			b.Fatal(errs[0])
		}
		if len(blocks) != 1 {
			b.Fatalf("got %d blocks", len(blocks))
		}
	}
}

func BenchmarkListFiles5k(b *testing.B) {
	dir := setupBenchDir(b, 5000, 0)
	benchList(b, dir, ListOptions{Sort: SortOptions{Field: SortByName}})
}

func BenchmarkListSymlinks2k(b *testing.B) {
	dir := setupBenchDir(b, 100, 2000)
	benchList(b, dir, ListOptions{Sort: SortOptions{Field: SortByName}})
}
