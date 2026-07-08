package listing

import (
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

// statusCode is one column of a porcelain-v1 XY pair; the byte is git's
// own display character.
type statusCode byte

const (
	statusUnmodified statusCode = ' '
	statusUntracked  statusCode = '?'
	statusIgnored    statusCode = '!'
)

type statusPair struct{ staging, worktree statusCode }

type repoGitInfo struct {
	// files maps repo-root-relative slash paths of changed, untracked, or
	// ignored files to their porcelain XY pair.
	files map[string]statusPair
	// dirs holds directories git reported collapsed because everything
	// beneath them is untracked or ignored (trailing "/" stripped); their
	// contents never appear in files.
	dirs map[string]statusPair
}

// decorate fills GitStatus for entries listed from dir and reports whether
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
	type dirAgg struct{ staging, worktree statusCode }
	aggs := make(map[int]*dirAgg)
	fold := func(key string, s statusPair) {
		if !strings.HasPrefix(key, prefix) {
			return
		}
		name, _, nested := strings.Cut(key[len(prefix):], "/")
		i, ok := byName[name]
		if !ok {
			return
		}
		if !nested {
			setEntryStatus(&entries[i], s)
			return
		}
		if s.staging == statusIgnored {
			return
		}
		a := aggs[i]
		if a == nil {
			a = &dirAgg{staging: statusUnmodified, worktree: statusUnmodified}
			aggs[i] = a
		}
		a.staging = foldStatusCode(a.staging, s.staging)
		a.worktree = foldStatusCode(a.worktree, s.worktree)
	}
	for key, s := range info.files {
		fold(key, s)
	}
	for key, s := range info.dirs {
		fold(key, s)
	}
	for i, a := range aggs {
		if entries[i].GitStatus == "" {
			setEntryStatus(&entries[i], statusPair{a.staging, a.worktree})
		}
	}

	// Entries absent from the status maps are tracked-and-clean
	// unless they live under a collapsed untracked or ignored directory,
	// which git reports only as the directory itself.
	for i := range entries {
		e := &entries[i]
		if e.GitStatus != "" {
			continue
		}
		// Dot entries get the neutral cell so the divider line stays
		// unbroken; their state is not meaningful, so it stays None.
		if e.Name == "." || e.Name == ".." {
			e.GitStatus = gitStatusDisplay(statusUnmodified, statusUnmodified)
			continue
		}
		if e.Name == ".git" {
			e.GitStatus = gitStatusIgnoredDisplay()
			e.GitState = GitStateIgnored
			continue
		}
		if s, ok := info.collapsed(prefix + e.Name); ok {
			setEntryStatus(e, s)
			continue
		}
		e.GitStatus = gitStatusDisplay(statusUnmodified, statusUnmodified)
		e.GitState = GitStateClean
	}
	return true
}

func setEntryStatus(e *Entry, s statusPair) {
	if s.staging == statusIgnored {
		e.GitStatus = gitStatusIgnoredDisplay()
		e.GitState = GitStateIgnored
		return
	}
	e.GitStatus = gitStatusDisplay(s.staging, s.worktree)
	e.GitState = gitStateOf(s.staging, s.worktree)
}

func gitStateOf(staging, worktree statusCode) GitState {
	switch {
	case staging == statusUntracked && worktree == statusUntracked:
		return GitStateUntracked
	case staging == statusUnmodified && worktree == statusUnmodified:
		return GitStateClean
	default:
		return GitStateModified
	}
}

// collapsed reports the status of the collapsed untracked or ignored
// directory that path is or lives under, if any.
func (info *repoGitInfo) collapsed(path string) (statusPair, bool) {
	for {
		if s, ok := info.dirs[path]; ok {
			return s, true
		}
		i := strings.LastIndexByte(path, '/')
		if i < 0 {
			return statusPair{}, false
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
		files: make(map[string]statusPair),
		dirs:  make(map[string]statusPair),
	}
	// Records are NUL-terminated "XY path"; staged renames and copies
	// append the source path as an extra NUL-terminated field.
	recs := strings.Split(string(out), "\x00")
	for i := 0; i < len(recs); i++ {
		rec := recs[i]
		if len(rec) < 4 || rec[2] != ' ' {
			continue
		}
		s := statusPair{statusCode(rec[0]), statusCode(rec[1])}
		path := rec[3:]
		if s.staging == 'R' || s.staging == 'C' {
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
// column: a real change beats untracked, untracked beats unmodified, and
// among real changes the first one seen wins.
func foldStatusCode(cur, next statusCode) statusCode {
	switch {
	case next == statusUnmodified:
		return cur
	case cur == statusUnmodified || cur == statusUntracked:
		return next
	default:
		return cur
	}
}

// GitStatusSeparator sits between the staging and worktree columns of the
// git status cell; exported so the table renderer can hook its borders
// into it when it is a box-drawing glyph. gitStatusIgnoredMark fills the
// staging slot for gitignored entries. Variables so the glyphs can change
// later.
var (
	GitStatusSeparator   rune = '│'
	gitStatusIgnoredMark rune = 'I'
)

func gitStatusDisplay(staging, worktree statusCode) string {
	return string([]rune{rune(staging), GitStatusSeparator, rune(worktree)})
}

func gitStatusIgnoredDisplay() string {
	return string([]rune{gitStatusIgnoredMark, GitStatusSeparator, ' '})
}
