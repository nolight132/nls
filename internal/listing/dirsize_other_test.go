//go:build !unix

package listing

import "os"

func sparseFixtureDiskUsage(info os.FileInfo) (int64, bool) {
	return 0, false
}
