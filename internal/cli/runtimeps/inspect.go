package runtimeps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/shirou/gopsutil/v3/process"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/gopsclient"
	"github.com/inovacc/omni/pkg/obfuscate"
	"github.com/inovacc/omni/pkg/procutil"
)

// InspectReport aggregates per-process info plus the obfuscation verdict.
// HasAgent is reserved for Phase 2B (agent infrastructure); currently always false.
type InspectReport struct {
	Process   ProcessInfo       `json:"process"`
	Obfuscate obfuscate.Verdict `json:"obfuscation"`
	HasAgent  bool              `json:"has_agent"`
}

// ProcessInfo is a JSON-friendly subset of procutil.Process flattened for the
// inspect report. We don't reuse procutil.Process directly so the JSON shape
// is stable even if procutil adds fields.
type ProcessInfo struct {
	PID       int32  `json:"pid"`
	PPID      int32  `json:"ppid"`
	Name      string `json:"name"`
	ExePath   string `json:"exe_path"`
	Username  string `json:"username,omitempty"`
	Runtime   string `json:"runtime"`
	GoVersion string `json:"go_version,omitempty"`
	Module    string `json:"module,omitempty"`
}

// InspectOptions configures the inspect verb.
type InspectOptions struct {
	Format string // "table" | "json"
}

// RunInspect renders detailed info + obfuscation verdict for a single PID.
func RunInspect(ctx context.Context, w io.Writer, pidStr string, opts InspectOptions) error {
	pid64, err := strconv.ParseInt(pidStr, 10, 32)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("invalid pid %q: %v", pidStr, err))
	}
	pid := int32(pid64)

	p, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("pid %d: %v", pid, err))
	}
	exe, err := p.ExeWithContext(ctx)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("resolve exe for pid %d: %v", pid, err))
	}
	name, _ := p.NameWithContext(ctx)
	ppid, _ := p.PpidWithContext(ctx)
	user, _ := p.UsernameWithContext(ctx)

	info, isGo, _ := procutil.ReadGoBinary(exe)
	runtime := procutil.RuntimeUnknown
	if isGo {
		runtime = procutil.RuntimeGo
	}

	v, vErr := obfuscate.Detect(exe)
	if vErr != nil {
		// Non-fatal — surface the partial verdict; caller can see the error in v.Path.
		_, _ = fmt.Fprintf(w, "warning: obfuscation probe failed: %v\n", vErr)
	}

	report := InspectReport{
		Process: ProcessInfo{
			PID:       pid,
			PPID:      ppid,
			Name:      name,
			ExePath:   exe,
			Username:  user,
			Runtime:   runtime.String(),
			GoVersion: info.GoVersion,
			Module:    info.Module,
		},
		Obfuscate: v,
		HasAgent:  gopsclient.HasAgent(int(pid)),
	}

	switch opts.Format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	default:
		return renderInspectTable(w, report)
	}
}

func renderInspectTable(w io.Writer, r InspectReport) error {
	_, _ = fmt.Fprintf(w, "PID:         %d\n", r.Process.PID)
	_, _ = fmt.Fprintf(w, "PPID:        %d\n", r.Process.PPID)
	_, _ = fmt.Fprintf(w, "Name:        %s\n", r.Process.Name)
	_, _ = fmt.Fprintf(w, "Exe:         %s\n", r.Process.ExePath)
	_, _ = fmt.Fprintf(w, "Username:    %s\n", r.Process.Username)
	_, _ = fmt.Fprintf(w, "Runtime:     %s\n", r.Process.Runtime)
	if r.Process.GoVersion != "" {
		_, _ = fmt.Fprintf(w, "Go version:  %s\n", r.Process.GoVersion)
	}
	if r.Process.Module != "" {
		_, _ = fmt.Fprintf(w, "Module:      %s\n", r.Process.Module)
	}
	_, _ = fmt.Fprintf(w, "Obfuscation: %s", r.Obfuscate.Verdict)
	if r.Obfuscate.Confidence != "" {
		_, _ = fmt.Fprintf(w, " (%s confidence)", r.Obfuscate.Confidence)
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "  buildinfo: %t  symbols: %t  main_mangled: %t\n",
		r.Obfuscate.BuildInfoFound, r.Obfuscate.SymbolsFound, r.Obfuscate.MainMangled)
	_, _ = fmt.Fprintf(w, "Agent:       %t\n", r.HasAgent)
	return nil
}
