package listing

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type dirSizeResult struct {
	bytes  int64
	newest time.Time
	approx bool
}

// dirSizeBudget returns the wall-clock budget shared by all directory size
// estimates of one listing; zero means unbounded.
func dirSizeBudget(opts ListOptions) time.Duration {
	if opts.Precise {
		return 0
	}
	switch strings.ToLower(strings.TrimSpace(opts.DirSizeTiming)) {
	case "unlimited":
		return 0
	case "strict":
		return 8 * time.Millisecond
	case "relaxed":
		return 100 * time.Millisecond
	default:
		return 20 * time.Millisecond
	}
}

// estimateDirectorySizes fills Size for directory entries by summing file
// contents. Bounded estimation shares one wall-clock budget across all
// directories of the listing: each job claims a fair share of the time left
// when a worker picks it up, so directories that finish early return their
// unused slice to the pool instead of wasting it.
func estimateDirectorySizes(parent string, entries []Entry, opts ListOptions) {
	type job struct {
		idx  int
		path string
	}

	bounded := opts.EstimateDepth == EstimateDepthBounded
	maxDepth := max(opts.EstimateDepth, 0)
	if bounded && opts.DirSizeDepth > 0 {
		maxDepth = opts.DirSizeDepth
	}

	jobs := make([]job, 0, len(entries))
	for i, e := range entries {
		if e.Kind != KindDirectory {
			continue
		}
		// "." re-walks the listed dir and ".." walks its parent; both
		// waste I/O and eat into the shared budget.
		if e.Name == "." || e.Name == ".." {
			continue
		}
		jobs = append(jobs, job{idx: i, path: filepath.Join(parent, e.Name)})
	}
	if len(jobs) == 0 {
		return
	}

	var deadline time.Time
	if bounded {
		if budget := dirSizeBudget(opts); budget > 0 {
			deadline = time.Now().Add(budget)
		}
	}
	workers := min(len(jobs), runtime.NumCPU())

	var pending atomic.Int64
	pending.Store(int64(len(jobs)))

	ch := make(chan job)
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for j := range ch {
				left := pending.Add(-1) + 1
				jobDeadline := deadline
				if !deadline.IsZero() {
					remaining := time.Until(deadline)
					if remaining <= 0 {
						// Not walked: stat size is only a lower bound.
						entries[j.idx].SizeApprox = true
						continue
					}
					// workers jobs run concurrently, so each may claim
					// workers/left of the remaining budget.
					if share := remaining * time.Duration(workers) / time.Duration(left); share < remaining {
						jobDeadline = time.Now().Add(share)
					}
				}
				result := sumDirSize(j.path, jobDeadline, maxDepth)
				entries[j.idx].Size = result.bytes
				entries[j.idx].SizeApprox = result.approx
				if result.newest.After(entries[j.idx].Modified) {
					entries[j.idx].Modified = result.newest
				}
			}
		}()
	}

	for _, j := range jobs {
		ch <- j
	}
	close(ch)
	wg.Wait()
}

func sumDirSize(root string, deadline time.Time, maxDepth int) dirSizeResult {
	var total int64
	var newest time.Time
	truncated := false
	root = filepath.Clean(root)

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if maxDepth > 0 && treeDepth(root, path) > maxDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			truncated = true
			return fs.SkipAll
		}

		if d.IsDir() {
			if info, err := d.Info(); err == nil && info.ModTime().After(newest) {
				newest = info.ModTime()
			}
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil {
			return nil
		}
		total += diskUsageOf(info)
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
		return nil
	})

	return dirSizeResult{bytes: total, newest: newest, approx: truncated}
}

func treeDepth(root, path string) int {
	path = filepath.Clean(path)
	if path == root {
		return 0
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator)) + 1
}
