//go:build windows

package ls

import "io/fs"

func fillUnixInfo(entry *Entry, info fs.FileInfo) {
	// Windows doesn't have Unix-style inodes, UIDs, GIDs
	// These fields will remain at their zero values
}
