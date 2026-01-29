//go:build unix

package kill

import (
	"fmt"
	"io"
	"syscall"
)

// signalMap maps signal names to syscall.Signal values
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

// defaultSignal returns the default signal (SIGTERM)
func defaultSignal() syscall.Signal {
	return syscall.SIGTERM
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
