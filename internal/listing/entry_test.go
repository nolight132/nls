package listing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDirSortsAlphabetically(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"zebra.txt", "alpha", "mango"} {
		path := filepath.Join(dir, name)
		if name == "alpha" || name == "mango" {
			if err := os.Mkdir(path, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := ReadDir(dir, ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("len = %d, want 3", len(entries))
	}
	if entries[0].Name != "alpha" || entries[1].Name != "mango" || entries[2].Name != "zebra.txt" {
		t.Fatalf("unexpected order: %#v", entries)
	}
}

func TestReadDirHidden(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "visible"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	without, err := ReadDir(dir, ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(without) != 1 || without[0].Name != "visible" {
		t.Fatalf("without all: %#v", without)
	}

	with, err := ReadDir(dir, ListOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(with) != 4 {
		t.Fatalf("with all: got %d entries, want 4 (. .. .hidden visible): %#v", len(with), with)
	}
}

func TestRecursiveAllSkipsDotEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	blocks, err := List([]string{dir}, ListOptions{All: true, Recursive: true, Sort: SortOptions{Field: SortByName}})
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want root and sub only: %#v", len(blocks), blocks)
	}
}

func TestFormatPermissions(t *testing.T) {
	cases := []struct {
		mode os.FileMode
		want string
	}{
		{0o755, "-rwxr-xr-x"},
		{os.ModeDir | 0o755, "drwxr-xr-x"},
		{os.ModeSymlink | 0o777, "lrwxrwxrwx"},
		{os.ModeNamedPipe | 0o644, "prw-r--r--"},
		{os.ModeSocket | 0o755, "srwxr-xr-x"},
		{os.ModeDevice | os.ModeCharDevice | 0o666, "crw-rw-rw-"},
		{os.ModeDevice | 0o660, "brw-rw----"},
		{os.ModeSetuid | 0o755, "-rwsr-xr-x"},
		{os.ModeSetuid | 0o644, "-rwSr--r--"},
		{os.ModeSetgid | 0o755, "-rwxr-sr-x"},
		{os.ModeSetgid | 0o745, "-rwxr-Sr-x"},
		{os.ModeDir | os.ModeSticky | 0o755, "drwxr-xr-t"},
		{os.ModeDir | os.ModeSticky | 0o754, "drwxr-xr-T"},
		{os.ModeSetuid | os.ModeSetgid | os.ModeSticky | 0o777, "-rwsrwsrwt"},
	}
	for _, tc := range cases {
		if got := formatPermissions(tc.mode); got != tc.want {
			t.Errorf("formatPermissions(%v) = %q, want %q", tc.mode, got, tc.want)
		}
	}
}

func TestSymlinkToDirOperand(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "real"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "real", "inside.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "dirlink")
	if err := os.Symlink("real", link); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name string
		path string
		opts ListOptions
	}{
		{"bare operand", link, ListOptions{}},
		{"trailing slash", link + "/", ListOptions{LongListing: true}},
	} {
		blocks, errs := List([]string{tc.path}, tc.opts)
		if len(errs) > 0 {
			t.Fatalf("%s: errs = %v", tc.name, errs)
		}
		if len(blocks) != 1 || len(blocks[0].Entries) != 1 || blocks[0].Entries[0].Name != "inside.txt" {
			t.Fatalf("%s: got %#v, want target contents", tc.name, blocks)
		}
	}

	for _, opts := range []ListOptions{{Directory: true}, {LongListing: true}, {Classify: true}} {
		blocks, errs := List([]string{link}, opts)
		if len(errs) > 0 {
			t.Fatalf("%+v: errs = %v", opts, errs)
		}
		if len(blocks) != 1 || len(blocks[0].Entries) != 1 || blocks[0].Entries[0].Kind != KindSymlink {
			t.Fatalf("%+v: got %#v, want the link itself", opts, blocks)
		}
	}
}
