package output

import (
	"fmt"
	"io"
	"io/fs"
	"testing"
	"time"

	"github.com/nolight132/nls/internal/icons"
	"github.com/nolight132/nls/internal/listing"
)

func benchBlocks(n int) []listing.Block {
	exts := []string{".go", ".txt", ".log", ".png", ".rs", ".md", ".json", ".tar.gz", "", ".yaml"}
	now := time.Now()
	entries := make([]listing.Entry, 0, n)
	for i := range n {
		e := listing.Entry{
			Name:        fmt.Sprintf("file-%05d%s", i, exts[i%len(exts)]),
			Kind:        listing.KindFile,
			Size:        int64(i) * 137,
			Modified:    now.Add(-time.Duration(i) * time.Minute),
			Permissions: "-rw-r--r--",
			Mode:        0o644,
		}
		switch {
		case i%7 == 0:
			e.Kind = listing.KindDirectory
			e.Mode = fs.ModeDir | 0o755
			e.Permissions = "drwxr-xr-x"
		case i%13 == 0:
			e.Kind = listing.KindSymlink
			e.Mode = fs.ModeSymlink | 0o777
			e.LinkTarget = "target"
		}
		entries = append(entries, e)
	}
	return []listing.Block{{Dir: "bench", Entries: entries, Directory: true}}
}

func BenchmarkRenderPlainOne(b *testing.B) {
	blocks := benchBlocks(10000)
	opts := RenderOptions{Plain: PlainOne, Now: time.Now()}
	for b.Loop() {
		if err := Render(io.Discard, blocks, opts); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderPlainColumns(b *testing.B) {
	blocks := benchBlocks(2000)
	opts := RenderOptions{
		Plain:   PlainLong,
		Long:    true,
		Human:   true,
		Columns: []string{"permissions", "size", "modified", "name"},
		Now:     time.Now(),
	}
	for b.Loop() {
		if err := Render(io.Discard, blocks, opts); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRenderTable(b *testing.B) {
	blocks := benchBlocks(2000)
	opts := RenderOptions{
		UseTable: true,
		Color:    true,
		IsTTY:    true,
		Human:    true,
		IconSet:  icons.SetNerd,
		Columns:  []string{"id", "name", "size", "modified"},
		Now:      time.Now(),
	}
	b.Run("wide", func(b *testing.B) {
		for b.Loop() {
			if err := Render(io.Discard, blocks, opts); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("fit80", func(b *testing.B) {
		narrow := opts
		narrow.Width = 80
		for b.Loop() {
			if err := Render(io.Discard, blocks, narrow); err != nil {
				b.Fatal(err)
			}
		}
	})
}
