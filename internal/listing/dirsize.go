package listing

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	maxDirWalkEntries        = 400
	maxDirWalkDuration       = 50 * time.Millisecond
	maxDirWorkers            = 3
	maxDirsPerListingDefault = 6
	maxListingEstimate       = 120 * time.Millisecond
)

type dirSizeResult struct {
	bytes  int64
	approx bool
}

// estimateDirectorySizes fills Size for directory entries by summing file contents.
func estimateDirectorySizes(parent string, entries []Entry, depth int, limits Limits) {
	type job struct {
		idx  int
		path string
	}

	bounded := depth == EstimateDepthBounded
	maxMode := depth == EstimateDepthMax
	maxWalkDepth := max(depth, 0)

	if maxMode {
		limits = SafetyLimits()
	}
	if bounded && limits == (Limits{}) {
		limits = DefaultBoundedLimits()
	}
	maxDirs := limits.MaxDirsPerListing
	if maxDirs <= 0 {
		maxDirs = maxDirsPerListingDefault
	}
	maxWalkEntries := limits.MaxWalkEntries
	if maxWalkEntries <= 0 {
		maxWalkEntries = maxDirWalkEntries
	}
	walkBudget := limits.WalkDuration
	if walkBudget <= 0 {
		walkBudget = maxDirWalkDuration
	}
	listingBudget := limits.ListingDuration
	if listingBudget <= 0 {
		listingBudget = maxListingEstimate
	}
	boundedMaxDepth := limits.MaxDepth

	jobs := make([]job, 0, len(entries))
	for i, e := range entries {
		if e.Kind != KindDirectory {
			continue
		}
		if (bounded || maxMode) && len(jobs) >= maxDirs {
			break
		}
		jobs = append(jobs, job{idx: i, path: filepath.Join(parent, e.Name)})
	}
	if len(jobs) == 0 {
		return
	}

	var listingDeadline time.Time
	if bounded {
		listingDeadline = time.Now().Add(listingBudget)
	}
	workers := min(len(jobs), maxDirWorkers)

	ch := make(chan job)
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for j := range ch {
				if bounded && time.Now().After(listingDeadline) {
					continue
				}
				result := sumDirSize(j.path, listingDeadline, bounded, maxWalkDepth, boundedMaxDepth, walkBudget, maxWalkEntries)
				entries[j.idx].Size = result.bytes
				entries[j.idx].SizeApprox = result.approx
			}
		}()
	}

	for _, j := range jobs {
		ch <- j
	}
	close(ch)
	wg.Wait()
}

func sumDirSize(root string, listingDeadline time.Time, bounded bool, maxWalkDepth, boundedMaxDepth int, walkBudget time.Duration, maxWalkEntries int) dirSizeResult {
	var walkDeadline time.Time
	if bounded {
		walkDeadline = time.Now().Add(walkBudget)
		if listingDeadline.Before(walkDeadline) {
			walkDeadline = listingDeadline
		}
	}

	var total int64
	var count int
	truncated := false
	root = filepath.Clean(root)

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		depth := treeDepth(root, path)
		if maxWalkDepth > 0 && depth > maxWalkDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if bounded && boundedMaxDepth > 0 && depth > boundedMaxDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if bounded && time.Now().After(walkDeadline) {
			truncated = true
			return fs.SkipAll
		}
		if count >= maxWalkEntries {
			truncated = true
			return fs.SkipAll
		}
		count++

		if d.IsDir() {
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil {
			return nil
		}
		total += diskUsageOf(info)
		return nil
	})

	return dirSizeResult{bytes: total, approx: truncated}
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
