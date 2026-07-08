package listing

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gitStatusCache shares one status computation per repository across all
// directories of a single List call: repo discovery is a cheap upward stat
// walk, but the git status subprocess scans the worktree and must run once.
// A nil info records a failed repo so it is not retried per directory.
type gitStatusCache map[string]*repoGitInfo

// decorate fills GitState for entries listed from dir and reports whether
// dir belongs to a git worktree. Any failure (not a repository, git missing)
// leaves the entries untouched and returns false.
func (c gitStatusCache) decorate(dir string, entries []Entry) bool {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	root := findRepoRoot(absDir)
	if root == "" {
		return false
	}

	info, ok := c[root]
	if !ok {
		info = loadRepoGitInfo(root)
		c[root] = info
	}
	if info == nil {
		return false
	}

	// Status keys are slash-separated paths relative to the repo root;
	// entry names are relative to the listed directory.
	rel, err := filepath.Rel(root, absDir)
	if err != nil || strings.HasPrefix(rel, "..") {
		return false
	}
	prefix := ""
	if rel != "." {
		prefix = filepath.ToSlash(rel) + "/"
	}

	byName := make(map[string]int, len(entries))
	for i := range entries {
		if entries[i].Name == "." || entries[i].Name == ".." {
			continue
		}
		byName[entries[i].Name] = i
	}

	// One pass over the status maps (only changed files appear in them;
	// clean files are absent). Exact keys set the entry directly; nested
	// keys fold into the directory entry they live under. Ignored children
	// are skipped when folding so an ignored file cannot dirty its parent.
	// The zero GitState marks entries not yet decorated: porcelain codes
	// are always printable bytes, so it cannot occur as a real status.
	aggs := make(map[int]*GitState)
	fold := func(key string, s GitState) {
		if !strings.HasPrefix(key, prefix) {
			return
		}
		name, _, nested := strings.Cut(key[len(prefix):], "/")
		i, ok := byName[name]
		if !ok {
			return
		}
		if !nested {
			entries[i].GitState = s
			return
		}
		if s.Staging == StatusIgnored {
			return
		}
		a := aggs[i]
		if a == nil {
			a = &GitState{StatusUnmodified, StatusUnmodified}
			aggs[i] = a
		}
		a.Staging = foldStatusCode(a.Staging, s.Staging)
		a.Worktree = foldStatusCode(a.Worktree, s.Worktree)
	}
	for key, s := range info.files {
		fold(key, s)
	}
	for key, s := range info.dirs {
		fold(key, s)
	}
	for i, a := range aggs {
		if entries[i].GitState == (GitState{}) {
			entries[i].GitState = *a
		}
	}

	// Entries absent from the status maps are tracked-and-clean unless
	// they live under a collapsed untracked or ignored directory, which
	// git reports only as the directory itself. Dot entries get the
	// neutral cell so the divider line stays unbroken, and .git is shown
	// ignored without traversing it.
	clean := GitState{StatusUnmodified, StatusUnmodified}
	for i := range entries {
		e := &entries[i]
		if e.GitState != (GitState{}) {
			continue
		}
		switch e.Name {
		case ".", "..":
			e.GitState = clean
		case ".git":
			e.GitState = GitState{StatusIgnored, StatusIgnored}
		default:
			if s, ok := info.collapsed(prefix + e.Name); ok {
				e.GitState = s
			} else {
				e.GitState = clean
			}
		}
	}
	return true
}

// StatusCode is one column of a porcelain-v1 XY pair; the byte is git's
// own display character.
type StatusCode byte

const (
	StatusUnmodified StatusCode = ' '
	StatusUntracked  StatusCode = '?'
	StatusIgnored    StatusCode = '!'
	StatusModified   StatusCode = 'M'
)

type GitState struct{ Staging, Worktree StatusCode }

func (s GitState) String() string {
	return string(s.Staging) + string(s.Worktree)
}

// MarshalJSON emits the porcelain XY pair as a string ("??", " M") rather
// than the raw byte values; encoding/json calls it via the interface when
// JSON output serializes entries.
func (s GitState) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type repoGitInfo struct {
	// files maps repo-root-relative slash paths of changed, untracked, or
	// ignored files to their porcelain XY pair.
	files map[string]GitState
	// dirs holds directories git reported collapsed because everything
	// beneath them is untracked or ignored (trailing "/" stripped); their
	// contents never appear in files.
	dirs map[string]GitState
}

// collapsed reports the status of the collapsed untracked or ignored
// directory that path is or lives under, if any.
func (info *repoGitInfo) collapsed(path string) (GitState, bool) {
	for {
		if s, ok := info.dirs[path]; ok {
			return s, true
		}
		i := strings.LastIndexByte(path, '/')
		if i < 0 {
			return GitState{}, false
		}
		path = path[:i]
	}
}

// findRepoRoot walks upward from dir looking for a .git entry (a directory
// in a normal checkout, a file in linked worktrees and submodules) and
// returns the directory containing it, or "" when dir is not in a worktree.
func findRepoRoot(dir string) string {
	for {
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// loadRepoGitInfo runs git status once for the repository rooted at root
// and indexes its porcelain output.
func loadRepoGitInfo(root string) *repoGitInfo {
	out, err := exec.Command("git", "--no-optional-locks", "-C", root,
		"status", "--porcelain", "-z", "--ignored").Output()
	if err != nil {
		return nil
	}
	info := &repoGitInfo{
		files: make(map[string]GitState),
		dirs:  make(map[string]GitState),
	}
	// Records are NUL-terminated "XY path"; staged renames and copies
	// append the source path as an extra NUL-terminated field.
	recs := strings.Split(string(out), "\x00")
	for i := 0; i < len(recs); i++ {
		rec := recs[i]
		if len(rec) < 4 || rec[2] != ' ' {
			continue
		}
		s := GitState{StatusCode(rec[0]), StatusCode(rec[1])}
		path := rec[3:]
		if s.Staging == 'R' || s.Staging == 'C' {
			i++ // skip the rename/copy source path
		}
		if p, ok := strings.CutSuffix(path, "/"); ok {
			info.dirs[p] = s
			continue
		}
		info.files[path] = s
	}
	return info
}

// foldStatusCode merges a child's code into a directory aggregate, per
// column: real changes collapse to M — the directory itself is modified,
// not deleted or renamed, and folding happens in map order so keeping a
// child's own code would be nondeterministic — and untracked beats
// unmodified.
func foldStatusCode(cur, next StatusCode) StatusCode {
	switch next {
	case StatusUnmodified:
		return cur
	case StatusUntracked:
		if cur == StatusUnmodified {
			return StatusUntracked
		}
		return cur
	default:
		return StatusModified
	}
}
