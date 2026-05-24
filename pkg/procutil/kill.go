package procutil

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// Signal represents a portable signal name. Windows only supports
// terminate-style signals (TERM, KILL); INT/HUP error there.
type Signal string

const (
	SigTerm Signal = "TERM"
	SigKill Signal = "KILL"
	SigInt  Signal = "INT"
	SigHup  Signal = "HUP"
)

// ParseSignal converts a user-supplied string to a Signal. Empty input
// defaults to TERM. Comparison is case-insensitive.
func ParseSignal(s string) (Signal, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "", "TERM":
		return SigTerm, nil
	case "KILL":
		return SigKill, nil
	case "INT":
		return SigInt, nil
	case "HUP":
		return SigHup, nil
	default:
		return "", fmt.Errorf("unknown signal %q (want TERM|KILL|INT|HUP)", s)
	}
}

// Kill delivers sig to a single PID. On Windows, only TERM/KILL are honoured
// (both map to TerminateProcess); INT/HUP return an unsupported-OS error.
func Kill(pid int, sig Signal) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid %d", pid)
	}
	return sendImpl(pid, sig)
}

// KillResult is one row of a KillAllMatching outcome.
type KillResult struct {
	PID  int32
	Name string
	Err  error // nil on success
}

// KillAllMatching enumerates processes (optionally filtered to runtimeFilter),
// matches each against target, and signals every match.
//
// Matching rules — target is treated as:
//   - A numeric string → single PID match (exact equality)
//   - Anything else    → matches if equal to (case-insensitive) the process
//     Name, or the basename of ExePath (with .exe stripped on Windows).
//
// runtimeFilter narrows the candidate set; pass empty string to match across all runtimes.
// Returns one KillResult per signaled process. The caller decides how to render aggregated errors.
func KillAllMatching(ctx context.Context, target string, sig Signal, runtimeFilter Runtime) ([]KillResult, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, errors.New("target is required (pid or process name)")
	}
	// Numeric → direct PID kill, no enumeration needed.
	if pid, err := strconv.Atoi(target); err == nil {
		if err := Kill(pid, sig); err != nil {
			return []KillResult{{PID: int32(pid), Err: err}}, err
		}
		return []KillResult{{PID: int32(pid)}}, nil
	}

	procs, err := List(ctx, ListOptions{Runtime: runtimeFilter})
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(target)
	matches := make([]Process, 0, 4)
	for _, p := range procs {
		base := strings.TrimSuffix(strings.ToLower(filepath.Base(p.ExePath)), ".exe")
		name := strings.TrimSuffix(strings.ToLower(p.Name), ".exe")
		if name == lower || base == lower {
			matches = append(matches, p)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no processes match %q (runtime=%q)", target, runtimeFilter)
	}

	results := make([]KillResult, 0, len(matches))
	var firstErr error
	for _, p := range matches {
		err := Kill(int(p.PID), sig)
		r := KillResult{PID: p.PID, Name: p.Name, Err: err}
		results = append(results, r)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return results, firstErr
}
