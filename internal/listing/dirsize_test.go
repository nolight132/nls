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
