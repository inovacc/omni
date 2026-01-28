//go:build unix

package df

import (
	"syscall"
)

// getDiskInfo returns disk usage information for a path on Unix systems
func getDiskInfo(path string) (DFInfo, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return DFInfo{}, err
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
