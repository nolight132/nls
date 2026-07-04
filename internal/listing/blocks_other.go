//go:build !unix

package listing

import "io/fs"

func blocksOf(info fs.FileInfo) int64 {
	return fallbackBlocks(info)
}

func diskUsageOf(info fs.FileInfo) int64 {
	return fallbackDiskUsage(info)
}

func linksOf(_ fs.FileInfo) uint64 {
	return 1
}
