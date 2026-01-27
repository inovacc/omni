//go:build unix

package cli

import "syscall"

// getKernelRelease returns the kernel release (e.g., "5.15.0-generic")
func getKernelRelease() string {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return "unknown"
	}

	return charsToString(uname.Release[:])
}

// getKernelVersion returns the kernel version
func getKernelVersion() string {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return "unknown"
	}

	return charsToString(uname.Version[:])
}

// charsToString converts a fixed-size byte array to a string
func charsToString(ca []int8) string {
	s := make([]byte, 0, len(ca))
	for _, c := range ca {
		if c == 0 {
			break
		}

		s = append(s, byte(c))
	}

	return string(s)
}
