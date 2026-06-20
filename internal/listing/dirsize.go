package listing

import (
	"io/fs"
	"path/filepath"
	"sync"
	"time"
)

const (
	maxDirWalkEntries  = 400
	maxDirWalkDuration = 50 * time.Millisecond
	maxDirWorkers      = 3
	maxDirsPerListing  = 6
	maxListingEstimate = 120 * time.Millisecond
)

type dirSizeResult struct {
	bytes  int64
	approx bool
}

// estimateDirectorySizes fills Size for directory entries by summing file contents.
func estimateDirectorySizes(parent string, entries []Entry) {
	type job struct {
		idx  int
		path string
	}

	jobs := make([]job, 0, maxDirsPerListing)
	for i, e := range entries {
		if e.Kind != KindDirectory {
			continue
		}
		if len(jobs) >= maxDirsPerListing {
			break
		}
		jobs = append(jobs, job{idx: i, path: filepath.Join(parent, e.Name)})
	}
	if len(jobs) == 0 {
		return
	}

	deadline := time.Now().Add(maxListingEstimate)
	workers := min(len(jobs), maxDirWorkers)

	ch := make(chan job)
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for j := range ch {
				if time.Now().After(deadline) {
					continue
				}
				result := sumDirSize(j.path, deadline)
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

func sumDirSize(root string, listingDeadline time.Time) dirSizeResult {
	dirDeadline := time.Now().Add(maxDirWalkDuration)
	if listingDeadline.Before(dirDeadline) {
		dirDeadline = listingDeadline
	}

	var total int64
	var count int
	truncated := false

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if time.Now().After(dirDeadline) {
			truncated = true
			return fs.SkipAll
		}
		if count >= maxDirWalkEntries {
			truncated = true
			return fs.SkipAll
		}
		count++

		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})

	return dirSizeResult{bytes: total, approx: truncated}
}
