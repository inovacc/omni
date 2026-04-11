//go:build unix

package df

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// getDiskInfo returns disk usage information for a path on Unix systems
func getDiskInfo(path string) (DFInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			return DFInfo{}, cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("df: %s: %v", path, err))
		case errors.Is(err, os.ErrPermission):
			return DFInfo{}, cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("df: %s: %v", path, err))
		default:
			return DFInfo{}, cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("df: %s: %v", path, err))
		}
	}

	blockSize := uint64(stat.Bsize)
	total := stat.Blocks * blockSize
	free := stat.Bavail * blockSize
	used := total - (stat.Bfree * blockSize)

	usePercent := 0
	if total > 0 {
		usePercent = int(float64(used) / float64(total) * 100)
	}

	iusePercent := 0
	if stat.Files > 0 {
		iusePercent = int(float64(stat.Files-stat.Ffree) / float64(stat.Files) * 100)
	}

	return DFInfo{
		Filesystem:  path,
		Size:        total,
		Used:        used,
		Available:   free,
		UsePercent:  usePercent,
		MountedOn:   path,
		Inodes:      stat.Files,
		IUsed:       stat.Files - stat.Ffree,
		IFree:       stat.Ffree,
		IUsePercent: iusePercent,
	}, nil
}
