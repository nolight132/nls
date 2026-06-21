package listing

import "os"

func fallbackBlocks(info os.FileInfo) int64 {
	const blockSize = 1024
	if info.Size() == 0 {
		return 0
	}
	return (info.Size() + blockSize - 1) / blockSize
}
