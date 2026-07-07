package listing

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v6/osfs"
	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/format/gitignore"
)

// gitStatusCache shares one status computation per repository across all
// directories of a single List call: repo discovery is a cheap upward stat
// walk, but Status and ReadPatterns scan the worktree and must run once.
// A nil info records a failed repo so it is not retried per directory.
type gitStatusCache map[string]*repoGitInfo

type repoGitInfo struct {
	status  git.Status
	matcher gitignore.Matcher
}

// decorate fills GitStatus for entries listed from dir and reports whether
// dir belongs to a git worktree. Any failure (not a repository, unreadable
// index) leaves the entries untouched and returns false.
func (c gitStatusCache) decorate(dir string, entries []Entry) bool {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	repo, err := git.PlainOpenWithOptions(absDir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return false
	}
	wt, err := repo.Worktree()
	if err != nil {
		return false
	}
	root := wt.Filesystem().Root()

	info, ok := c[root]
	if !ok {
		info = loadRepoGitInfo(wt)
		c[root] = info
	}
	if info == nil {
		return false
	}
	status, matcher := info.status, info.matcher

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

	// One pass over the status map (only changed files appear in it; clean
	// files are absent, which is also why Status.File is unusable here — it
	// fabricates untracked placeholders for unknown paths). Exact keys are
	// files; nested keys fold into the directory entry they live under,
	// since git tracks files only and "sub/" never appears as a key.
	type dirAgg struct{ staging, worktree git.StatusCode }
	aggs := make(map[int]*dirAgg)
	for key, s := range status {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		name, _, nested := strings.Cut(key[len(prefix):], "/")
		i, ok := byName[name]
		if !ok {
			continue
		}
		if !nested {
			entries[i].GitStatus = gitStatusDisplay(s.Staging, s.Worktree)
			continue
		}
		a := aggs[i]
		if a == nil {
			a = &dirAgg{staging: git.Unmodified, worktree: git.Unmodified}
			aggs[i] = a
		}
		a.staging = foldStatusCode(a.staging, s.Staging)
		a.worktree = foldStatusCode(a.worktree, s.Worktree)
	}
	for i, a := range aggs {
		if entries[i].GitStatus == "" {
			entries[i].GitStatus = gitStatusDisplay(a.staging, a.worktree)
		}
	}

	// Entries absent from the status map are tracked-and-clean ("-:-") or
	// gitignored ("-I-"). The same patterns Status used decide which.
	var base []string
	if prefix != "" {
		base = strings.Split(strings.TrimSuffix(prefix, "/"), "/")
	}
	for i := range entries {
		e := &entries[i]
		if e.Name == "." || e.Name == ".." || e.GitStatus != "" {
			continue
		}
		segs := append(base[:len(base):len(base)], e.Name)
		if e.Name == ".git" || matcher.Match(segs, e.Kind == KindDirectory) {
			e.GitStatus = gitStatusIgnoredDisplay()
			continue
		}
		e.GitStatus = gitStatusDisplay(git.Unmodified, git.Unmodified)
	}
	return true
}

// loadRepoGitInfo computes the worktree status and ignore matcher for one
// repository; nil means the repo is unusable.
func loadRepoGitInfo(wt *git.Worktree) *repoGitInfo {
	wt.Excludes = globalIgnorePatterns()
	status, err := wt.Status()
	if err != nil {
		return nil
	}
	patterns, _ := gitignore.ReadPatterns(wt.Filesystem(), nil)
	return &repoGitInfo{
		status:  status,
		matcher: gitignore.NewMatcher(append(patterns, wt.Excludes...)),
	}
}

// globalIgnorePatterns loads the user's global gitignore the way git does:
// core.excludesFile from the global config when set, otherwise the XDG
// default ($XDG_CONFIG_HOME/git/ignore, falling back to ~/.config/git/ignore).
// go-git only implements the former.
func globalIgnorePatterns() []gitignore.Pattern {
	if ps, err := gitignore.LoadGlobalPatterns(osfs.New("/")); err == nil && len(ps) > 0 {
		return ps
	}

	confDir := os.Getenv("XDG_CONFIG_HOME")
	if confDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		confDir = filepath.Join(home, ".config")
	}
	data, err := os.ReadFile(filepath.Join(confDir, "git", "ignore"))
	if err != nil {
		return nil
	}

	var ps []gitignore.Pattern
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ps = append(ps, gitignore.ParsePattern(line, nil))
	}
	return ps
}

// foldStatusCode merges a child's code into a directory aggregate, per
// column: a real change beats untracked, untracked beats unmodified, and
// among real changes the first one seen wins.
func foldStatusCode(cur, next git.StatusCode) git.StatusCode {
	switch {
	case next == git.Unmodified:
		return cur
	case cur == git.Unmodified || cur == git.Untracked:
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

// gitStatusDisplay renders a status pair as three fixed characters
func gitStatusDisplay(staging, worktree git.StatusCode) string {
	return string([]rune{rune(displayCode(staging)), GitStatusSeparator, rune(displayCode(worktree))})
}

func gitStatusIgnoredDisplay() string {
	return string([]rune{gitStatusIgnoredMark, GitStatusSeparator, ' '})
}

func displayCode(c git.StatusCode) byte {
	if c == git.Unmodified {
		return ' '
	}
	return byte(c)
}
