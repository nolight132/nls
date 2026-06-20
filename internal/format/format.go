package format

import (
	"fmt"
	"strings"
	"time"
)

// Size formats nbytes for display.
func Size(nbytes int64, human bool, approx bool) string {
	prefix := ""
	if approx {
		prefix = ">" // Safe assumption. Avoids under-estimating size.
	}
	if !human {
		return fmt.Sprintf("%s%d", prefix, nbytes)
	}
	return prefix + humanSize(nbytes)
}

// IsRelativeModified reports whether s is a relative mtime label from Modified().
func IsRelativeModified(s string) bool {
	switch s {
	case "-":
		return false
	case "just now", "yesterday":
		return true
	}
	return strings.HasSuffix(s, " ago")
}

func humanSize(nbytes int64) string {
	const unit = 1024
	if nbytes < unit {
		return fmt.Sprintf("%d B", nbytes)
	}

	div, exp := int64(unit), 0
	for n := nbytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	suffixes := []string{"kB", "MB", "GB", "TB", "PB", "EB"}
	if exp >= len(suffixes) {
		exp = len(suffixes) - 1
		div = 1
		for range exp + 1 {
			div *= unit
		}
	}
	suffix := suffixes[exp]
	return fmt.Sprintf("%.1f %s", float64(nbytes)/float64(div), suffix)
}

// LsTime formats mtime like GNU ls -l.
func LsTime(t time.Time, now time.Time, full bool) string {
	if t.IsZero() {
		return "-"
	}
	if full {
		return t.Format("2006-01-02 15:04:05.000000000 -0700")
	}
	if now.Year() != t.Year() {
		return t.Format("Jan _2  2006")
	}
	return t.Format("Jan _2 15:04")
}

// Modified formats mtime for display.
func Modified(t time.Time, now time.Time) string {
	if t.IsZero() {
		return "-"
	}

	diff := now.Sub(t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 48*time.Hour:
		return "yesterday"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / (24 * 30))
		if months <= 1 {
			return "a month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(diff.Hours() / (24 * 365))
		if years <= 1 {
			return "a year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
