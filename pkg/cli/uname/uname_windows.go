//go:build windows

package uname

import (
	"os/exec"
	"strings"
)

// getKernelRelease returns the Windows version
func getKernelRelease() string {
	// Try to get Windows version info
	out, err := exec.Command("cmd", "/c", "ver").Output()
	if err != nil {
		return "unknown"
	}
	// Parse output like "Microsoft Windows [Version 10.0.19041.1]"
	version := strings.TrimSpace(string(out))
	if idx := strings.Index(version, "["); idx != -1 {
		if endIdx := strings.Index(version, "]"); endIdx != -1 {
			return strings.TrimPrefix(version[idx+1:endIdx], "Version ")
		}
	}

	return "unknown"
}

// getKernelVersion returns the Windows build info
func getKernelVersion() string {
	return getKernelRelease()
}
