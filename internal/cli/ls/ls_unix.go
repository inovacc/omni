//go:build unix

package ls

import (
	"io/fs"
	"syscall"
)

func fillUnixInfo(entry *Entry, info fs.FileInfo) {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		// syscall.Stat_t field widths vary by GOOS/GOARCH (e.g. Nlink is
		// uint16 on darwin, uint64 on linux/amd64, int* on some BSDs).
		// Convert explicitly so this compiles on every supported platform.
		entry.Inode = uint64(stat.Ino)
		entry.NLink = uint64(stat.Nlink)
		entry.UID = uint32(stat.Uid)
		entry.GID = uint32(stat.Gid)
	}
}
