//go:build windows

package kill

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// isSupportedSignal reports whether sig is one of the 3 signals Windows can
// deliver via os.Process.Signal: SIGINT, SIGKILL, SIGTERM.
func isSupportedSignal(sig syscall.Signal) bool {
	switch sig {
	case syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM:
		return true
	default:
		return false
	}
}

// signalDisplayName returns a stable "SIG<NAME>" string for error messages,
// or the numeric form if the signal is not in the known Windows set.
func signalDisplayName(sig syscall.Signal) string {
	switch sig {
	case syscall.SIGINT:
		return "SIGINT"
	case syscall.SIGKILL:
		return "SIGKILL"
	case syscall.SIGTERM:
		return "SIGTERM"
	default:
		return fmt.Sprintf("signal %d", int(sig))
	}
}

// sendSignal dispatches a signal on Windows. Unsupported signals are rejected
// with cmderr.ErrUnsupported using a locked message format. The message is
// pinned by a golden snapshot — do not change without updating the snapshot.
func sendSignal(process *os.Process, sig syscall.Signal) error {
	if !isSupportedSignal(sig) {
		return cmderr.Wrap(cmderr.ErrUnsupported,
			fmt.Sprintf("kill: signal %s not supported on windows (INT/KILL/TERM only)", signalDisplayName(sig)))
	}
	return process.Signal(sig)
}

// isNoSuchProcess reports whether err represents a missing process on Windows.
// Windows does not have ESRCH; os.FindProcess always succeeds, so this only
// matches wrapped os.ErrProcessDone from process.Signal.
func isNoSuchProcess(err error) bool {
	return errors.Is(err, os.ErrProcessDone)
}

// posixSignalNames is the set of POSIX signal mnemonics recognised on Unix
// but unavailable on Windows. Used to produce ErrUnsupported for requests
// like "kill -USR1 <pid>" instead of a confusing "unknown signal" error.
var posixSignalNames = map[string]struct{}{
	"HUP": {}, "QUIT": {}, "ILL": {}, "TRAP": {}, "ABRT": {},
	"BUS": {}, "FPE": {}, "USR1": {}, "SEGV": {}, "USR2": {},
	"PIPE": {}, "ALRM": {}, "CHLD": {}, "CONT": {}, "STOP": {},
	"TSTP": {}, "TTIN": {}, "TTOU": {}, "URG": {}, "XCPU": {},
	"XFSZ": {}, "VTALRM": {}, "PROF": {}, "WINCH": {}, "IO": {},
	"SYS": {},
}

// isPlatformUnsupportedSignal reports whether the given POSIX signal name
// exists in the Unix world but not in the Windows signalMap. Callers should
// produce cmderr.ErrUnsupported when this returns true.
func isPlatformUnsupportedSignal(name string) bool {
	_, ok := posixSignalNames[name]
	return ok
}

// signalMap maps signal names to syscall.Signal values
// Windows only supports a limited set of signals
var signalMap = map[string]syscall.Signal{
	"INT":  syscall.SIGINT,
	"KILL": syscall.SIGKILL,
	"TERM": syscall.SIGTERM,
	// Numeric aliases
	"2":  syscall.SIGINT,
	"9":  syscall.SIGKILL,
	"15": syscall.SIGTERM,
}

// defaultSignal returns the default signal (SIGTERM)
func defaultSignal() syscall.Signal {
	return syscall.SIGTERM
}

func listSignals(w io.Writer) {
	signals := []struct {
		num  int
		name string
	}{
		{2, "INT"},
		{9, "KILL"},
		{15, "TERM"},
	}

	for i, sig := range signals {
		_, _ = fmt.Fprintf(w, "%2d) SIG%-8s", sig.num, sig.name)
		if (i+1)%4 == 0 {
			_, _ = fmt.Fprintln(w)
		}
	}

	_, _ = fmt.Fprintln(w)
}
