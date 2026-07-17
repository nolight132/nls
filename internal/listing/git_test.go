package listing

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// gitTestEnv skips the test when git is unavailable and isolates it from
// the developer's real global/system config (core.excludesFile would
// otherwise leak into ignore decisions).
func gitTestEnv(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("XDG_CONFIG_HOME", "")
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{
		"-C", dir,
		"-c", "user.name=test",
		"-c", "user.email=test@test",
	}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// Covers the two easy regressions: clean files must stay blank and lookups
// from a subdirectory must use repo-root-relative keys.
func TestGitStatusInSubdirectory(t *testing.T) {
	gitTestEnv(t)
	root := t.TempDir()
	gitRun(t, root, "init")
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "clean.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", "sub/clean.txt")
	gitRun(t, root, "commit", "-m", "init")

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
	gitRun(t, root, "add", "sub/committed.txt")
	gitRun(t, root, "commit", "-m", "second")

	blocks, errs := List([]string{sub}, ListOptions{GitStatus: true, All: true})
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	if !blocks[0].GitRepo {
		t.Error("block inside a repo should set GitRepo")
	}
	entries := blocks[0].Entries
	assertGitState(t, entries, "clean.txt", GitState{StatusUnmodified, StatusModified})
	assertGitState(t, entries, "new.txt", GitState{StatusUntracked, StatusUntracked})
	assertGitState(t, entries, "debug.log", GitState{StatusIgnored, StatusIgnored})
	assertGitState(t, entries, "committed.txt", GitState{StatusUnmodified, StatusUnmodified})
}

func assertGitState(t *testing.T, entries []Entry, name string, want GitState) {
	t.Helper()
	if got := findEntry(t, entries, name).GitState; got != want {
		t.Errorf("%s state = %q, want %q", name, got, want)
	}
}

func TestGitStatusGlobalIgnore(t *testing.T) {
	gitTestEnv(t)
	xdg := t.TempDir()
	if err := os.MkdirAll(filepath.Join(xdg, "git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(xdg, "git", "ignore"), []byte("*.tmp\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdg)

	root := t.TempDir()
	gitRun(t, root, "init")
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
	assertGitState(t, blocks[0].Entries, "junk.tmp", GitState{StatusIgnored, StatusIgnored})
	assertGitState(t, blocks[0].Entries, "real.txt", GitState{StatusUntracked, StatusUntracked})
}

// A tracked file stays part of the worktree even when an ignore pattern
// matches it, so it must not render as ignored.
func TestGitStatusTrackedFileMatchingIgnorePattern(t *testing.T) {
	gitTestEnv(t)
	root := t.TempDir()
	gitRun(t, root, "init")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "keep.log"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "debug.log"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", "-f", "keep.log")
	gitRun(t, root, "add", ".gitignore")
	gitRun(t, root, "commit", "-m", "init")

	blocks, errs := List([]string{root}, ListOptions{GitStatus: true})
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	assertGitState(t, blocks[0].Entries, "keep.log", GitState{StatusUnmodified, StatusUnmodified})
	assertGitState(t, blocks[0].Entries, "debug.log", GitState{StatusIgnored, StatusIgnored})
}

// Listing a directory inside a fully-untracked or fully-ignored tree must
// inherit that state: git reports such directories collapsed ("dir/"), so
// their contents never appear in the porcelain output individually.
func TestGitStatusCollapsedDirectories(t *testing.T) {
	gitTestEnv(t)
	root := t.TempDir()
	gitRun(t, root, "init")
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("build/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", ".gitignore")
	gitRun(t, root, "commit", "-m", "init")
	for _, d := range []string{"build/pkg", "newdir/inner"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, d, "f.txt"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	blocks, errs := List([]string{filepath.Join(root, "build", "pkg")}, ListOptions{GitStatus: true})
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	assertGitState(t, blocks[0].Entries, "f.txt", GitState{StatusIgnored, StatusIgnored})

	blocks, errs = List([]string{filepath.Join(root, "newdir", "inner")}, ListOptions{GitStatus: true})
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	assertGitState(t, blocks[0].Entries, "f.txt", GitState{StatusUntracked, StatusUntracked})
}

func TestGitStatusOutsideRepo(t *testing.T) {
	gitTestEnv(t)
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
