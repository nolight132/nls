//go:build !linux && !darwin

package listing

import (
	"os"
	"time"
)

func fileTimes(info os.FileInfo) (accessed, changed time.Time) {
	t := info.ModTime()
	return t, t
}
