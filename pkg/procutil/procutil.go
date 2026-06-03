// Package procutil enumerates and classifies running processes by their
// language runtime (Go, Node.js, Python, Java) and signals them. It uses
// pure-Go process introspection via gopsutil and stdlib debug/buildinfo;
// no external processes are ever spawned (matches omni's "no exec" rule).
//
// The CLI wrappers in internal/cli/{gops,nodeps,pyps,javaps}/ are thin
// runtime-filtered views over this package.
//
// Portions of the classification + kill logic are adapted from
// github.com/inovacc/gops (MIT) — see THIRD_PARTY_LICENSES/gops-MIT.txt.
//
// Experimental: this package's API may change before a stable release and is
// not covered by the v1.0 compatibility guarantee.
package procutil

import (
	"context"
	"os"
	"path/filepath"

	"github.com/shirou/gopsutil/v3/process"
)

// Runtime identifies the language/runtime of a process.
type Runtime string

const (
	RuntimeGo      Runtime = "go"
	RuntimeNode    Runtime = "node"
	RuntimePython  Runtime = "python"
	RuntimeJava    Runtime = "java"
	RuntimeUnknown Runtime = "unknown"
)

// String returns the canonical lowercase identifier.
func (r Runtime) String() string { return string(r) }

// Process is a runtime-classified snapshot of a running process.
type Process struct {
	PID       int32   `json:"pid"`
	PPID      int32   `json:"ppid"`
	Name      string  `json:"name"`
	ExePath   string  `json:"exe_path"`
	Username  string  `json:"username,omitempty"`
	Runtime   Runtime `json:"runtime"`
	GoVersion string  `json:"go_version,omitempty"` // populated only for RuntimeGo
	Module    string  `json:"module,omitempty"`     // populated only for RuntimeGo
}

// ListOptions configures process enumeration.
type ListOptions struct {
	// IncludeSelf includes the calling process in the result. Off by default
	// so callers can list "other" processes without filtering themselves.
	IncludeSelf bool
	// Runtime restricts results to one runtime; zero value (empty string) means all.
	Runtime Runtime
}

// ListAll enumerates every running process and classifies each by runtime.
// Unreadable / inaccessible processes are silently skipped (typical for
// short-lived or higher-privileged processes).
func ListAll(ctx context.Context) ([]Process, error) {
	return List(ctx, ListOptions{})
}

// ListByRuntime is shorthand for List with Runtime filter set.
func ListByRuntime(ctx context.Context, rt Runtime) ([]Process, error) {
	return List(ctx, ListOptions{Runtime: rt})
}

// List enumerates processes with the given options.
func List(ctx context.Context, opts ListOptions) ([]Process, error) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	selfPID := int32(os.Getpid())
	out := make([]Process, 0, len(procs))
	for _, p := range procs {
		if !opts.IncludeSelf && p.Pid == selfPID {
			continue
		}
		exe, err := p.ExeWithContext(ctx)
		if err != nil || exe == "" || !filepath.IsAbs(exe) {
			continue
		}
		fi, err := os.Stat(exe)
		if err != nil || fi.IsDir() {
			continue
		}
		rt, info := classifyExe(exe)
		if opts.Runtime != "" && rt != opts.Runtime {
			continue
		}
		if opts.Runtime == "" && rt == RuntimeUnknown {
			// Caller asked for everything; skip unclassified noise.
			continue
		}
		name, _ := p.NameWithContext(ctx)
		ppid, _ := p.PpidWithContext(ctx)
		user, _ := p.UsernameWithContext(ctx)
		proc := Process{
			PID:      p.Pid,
			PPID:     ppid,
			Name:     name,
			ExePath:  exe,
			Username: user,
			Runtime:  rt,
		}
		if info != nil {
			proc.GoVersion = info.GoVersion
			proc.Module = info.Module
		}
		out = append(out, proc)
	}
	return out, nil
}
