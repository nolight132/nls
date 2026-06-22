package listing

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
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

	entries, err := ReadDir(dir, Options{EstimateDepth: EstimateDepthMax})
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

	entries, err := ReadDir(dir, Options{EstimateDepth: 1})
	if err != nil {
		t.Fatal(err)
	}
	docs := findEntry(t, entries, "docs")
	want := diskUsageAt(t, filepath.Join(sub, "a.txt"))
	if docs.Size != want {
		t.Fatalf("depth 1 size = %d, want %d", docs.Size, want)
	}

	entries, err = ReadDir(dir, Options{EstimateDepth: 2})
	if err != nil {
		t.Fatal(err)
	}
	docs = findEntry(t, entries, "docs")
	want = diskUsageAt(t, filepath.Join(sub, "a.txt")) + diskUsageAt(t, filepath.Join(inner, "b.txt"))
	if docs.Size != want {
		t.Fatalf("depth 2 size = %d, want %d", docs.Size, want)
	}

	entries, err = ReadDir(dir, Options{EstimateDepth: 3})
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

	entries, err := ReadDir(dir, Options{EstimateDepth: EstimateDepthMax})
	if err != nil {
		t.Fatal(err)
	}
	data := findEntry(t, entries, "data")
	if data.Size >= (1 << 30) {
		t.Fatalf("estimated size = %d, should not use apparent size", data.Size)
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

func TestEstimateMaxSafetyNetTruncatesHugeTree(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "huge")
	if err := os.Mkdir(big, 0o755); err != nil {
		t.Fatal(err)
	}

	cap := 500
	for i := range cap + 100 {
		if err := os.WriteFile(filepath.Join(big, "f"+strconv.Itoa(i)), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := ReadDir(dir, Options{EstimateDepth: EstimateDepthBounded, BoundedLimits: Limits{
		MaxWalkEntries:    cap,
		MaxDirsPerListing: 10,
		WalkDuration:      0,
		ListingDuration:   0,
	}})
	if err != nil {
		t.Fatal(err)
	}
	huge := findEntry(t, entries, "huge")
	if !huge.SizeApprox {
		t.Fatal("should mark approx when entry cap is exceeded")
	}
}

func TestSafetyLimitsHasNoTimeLimits(t *testing.T) {
	s := SafetyLimits()
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
