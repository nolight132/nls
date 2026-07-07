package termcolor

import (
	"strings"
	"testing"

	"github.com/nolight132/nls/internal/listing"
)

func TestNameGitOverridesOnlyChangedStates(t *testing.T) {
	s := New(true)

	plain := s.Name("file.txt", listing.KindFile)
	if got := s.NameGit("file.txt", listing.KindFile, listing.GitStateClean); got != plain {
		t.Errorf("clean entry should use LS_COLORS path: %q vs %q", got, plain)
	}
	if got := s.NameGit("file.txt", listing.KindFile, listing.GitStateNone); got != plain {
		t.Errorf("entry outside repo should use LS_COLORS path: %q vs %q", got, plain)
	}
	for _, state := range []listing.GitState{
		listing.GitStateModified, listing.GitStateUntracked, listing.GitStateIgnored,
	} {
		got := s.NameGit("file.txt", listing.KindFile, state)
		if got == plain || !strings.Contains(got, "\x1b[") {
			t.Errorf("state %d should override with its own color, got %q", state, got)
		}
	}

	disabled := New(false)
	if got := disabled.NameGit("file.txt", listing.KindFile, listing.GitStateModified); got != "file.txt" {
		t.Errorf("disabled style should not color: %q", got)
	}
}
