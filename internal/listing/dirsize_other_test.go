//go:build !unix

package listing

import "io/fs"

func sparseFixtureDiskUsage(_ fs.FileInfo) (int64, bool) {
	return 0, false
}
