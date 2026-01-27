//go:build windows

package cli

import (
	"fmt"
	"io"
	"syscall"
)

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
