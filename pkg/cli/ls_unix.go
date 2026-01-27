//go:build unix

package cli

import (
	"io/fs"
	"syscall"
)

func fillUnixInfo(entry *FileEntry, info fs.FileInfo) {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		entry.Inode = stat.Ino
		entry.NLink = stat.Nlink
		entry.UID = stat.Uid
		entry.GID = stat.Gid
	}
}
