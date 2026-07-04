package listing

import "io/fs"

func fallbackBlocks(info fs.FileInfo) int64 {
	const blockSize = 1024
	if info.Size() == 0 {
		return 0
	}
	return (info.Size() + blockSize - 1) / blockSize
}

func fallbackDiskUsage(info fs.FileInfo) int64 {
	return fallbackBlocks(info) * 1024
}
