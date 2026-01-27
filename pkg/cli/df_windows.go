//go:build windows

package cli

import (
	"syscall"
	"unsafe"
)

// getDiskInfo returns disk usage information for a path on Windows
func getDiskInfo(path string) (DFInfo, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return DFInfo{}, err
	}

	ret, _, err := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		return DFInfo{}, err
	}

	used := totalBytes - totalFreeBytes
	usePercent := 0
	if totalBytes > 0 {
		usePercent = int(float64(used) / float64(totalBytes) * 100)
	}

	return DFInfo{
		Filesystem: path,
		Size:       totalBytes,
		Used:       used,
		Available:  freeBytesAvailable,
		UsePercent: usePercent,
		MountedOn:  path,
		// Windows doesn't expose inode info in the same way
		Inodes:      0,
		IUsed:       0,
		IFree:       0,
		IUsePercent: 0,
	}, nil
}
