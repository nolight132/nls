//go:build !linux

package listing

import "os"

func readDirNamesUnsorted(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries)+2)
	names = append(names, ".", "..")
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}
