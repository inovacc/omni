package runtimeps

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/gopsclient"
)

// RunAgentCmd sends a named opcode (stack/gc/memstats/version/stats/snapshot)
// to the agent embedded in the target pid and writes the raw response to w.
func RunAgentCmd(ctx context.Context, w io.Writer, pidStr, opName string) error {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("invalid pid %q: %v", pidStr, err))
	}
	op, err := OpcodeForName(opName)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, err.Error())
	}
	addr, err := gopsclient.AddrForPID(pid)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrNotFound,
			fmt.Sprintf("agent not found for pid %d: %v — target must embed pkg/gopsagent and call Listen()", pid, err))
	}
	resp, err := gopsclient.NewClient(addr).Call(ctx, op)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, err.Error())
	}
	_, _ = w.Write(resp)
	return nil
}

// OpcodeForName maps a user-friendly command name to its byte opcode.
// Used by both the agent-cmd CLI and tests.
func OpcodeForName(name string) (byte, error) {
	switch name {
	case "stack":
		return gopsclient.OpStack, nil
	case "gc":
		return gopsclient.OpGC, nil
	case "memstats":
		return gopsclient.OpMemStats, nil
	case "version":
		return gopsclient.OpVersion, nil
	case "stats":
		return gopsclient.OpStats, nil
	case "snapshot":
		return gopsclient.OpRuntimeSnapshot, nil
	default:
		return 0, fmt.Errorf("unknown agent cmd %q (supported: stack, gc, memstats, version, stats, snapshot)", name)
	}
}

// TraceOptions configures the trace verb.
type TraceOptions struct {
	Duration time.Duration
	OutFile  string // empty → stdout
}

// RunTrace captures a runtime trace from the target's embedded agent.
func RunTrace(ctx context.Context, w io.Writer, pidStr string, opts TraceOptions) error {
	return runProfileVerb(ctx, w, pidStr, opts.Duration, opts.OutFile, gopsclient.OpTrace, "tracing", "go tool trace")
}

// ProfileOptions configures the profile verb.
type ProfileOptions struct {
	Duration time.Duration
	OutFile  string
}

// RunProfile captures a CPU profile from the target's embedded agent.
func RunProfile(ctx context.Context, w io.Writer, pidStr string, opts ProfileOptions) error {
	return runProfileVerb(ctx, w, pidStr, opts.Duration, opts.OutFile, gopsclient.OpCPUProfile, "profiling", "go tool pprof")
}

func runProfileVerb(ctx context.Context, w io.Writer, pidStr string, dur time.Duration, outFile string, op byte, gerund, analyzeTool string) error {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("invalid pid %q: %v", pidStr, err))
	}
	addr, err := gopsclient.AddrForPID(pid)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrNotFound,
			fmt.Sprintf("agent not found for pid %d: %v — target must embed pkg/gopsagent and call Listen()", pid, err))
	}
	out := w
	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrIO, err.Error())
		}
		defer func() { _ = f.Close() }()
		out = f
	}
	_, _ = fmt.Fprintf(os.Stderr, "%s pid %d for %s...\n", gerund, pid, dur)
	if err := gopsclient.NewClient(addr).CallProfile(ctx, op, dur, out); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, err.Error())
	}
	if outFile != "" {
		_, _ = fmt.Fprintf(os.Stderr, "wrote %s; analyze with: %s %s\n", outFile, analyzeTool, outFile)
	}
	return nil
}

// StreamOptions configures the stream verb.
type StreamOptions struct {
	Interval time.Duration
}

// RunStream copies NDJSON runtime snapshots from the agent to w until ctx ends.
func RunStream(ctx context.Context, w io.Writer, pidStr string, opts StreamOptions) error {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("invalid pid %q: %v", pidStr, err))
	}
	addr, err := gopsclient.AddrForPID(pid)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrNotFound,
			fmt.Sprintf("agent not found for pid %d: %v — target must embed pkg/gopsagent and call Listen()", pid, err))
	}
	if opts.Interval <= 0 {
		opts.Interval = time.Second
	}
	ms := uint32(opts.Interval / time.Millisecond)
	if err := gopsclient.NewClient(addr).Stream(ctx, ms, w); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, err.Error())
	}
	return nil
}
