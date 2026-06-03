//go:build windows

package uname

import (
	"strconv"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// currentVersionKey is the registry path that holds Windows version metadata.
const currentVersionKey = `SOFTWARE\Microsoft\Windows NT\CurrentVersion`

// getKernelRelease returns the Windows version in the traditional
// "Major.Minor.Build.UBR" form (e.g., "10.0.19041.1"), matching the format
// previously produced by parsing "cmd /c ver" output.
//
// Implemented purely in Go with no os/exec: the Major/Minor/Build numbers come
// from windows.RtlGetVersion (an ntdll syscall, not a spawned process) and the
// Update Build Revision (UBR) is read from the registry.
func getKernelRelease() string {
	// RtlGetVersion always succeeds (documented to return STATUS_SUCCESS) and is
	// not subject to application manifest version shimming, unlike GetVersionEx.
	info := windows.RtlGetVersion()
	if info == nil {
		return "unknown"
	}

	release := strconv.FormatUint(uint64(info.MajorVersion), 10) + "." +
		strconv.FormatUint(uint64(info.MinorVersion), 10) + "." +
		strconv.FormatUint(uint64(info.BuildNumber), 10)

	if ubr, ok := getUBR(); ok {
		release += "." + strconv.FormatUint(uint64(ubr), 10)
	}

	return release
}

// getKernelVersion returns the Windows build info. Mirrors the historical
// behavior where kernel version and release reported the same value.
func getKernelVersion() string {
	return getKernelRelease()
}

// getUBR reads the Update Build Revision from the registry. Returns false when
// the value is unavailable so the caller can omit the fourth version segment.
func getUBR() (uint32, bool) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, currentVersionKey, registry.QUERY_VALUE)
	if err != nil {
		return 0, false
	}
	defer func() { _ = k.Close() }()

	ubr, _, err := k.GetIntegerValue("UBR")
	if err != nil {
		return 0, false
	}

	return uint32(ubr), true
}
