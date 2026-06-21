package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestGNUCompatNameOutput(t *testing.T) {
	lsPath := requireGNULS(t)
	nlsPath := buildNLS(t)
	fixture := newCompatFixture(t)

	cases := []struct {
		name string
		args []string
	}{
		{name: "default", args: []string{"."}},
		{name: "one", args: []string{"-1", "."}},
		{name: "all", args: []string{"-a", "."}},
		{name: "almost all", args: []string{"-A", "."}},
		{name: "ignore backups", args: []string{"-B", "."}},
		{name: "reverse", args: []string{"-r", "."}},
		{name: "sort size", args: []string{"-S", "."}},
		{name: "sort extension", args: []string{"-X", "."}},
		{name: "sort time", args: []string{"-t", "."}},
		{name: "sort time reverse", args: []string{"-tr", "."}},
		{name: "access time sort", args: []string{"-u", "."}},
		{name: "ctime sort", args: []string{"-c", "."}},
		{name: "unsorted", args: []string{"-U", "."}},
		{name: "fast", args: []string{"-f", "."}},
		{name: "classify", args: []string{"-F", "."}},
		{name: "slash", args: []string{"-p", "."}},
		{name: "quote", args: []string{"-Q", "."}},
		{name: "inode", args: []string{"-i", "."}},
		{name: "blocks", args: []string{"-s", "."}},
		{name: "human blocks", args: []string{"-sh", "."}},
		{name: "group dirs", args: []string{"--group-directories-first", "."}},
		{name: "directory operand", args: []string{"-d", "subdir"}},
		{name: "file operand", args: []string{"alpha.txt"}},
		{name: "nested file operand", args: []string{filepath.Join("subdir", "nested.txt")}},
		{name: "multiple files", args: []string{"beta.go", "alpha.txt"}},
		{name: "files and directory", args: []string{"beta.go", "subdir", "alpha.txt"}},
		{name: "multiple directories", args: []string{"subdir", "otherdir"}},
		{name: "recursive", args: []string{"-R", "."}},
		{name: "recursive all", args: []string{"-aR", "."}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := runCompatCommand(t, fixture, lsPath, tc.args...)
			got := runCompatCommand(t, fixture, nlsPath, tc.args...)
			if got != want {
				t.Fatalf("nls differs from GNU ls for args %q\n--- want ---\n%s--- got ---\n%s", strings.Join(tc.args, " "), want, got)
			}
		})
	}
}

func TestGNUCompatLongOutput(t *testing.T) {
	lsPath := requireGNULS(t)
	nlsPath := buildNLS(t)
	fixture := newCompatFixture(t)

	cases := []struct {
		name string
		args []string
	}{
		{name: "long directory", args: []string{"-l", "."}},
		{name: "long all", args: []string{"-la", "."}},
		{name: "long human", args: []string{"-lh", "."}},
		{name: "long classify", args: []string{"-lF", "."}},
		{name: "long slash", args: []string{"-lp", "."}},
		{name: "long full time", args: []string{"-l", "--full-time", "."}},
		{name: "full time implies long", args: []string{"--full-time", "."}},
		{name: "long inode", args: []string{"-li", "."}},
		{name: "long blocks", args: []string{"-ls", "."}},
		{name: "long human inode blocks", args: []string{"-lhis", "."}},
		{name: "long access", args: []string{"-lu", "."}},
		{name: "long access sort", args: []string{"-ltu", "."}},
		{name: "long ctime", args: []string{"-lc", "."}},
		{name: "long ctime sort", args: []string{"-ltc", "."}},
		{name: "long group dirs", args: []string{"-l", "--group-directories-first", "."}},
		{name: "long file", args: []string{"-l", "alpha.txt"}},
		{name: "long multiple files", args: []string{"-l", "beta.go", "alpha.txt"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := runCompatCommand(t, fixture, lsPath, tc.args...)
			got := runCompatCommand(t, fixture, nlsPath, tc.args...)
			if got != want {
				t.Fatalf("nls differs from GNU ls for args %q\n--- want ---\n%s--- got ---\n%s", strings.Join(tc.args, " "), want, got)
			}
		})
	}
}

func requireGNULS(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("GNU ls comparison requires Unix-like filesystem semantics")
	}
	lsPath, err := exec.LookPath("ls")
	if err != nil {
		t.Skip("ls not found")
	}
	cmd := exec.Command(lsPath, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil || !bytes.Contains(out, []byte("GNU coreutils")) {
		t.Skip("GNU coreutils ls not found")
	}
	return lsPath
}

func buildNLS(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	repoRoot := filepath.Clean(filepath.Join(wd, "..", ".."))
	bin := filepath.Join(t.TempDir(), "nls")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/nls")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build nls: %v\n%s", err, out)
	}
	return bin
}

func newCompatFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "alpha.txt"), "alpha\n", 0o644)
	mustWriteFile(t, filepath.Join(dir, "beta.go"), "beta beta\n", 0o644)
	mustWriteFile(t, filepath.Join(dir, ".hidden"), "hidden\n", 0o644)
	mustWriteFile(t, filepath.Join(dir, "backup~"), "backup\n", 0o644)
	mustWriteFile(t, filepath.Join(dir, "run.sh"), "#!/bin/sh\n", 0o755)
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "otherdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(dir, "subdir", "nested.txt"), "nested\n", 0o644)
	mustWriteFile(t, filepath.Join(dir, "otherdir", "other.txt"), "other\n", 0o644)
	if err := os.Symlink("alpha.txt", filepath.Join(dir, "link-alpha")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if err := os.Symlink("subdir", filepath.Join(dir, "link-subdir")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}

	base := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	for i, name := range []string{"alpha.txt", "beta.go", ".hidden", "backup~", "run.sh", "subdir", "subdir/nested.txt", "otherdir", "otherdir/other.txt", "link-alpha", "link-subdir"} {
		when := base.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(filepath.Join(dir, name), when, when); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func mustWriteFile(t *testing.T, path, contents string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), mode); err != nil {
		t.Fatal(err)
	}
}

func runCompatCommand(t *testing.T, dir, bin string, args ...string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"LC_ALL=C",
		"TZ=UTC",
		"LS_COLORS=",
		"CLICOLOR=0",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", bin, strings.Join(args, " "), err, out)
	}
	return fmt.Sprintf("%s", out)
}
