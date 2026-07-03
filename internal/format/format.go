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

// LsSize formats nbytes for GNU ls-compatible long output.
func LsSize(nbytes int64, human bool, approx bool) string {
	prefix := ""
	if approx {
		prefix = ">"
	}
	if !human {
		return fmt.Sprintf("%s%d", prefix, nbytes)
	}
	return prefix + humanLsSize(nbytes)
}

// LsTotalSize formats the block total line for GNU ls-compatible output.
func LsTotalSize(nbytes int64, human bool) string {
	if !human {
		return fmt.Sprintf("%d", nbytes)
	}
	return humanLsTotalSize(nbytes)
}

// LsBlockSize formats 1K block counts for GNU ls -s-compatible columns.
func LsBlockSize(blocks int64, human bool) string {
	if !human {
		return fmt.Sprintf("%d", blocks)
	}
	return humanLsBlockSize(blocks * 1024)
}

// IsRelativeModified reports whether s is a relative mtime label from Modified().
func IsRelativeModified(s string) bool {
	switch s {
	case "-":
		return false
	case "just now", "a day ago":
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

	suffixes := []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	suffix := suffixes[exp]
	return fmt.Sprintf("%.1f %s", float64(nbytes)/float64(div), suffix)
}

func humanLsTotalSize(nbytes int64) string {
	const unit = 1024
	if nbytes < unit {
		return fmt.Sprintf("%d", nbytes)
	}

	div := int64(unit)
	exp := 0
	for n := nbytes / unit; n >= unit && exp < 5; n /= unit {
		div *= unit
		exp++
	}

	suffixes := []string{"K", "M", "G", "T", "P", "E"}
	value := float64(nbytes) / float64(div)
	if value >= 10 || value == float64(int64(value)) {
		return fmt.Sprintf("%.0f%s", value, suffixes[exp])
	}
	return fmt.Sprintf("%.1f%s", value, suffixes[exp])
}

func humanLsSize(nbytes int64) string {
	const unit = 1024
	if nbytes < unit {
		return fmt.Sprintf("%d", nbytes)
	}

	div := int64(unit)
	exp := 0
	for n := nbytes / unit; n >= unit && exp < 5; n /= unit {
		div *= unit
		exp++
	}

	suffixes := []string{"K", "M", "G", "T", "P", "E"}
	return fmt.Sprintf("%.1f%s", float64(nbytes)/float64(div), suffixes[exp])
}

func humanLsBlockSize(nbytes int64) string {
	const unit = 1024
	if nbytes == 0 {
		return "0"
	}
	if nbytes < unit {
		return fmt.Sprintf("%d", nbytes)
	}

	div := int64(unit)
	exp := 0
	for n := nbytes / unit; n >= unit && exp < 5; n /= unit {
		div *= unit
		exp++
	}

	suffixes := []string{"K", "M", "G", "T", "P", "E"}
	return fmt.Sprintf("%.1f%s", float64(nbytes)/float64(div), suffixes[exp])
}

// LsTime formats mtime like GNU ls -l.
func LsTime(t time.Time, now time.Time, full bool) string {
	if t.IsZero() {
		return "-"
	}
	if full {
		return t.Format("2006-01-02 15:04:05.000000000 -0700")
	}
	const sixMonths = time.Duration(31556952/2) * time.Second
	age := now.Sub(t)
	if age < 0 || age >= sixMonths {
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
		return "a day ago"
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
