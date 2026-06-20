package listing

import (
	"strings"
)

func includeName(name string, opts Options) bool {
	if opts.IgnoreBackups && strings.HasSuffix(name, "~") {
		return false
	}
	if name == "." || name == ".." {
		return opts.All
	}
	if strings.HasPrefix(name, ".") {
		return opts.All || opts.AlmostAll
	}
	return true
}
