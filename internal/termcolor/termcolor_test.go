package termcolor

import (
	"testing"

	"github.com/fatih/color"
	"github.com/nolight132/nls/internal/listing"
)

func TestStyleDisabledDoesNotPolluteGlobalColor(t *testing.T) {
	enabled := New(true)
	got := enabled.Size("42")
	if got == "42" {
		t.Fatalf("enabled style produced uncolored output")
	}

	New(false)

	got = enabled.Size("42")
	if got == "42" {
		t.Errorf("enabled style output was disabled by a later New(false) call: global state pollution")
	}
}

func TestStyleEnabledDoesNotPolluteGlobalColor(t *testing.T) {
	color.NoColor = true
	disabled := New(false)

	New(true)

	got := disabled.Heading("title")
	if got != "title" {
		t.Errorf("disabled style produced colored output %q after New(true)", got)
	}
}

func TestNameRespectsEnabledAcrossInstances(t *testing.T) {
	enabled := New(true)
	New(false)

	got := enabled.Name("docs", listing.KindDirectory)
	if got == "docs" {
		t.Errorf("enabled Name() lost color after New(false): %q", got)
	}
}
