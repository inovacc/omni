//go:build unix

package cli

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func getMemInfo() (MemInfo, error) {
	var info MemInfo

	// Try reading from /proc/meminfo first (Linux)
	file, err := os.Open("/proc/meminfo")
	if err == nil {
		defer func() {
			_ = file.Close()
		}()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			value, _ := strconv.ParseUint(fields[1], 10, 64)
			value *= 1024 // Convert from KB to bytes

			switch fields[0] {
			case "MemTotal:":
				info.MemTotal = value
			case "MemFree:":
				info.MemFree = value
			case "MemAvailable:":
				info.MemAvailable = value
			case "Buffers:":
				info.Buffers = value
			case "Cached:":
				info.Cached = value
			case "SwapTotal:":
				info.SwapTotal = value
			case "SwapFree:":
				info.SwapFree = value
			}
		}

		return info, nil
	}

	// Fallback to sysinfo (works on Linux but less detailed)
	var sysinfo syscall.Sysinfo_t
	if err := syscall.Sysinfo(&sysinfo); err != nil {
		return info, err
	}

	unit := uint64(sysinfo.Unit)
	info.MemTotal = uint64(sysinfo.Totalram) * unit
	info.MemFree = uint64(sysinfo.Freeram) * unit
	info.MemAvailable = info.MemFree // Approximation
	info.Buffers = uint64(sysinfo.Bufferram) * unit
	info.Cached = 0 // Not available from sysinfo
	info.SwapTotal = uint64(sysinfo.Totalswap) * unit
	info.SwapFree = uint64(sysinfo.Freeswap) * unit

	return info, nil
}
