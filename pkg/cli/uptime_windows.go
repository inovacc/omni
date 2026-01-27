//go:build windows

package cli

import (
	"syscall"
	"time"
	"unsafe"
)

func getUptimeInfo() (UptimeInfo, error) {
	var info UptimeInfo

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getTickCount64 := kernel32.NewProc("GetTickCount64")

	ret, _, _ := getTickCount64.Call()
	uptimeMs := uint64(ret)

	info.Uptime = time.Duration(uptimeMs) * time.Millisecond
	info.BootTime = time.Now().Add(-info.Uptime)

	// Windows doesn't have load average concept
	info.LoadAvg1 = 0
	info.LoadAvg5 = 0
	info.LoadAvg15 = 0

	// Get number of active sessions
	info.Users = getWindowsUserCount()

	return info, nil
}

func getWindowsUserCount() int {
	wtsapi32 := syscall.NewLazyDLL("wtsapi32.dll")
	wtsEnumerateSessions := wtsapi32.NewProc("WTSEnumerateSessionsW")
	wtsFreeMemory := wtsapi32.NewProc("WTSFreeMemory")

	var pSessionInfo uintptr
	var count uint32

	ret, _, _ := wtsEnumerateSessions.Call(
		0, // WTS_CURRENT_SERVER_HANDLE
		0,
		1,
		uintptr(unsafe.Pointer(&pSessionInfo)),
		uintptr(unsafe.Pointer(&count)),
	)

	if ret == 0 {
		return 1 // Default to 1 user
	}

	defer func() {
		_, _, _ = wtsFreeMemory.Call(pSessionInfo)
	}()

	// Count active sessions
	activeCount := 0
	const sessionInfoSize = 24 // Size of WTS_SESSION_INFO_1W on 64-bit
	for i := uint32(0); i < count; i++ {
		// Check session state (offset 16 is State field)
		statePtr := pSessionInfo + uintptr(i*sessionInfoSize) + 16
		state := *(*uint32)(unsafe.Pointer(statePtr))
		if state == 0 { // WTSActive
			activeCount++
		}
	}

	if activeCount == 0 {
		activeCount = 1
	}

	return activeCount
}
