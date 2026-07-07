package listing

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// Covers the two easy regressions: clean files must stay blank (go-git's
// Status.File fabricates untracked entries for unknown paths) and lookups
// from a subdirectory must use repo-root-relative keys.
func TestGitStatusInSubdirectory(t *testing.T) {
	root := t.TempDir()
	repo, err := git.PlainInit(root, false)
	if err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "clean.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wt.Add("sub/clean.txt"); err != nil {
		t.Fatal(err)
	}
	sig := &object.Signature{Name: "test", Email: "test@test", When: time.Now()}
	if _, err := wt.Commit("init", &git.CommitOptions{Author: sig}); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(sub, "clean.txt"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "new.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "committed.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "debug.log"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := wt.Add("sub/committed.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := wt.Commit("second", &git.CommitOptions{Author: sig}); err != nil {
		t.Fatal(err)
	}

	blocks, errs := List([]string{sub}, ListOptions{GitStatus: true})
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	if !blocks[0].GitRepo {
		t.Error("block inside a repo should set GitRepo")
	}
	entries := blocks[0].Entries
	if got := findEntry(t, entries, "new.txt").GitStatus; got != "?│?" {
		t.Errorf("untracked = %q, want ?│?", got)
	}
	if got := findEntry(t, entries, "clean.txt").GitStatus; got != " │M" {
		t.Errorf("modified = %q, want \" │M\"", got)
	}
	if got := findEntry(t, entries, "committed.txt").GitStatus; got != " │ " {
		t.Errorf("committed clean file = %q, want \" │ \"", got)
	}
	if got := findEntry(t, entries, "debug.log").GitStatus; got != "I│ " {
		t.Errorf("ignored file = %q, want \"I│ \"", got)
	}

	// The directory itself aggregates its children per column: untracked
	// new.txt contributes '?' to staging, modified clean.txt 'M' to worktree.
	rootBlocks, errs := List([]string{root}, ListOptions{GitStatus: true})
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	if got := findEntry(t, rootBlocks[0].Entries, "sub").GitStatus; got != "?│M" {
		t.Errorf("dir aggregate = %q, want ?│M", got)
	}
}

func TestGitStatusGlobalIgnore(t *testing.T) {
	xdg := t.TempDir()
	if err := os.MkdirAll(filepath.Join(xdg, "git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xdg, "git", "ignore"), []byte("*.tmp\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdg)

	root := t.TempDir()
	if _, err := git.PlainInit(root, false); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "junk.tmp"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "real.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	blocks, errs := List([]string{root}, ListOptions{GitStatus: true})
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	if got := findEntry(t, blocks[0].Entries, "junk.tmp").GitStatus; got != "I│ " {
		t.Errorf("globally ignored file = %q, want \"I│ \"", got)
	}
	if got := findEntry(t, blocks[0].Entries, "real.txt").GitStatus; got != "?│?" {
		t.Errorf("untracked file = %q, want ?│?", got)
	}
}

func TestGitStatusOutsideRepo(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	blocks, errs := List([]string{dir}, ListOptions{GitStatus: true})
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	if blocks[0].GitRepo {
		t.Error("block outside a repo should not set GitRepo")
	}
}
