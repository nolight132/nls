//go:build !unix

package listing

import "io/fs"

func inodeOf(_ fs.FileInfo) uint64 {
	return 0
}
