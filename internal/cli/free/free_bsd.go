//go:build freebsd || netbsd || openbsd || dragonfly

package free

import (
	"golang.org/x/sys/unix"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// getMemInfo gathers memory statistics on the BSDs via sysctl, staying pure-Go
// (no exec). The BSDs lack /proc/meminfo and the Linux-only syscall.Sysinfo, and
// sysctl key names vary across the family, so total memory is resolved from a
// list of candidate keys.
func getMemInfo() (MemInfo, error) {
	var info MemInfo

	// Total physical memory in bytes. Key names differ across BSDs:
	//   FreeBSD/DragonFly: hw.physmem
	//   NetBSD/OpenBSD:    hw.physmem64 (hw.physmem is 32-bit and may truncate)
	for _, key := range []string{"hw.physmem64", "hw.physmem", "hw.realmem"} {
		if total, err := unix.SysctlUint64(key); err == nil && total != 0 {
			info.MemTotal = total
			break
		}
	}

	if info.MemTotal == 0 {
		return info, cmderr.Wrap(cmderr.ErrIO, "free: unable to determine total memory via sysctl")
	}

	// Free memory is not uniformly exposed across the BSDs through a single
	// portable sysctl key, so it is left as an approximation here.
	info.MemFree = 0
	info.MemAvailable = 0
	info.Buffers = 0
	info.Cached = 0
	info.SwapTotal = 0
	info.SwapFree = 0

	return info, nil
}
