//go:build unix

package cli

import (
	"os"
	"syscall"
)

// getFileOwner returns the UID and GID of a file
func getFileOwner(info os.FileInfo) (int, int, error) {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return int(stat.Uid), int(stat.Gid), nil
	}
	return -1, -1, nil
}
