package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/nolight132/nls/internal/listing"
)

func TestVisibleWidthCJK(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"file.txt", 8},
		{"文件.txt", 8},
		{"文", 2},
		{"abc", 3},
		{"日本語", 6},
		{"🎉", 2},
	}
	for _, tt := range tests {
		got := visibleWidth(tt.name)
		if got != tt.want {
			t.Errorf("visibleWidth(%q) = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestTableCJKBorderAlignment(t *testing.T) {
	entries := []listing.Entry{
		{Name: "文件.txt", Kind: listing.KindFile},
		{Name: "file.txt", Kind: listing.KindFile},
	}
	var buf bytes.Buffer
	if err := Render(&buf, []listing.Block{{Entries: entries}}, RenderOptions{
		UseTable: true, IsTTY: true, Color: false,
		Columns: []string{"name"},
		Now:     time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(buf.String(), "\n")
	var cjkRight, asciiRight int
	for _, line := range lines {
		if strings.Contains(line, "文件.txt") {
			idx := strings.Index(line, "文件.txt")
			cjkRight = idx + visibleWidth("文件.txt")
		}
		if strings.Contains(line, "file.txt") {
			idx := strings.Index(line, "file.txt")
			asciiRight = idx + visibleWidth("file.txt")
		}
	}
	if cjkRight != asciiRight {
		t.Errorf("CJK right edge at display col %d, ASCII at %d; borders misaligned", cjkRight, asciiRight)
	}
}
