package runtimeps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/shirou/gopsutil/v3/process"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/obfuscate"
)

// ObfuscationOptions configures the obfuscation verb.
type ObfuscationOptions struct {
	Format string // "table" | "json"
}

// RunObfuscation runs the obfuscate heuristics against a target that is
// either a numeric PID (resolved to its executable via gopsutil) or a path
// to a binary on disk. Renders the verdict in the chosen format.
func RunObfuscation(ctx context.Context, w io.Writer, target string, opts ObfuscationOptions) error {
	path := target
	if pid64, err := strconv.ParseInt(target, 10, 32); err == nil {
		p, err := process.NewProcessWithContext(ctx, int32(pid64))
		if err != nil {
			return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("pid %d: %v", pid64, err))
		}
		exe, err := p.ExeWithContext(ctx)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("resolve exe for pid %d: %v", pid64, err))
		}
		path = exe
	}
	v, err := obfuscate.Detect(path)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("obfuscation probe: %v", err))
	}

	switch opts.Format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	default:
		_, _ = fmt.Fprintf(w, "Path:        %s\n", v.Path)
		_, _ = fmt.Fprintf(w, "Verdict:     %s", v.Verdict)
		if v.Confidence != "" {
			_, _ = fmt.Fprintf(w, " (%s confidence)", v.Confidence)
		}
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintf(w, "Buildinfo:   %t\n", v.BuildInfoFound)
		_, _ = fmt.Fprintf(w, "Symbols:     %t\n", v.SymbolsFound)
		if v.GoVersion != "" {
			_, _ = fmt.Fprintf(w, "Go version:  %s\n", v.GoVersion)
		}
		_, _ = fmt.Fprintf(w, "Main mangled: %t\n", v.MainMangled)
		return nil
	}
}
