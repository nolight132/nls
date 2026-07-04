//go:build unix

package listing

import (
	"io/fs"
	"syscall"
)

func inodeOf(info fs.FileInfo) uint64 {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0
	}
	return uint64(st.Ino)
}
