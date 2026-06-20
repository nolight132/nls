//go:build !unix

package listing

import "os"

func inodeOf(info os.FileInfo) uint64 {
	return 0
}
