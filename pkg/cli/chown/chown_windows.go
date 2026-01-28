//go:build windows

package chown

import (
	"os"
)

// getFileOwner returns the UID and GID of a file (not supported on Windows)
func getFileOwner(info os.FileInfo) (int, int, error) {
	// Windows doesn't have Unix-style UID/GID
	return -1, -1, nil
}
