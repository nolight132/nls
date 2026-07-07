package listing

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestEstimateDirectorySizesUnlimited(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "a.txt"), make([]byte, 1000), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), make([]byte, 500), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), make([]byte, 100), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err := ReadDir(dir, ListOptions{EstimateSizes: true, EstimateDepth: EstimateDepthMax})
	if err != nil {
		t.Fatal(err)
	}

	docs := findEntry(t, entries, "docs")
	want := diskUsageAt(t, filepath.Join(sub, "a.txt")) + diskUsageAt(t, filepath.Join(sub, "b.txt"))
	if docs.Size != want {
		t.Fatalf("docs size = %d, want %d", docs.Size, want)
	}
	if docs.SizeApprox {
		t.Fatal("small tree should not hit safety caps")
	}
}

func TestEstimateDirectorySizesRespectsWalkDepth(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	inner := filepath.Join(sub, "nested")
	deep := filepath.Join(inner, "deep")
	for _, path := range []string{sub, inner, deep} {
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(sub, "a.txt"), make([]byte, 1000), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inner, "b.txt"), make([]byte, 500), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deep, "c.txt"), make([]byte, 10000), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err := ReadDir(dir, ListOptions{EstimateSizes: true, EstimateDepth: 1})
	if err != nil {
		t.Fatal(err)
	}
	docs := findEntry(t, entries, "docs")
	want := diskUsageAt(t, filepath.Join(sub, "a.txt"))
	if docs.Size != want {
		t.Fatalf("depth 1 size = %d, want %d", docs.Size, want)
	}

	entries, err = ReadDir(dir, ListOptions{EstimateSizes: true, EstimateDepth: 2})
	if err != nil {
		t.Fatal(err)
	}
	docs = findEntry(t, entries, "docs")
	want = diskUsageAt(t, filepath.Join(sub, "a.txt")) + diskUsageAt(t, filepath.Join(inner, "b.txt"))
	if docs.Size != want {
		t.Fatalf("depth 2 size = %d, want %d", docs.Size, want)
	}

	entries, err = ReadDir(dir, ListOptions{EstimateSizes: true, EstimateDepth: 3})
	if err != nil {
		t.Fatal(err)
	}
	docs = findEntry(t, entries, "docs")
	want = diskUsageAt(t, filepath.Join(sub, "a.txt")) +
		diskUsageAt(t, filepath.Join(inner, "b.txt")) +
		diskUsageAt(t, filepath.Join(deep, "c.txt"))
	if docs.Size != want {
		t.Fatalf("depth 3 size = %d, want %d", docs.Size, want)
	}
}

func TestEstimateDirectorySizesBubblesNewestMtime(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	inner := filepath.Join(sub, "nested")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(inner, "a.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Deep file changed recently; the directories themselves have not,
	// mirroring an edit that adds no direct children.
	fileTime := time.Now().Add(-time.Hour).Truncate(time.Second)
	dirTime := fileTime.Add(-24 * time.Hour)
	if err := os.Chtimes(file, fileTime, fileTime); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{inner, sub} {
		if err := os.Chtimes(p, dirTime, dirTime); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := ReadDir(dir, ListOptions{EstimateSizes: true, EstimateDepth: EstimateDepthMax})
	if err != nil {
		t.Fatal(err)
	}
	docs := findEntry(t, entries, "docs")
	if !docs.Modified.Equal(fileTime) {
		t.Fatalf("docs modified = %v, want newest nested change %v", docs.Modified, fileTime)
	}
}

func diskUsageAt(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	return diskUsageOf(info)
}

func findEntry(t *testing.T, entries []Entry, name string) Entry {
	t.Helper()
	for _, e := range entries {
		if e.Name == name {
			return e
		}
	}
	t.Fatalf("%q not found", name)
	return Entry{}
}

func TestSumDirSizeMarksApproxWhenEntryCapExceeded(t *testing.T) {
	dir := t.TempDir()
	cap := 500
	for i := range cap + 100 {
		if err := os.WriteFile(filepath.Join(dir, "f"+strconv.Itoa(i)), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got := sumDirSize(dir, time.Time{}, false, 0, 0, 0, cap)
	if !got.approx {
		t.Fatal("should mark approx when entry cap is exceeded")
	}
}

func TestEstimateDirectorySizesMarksSkippedDirsApprox(t *testing.T) {
	dir := t.TempDir()
	// Strict timing caps MaxDirsPerListing at 4; dirs 5-6 are never
	// walked and must be flagged so their stat size reads as a lower bound.
	entries := make([]Entry, 0, 6)
	for i := range 6 {
		name := "d" + strconv.Itoa(i)
		sub := filepath.Join(dir, name)
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sub, "a.txt"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		entries = append(entries, Entry{Name: name, Kind: KindDirectory})
	}

	estimateDirectorySizes(dir, entries, ListOptions{
		EstimateSizes: true,
		EstimateDepth: EstimateDepthBounded,
		DirSizeTiming: "strict",
	})

	for _, e := range entries[4:] {
		if !e.SizeApprox {
			t.Fatalf("%s skipped by dir cap, should be marked approx", e.Name)
		}
	}
}

func TestDirSizeCapsUnlimitedTimingHasNoLimits(t *testing.T) {
	caps := dirSizeCapsFor(ListOptions{EstimateDepth: EstimateDepthBounded, DirSizeTiming: "unlimited"})
	if caps != (dirSizeCaps{}) {
		t.Fatalf("unlimited caps = %+v, want zero caps", caps)
	}
}

func TestPreciseIgnoresTimingLimits(t *testing.T) {
	caps := dirSizeCapsFor(ListOptions{EstimateDepth: EstimateDepthMax, Precise: true, DirSizeDepth: 2, DirSizeTiming: "strict"})
	if caps != (dirSizeCaps{}) {
		t.Fatalf("precise caps = %+v, want zero caps", caps)
	}
}

func TestTreeDepth(t *testing.T) {
	root := "/home/nolight"
	cases := map[string]int{
		root:                          0,
		"/home/nolight/file":          1,
		"/home/nolight/a/b":           2,
		"/home/nolight/a/b/c.txt":     3,
		"/home/nolight/a/b/c/d/e.txt": 5,
	}
	for path, want := range cases {
		if got := treeDepth(root, path); got != want {
			t.Fatalf("treeDepth(%q) = %d, want %d", path, got, want)
		}
	}
}
