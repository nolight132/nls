//go:build !unix

package listing

import "os"

func blocksOf(info os.FileInfo) int64 {
	return fallbackBlocks(info)
}

func diskUsageOf(info os.FileInfo) int64 {
	return fallbackDiskUsage(info)
}

func linksOf(info os.FileInfo) uint64 {
	return 1
}
