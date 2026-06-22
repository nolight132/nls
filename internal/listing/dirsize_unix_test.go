//go:build unix

package listing

import (
	"os"
	"syscall"
)

func sparseFixtureDiskUsage(info os.FileInfo) (int64, bool) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return int64(st.Blocks) * 512, true
}
