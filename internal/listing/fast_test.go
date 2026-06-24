package listing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFastListNames(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"zebra", "alpha", "mango"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	names, err := FastListNames(dir, ListOptions{Sort: SortOptions{Field: SortByName}})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 3 || names[0] != "alpha" || names[2] != "zebra" {
		t.Fatalf("got %v", names)
	}
}

func TestFastListNamesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	names, err := FastListNames(path, ListOptions{Sort: SortOptions{Field: SortByName}})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != path {
		t.Fatalf("got %v", names)
	}
}

func TestCanFastListDefault(t *testing.T) {
	if !CanFastList(ListOptions{Sort: SortOptions{Field: SortByName}}) {
		t.Fatal("expected fast list for default name sort")
	}
	if CanFastList(ListOptions{LongListing: true}) {
		t.Fatal("long listing should not fast list")
	}
}
