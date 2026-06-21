//go:build !unix

package listing

import "os"

func ownerGroupOf(info os.FileInfo) (string, string) {
	return "-", "-"
}
