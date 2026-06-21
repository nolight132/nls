//go:build linux

package listing

import (
	"encoding/binary"
	"os"
	"syscall"
)

func readDirNamesUnsorted(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, 8192)
	var names []string
	for {
		n, err := syscall.Getdents(int(f.Fd()), buf)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			break
		}
		for off := 0; off < n; {
			if off+19 > n {
				break
			}
			reclen := int(binary.LittleEndian.Uint16(buf[off+16:]))
			if reclen <= 0 || off+reclen > n {
				break
			}
			nameBytes := buf[off+19 : off+reclen]
			for i, b := range nameBytes {
				if b == 0 {
					nameBytes = nameBytes[:i]
					break
				}
			}
			if len(nameBytes) > 0 {
				names = append(names, string(nameBytes))
			}
			off += reclen
		}
	}
	return names, nil
}
