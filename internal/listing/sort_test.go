package listing

import (
	"reflect"
	"testing"
)

func TestSortNamesUsesCLocaleByteOrder(t *testing.T) {
	t.Setenv("LC_ALL", "C")
	names := []string{"CHANGELOG.md", "LICENSE", "README.md", "cmd", "go.mod", "go.sum", "internal"}
	sortNames(names, SortOptions{Field: SortByName})

	want := []string{"CHANGELOG.md", "LICENSE", "README.md", "cmd", "go.mod", "go.sum", "internal"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("got %v, want %v", names, want)
	}
}

func TestSortNamesUsesLocaleCollation(t *testing.T) {
	t.Setenv("LC_ALL", "en_US.UTF-8")
	names := []string{"CHANGELOG.md", "LICENSE", "README.md", "cmd", "go.mod", "go.sum", "internal", ".hidden"}
	sortNames(names, SortOptions{Field: SortByName})

	want := []string{"CHANGELOG.md", "cmd", "go.mod", "go.sum", ".hidden", "internal", "LICENSE", "README.md"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("got %v, want %v", names, want)
	}
}

func TestSortEntriesGroupsSymlinkedDirectoriesFirst(t *testing.T) {
	t.Setenv("LC_ALL", "C")
	entries := []Entry{
		{Name: "file", Kind: KindFile},
		{Name: "link-dir", Kind: KindSymlink, LinkTargetDir: true},
		{Name: "dir", Kind: KindDirectory},
	}
	sortEntries(entries, SortOptions{Field: SortByName, DirsFirst: true})

	want := []string{"dir", "link-dir", "file"}
	got := []string{entries[0].Name, entries[1].Name, entries[2].Name}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
