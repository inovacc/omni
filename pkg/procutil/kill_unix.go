//go:build !windows

package procutil

import (
	"fmt"
	"os"
	"syscall"
)

func sendImpl(pid int, sig Signal) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	var s syscall.Signal
	switch sig {
	case SigTerm:
		s = syscall.SIGTERM
	case SigKill:
		s = syscall.SIGKILL
	case SigInt:
		s = syscall.SIGINT
	case SigHup:
		s = syscall.SIGHUP
	default:
		return fmt.Errorf("unsupported signal %s", sig)
	}
	if err := p.Signal(s); err != nil {
		return fmt.Errorf("signal pid %d (%s): %w", pid, sig, err)
	}
	return nil
}
