package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Expand resolves ~ and cleans the path. A trailing separator is kept so
// downstream stats follow a final symlink (POSIX "link/").
func Expand(raw string) (string, error) {
	if raw == "" {
		_, err := os.Lstat(raw)
		return "", err
	}

	if raw == "~" || strings.HasPrefix(raw, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home directory: %w", err)
		}
		if raw == "~" {
			return home, nil
		}
		return keepTrailingSep(raw, filepath.Join(home, raw[2:])), nil
	}

	return keepTrailingSep(raw, filepath.Clean(raw)), nil
}

func keepTrailingSep(raw, cleaned string) string {
	if os.IsPathSeparator(raw[len(raw)-1]) && !os.IsPathSeparator(cleaned[len(cleaned)-1]) {
		return cleaned + string(os.PathSeparator)
	}
	return cleaned
}
