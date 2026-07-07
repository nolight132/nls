package lscolor

import (
	"testing"

	"github.com/nolight132/nls/internal/listing"
)

func TestDefaultRegularFileUsesNoColor(t *testing.T) {
	s := New()
	got := s.matchSequence("README.md", listing.KindFile)
	want := "0"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDefaultDirectoryUsesSimpleAnsiBlue(t *testing.T) {
	s := New()
	got := s.matchSequence("cmd", listing.KindDirectory)
	want := "34"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestDefaultSymlinkAndExecutableUseSimpleAnsi(t *testing.T) {
	s := New()
	if got := s.matchSequence("link", listing.KindSymlink); got != "36" {
		t.Fatalf("symlink: got %q", got)
	}
	if got := s.matchSequence("tool", listing.KindExecutable); got != "32" {
		t.Fatalf("executable: got %q", got)
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

func TestLnTargetKeywordIsNotEmittedRaw(t *testing.T) {
	s := parse("ln=target:di=34")
	got := s.Colorize("mylink", listing.KindSymlink)
	if got != "mylink" {
		t.Fatalf("got %q, want plain name", got)
	}
}
