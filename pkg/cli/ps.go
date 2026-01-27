package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// PsOptions configures the ps command behavior
type PsOptions struct {
	All       bool   // -a: show processes for all users
	Full      bool   // -f: full-format listing
	Long      bool   // -l: long format
	User      string // -u: show processes for specified user
	Pid       int    // -p: show process with specified PID
	Forest    bool   // --forest: show process tree
	NoHeaders bool   // --no-headers: don't print header line
	Sort      string // --sort: sort by column (pid, cpu, mem, time)
	Aux       bool   // aux: BSD-style all processes with user info
}

// ProcessInfo represents information about a running process
type ProcessInfo struct {
	PID     int
	PPID    int
	UID     int
	User    string
	CPU     float64
	MEM     float64
	VSZ     int64 // Virtual memory size in KB
	RSS     int64 // Resident set size in KB
	TTY     string
	Stat    string
	Start   string
	Time    string // CPU time
	Command string
}

// RunPs lists running processes
func RunPs(w io.Writer, opts PsOptions) error {
	processes, err := GetProcessList(opts)
	if err != nil {
		return fmt.Errorf("ps: %w", err)
	}

	// Sort if requested
	if opts.Sort != "" {
		sortProcesses(processes, opts.Sort)
	}

	// Print output
	if opts.Long || opts.Full || opts.Aux {
		return printPsLong(w, processes, opts)
	}
	return printPsSimple(w, processes, opts)
}

func sortProcesses(procs []ProcessInfo, sortBy string) {
	switch strings.ToLower(sortBy) {
	case "pid":
		sort.Slice(procs, func(i, j int) bool {
			return procs[i].PID < procs[j].PID
		})
	case "cpu":
		sort.Slice(procs, func(i, j int) bool {
			return procs[i].CPU > procs[j].CPU
		})
	case "mem":
		sort.Slice(procs, func(i, j int) bool {
			return procs[i].MEM > procs[j].MEM
		})
	case "time":
		sort.Slice(procs, func(i, j int) bool {
			return procs[i].Time > procs[j].Time
		})
	}
}

func printPsSimple(w io.Writer, procs []ProcessInfo, opts PsOptions) error {
	if !opts.NoHeaders {
		_, _ = fmt.Fprintf(w, "%5s %-8s %8s %s\n", "PID", "TTY", "TIME", "CMD")
	}

	for _, p := range procs {
		cmd := p.Command
		if len(cmd) > 60 {
			cmd = cmd[:60]
		}
		_, _ = fmt.Fprintf(w, "%5d %-8s %8s %s\n", p.PID, p.TTY, p.Time, cmd)
	}

	return nil
}

func printPsLong(w io.Writer, procs []ProcessInfo, opts PsOptions) error {
	if !opts.NoHeaders {
		if opts.Aux {
			_, _ = fmt.Fprintf(w, "%-8s %5s %4s %4s %8s %8s %-8s %-4s %-5s %8s %s\n",
				"USER", "PID", "%CPU", "%MEM", "VSZ", "RSS", "TTY", "STAT", "START", "TIME", "COMMAND")
		} else if opts.Long {
			_, _ = fmt.Fprintf(w, "F S %5s %5s %5s  C PRI NI %8s %8s %-8s %8s %s\n",
				"UID", "PID", "PPID", "SZ", "RSS", "TTY", "TIME", "CMD")
		} else {
			_, _ = fmt.Fprintf(w, "%-8s %5s %5s  C %-5s %8s %s\n",
				"UID", "PID", "PPID", "STIME", "TIME", "CMD")
		}
	}

	for _, p := range procs {
		cmd := p.Command
		if len(cmd) > 50 {
			cmd = cmd[:50]
		}

		if opts.Aux {
			_, _ = fmt.Fprintf(w, "%-8s %5d %4.1f %4.1f %8d %8d %-8s %-4s %-5s %8s %s\n",
				p.User, p.PID, p.CPU, p.MEM, p.VSZ, p.RSS, p.TTY, p.Stat, p.Start, p.Time, cmd)
		} else if opts.Long {
			_, _ = fmt.Fprintf(w, "0 %s %5d %5d %5d  0  80  0 %8d %8d %-8s %8s %s\n",
				p.Stat[:1], p.UID, p.PID, p.PPID, p.VSZ, p.RSS, p.TTY, p.Time, cmd)
		} else {
			_, _ = fmt.Fprintf(w, "%-8d %5d %5d  0 %-5s %8s %s\n",
				p.UID, p.PID, p.PPID, p.Start, p.Time, cmd)
		}
	}

	return nil
}
