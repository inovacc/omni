package pkill

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/inovacc/omni/internal/cli/output"
	"github.com/shirou/gopsutil/v3/process"
)

// Options configures the pkill command behavior
type Options struct {
	Signal     string // -signal: signal to send (default: TERM)
	Exact      bool   // -x: match exactly
	Full       bool   // -f: match against full command line
	Newest     bool   // -n: select only the newest process
	Oldest     bool   // -o: select only the oldest process
	Count      bool   // -c: count matching processes
	ListOnly   bool   // -l: list matching processes (don't kill)
	User       string // -u: only match processes owned by user
	Parent     int    // -P: only match processes with given parent PID
	Terminal   string // -t: only match processes on terminal
	Verbose    bool   // -v: verbose output
	OutputFormat output.Format // output format (text/json/table)
	IgnoreCase bool   // -i: case insensitive matching
}

// Result represents the result of a pkill operation
type Result struct {
	PID     int    `json:"pid"`
	Name    string `json:"name"`
	Cmdline string `json:"cmdline,omitempty"`
	Signal  int    `json:"signal,omitempty"`
	Matched bool   `json:"matched"`
	Killed  bool   `json:"killed,omitempty"`
	Error   string `json:"error,omitempty"`
}

// signalMap maps signal names to syscall.Signal
var signalMap = map[string]syscall.Signal{
	"HUP":  syscall.SIGHUP,
	"INT":  syscall.SIGINT,
	"QUIT": syscall.SIGQUIT,
	"ILL":  syscall.SIGILL,
	"TRAP": syscall.SIGTRAP,
	"ABRT": syscall.SIGABRT,
	"KILL": syscall.SIGKILL,
	"SEGV": syscall.SIGSEGV,
	"PIPE": syscall.SIGPIPE,
	"ALRM": syscall.SIGALRM,
	"TERM": syscall.SIGTERM,
}

// Run executes the pkill command
func Run(w io.Writer, pattern string, opts Options) error {
	if pattern == "" {
		return fmt.Errorf("pkill: no pattern specified")
	}

	// Compile pattern
	var (
		re  *regexp.Regexp
		err error
	)

	patternStr := pattern
	if opts.Exact {
		patternStr = "^" + regexp.QuoteMeta(pattern) + "$"
	}

	if opts.IgnoreCase {
		re, err = regexp.Compile("(?i)" + patternStr)
	} else {
		re, err = regexp.Compile(patternStr)
	}

	if err != nil {
		return fmt.Errorf("pkill: invalid pattern: %w", err)
	}

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	// Get signal
	sig := syscall.SIGTERM

	if opts.Signal != "" {
		sigName := strings.ToUpper(strings.TrimPrefix(opts.Signal, "SIG"))
		if s, ok := signalMap[sigName]; ok {
			sig = s
		} else {
			sigNum, err := strconv.Atoi(opts.Signal)
			if err != nil {
				return fmt.Errorf("pkill: invalid signal: %s", opts.Signal)
			}

			sig = syscall.Signal(sigNum)
		}
	}

	// Get all processes
	procs, err := process.Processes()
	if err != nil {
		return fmt.Errorf("pkill: failed to get processes: %w", err)
	}

	var matched []Result

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		// Get cmdline for full matching
		cmdline := ""

		if opts.Full {
			cmd, err := p.Cmdline()
			if err == nil {
				cmdline = cmd
			}
		}

		// Match
		matchTarget := name
		if opts.Full && cmdline != "" {
			matchTarget = cmdline
		}

		if !re.MatchString(matchTarget) {
			continue
		}

		// Filter by user
		if opts.User != "" {
			username, err := p.Username()
			if err != nil || !strings.EqualFold(username, opts.User) {
				continue
			}
		}

		// Filter by parent PID
		if opts.Parent > 0 {
			ppid, err := p.Ppid()
			if err != nil || int(ppid) != opts.Parent {
				continue
			}
		}

		// Filter by terminal
		if opts.Terminal != "" {
			terminal, err := p.Terminal()
			if err != nil || !strings.Contains(terminal, opts.Terminal) {
				continue
			}
		}

		result := Result{
			PID:     int(p.Pid),
			Name:    name,
			Cmdline: cmdline,
			Matched: true,
		}

		matched = append(matched, result)
	}

	if len(matched) == 0 {
		if jsonMode {
			_, _ = fmt.Fprintln(w, "[]")
		}

		return nil
	}

	// Select newest/oldest if requested
	if opts.Newest {
		// Keep only the process with highest PID (newest)
		newest := matched[0]
		for _, m := range matched[1:] {
			if m.PID > newest.PID {
				newest = m
			}
		}

		matched = []Result{newest}
	} else if opts.Oldest {
		// Keep only the process with lowest PID (oldest)
		oldest := matched[0]
		for _, m := range matched[1:] {
			if m.PID < oldest.PID {
				oldest = m
			}
		}

		matched = []Result{oldest}
	}

	// Count mode
	if opts.Count {
		if jsonMode {
			_, _ = fmt.Fprintf(w, `{"count": %d}`+"\n", len(matched))
		} else {
			_, _ = fmt.Fprintf(w, "%d\n", len(matched))
		}

		return nil
	}

	// List mode (pgrep behavior)
	if opts.ListOnly {
		if jsonMode {
			return f.Print(matched)
		}

		for _, m := range matched {
			if opts.Full {
				_, _ = fmt.Fprintf(w, "%d %s\n", m.PID, m.Cmdline)
			} else {
				_, _ = fmt.Fprintf(w, "%d %s\n", m.PID, m.Name)
			}
		}

		return nil
	}

	// Kill mode
	var results []Result

	for _, m := range matched {
		m.Signal = int(sig)

		proc, err := os.FindProcess(m.PID)
		if err != nil {
			m.Error = "no such process"
			results = append(results, m)

			continue
		}

		if err := proc.Signal(sig); err != nil {
			m.Error = err.Error()
		} else {
			m.Killed = true
			if opts.Verbose {
				_, _ = fmt.Fprintf(w, "pkill: killed %d (%s)\n", m.PID, m.Name)
			}
		}

		results = append(results, m)
	}

	if jsonMode {
		return f.Print(results)
	}

	return nil
}
