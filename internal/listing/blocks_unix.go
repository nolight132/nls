//go:build unix

package listing

import (
	"io/fs"
	"syscall"
)

func blocksOf(info fs.FileInfo) int64 {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fallbackBlocks(info)
	}
	return (st.Blocks + 1) / 2
}

func diskUsageOf(info fs.FileInfo) int64 {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fallbackDiskUsage(info)
	}
	return int64(st.Blocks) * 512
}

func linksOf(info fs.FileInfo) uint64 {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 1
	}
	return uint64(st.Nlink)
}
