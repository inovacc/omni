//go:build windows

package kill

import (
	"errors"
	"os"
	"syscall"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// runWindowsSignalHelperChecks exercises the windows-only signal helpers:
// isSupportedSignal, signalDisplayName, sendSignal (reject path), and
// isNoSuchProcess. It is invoked from the cross-platform TestWindowsSignalHelpers.
func runWindowsSignalHelperChecks(t *testing.T) {
	t.Helper()

	// isSupportedSignal: only INT/KILL/TERM are deliverable on windows.
	supported := []syscall.Signal{syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM}
	for _, s := range supported {
		if !isSupportedSignal(s) {
			t.Errorf("isSupportedSignal(%v) = false, want true", s)
		}
	}
	if isSupportedSignal(syscall.Signal(99)) {
		t.Error("isSupportedSignal(99) = true, want false")
	}

	// signalDisplayName: stable names for known signals, numeric fallback otherwise.
	cases := map[syscall.Signal]string{
		syscall.SIGINT:    "SIGINT",
		syscall.SIGKILL:   "SIGKILL",
		syscall.SIGTERM:   "SIGTERM",
		syscall.Signal(7): "signal 7",
	}
	for sig, want := range cases {
		if got := signalDisplayName(sig); got != want {
			t.Errorf("signalDisplayName(%d) = %q, want %q", int(sig), got, want)
		}
	}

	// sendSignal: an unsupported signal is rejected with ErrUnsupported and the
	// process is never touched. os.FindProcess on windows always succeeds.
	self, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("find self process: %v", err)
	}
	if err := sendSignal(self, syscall.Signal(99)); err == nil {
		t.Error("sendSignal(unsupported) = nil, want ErrUnsupported")
	} else if !errors.Is(err, cmderr.ErrUnsupported) {
		t.Errorf("sendSignal(unsupported) = %v, want ErrUnsupported", err)
	}

	// isNoSuchProcess matches the windows sentinel only.
	if !isNoSuchProcess(os.ErrProcessDone) {
		t.Error("isNoSuchProcess(ErrProcessDone) = false, want true")
	}
	if isNoSuchProcess(errors.New("nope")) {
		t.Error("isNoSuchProcess(unrelated) = true, want false")
	}
}
