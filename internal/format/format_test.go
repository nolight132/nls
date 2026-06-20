package format

import (
	"testing"
	"time"
)

func TestHumanSize(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 kB"},
		{1536, "1.5 kB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
		{1125899906842624, "1.0 PB"},
	}

	for _, tt := range tests {
		got := Size(tt.n, true, false)
		if got != tt.want {
			t.Errorf("Size(%d, true, false) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestSizeApprox(t *testing.T) {
	if got := Size(1024, true, true); got != ">1.0 kB" {
		t.Fatalf("got %q", got)
	}
}

func TestIsRelativeModified(t *testing.T) {
	if !IsRelativeModified("22 minutes ago") {
		t.Fatal("expected relative")
	}
	if IsRelativeModified("2026-01-01 08:30") {
		t.Fatal("expected absolute")
	}
}

func TestSizeRaw(t *testing.T) {
	if got := Size(42, false, false); got != "42" {
		t.Fatalf("Size(42, false, false) = %q, want 42", got)
	}
}

func TestModifiedRelative(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)

	if got := Modified(now.Add(-30*time.Second), now); got != "just now" {
		t.Fatalf("got %q", got)
	}
	if got := Modified(now.Add(-5*time.Minute), now); got != "5 minutes ago" {
		t.Fatalf("got %q", got)
	}
	if got := Modified(now.Add(-26*time.Hour), now); got != "yesterday" {
		t.Fatalf("got %q", got)
	}
}

func TestModifiedAbsolute(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	old := time.Date(2026, 1, 1, 8, 30, 0, 0, time.UTC)
	if got := Modified(old, now); got != "2026-01-01 08:30" {
		t.Fatalf("got %q", got)
	}
}
