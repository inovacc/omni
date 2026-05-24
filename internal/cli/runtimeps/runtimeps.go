// Package runtimeps is the shared CLI plumbing for the runtime-filtered
// process commands (omni gops, nodeps, pyps, javaps). It defers all process
// inspection to pkg/procutil and only handles flag parsing, output rendering,
// and the kill confirmation flow.
package runtimeps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/procutil"
)

// ListOptions configures the "list" verb for a runtime-filtered process command.
type ListOptions struct {
	All    bool   // include the omni process itself
	Format string // "table" (default) or "json"
}

// RunList enumerates processes for the given runtime and renders them.
func RunList(ctx context.Context, w io.Writer, rt procutil.Runtime, opts ListOptions) error {
	procs, err := procutil.List(ctx, procutil.ListOptions{
		Runtime:     rt,
		IncludeSelf: opts.All,
	})
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("list %s processes: %v", rt, err))
	}
	sort.Slice(procs, func(i, j int) bool { return procs[i].PID < procs[j].PID })

	switch strings.ToLower(opts.Format) {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(procs)
	case "", "table":
		return renderTable(w, rt, procs)
	default:
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("unknown format %q (want table|json)", opts.Format))
	}
}

func renderTable(w io.Writer, rt procutil.Runtime, procs []procutil.Process) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if rt == procutil.RuntimeGo {
		_, _ = fmt.Fprintln(tw, "PID\tPPID\tNAME\tGO VERSION\tMODULE\tEXE")
		for _, p := range procs {
			_, _ = fmt.Fprintf(tw, "%d\t%d\t%s\t%s\t%s\t%s\n", p.PID, p.PPID, p.Name, p.GoVersion, p.Module, p.ExePath)
		}
	} else {
		_, _ = fmt.Fprintln(tw, "PID\tPPID\tNAME\tEXE")
		for _, p := range procs {
			_, _ = fmt.Fprintf(tw, "%d\t%d\t%s\t%s\n", p.PID, p.PPID, p.Name, p.ExePath)
		}
	}
	if len(procs) == 0 {
		_, _ = fmt.Fprintf(tw, "(no %s processes found)\n", rt)
	}
	return tw.Flush()
}

// KillOptions configures the "kill" verb.
type KillOptions struct {
	Signal    string // TERM|KILL|INT|HUP (default TERM)
	Recursive bool   // kill every process matching the target name
	Yes       bool   // skip the confirmation prompt required for --recursive
}

// RunKill signals one PID (numeric target) or one-or-many processes matching
// a name. When target is non-numeric and Recursive is false, RunKill returns
// an error if more than one process matches — the caller must pass
// --recursive to confirm bulk action. Recursive also requires --yes.
func RunKill(ctx context.Context, w io.Writer, rt procutil.Runtime, target string, opts KillOptions) error {
	sig, err := procutil.ParseSignal(opts.Signal)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, err.Error())
	}

	// For named targets, count matches first so we can enforce the
	// --recursive guard before sending any signal.
	if !isNumeric(target) {
		procs, err := procutil.List(ctx, procutil.ListOptions{Runtime: rt})
		if err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("list %s processes: %v", rt, err))
		}
		var matches []procutil.Process
		lower := strings.ToLower(strings.TrimSpace(target))
		for _, p := range procs {
			if strings.EqualFold(stripExe(p.Name), lower) || strings.EqualFold(stripExe(baseExe(p.ExePath)), lower) {
				matches = append(matches, p)
			}
		}
		if len(matches) == 0 {
			return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("no %s processes match %q", rt, target))
		}
		if len(matches) > 1 && !opts.Recursive {
			return cmderr.Wrap(cmderr.ErrConflict,
				fmt.Sprintf("%d processes match %q — pass --recursive to kill all", len(matches), target))
		}
		if opts.Recursive && !opts.Yes {
			return cmderr.Wrap(cmderr.ErrConflict, "--recursive requires --yes (destructive)")
		}
	}

	results, err := procutil.KillAllMatching(ctx, target, sig, rt)
	for _, r := range results {
		if r.Err != nil {
			_, _ = fmt.Fprintf(w, "FAIL pid=%d name=%s err=%v\n", r.PID, r.Name, r.Err)
			continue
		}
		_, _ = fmt.Fprintf(w, "sent %s pid=%d name=%s\n", sig, r.PID, r.Name)
	}
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("kill: %v", err))
	}
	return nil
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func stripExe(name string) string { return strings.TrimSuffix(strings.ToLower(name), ".exe") }

func baseExe(path string) string {
	// minimal basename (avoid an extra import here — path/filepath would work too).
	i := strings.LastIndexAny(path, `/\`)
	if i < 0 {
		return path
	}
	return path[i+1:]
}
