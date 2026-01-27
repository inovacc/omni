package cli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// KillOptions configures the kill command behavior
type KillOptions struct {
	Signal  string // -s: specify a signal to send
	List    bool   // -l: list signal names
	Verbose bool   // -v: verbose output
}

// RunKill sends a signal to a process
func RunKill(w io.Writer, args []string, opts KillOptions) error {
	if opts.List {
		listSignals(w)
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("kill: usage: kill [-s signal | -signal] pid ")
	}

	// Determine signal
	sig := defaultSignal()

	if opts.Signal != "" {
		var ok bool

		sigName := strings.ToUpper(strings.TrimPrefix(opts.Signal, "SIG"))

		sig, ok = signalMap[sigName]
		if !ok {
			// Try parsing as a number
			sigNum, err := strconv.Atoi(opts.Signal)
			if err != nil {
				return fmt.Errorf("kill: invalid signal: %s", opts.Signal)
			}

			sig = syscall.Signal(sigNum)
		}
	}

	// Process each PID
	var lastErr error

	for _, arg := range args {
		// Check for signal specification in argument (-9, -KILL, etc.)
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			sigSpec := arg[1:]

			sigName := strings.ToUpper(strings.TrimPrefix(sigSpec, "SIG"))
			if s, ok := signalMap[sigName]; ok {
				sig = s
				continue
			}
		}

		pid, err := strconv.Atoi(arg)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "kill: invalid PID: %s\n", arg)
			lastErr = err

			continue
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "kill: no such process: %d\n", pid)
			lastErr = err

			continue
		}

		if err := process.Signal(sig); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "kill: (%d) - %v\n", pid, err)
			lastErr = err

			continue
		}

		if opts.Verbose {
			_, _ = fmt.Fprintf(w, "Sent signal %d to process %d\n", sig, pid)
		}
	}

	return lastErr
}

// Kill sends a signal to a process
func Kill(pid int, sig syscall.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(sig)
}
