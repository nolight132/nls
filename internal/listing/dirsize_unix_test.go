//go:build unix

package listing

import (
	"io/fs"
	"syscall"
)

func sparseFixtureDiskUsage(info fs.FileInfo) (int64, bool) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return int64(st.Blocks) * 512, true
}
