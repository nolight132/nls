//go:build darwin

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
	accessed = time.Unix(st.Atimespec.Sec, st.Atimespec.Nsec)
	changed = time.Unix(st.Ctimespec.Sec, st.Ctimespec.Nsec)
	return accessed, changed
}
