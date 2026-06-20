//go:build !windows

package kill

import "testing"

// runWindowsSignalHelperChecks is a no-op stub on non-windows platforms. The
// windows-tagged helpers it would exercise do not exist here; the cross-platform
// TestWindowsSignalHelpers skips before calling this on non-windows anyway.
func runWindowsSignalHelperChecks(t *testing.T) {
	t.Helper()
	t.Skip("windows-only helpers not present on this platform")
}
