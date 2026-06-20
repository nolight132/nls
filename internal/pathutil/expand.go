package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Expand resolves ~ and cleans the path.
func Expand(raw string) (string, error) {
	if raw == "" {
		return ".", nil
	}

	if raw == "~" || strings.HasPrefix(raw, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home directory: %w", err)
		}
		if raw == "~" {
			return home, nil
		}
		return filepath.Join(home, raw[2:]), nil
	}

	return filepath.Clean(raw), nil
}
