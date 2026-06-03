//go:build unix && !linux && !darwin && !freebsd && !netbsd && !openbsd && !dragonfly

package uptime

import (
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// getUptimeInfo provides a safe fallback for Unix platforms without a
// dedicated implementation (e.g. solaris, aix, illumos). It reports the
// operation as unsupported rather than returning fabricated values.
func getUptimeInfo() (UptimeInfo, error) {
	return UptimeInfo{
		BootTime: time.Now(),
		Users:    1,
	}, cmderr.Wrap(cmderr.ErrUnsupported, "uptime: not supported on this platform")
}
