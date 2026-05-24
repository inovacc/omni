//go:build windows

package procutil

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func sendImpl(pid int, sig Signal) error {
	switch sig {
	case SigTerm, SigKill:
		h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(pid))
		if err != nil {
			return fmt.Errorf("open process %d: %w", pid, err)
		}
		defer func() { _ = windows.CloseHandle(h) }()
		if err := windows.TerminateProcess(h, 1); err != nil {
			return fmt.Errorf("terminate process %d: %w", pid, err)
		}
		return nil
	case SigInt, SigHup:
		return fmt.Errorf("signal %s not supported on windows (try TERM or KILL)", sig)
	default:
		return fmt.Errorf("unsupported signal %s", sig)
	}
}
