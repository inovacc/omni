//go:build windows

package cli

import "io/fs"

func fillUnixInfo(entry *FileEntry, info fs.FileInfo) {
	// Windows doesn't have Unix-style inodes, UIDs, GIDs
	// These fields will remain at their zero values
}
