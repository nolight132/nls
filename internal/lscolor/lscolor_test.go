package lscolor

import (
	"testing"

	"github.com/nolight132/nls/internal/listing"
)

func TestREADMEUsesNushellHighlight(t *testing.T) {
	s := New()
	got := s.matchSequence("README.md", listing.KindFile)
	want := "0;38;5;16;48;5;186"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDirectoryUsesCyan(t *testing.T) {
	s := New()
	got := s.matchSequence("cmd", listing.KindDirectory)
	want := "0;38;5;81"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRegularGoFile(t *testing.T) {
	s := New()
	got := s.matchSequence("go.mod", listing.KindFile)
	if got != "" && got != "0" {
		// Extension-less regular files fall back to fi=0.
		t.Fatalf("expected default file color, got %q", got)
	}
}

func TestRespectsLS_COLORSOverride(t *testing.T) {
	t.Setenv("LS_COLORS", "di=01;31:*.md=01;32")
	s := New()
	if got := s.matchSequence("docs", listing.KindDirectory); got != "01;31" {
		t.Fatalf("dir: got %q", got)
	}
	if got := s.matchSequence("notes.md", listing.KindFile); got != "01;32" {
		t.Fatalf("md: got %q", got)
	}
}
