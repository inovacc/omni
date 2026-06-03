//go:build darwin

package free

import (
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// getMemInfo gathers memory statistics on macOS via sysctl, staying pure-Go
// (no exec of vm_stat/sysctl binaries). Darwin lacks /proc/meminfo and the
// Linux-only syscall.Sysinfo, so values are derived from standard sysctl keys.
func getMemInfo() (MemInfo, error) {
	var info MemInfo

	// Total physical memory in bytes.
	total, err := unix.SysctlUint64("hw.memsize")
	if err != nil {
		return info, cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("free: sysctl hw.memsize: %v", err))
	}
	info.MemTotal = total

	// Page size in bytes; fall back to a sane default if unavailable.
	pageSize, err := unix.SysctlUint64("hw.pagesize")
	if err != nil || pageSize == 0 {
		pageSize = 4096
	}

	// Free physical pages. vm.page_free_count is exposed as a 32-bit value.
	freePages, err := unix.SysctlUint32("vm.page_free_count")
	if err == nil {
		free := uint64(freePages) * pageSize
		if free > info.MemTotal {
			free = info.MemTotal
		}
		info.MemFree = free
	}
	info.MemAvailable = info.MemFree // Approximation

	// Darwin does not expose Linux-style buffers/cached or swap totals via a
	// single portable sysctl; leave them zeroed rather than guess.
	info.Buffers = 0
	info.Cached = 0
	info.SwapTotal = 0
	info.SwapFree = 0

	return info, nil
}
