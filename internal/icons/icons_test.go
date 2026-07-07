package icons

import (
	"testing"

	"github.com/nolight132/nls/internal/listing"
)

func TestResolveNoIconsFlagWins(t *testing.T) {
	if got := Resolve(true, true, true); got != SetNone {
		t.Fatalf("no-icons flag should win, got %v", got)
	}
}

func TestResolveDefaultsOff(t *testing.T) {
	if got := Resolve(false, false, true); got != SetNone {
		t.Fatalf("everything off should yield SetNone, got %v", got)
	}
}

func TestResolveCanDisableSpecialIcons(t *testing.T) {
	if got := Resolve(false, true, false); got != SetNerdBasic {
		t.Fatalf("special icons disabled should yield basic nerd set, got %v", got)
	}
}

// Matchers are an ordered list where the first hit wins, so the risky
// regressions are ordering ones: a specific rule falling behind a generic
// extension rule. Each case here exercises one matching mechanism.
func TestMatchIconPrecedence(t *testing.T) {
	tests := []struct {
		name string
		file string
		icon Icon
	}{
		{"plain extension", "server.go", IconGo},
		{"extension is case-insensitive", "PHOTO.PNG", IconImage},
		{"exact name without extension", "Dockerfile", IconDockerfile},
		{"dotfile", ".gitignore", IconGit},
		{"exact name beats json extension", "package.json", IconNode},
		{"lockfile beats json extension", "package-lock.json", IconLockfile},
		{"kubernetes beats yaml extension", "deployment.yaml", IconKubernetes},
		{"tailwind config beats js extension", "tailwind.config.js", IconTailwind},
		{"vite config beats ts extension", "vite.config.ts", IconVite},
		{"pyproject beats toml extension", "pyproject.toml", IconPythonPackage},
		{"multi-part suffix", ".env.local", IconEnvironment},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchIcon(tt.file); got != tt.icon.Char {
				t.Fatalf("MatchIcon(%q) = %q, want %q (%s)", tt.file, got, tt.icon.Char, tt.icon.Name)
			}
		})
	}
}

func TestForHonorsSpecialIconSet(t *testing.T) {
	entry := listing.Entry{Name: ".env", Kind: listing.KindFile}
	if got := For(entry, SetNerd); got != IconEnvironment.Char {
		t.Fatalf("special set icon = %q, want %q", got, IconEnvironment.Char)
	}
	if got := For(entry, SetNerdBasic); got != basicIcon(listing.KindFile) {
		t.Fatalf("basic set icon = %q, want file fallback", got)
	}
	if got := For(entry, SetNone); got != "" {
		t.Fatalf("no icon set = %q, want empty", got)
	}
}

func TestMatchIconUsesBaseName(t *testing.T) {
	if got := MatchIcon("/tmp/project/main.go"); got != IconGo.Char {
		t.Fatalf("MatchIcon should match against the base name, got %q, want %q", got, IconGo.Char)
	}
}

func TestMatchIconUnknown(t *testing.T) {
	if got := MatchIcon("unknown-file"); got != "" {
		t.Fatalf("unknown file should not match an icon, got %q", got)
	}
}
