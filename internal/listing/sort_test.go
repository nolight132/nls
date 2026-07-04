package listing

import (
	"reflect"
	"testing"
)

func entriesFromNames(names ...string) []Entry {
	out := make([]Entry, len(names))
	for i, n := range names {
		out[i] = Entry{Name: n}
	}
	return out
}

func namesOf(entries []Entry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Name
	}
	return out
}

func TestSortEntriesUsesCLocaleByteOrder(t *testing.T) {
	t.Setenv("LC_ALL", "C")
	entries := entriesFromNames("CHANGELOG.md", "LICENSE", "README.md", "cmd", "go.mod", "go.sum", "internal")
	sortEntries(entries, SortOptions{Field: SortByName})

	want := []string{"CHANGELOG.md", "LICENSE", "README.md", "cmd", "go.mod", "go.sum", "internal"}
	if got := namesOf(entries); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestSortEntriesUsesLocaleCollation(t *testing.T) {
	t.Setenv("LC_ALL", "en_US.UTF-8")
	entries := entriesFromNames("CHANGELOG.md", "LICENSE", "README.md", "cmd", "go.mod", "go.sum", "internal", ".hidden")
	sortEntries(entries, SortOptions{Field: SortByName})

	want := []string{"CHANGELOG.md", "cmd", "go.mod", "go.sum", ".hidden", "internal", "LICENSE", "README.md"}
	if got := namesOf(entries); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
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
