//go:build !unix

package listing

import "io/fs"

func ownerGroupOf(_ fs.FileInfo) (string, string) {
	return "-", "-"
}
