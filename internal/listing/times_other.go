//go:build !linux && !darwin

package listing

import (
	"io/fs"
	"time"
)

func fileTimes(info fs.FileInfo) (accessed, changed time.Time) {
	t := info.ModTime()
	return t, t
}
