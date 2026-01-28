//go:build windows

package free

import (
	"syscall"
	"unsafe"
)

type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

func getMemInfo() (MemInfo, error) {
	var info MemInfo

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx := kernel32.NewProc("GlobalMemoryStatusEx")

	var memStatus memoryStatusEx

	memStatus.dwLength = uint32(unsafe.Sizeof(memStatus))

	ret, _, err := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret == 0 {
		return info, err
	}

	info.MemTotal = memStatus.ullTotalPhys
	info.MemFree = memStatus.ullAvailPhys
	info.MemAvailable = memStatus.ullAvailPhys

	// Windows doesn't separate buffers/cached like Linux
	info.Buffers = 0
	info.Cached = 0

	// Page file is roughly equivalent to swap
	info.SwapTotal = memStatus.ullTotalPageFile - memStatus.ullTotalPhys
	info.SwapFree = memStatus.ullAvailPageFile - memStatus.ullAvailPhys

	if info.SwapTotal > memStatus.ullTotalPageFile {
		info.SwapTotal = 0
	}

	if info.SwapFree > memStatus.ullAvailPageFile {
		info.SwapFree = 0
	}

	return info, nil
}
