//go:build unix

package cli

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func getUptimeInfo() (UptimeInfo, error) {
	var info UptimeInfo

	// Get uptime from sysinfo
	var sysinfo syscall.Sysinfo_t
	if err := syscall.Sysinfo(&sysinfo); err != nil {
		return info, err
	}

	info.Uptime = time.Duration(sysinfo.Uptime) * time.Second
	info.BootTime = time.Now().Add(-info.Uptime)

	// Get load averages
	// sysinfo.Loads is [3]uint64, scaled by 65536
	info.LoadAvg1 = float64(sysinfo.Loads[0]) / 65536.0
	info.LoadAvg5 = float64(sysinfo.Loads[1]) / 65536.0
	info.LoadAvg15 = float64(sysinfo.Loads[2]) / 65536.0

	// Count users from /var/run/utmp or similar
	info.Users = countUsers()

	return info, nil
}

func countUsers() int {
	// Try to read /proc/1/stat to see how many login processes
	// This is a simplified approach
	content, err := os.ReadFile("/var/run/utmp")
	if err != nil {
		return 1 // Assume at least 1 user
	}

	// Very simplified utmp parsing - count non-empty records
	// Real utmp records are 384 bytes on 64-bit systems
	recordSize := 384
	count := 0
	for i := 0; i+recordSize <= len(content); i += recordSize {
		// Check if record type is USER_PROCESS (7)
		if content[i] == 7 {
			count++
		}
	}

	if count == 0 {
		count = 1
	}

	return count
}

// getLoadAvg reads load average from /proc/loadavg (alternative method)
func getLoadAvg() (float64, float64, float64, error) {
	content, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}

	fields := strings.Fields(string(content))
	if len(fields) < 3 {
		return 0, 0, 0, nil
	}

	load1, _ := strconv.ParseFloat(fields[0], 64)
	load5, _ := strconv.ParseFloat(fields[1], 64)
	load15, _ := strconv.ParseFloat(fields[2], 64)

	return load1, load5, load15, nil
}
