package kill

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/inovacc/omni/internal/cli/output"
)

// KillOptions configures the kill command behavior
type KillOptions struct {
	Signal  string // -s: specify a signal to send
	List    bool   // -l: list signal names
	Verbose bool   // -v: verbose output
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

	var results []KillResult

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
			if jsonMode {
				results = append(results, KillResult{
					PID:     0,
					Signal:  int(sig),
					Success: false,
					Error:   fmt.Sprintf("invalid PID: %s", arg),
				})
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "kill: invalid PID: %s\n", arg)
			}

			lastErr = err

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
				_, _ = fmt.Fprintf(os.Stderr, "kill: no such process: %d\n", pid)
			}

			lastErr = err

			continue
		}

		if err := process.Signal(sig); err != nil {
			if jsonMode {
				results = append(results, KillResult{
					PID:     pid,
					Signal:  int(sig),
					Success: false,
					Error:   err.Error(),
				})
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "kill: (%d) - %v\n", pid, err)
			}

			lastErr = err

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
