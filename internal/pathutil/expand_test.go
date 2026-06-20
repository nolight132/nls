package pathutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpand(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		in   string
		want string
	}{
		{"", "."},
		{"~", home},
		{"~/tmp", filepath.Join(home, "tmp")},
		{"/tmp", filepath.Clean("/tmp")},
		{"./foo", "foo"},
	}

	for _, tt := range tests {
		got, err := Expand(tt.in)
		if err != nil {
			t.Fatalf("Expand(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("Expand(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
