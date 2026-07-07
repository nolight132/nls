//go:build unix

package listing

import (
	"io/fs"
	"os/user"
	"strconv"
	"syscall"
)

// Lookups go through NSS and can cost an IPC round-trip each; a listing
// rarely has more than a handful of distinct ids, so memoize per process.
var (
	userNames  = map[uint32]string{}
	groupNames = map[uint32]string{}
)

func ownerGroupOf(info fs.FileInfo) (string, string) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "-", "-"
	}
	return userName(uint32(st.Uid)), groupName(uint32(st.Gid))
}

func userName(uid uint32) string {
	if name, ok := userNames[uid]; ok {
		return name
	}
	id := strconv.FormatUint(uint64(uid), 10)
	name := id
	if u, err := user.LookupId(id); err == nil {
		name = u.Username
	}
	userNames[uid] = name
	return name
}

func groupName(gid uint32) string {
	if name, ok := groupNames[gid]; ok {
		return name
	}
	id := strconv.FormatUint(uint64(gid), 10)
	name := id
	if g, err := user.LookupGroupId(id); err == nil {
		name = g.Name
	}
	groupNames[gid] = name
	return name
}
