package free

import (
	"fmt"
	"io"
)

// FreeOptions configures the free command behavior
type FreeOptions struct {
	Bytes     bool // -b: show output in bytes
	Kibibytes bool // -k: show output in kibibytes (default)
	Mebibytes bool // -m: show output in mebibytes
	Gibibytes bool // -g: show output in gibibytes
	Human     bool // -h: show human-readable output
	Wide      bool // -w: wide output
	Total     bool // -t: show total for RAM + swap
	Seconds   int  // -s: continuously display every N seconds
	Count     int  // -c: display N times, then exit
}

// MemInfo contains memory information
type MemInfo struct {
	MemTotal     uint64 `json:"memTotal"`
	MemFree      uint64 `json:"memFree"`
	MemAvailable uint64 `json:"memAvailable"`
	Buffers      uint64 `json:"buffers"`
	Cached       uint64 `json:"cached"`
	SwapTotal    uint64 `json:"swapTotal"`
	SwapFree     uint64 `json:"swapFree"`
}

// RunFree displays amount of free and used memory in the system
func RunFree(w io.Writer, opts FreeOptions) error {
	info, err := getMemInfo()
	if err != nil {
		return fmt.Errorf("free: %w", err)
	}

	// Determine unit divisor and suffix
	var divisor uint64 = 1024 // Default kibibytes

	suffix := ""

	switch {
	case opts.Bytes:
		divisor = 1
		suffix = "B"
	case opts.Mebibytes:
		divisor = 1024 * 1024
		suffix = "Mi"
	case opts.Gibibytes:
		divisor = 1024 * 1024 * 1024
		suffix = "Gi"
	case opts.Human:
		// Will format each value individually
		divisor = 0
	default:
		suffix = "Ki"
	}

	// Calculate values
	memUsed := info.MemTotal - info.MemFree - info.Buffers - info.Cached
	swapUsed := info.SwapTotal - info.SwapFree

	// Print header
	if opts.Human {
		_, _ = fmt.Fprintf(w, "%15s %10s %10s %10s %10s %10s\n",
			"", "total", "used", "free", "shared", "available")
	} else {
		_, _ = fmt.Fprintf(w, "%15s %12s %12s %12s %12s %12s\n",
			"", "total", "used", "free", "shared", "available")
	}

	// Print memory line
	if opts.Human {
		_, _ = fmt.Fprintf(w, "%-15s %10s %10s %10s %10s %10s\n",
			"Mem:",
			formatBytes(info.MemTotal),
			formatBytes(memUsed),
			formatBytes(info.MemFree),
			formatBytes(0), // shared not easily available
			formatBytes(info.MemAvailable))
	} else {
		_, _ = fmt.Fprintf(w, "%-15s %12d %12d %12d %12d %12d\n",
			"Mem:",
			info.MemTotal/divisor,
			memUsed/divisor,
			info.MemFree/divisor,
			0, // shared
			info.MemAvailable/divisor)
	}

	// Print swap line
	if opts.Human {
		_, _ = fmt.Fprintf(w, "%-15s %10s %10s %10s\n",
			"Swap:",
			formatBytes(info.SwapTotal),
			formatBytes(swapUsed),
			formatBytes(info.SwapFree))
	} else {
		_, _ = fmt.Fprintf(w, "%-15s %12d %12d %12d\n",
			"Swap:",
			info.SwapTotal/divisor,
			swapUsed/divisor,
			info.SwapFree/divisor)
	}

	// Print total if requested
	if opts.Total {
		totalMem := info.MemTotal + info.SwapTotal
		totalUsed := memUsed + swapUsed
		totalFree := info.MemFree + info.SwapFree

		if opts.Human {
			_, _ = fmt.Fprintf(w, "%-15s %10s %10s %10s\n",
				"Total:",
				formatBytes(totalMem),
				formatBytes(totalUsed),
				formatBytes(totalFree))
		} else {
			_, _ = fmt.Fprintf(w, "%-15s %12d %12d %12d\n",
				"Total:",
				totalMem/divisor,
				totalUsed/divisor,
				totalFree/divisor)
		}
	}

	_ = suffix // Suppress unused warning

	return nil
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f%ci", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetMemInfo returns system memory information
func GetMemInfo() (MemInfo, error) {
	return getMemInfo()
}
