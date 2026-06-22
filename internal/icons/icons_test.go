package icons

import (
	"testing"
)

func TestResolveNoIconsFlagWins(t *testing.T) {
	t.Setenv("NLS_ICONS", "1")
	t.Setenv("NERD_FONT", "1")
	if got := Resolve(true, true); got != SetNone {
		t.Fatalf("no-icons flag should win, got %v", got)
	}
}

func TestResolveEnvOverridesConfigOff(t *testing.T) {
	t.Setenv("NLS_ICONS", "1")
	t.Setenv("NERD_FONT", "1")
	if got := Resolve(false, false); got != SetNerd {
		t.Fatalf("env on + config off should enable, got %v", got)
	}
}

func TestResolveConfigEnablesWhenEnvUnset(t *testing.T) {
	t.Setenv("NLS_ICONS", "")
	t.Setenv("NERD_FONT", "1")
	if got := Resolve(false, true); got != SetNerd {
		t.Fatalf("config on + env unset should enable, got %v", got)
	}
}

func TestResolveDefaultsOff(t *testing.T) {
	t.Setenv("NLS_ICONS", "")
	t.Setenv("NERD_FONT", "1")
	if got := Resolve(false, false); got != SetNone {
		t.Fatalf("everything off should yield SetNone, got %v", got)
	}
}

func TestResolveNeedsNerdFont(t *testing.T) {
	t.Setenv("NLS_ICONS", "")
	t.Setenv("NERD_FONT", "0")
	t.Setenv("NLS_NERD_FONT", "0")
	t.Setenv("TERM", "")
	t.Setenv("FONT", "")
	t.Setenv("FONTFACE", "")
	t.Setenv("FONT_FAMILY", "")
	if got := Resolve(false, true); got != SetNone {
		t.Fatalf("no nerd font available should yield SetNone, got %v", got)
	}
}
