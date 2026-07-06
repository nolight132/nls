package termcolor

import (
	"testing"
)

func TestModifiedColoringByDesign(t *testing.T) {
	enabled := New(true)

	cases := []struct {
		value string
		want  string
	}{
		{"just now", "\x1b[35mjust now\x1b[0m"},
		{"5 minutes ago", "\x1b[35m5 minutes ago\x1b[0m"},
		{"a year ago", "\x1b[35ma year ago\x1b[0m"},
		{"-", "-"},
	}
	for _, c := range cases {
		got := enabled.Modified(c.value)
		if got != c.want {
			t.Errorf("Modified(%q): got %q, want %q", c.value, got, c.want)
		}
	}

	disabled := New(false)
	if got := disabled.Modified("just now"); got != "just now" {
		t.Errorf("disabled Modified: got %q, want %q", got, "just now")
	}
}
