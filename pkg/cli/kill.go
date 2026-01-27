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

// Signal constants
var signalMap = map[string]syscall.Signal{
	"HUP":    syscall.SIGHUP,
	"INT":    syscall.SIGINT,
	"QUIT":   syscall.SIGQUIT,
	"ILL":    syscall.SIGILL,
	"TRAP":   syscall.SIGTRAP,
	"ABRT":   syscall.SIGABRT,
	"BUS":    syscall.SIGBUS,
	"FPE":    syscall.SIGFPE,
	"KILL":   syscall.SIGKILL,
	"USR1":   syscall.SIGUSR1,
	"SEGV":   syscall.SIGSEGV,
	"USR2":   syscall.SIGUSR2,
	"PIPE":   syscall.SIGPIPE,
	"ALRM":   syscall.SIGALRM,
	"TERM":   syscall.SIGTERM,
	"CHLD":   syscall.SIGCHLD,
	"CONT":   syscall.SIGCONT,
	"STOP":   syscall.SIGSTOP,
	"TSTP":   syscall.SIGTSTP,
	"TTIN":   syscall.SIGTTIN,
	"TTOU":   syscall.SIGTTOU,
	"URG":    syscall.SIGURG,
	"XCPU":   syscall.SIGXCPU,
	"XFSZ":   syscall.SIGXFSZ,
	"VTALRM": syscall.SIGVTALRM,
	"PROF":   syscall.SIGPROF,
	"WINCH":  syscall.SIGWINCH,
	"IO":     syscall.SIGIO,
	"SYS":    syscall.SIGSYS,
	// Numeric aliases
	"1":  syscall.SIGHUP,
	"2":  syscall.SIGINT,
	"3":  syscall.SIGQUIT,
	"9":  syscall.SIGKILL,
	"15": syscall.SIGTERM,
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
	sig := syscall.SIGTERM // Default signal

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

func listSignals(w io.Writer) {
	signals := []struct {
		num  int
		name string
	}{
		{1, "HUP"},
		{2, "INT"},
		{3, "QUIT"},
		{4, "ILL"},
		{5, "TRAP"},
		{6, "ABRT"},
		{7, "BUS"},
		{8, "FPE"},
		{9, "KILL"},
		{10, "USR1"},
		{11, "SEGV"},
		{12, "USR2"},
		{13, "PIPE"},
		{14, "ALRM"},
		{15, "TERM"},
		{17, "CHLD"},
		{18, "CONT"},
		{19, "STOP"},
		{20, "TSTP"},
		{21, "TTIN"},
		{22, "TTOU"},
	}

	for i, sig := range signals {
		_, _ = fmt.Fprintf(w, "%2d) SIG%-8s", sig.num, sig.name)
		if (i+1)%4 == 0 {
			_, _ = fmt.Fprintln(w)
		}
	}

	_, _ = fmt.Fprintln(w)
}

// Kill sends a signal to a process
func Kill(pid int, sig syscall.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(sig)
}
