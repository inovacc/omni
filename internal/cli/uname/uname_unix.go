//go:build unix

package uname

import "golang.org/x/sys/unix"

// getKernelRelease returns the kernel release (e.g., "5.15.0-generic").
//
// Implemented via the uname(2) syscall through golang.org/x/sys/unix (no
// os/exec). Unlike syscall.Uname/syscall.Utsname — which only exist on Linux
// and a subset of BSDs — unix.Uname is portable across all unix GOOS targets
// (linux, darwin, freebsd, openbsd, netbsd, ...), so this file builds cleanly
// under the broad `unix` build tag.
func getKernelRelease() string {
	var uts unix.Utsname
	if err := unix.Uname(&uts); err != nil {
		return "unknown"
	}

	return charsToString(uts.Release[:])
}

// getKernelVersion returns the kernel version.
func getKernelVersion() string {
	var uts unix.Utsname
	if err := unix.Uname(&uts); err != nil {
		return "unknown"
	}

	return charsToString(uts.Version[:])
}

// charsToString converts a NUL-terminated, fixed-size byte array (as returned
// in unix.Utsname fields) to a Go string, stopping at the first NUL.
func charsToString(ca []byte) string {
	for i, c := range ca {
		if c == 0 {
			return string(ca[:i])
		}
	}

	return string(ca)
}
