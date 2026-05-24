package runtimeps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/procmetrics"
)

// MonitorOptions configures the monitor verb.
type MonitorOptions struct {
	Watch    bool          // stream metrics continuously
	Interval time.Duration // sampling interval when Watch=true
	Format   string        // "table" | "json" (single-shot); "ndjson" forced when Watch=true
}

// RunMonitor samples gopsutil metrics for a single PID. When opts.Watch is
// true it streams NDJSON until the context is cancelled. The single-shot
// path renders one record in the chosen format.
func RunMonitor(ctx context.Context, w io.Writer, pidStr string, opts MonitorOptions) error {
	pid64, err := strconv.ParseInt(pidStr, 10, 32)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("invalid pid %q: %v", pidStr, err))
	}
	pid := int32(pid64)
	c := procmetrics.NewCollector()

	if !opts.Watch {
		m, err := c.Collect(ctx, pid)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("monitor pid %d: %v", pid, err))
		}
		return renderMetrics(w, m, opts.Format)
	}

	if opts.Interval <= 0 {
		opts.Interval = time.Second
	}
	enc := json.NewEncoder(w)
	t := time.NewTicker(opts.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			m, err := c.Collect(ctx, pid)
			if err != nil {
				return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("monitor pid %d: %v", pid, err))
			}
			if err := enc.Encode(m); err != nil {
				return err
			}
		}
	}
}

func renderMetrics(w io.Writer, m procmetrics.Metrics, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(m)
	default:
		_, _ = fmt.Fprintf(w, "PID:           %d\n", m.PID)
		_, _ = fmt.Fprintf(w, "CPU:           %.2f%%\n", m.CPUPercent)
		_, _ = fmt.Fprintf(w, "Mem RSS:       %s\n", humanBytes(m.MemRSS))
		_, _ = fmt.Fprintf(w, "Mem VMS:       %s\n", humanBytes(m.MemVMS))
		_, _ = fmt.Fprintf(w, "Disk read:     %s\n", humanBytes(m.DiskReadBytes))
		_, _ = fmt.Fprintf(w, "Disk write:    %s\n", humanBytes(m.DiskWriteBytes))
		_, _ = fmt.Fprintf(w, "Open FDs:      %d\n", m.OpenFDs)
		return nil
	}
}

func humanBytes(n uint64) string {
	const (
		KiB = 1 << 10
		MiB = 1 << 20
		GiB = 1 << 30
	)
	switch {
	case n >= GiB:
		return fmt.Sprintf("%.2f GiB", float64(n)/GiB)
	case n >= MiB:
		return fmt.Sprintf("%.2f MiB", float64(n)/MiB)
	case n >= KiB:
		return fmt.Sprintf("%.2f KiB", float64(n)/KiB)
	default:
		return fmt.Sprintf("%d B", n)
	}
}
