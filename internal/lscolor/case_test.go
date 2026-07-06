package lscolor

import (
	"testing"

	"github.com/nolight132/nls/internal/listing"
)

func TestExtRuleCaseSensitive(t *testing.T) {
	t.Setenv("LS_COLORS", "*.MD=01;33:*.go=01;34")
	s := New()

	cases := []struct {
		name string
		kind listing.Kind
		want string
	}{
		{"file.MD", listing.KindFile, "01;33"},
		{"file.md", listing.KindFile, ""},
		{"file.Go", listing.KindFile, ""},
		{"file.go", listing.KindFile, "01;34"},
	}
	for _, c := range cases {
		got := s.matchSequence(c.name, c.kind)
		if got != c.want {
			t.Errorf("matchSequence(%q): got %q, want %q", c.name, got, c.want)
		}
	}
}
