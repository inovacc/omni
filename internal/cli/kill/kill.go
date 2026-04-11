package kill

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

// KillOptions configures the kill command behavior
type KillOptions struct {
	Signal       string        // -s: specify a signal to send
	List         bool          // -l: list signal names
	Verbose      bool          // -v: verbose output
	OutputFormat output.Format // output format (text/json/table)
}

// KillResult represents the result of a kill operation for JSON output
type KillResult struct {
	PID     int    `json:"pid"`
	Signal  int    `json:"signal"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// RunKill sends a signal to a process
func RunKill(w io.Writer, args []string, opts KillOptions) error {
	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	if opts.List {
		if jsonMode {
			return listSignalsJSON(w, f)
		}

		listSignals(w)

		return nil
	}

	if len(args) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "kill: usage: kill [-s signal | -signal] pid")
	}

	// Determine signal
	sig := defaultSignal()

	if opts.Signal != "" {
		sigName := strings.ToUpper(strings.TrimPrefix(opts.Signal, "SIG"))

		s, ok := signalMap[sigName]
		switch {
		case ok:
			sig = s
		case isPlatformUnsupportedSignal(sigName):
			return cmderr.Wrap(cmderr.ErrUnsupported,
				fmt.Sprintf("kill: signal SIG%s not supported on windows (INT/KILL/TERM only)", sigName))
		default:
			// Try parsing as a number
			sigNum, err := strconv.Atoi(opts.Signal)
			if err != nil {
				return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("kill: unknown signal: %s", opts.Signal))
			}

			sig = syscall.Signal(sigNum)
		}
	}

	var results []KillResult

	// Process each PID
	var lastErr error

	for _, arg := range args {
		// Check for signal specification in argument (-9, -KILL, etc.)
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 1 {
			sigSpec := arg[1:]

			sigName := strings.ToUpper(strings.TrimPrefix(sigSpec, "SIG"))
			if s, ok := signalMap[sigName]; ok {
				sig = s
				continue
			}
			if isPlatformUnsupportedSignal(sigName) {
				return cmderr.Wrap(cmderr.ErrUnsupported,
					fmt.Sprintf("kill: signal SIG%s not supported on windows (INT/KILL/TERM only)", sigName))
			}
			// If it looks like a signal name (alphabetic), reject as unknown.
			if isAlphaSignalName(sigName) {
				return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("kill: unknown signal: %s", sigSpec))
			}
		}

		pid, err := strconv.Atoi(arg)
		if err != nil {
			if jsonMode {
				results = append(results, KillResult{
					PID:     0,
					Signal:  int(sig),
					Success: false,
					Error:   fmt.Sprintf("invalid pid: %s", arg),
				})
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "kill: invalid pid: %s\n", arg)
			}

			lastErr = cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("kill: invalid pid: %s", arg))

			continue
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			if jsonMode {
				results = append(results, KillResult{
					PID:     pid,
					Signal:  int(sig),
					Success: false,
					Error:   "no such process",
				})
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "kill: no such process: pid %d\n", pid)
			}

			lastErr = cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("kill: no such process: pid %d", pid))

			continue
		}

		if err := sendSignal(process, sig); err != nil {
			if jsonMode {
				results = append(results, KillResult{
					PID:     pid,
					Signal:  int(sig),
					Success: false,
					Error:   err.Error(),
				})
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "kill: %v\n", err)
			}

			lastErr = classifySignalErr(err, pid)

			continue
		}

		if jsonMode {
			results = append(results, KillResult{
				PID:     pid,
				Signal:  int(sig),
				Success: true,
			})
		} else if opts.Verbose {
			_, _ = fmt.Fprintf(w, "Sent signal %d to process %d\n", sig, pid)
		}
	}

	if jsonMode {
		return f.Print(results)
	}

	return lastErr
}

// listSignalsJSON outputs signal list as JSON
func listSignalsJSON(_ io.Writer, f *output.Formatter) error {
	type SignalInfo struct {
		Number int    `json:"number"`
		Name   string `json:"name"`
	}

	signals := make([]SignalInfo, 0, len(signalMap))

	for name, sig := range signalMap {
		signals = append(signals, SignalInfo{
			Number: int(sig),
			Name:   name,
		})
	}

	return f.Print(signals)
}

// Kill sends a signal to a process
func Kill(pid int, sig syscall.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(sig)
}

// isAlphaSignalName reports whether name looks like a POSIX signal mnemonic
// (purely alphabetic, non-empty). Used to distinguish "-USR1" (unknown signal)
// from "-1" (numeric PID).
func isAlphaSignalName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}
	return true
}

// classifySignalErr maps a raw os/syscall error from process.Signal to a
// cmderr-classified error. It recognises permission errors and "no such
// process" (ESRCH). Errors already wrapped with a cmderr sentinel (e.g.
// ErrUnsupported from the Windows branch) are passed through unchanged.
func classifySignalErr(err error, pid int) error {
	if err == nil {
		return nil
	}

	// Pass through already-classified cmderr errors (e.g. ErrUnsupported
	// from kill_windows.go sendSignal).
	if errors.Is(err, cmderr.ErrUnsupported) ||
		errors.Is(err, cmderr.ErrInvalidInput) ||
		errors.Is(err, cmderr.ErrPermission) ||
		errors.Is(err, cmderr.ErrNotFound) {
		return err
	}

	if errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.EPERM) {
		return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("kill: permission denied: pid %d", pid))
	}

	if isNoSuchProcess(err) {
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("kill: no such process: pid %d", pid))
	}

	return fmt.Errorf("kill: %w", err)
}
