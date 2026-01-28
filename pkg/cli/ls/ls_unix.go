//go:build unix

package ls

import (
	"io/fs"
	"syscall"
)

func fillUnixInfo(entry *Entry, info fs.FileInfo) {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		entry.Inode = stat.Ino
		entry.NLink = stat.Nlink
		entry.UID = stat.Uid
		entry.GID = stat.Gid
	}
}
