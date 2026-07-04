//go:build linux

package listing

import (
	"io/fs"
	"syscall"
	"time"
)

func fileTimes(info fs.FileInfo) (accessed, changed time.Time) {
	accessed = info.ModTime()
	changed = info.ModTime()
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return accessed, changed
	}
	accessed = time.Unix(st.Atim.Sec, st.Atim.Nsec)
	changed = time.Unix(st.Ctim.Sec, st.Ctim.Nsec)
	return accessed, changed
}
