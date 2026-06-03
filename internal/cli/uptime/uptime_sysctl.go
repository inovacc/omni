//go:build darwin || freebsd || netbsd || openbsd || dragonfly

package uptime

import (
	"encoding/binary"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// nativeOrder is the host byte order, used to decode the raw vm.loadavg
// sysctl buffer. It is resolved once at init rather than per call.
var nativeOrder binary.ByteOrder = func() binary.ByteOrder {
	var x uint16 = 1
	if (*[2]byte)(unsafe.Pointer(&x))[0] == 1 {
		return binary.LittleEndian
	}

	return binary.BigEndian
}()

func getUptimeInfo() (UptimeInfo, error) {
	var info UptimeInfo

	// Boot time comes from the kern.boottime sysctl, returned as a timeval.
	bootTv, err := unix.SysctlTimeval("kern.boottime")
	if err != nil {
		return info, err
	}

	sec, nsec := bootTv.Unix()
	info.BootTime = time.Unix(sec, nsec)
	info.Uptime = time.Since(info.BootTime)

	// Load averages come from the vm.loadavg sysctl. The raw struct is:
	//   struct loadavg { fixed_t ldavg[3]; long fscale; }
	// where ldavg values are scaled by fscale. Parse defensively: if the
	// buffer is shorter than expected, leave load averages at zero rather
	// than indexing out of bounds.
	info.LoadAvg1, info.LoadAvg5, info.LoadAvg15, _ = readLoadAvg()

	// utmpx parsing on Darwin/BSD differs from Linux; default to at least one
	// user rather than attempting fragile binary parsing.
	info.Users = 1

	return info, nil
}

// readLoadAvg reads the vm.loadavg sysctl and returns the 1, 5, and 15
// minute load averages. It returns zeros if the sysctl is unavailable or the
// returned buffer is malformed.
func readLoadAvg() (float64, float64, float64, error) { //nolint:unparam // error retained for signature symmetry
	raw, err := unix.SysctlRaw("vm.loadavg")
	if err != nil {
		return 0, 0, 0, err
	}

	// Layout: ldavg[3] (uint32 each) followed by fscale. The fixed_t entries
	// are 32-bit on Darwin and the BSDs. Treat the word after the three
	// ldavg entries as the scale. Require at least the three ldavg words plus
	// a 4-byte scale before reading anything to avoid indexing out of bounds.
	const ldavgWord = 4
	if len(raw) < ldavgWord*4 {
		return 0, 0, 0, nil
	}

	order := nativeOrder
	ld1 := order.Uint32(raw[0:ldavgWord])
	ld5 := order.Uint32(raw[ldavgWord : ldavgWord*2])
	ld15 := order.Uint32(raw[ldavgWord*2 : ldavgWord*3])
	scale := order.Uint32(raw[ldavgWord*3 : ldavgWord*4])

	if scale == 0 {
		return 0, 0, 0, nil
	}

	div := float64(scale)

	return float64(ld1) / div, float64(ld5) / div, float64(ld15) / div, nil
}
