package pathutil

import (
	"errors"
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

func TestExpandEmptyPath(t *testing.T) {
	_, err := Expand("")
	if err == nil {
		t.Fatal("Expand(\"\") expected error")
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("Expand(\"\") error = %T, want *os.PathError", err)
	}
	if pathErr.Path != "" {
		t.Fatalf("Expand(\"\") error path = %q, want empty", pathErr.Path)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Expand(\"\") error = %v, want not-exist error", err)
	}
}
