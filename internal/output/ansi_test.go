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

func TestStripANSINonSGR(t *testing.T) {
	sgr := "\x1b[31mred\x1b[0m"
	if got := stripANSI(sgr); got != "red" {
		t.Errorf("SGR: got %q, want %q", got, "red")
	}

	nonSGR := "\x1b[2Jclear"
	if got := stripANSI(nonSGR); got != "clear" {
		t.Errorf("non-SGR CSI (\\x1b[2J): got %q, want %q", got, "clear")
	}

	cursor := "before\x1b[10;20Hafter"
	if got := stripANSI(cursor); got != "beforeafter" {
		t.Errorf("cursor CSI: got %q, want %q", got, "beforeafter")
	}

	priv := "\x1b[?25lhidden\x1b[?25hshown"
	if got := stripANSI(priv); got != "hiddenshown" {
		t.Errorf("private CSI: got %q, want %q", got, "hiddenshown")
	}

	mixed := "\x1b[31m\x1b[2Jtext\x1b[0m"
	if got := stripANSI(mixed); got != "text" {
		t.Errorf("mixed: got %q, want %q", got, "text")
	}
}

func TestVisibleWidthWithNonSGR(t *testing.T) {
	got := visibleWidth("\x1b[2J\x1b[31mhello\x1b[0m")
	if got != 5 {
		t.Errorf("got %d, want 5", got)
	}
}

func TestTruncateANSI(t *testing.T) {
	tests := []struct {
		in   string
		max  int
		want string
	}{
		{"short", 10, "short"},
		{"exact", 5, "exact"},
		{"longername", 5, "long…"},
		{"\x1b[31mredname\x1b[0m", 4, "\x1b[31mred\x1b[0m…"},
		{"日本語ファイル", 5, "日本…"},
		{"日本語", 4, "日…"},
		{"anything", 0, "anything"},
	}
	for _, tt := range tests {
		if got := truncateANSI(tt.in, tt.max); got != tt.want {
			t.Errorf("truncateANSI(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
		}
		if tt.max > 0 {
			if w := visibleWidth(truncateANSI(tt.in, tt.max)); w > tt.max {
				t.Errorf("truncateANSI(%q, %d) has width %d", tt.in, tt.max, w)
			}
		}
	}
}

func TestTableCapsWidthToTerminal(t *testing.T) {
	entries := []listing.Entry{
		{Name: strings.Repeat("x", 120) + ".txt", Kind: listing.KindFile},
		{Name: "short.txt", Kind: listing.KindFile},
	}
	var buf bytes.Buffer
	if err := Render(&buf, []listing.Block{{Entries: entries}}, RenderOptions{
		UseTable: true, IsTTY: true, Color: false,
		Columns: []string{"id", "name", "size", "modified"},
		Width:   60,
		Now:     time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	for line := range strings.SplitSeq(strings.TrimRight(buf.String(), "\n"), "\n") {
		if w := visibleWidth(line); w > 60 {
			t.Errorf("line exceeds width 60 (got %d): %q", w, line)
		}
	}
	if !strings.Contains(buf.String(), "…") {
		t.Error("expected truncated name with ellipsis")
	}
}
