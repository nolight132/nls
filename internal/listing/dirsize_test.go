package listing

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/nolight132/nls/internal/config"
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

func diskUsageAt(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	return diskUsageOf(info)
}

func TestDiskUsageIgnoresSparseApparentSize(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "data")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(sub, "sparse.bin")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteAt([]byte{1}, (1<<30)-1); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() < (1 << 30) {
		t.Fatalf("apparent size = %d, want >= 1GiB", info.Size())
	}
	want, ok := sparseFixtureDiskUsage(info)
	if !ok || want >= (1<<30) {
		t.Skipf("filesystem does not report sparse disk usage below apparent size")
	}
	if got := diskUsageOf(info); got != want {
		t.Fatalf("disk usage = %d, want %d", got, want)
	}

	entries, err := ReadDir(dir, ListOptions{EstimateSizes: true, EstimateDepth: EstimateDepthMax})
	if err != nil {
		t.Fatal(err)
	}
	data := findEntry(t, entries, "data")
	if data.Size >= (1 << 30) {
		t.Fatalf("estimated size = %d, should not use apparent size", data.Size)
	}
}

func TestUnlimitedTimingEstimatesAllDirectories(t *testing.T) {
	setConfigUserForTest(t, config.Config{DirSize: config.DirSizeConfig{DefaultDepth: 0, Timing: "unlimited"}})
	dir := t.TempDir()
	for i := range maxDirsPerListingDefault + 2 {
		sub := filepath.Join(dir, "dir"+strconv.Itoa(i))
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sub, "file"), make([]byte, 1000), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := ReadDir(dir, ListOptions{EstimateSizes: true, EstimateDepth: EstimateDepthBounded})
	if err != nil {
		t.Fatal(err)
	}
	last := findEntry(t, entries, "dir"+strconv.Itoa(maxDirsPerListingDefault+1))
	if last.Size == 0 {
		t.Fatal("unlimited timing should estimate directories beyond the default cap")
	}
	if last.SizeApprox {
		t.Fatal("unlimited timing should not mark small trees approximate")
	}
}

func TestPreciseEstimatesBeyondEntryCap(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "huge")
	if err := os.Mkdir(big, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := range maxDirWalkEntries + 50 {
		if err := os.WriteFile(filepath.Join(big, "f"+strconv.Itoa(i)), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := ReadDir(dir, ListOptions{EstimateSizes: true, EstimateDepth: EstimateDepthMax, Precise: true})
	if err != nil {
		t.Fatal(err)
	}
	huge := findEntry(t, entries, "huge")
	if huge.Size == 0 {
		t.Fatal("precise should compute a non-zero size")
	}
	if huge.SizeApprox {
		t.Fatal("precise should not mark entry-capped trees approximate")
	}
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

func TestDirSizeCapsForMaxHaveNoTimeLimits(t *testing.T) {
	s := dirSizeCapsFor(EstimateDepthMax, false)
	if s.WalkDuration != 0 {
		t.Fatalf("safety WalkDuration = %v, want 0 (no time limit)", s.WalkDuration)
	}
	if s.ListingDuration != 0 {
		t.Fatalf("safety ListingDuration = %v, want 0 (no time limit)", s.ListingDuration)
	}
	if s.MaxDepth != 0 {
		t.Fatalf("safety MaxDepth = %d, want 0 (unlimited)", s.MaxDepth)
	}
	if s.MaxWalkEntries < 100000 {
		t.Fatalf("safety MaxWalkEntries = %d, want >= 100000", s.MaxWalkEntries)
	}
}

func TestDirSizeCapsUseConfigTimingAndDepth(t *testing.T) {
	setConfigUserForTest(t, config.Config{DirSize: config.DirSizeConfig{DefaultDepth: 3, Timing: "relaxed"}})
	caps := dirSizeCapsFor(EstimateDepthBounded, false)
	if caps.MaxDepth != 3 {
		t.Fatalf("MaxDepth = %d, want 3", caps.MaxDepth)
	}
	if caps.WalkDuration <= maxDirWalkDuration {
		t.Fatalf("relaxed walk budget = %v, want > %v", caps.WalkDuration, maxDirWalkDuration)
	}
}

func TestDirSizeCapsUnlimitedTimingHasNoLimits(t *testing.T) {
	setConfigUserForTest(t, config.Config{DirSize: config.DirSizeConfig{DefaultDepth: 0, Timing: "unlimited"}})
	caps := dirSizeCapsFor(EstimateDepthBounded, false)
	if caps != (dirSizeCaps{}) {
		t.Fatalf("unlimited caps = %+v, want zero caps", caps)
	}
}

func TestPreciseIgnoresTimingLimits(t *testing.T) {
	setConfigUserForTest(t, config.Config{DirSize: config.DirSizeConfig{DefaultDepth: 2, Timing: "strict"}})
	caps := dirSizeCapsFor(EstimateDepthMax, true)
	if caps != (dirSizeCaps{}) {
		t.Fatalf("precise caps = %+v, want zero caps", caps)
	}
}

func setConfigUserForTest(t *testing.T, cfg config.Config) {
	t.Helper()
	prev := config.User
	config.User = cfg
	t.Cleanup(func() { config.User = prev })
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
