package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/google/gops/goprocess"
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
	JSON      bool   // -j: output as JSON
	GoOnly    bool   // --go: show only Go processes
}

// ProcessInfo represents information about a running process
type ProcessInfo struct {
	PID       int     `json:"pid"`
	PPID      int     `json:"ppid"`
	UID       int     `json:"uid,omitempty"`
	User      string  `json:"user,omitempty"`
	CPU       float64 `json:"cpu"`
	MEM       float64 `json:"mem"`
	VSZ       int64   `json:"vsz"` // Virtual memory size in KB
	RSS       int64   `json:"rss"` // Resident set size in KB
	TTY       string  `json:"tty,omitempty"`
	Stat      string  `json:"stat,omitempty"`
	Start     string  `json:"start,omitempty"`
	Time      string  `json:"time"` // CPU time
	Command   string  `json:"command"`
	IsGo      bool    `json:"is_go"`                // true if this is a Go process
	GoVersion string  `json:"go_version,omitempty"` // Go version if IsGo
	BuildInfo string  `json:"build_info,omitempty"` // Go build path if IsGo
}

// RunPs lists running processes
func RunPs(w io.Writer, opts PsOptions) error {
	processes, err := GetProcessList(opts)
	if err != nil {
		return fmt.Errorf("ps: %w", err)
	}

	// Enrich with Go process information
	enrichWithGoInfo(processes)

	// Filter Go-only if requested
	if opts.GoOnly {
		processes = filterGoProcesses(processes)
	}

	// Sort if requested
	if opts.Sort != "" {
		sortProcesses(processes, opts.Sort)
	}

	// JSON output
	if opts.JSON {
		return printPsJSON(w, processes)
	}

	// Print output
	if opts.Long || opts.Full || opts.Aux {
		return printPsLong(w, processes, opts)
	}

	return printPsSimple(w, processes, opts)
}

// enrichWithGoInfo detects Go processes and adds Go-specific information
func enrichWithGoInfo(processes []ProcessInfo) {
	// Get list of Go processes from gops
	goProcs := goprocess.FindAll()

	// Create a map for fast lookup
	goProcMap := make(map[int]goprocess.P)
	for _, gp := range goProcs {
		goProcMap[gp.PID] = gp
	}

	// Enrich process info
	for i := range processes {
		if gp, ok := goProcMap[processes[i].PID]; ok {
			processes[i].IsGo = true
			processes[i].GoVersion = gp.BuildVersion
			processes[i].BuildInfo = gp.Path
		}
	}
}

// filterGoProcesses returns only Go processes
func filterGoProcesses(processes []ProcessInfo) []ProcessInfo {
	var result []ProcessInfo

	for _, p := range processes {
		if p.IsGo {
			result = append(result, p)
		}
	}

	return result
}

// printPsJSON outputs processes as JSON
func printPsJSON(w io.Writer, processes []ProcessInfo) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	return encoder.Encode(processes)
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

// RunTop shows top N processes sorted by resource usage
func RunTop(w io.Writer, opts PsOptions, n int) error {
	// Force show all for top
	opts.All = true

	processes, err := GetProcessList(opts)
	if err != nil {
		return fmt.Errorf("top: %w", err)
	}

	// Enrich with Go process information
	enrichWithGoInfo(processes)

	// Filter Go-only if requested
	if opts.GoOnly {
		processes = filterGoProcesses(processes)
	}

	// Sort (default by CPU)
	if opts.Sort == "" {
		opts.Sort = "cpu"
	}

	sortProcesses(processes, opts.Sort)

	// Limit to top N
	if n > 0 && len(processes) > n {
		processes = processes[:n]
	}

	// JSON output
	if opts.JSON {
		return printPsJSON(w, processes)
	}

	// Print top-style output
	return printTopOutput(w, processes)
}

func printTopOutput(w io.Writer, procs []ProcessInfo) error {
	// Header
	_, _ = fmt.Fprintf(w, "%5s %-10s %6s %6s %10s %10s  %s\n",
		"PID", "USER", "%CPU", "%MEM", "VSZ", "RSS", "COMMAND")

	for _, p := range procs {
		cmd := p.Command
		if len(cmd) > 50 {
			cmd = cmd[:50] + "..."
		}

		user := p.User
		if len(user) > 10 {
			user = user[:10]
		}

		goMarker := ""
		if p.IsGo {
			goMarker = " [go]"
		}

		_, _ = fmt.Fprintf(w, "%5d %-10s %6.1f %6.1f %10d %10d  %s%s\n",
			p.PID, user, p.CPU, p.MEM, p.VSZ, p.RSS, cmd, goMarker)
	}

	return nil
}
