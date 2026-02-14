package exist

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/inovacc/omni/internal/cli/ps"
	gnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// ErrNotFound is returned when the target does not exist.
var ErrNotFound = errors.New("not found")

// Options configures the exist command behavior.
type Options struct {
	Quiet bool // -q: suppress output
	JSON  bool // --json: output as JSON
}

// Result represents the existence check result.
type Result struct {
	Target  string `json:"target"`
	Exists  bool   `json:"exists"`
	Type    string `json:"type"`
	Details any    `json:"details,omitempty"`
}

// FileDetails contains metadata for file checks.
type FileDetails struct {
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"mod_time"`
}

// CommandDetails contains metadata for command checks.
type CommandDetails struct {
	Path string `json:"path"`
}

// EnvDetails contains metadata for env var checks.
type EnvDetails struct {
	Value string `json:"value"`
}

// ProcessDetails contains metadata for process checks.
type ProcessDetails struct {
	PID     int    `json:"pid"`
	Name    string `json:"name"`
	Command string `json:"command,omitempty"`
}

// PortDetails contains metadata for port checks.
type PortDetails struct {
	Port        uint32 `json:"port"`
	Protocol    string `json:"protocol"`
	PID         int32  `json:"pid,omitempty"`
	ProcessName string `json:"process_name,omitempty"`
}

// RunFile checks if a regular file exists.
func RunFile(w io.Writer, target string, opts Options) error {
	info, err := os.Stat(target)
	if err != nil || !info.Mode().IsRegular() {
		return outputResult(w, Result{
			Target: target,
			Exists: false,
			Type:   "file",
		}, opts)
	}

	return outputResult(w, Result{
		Target: target,
		Exists: true,
		Type:   "file",
		Details: FileDetails{
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			ModTime: info.ModTime().Format(time.RFC3339),
		},
	}, opts)
}

// RunDir checks if a directory exists.
func RunDir(w io.Writer, target string, opts Options) error {
	info, err := os.Stat(target)
	if err != nil || !info.IsDir() {
		return outputResult(w, Result{
			Target: target,
			Exists: false,
			Type:   "dir",
		}, opts)
	}

	return outputResult(w, Result{
		Target: target,
		Exists: true,
		Type:   "dir",
		Details: FileDetails{
			Mode:    info.Mode().String(),
			ModTime: info.ModTime().Format(time.RFC3339),
		},
	}, opts)
}

// RunPath checks if any path exists (file, dir, symlink, etc.).
func RunPath(w io.Writer, target string, opts Options) error {
	info, err := os.Lstat(target)
	if err != nil {
		return outputResult(w, Result{
			Target: target,
			Exists: false,
			Type:   "path",
		}, opts)
	}

	kind := "file"
	if info.IsDir() {
		kind = "dir"
	} else if info.Mode()&os.ModeSymlink != 0 {
		kind = "symlink"
	}

	return outputResult(w, Result{
		Target: target,
		Exists: true,
		Type:   kind,
		Details: FileDetails{
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			ModTime: info.ModTime().Format(time.RFC3339),
		},
	}, opts)
}

// RunCommand checks if a command exists in PATH.
func RunCommand(w io.Writer, target string, opts Options) error {
	path := findCommand(target)
	if path == "" {
		return outputResult(w, Result{
			Target: target,
			Exists: false,
			Type:   "command",
		}, opts)
	}

	return outputResult(w, Result{
		Target: target,
		Exists: true,
		Type:   "command",
		Details: CommandDetails{
			Path: path,
		},
	}, opts)
}

// RunEnv checks if an environment variable is set.
func RunEnv(w io.Writer, target string, opts Options) error {
	val, ok := os.LookupEnv(target)
	if !ok {
		return outputResult(w, Result{
			Target: target,
			Exists: false,
			Type:   "env",
		}, opts)
	}

	return outputResult(w, Result{
		Target: target,
		Exists: true,
		Type:   "env",
		Details: EnvDetails{
			Value: val,
		},
	}, opts)
}

// RunProcess checks if a process is running by name or PID.
func RunProcess(w io.Writer, target string, opts Options) error {
	pid, err := strconv.Atoi(target)
	if err == nil {
		return runProcessByPID(w, target, pid, opts)
	}

	return runProcessByName(w, target, opts)
}

func runProcessByPID(w io.Writer, target string, pid int, opts Options) error {
	procs, err := ps.GetProcessList(ps.Options{All: true, Pid: pid})
	if err != nil || len(procs) == 0 {
		return outputResult(w, Result{
			Target: target,
			Exists: false,
			Type:   "process",
		}, opts)
	}

	p := procs[0]

	return outputResult(w, Result{
		Target: target,
		Exists: true,
		Type:   "process",
		Details: ProcessDetails{
			PID:     p.PID,
			Name:    p.Command,
			Command: p.Command,
		},
	}, opts)
}

func runProcessByName(w io.Writer, target string, opts Options) error {
	procs, err := ps.GetProcessList(ps.Options{All: true})
	if err != nil {
		return outputResult(w, Result{
			Target: target,
			Exists: false,
			Type:   "process",
		}, opts)
	}

	search := strings.ToLower(target)

	for _, p := range procs {
		if strings.Contains(strings.ToLower(p.Command), search) {
			return outputResult(w, Result{
				Target: target,
				Exists: true,
				Type:   "process",
				Details: ProcessDetails{
					PID:     p.PID,
					Name:    p.Command,
					Command: p.Command,
				},
			}, opts)
		}
	}

	return outputResult(w, Result{
		Target: target,
		Exists: false,
		Type:   "process",
	}, opts)
}

// RunPort checks if a TCP port is listening.
func RunPort(w io.Writer, target string, opts Options) error {
	port, err := strconv.Atoi(target)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("exist: invalid port number: %s", target)
	}

	// Try gopsutil first for LISTEN state detection
	conns, err := gnet.Connections("tcp")
	if err == nil {
		for _, conn := range conns {
			state := strings.ToUpper(conn.Status)
			if conn.Laddr.Port == uint32(port) && (state == "LISTEN" || state == "ESTABLISHED") {
				details := PortDetails{
					Port:     conn.Laddr.Port,
					Protocol: "tcp",
					PID:      conn.Pid,
				}

				// Try to resolve process name
				if conn.Pid > 0 {
					if p, pErr := process.NewProcess(conn.Pid); pErr == nil {
						if name, nErr := p.Name(); nErr == nil {
							details.ProcessName = name
						}
					}
				}

				return outputResult(w, Result{
					Target:  target,
					Exists:  true,
					Type:    "port",
					Details: details,
				}, opts)
			}
		}
	}

	// Fallback: try dialing the port
	conn, dialErr := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
	if dialErr == nil {
		_ = conn.Close()

		return outputResult(w, Result{
			Target: target,
			Exists: true,
			Type:   "port",
			Details: PortDetails{
				Port:     uint32(port),
				Protocol: "tcp",
			},
		}, opts)
	}

	return outputResult(w, Result{
		Target: target,
		Exists: false,
		Type:   "port",
	}, opts)
}

// outputResult writes the result and returns ErrNotFound if the target doesn't exist.
func outputResult(w io.Writer, result Result, opts Options) error {
	if opts.Quiet {
		if !result.Exists {
			return ErrNotFound
		}

		return nil
	}

	if opts.JSON {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")

		if err := encoder.Encode(result); err != nil {
			return fmt.Errorf("exist: json encode: %w", err)
		}
	} else if result.Exists {
		_, _ = fmt.Fprintf(w, "%s: exists (%s)\n", result.Target, result.Type)
	} else {
		_, _ = fmt.Fprintf(w, "%s: not found\n", result.Target)
	}

	if !result.Exists {
		return ErrNotFound
	}

	return nil
}

// findCommand searches for an executable in PATH.
func findCommand(name string) string {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return ""
	}

	exts := []string{""}

	if runtime.GOOS == "windows" {
		pathExt := os.Getenv("PATHEXT")
		if pathExt == "" {
			pathExt = ".COM;.EXE;.BAT;.CMD"
		}

		exts = append(exts, strings.Split(strings.ToLower(pathExt), ";")...)
	}

	for dir := range strings.SplitSeq(pathEnv, string(os.PathListSeparator)) {
		fullPath := filepath.Join(dir, name)

		for _, ext := range exts {
			if isExec(fullPath + ext) {
				return fullPath + ext
			}
		}
	}

	return ""
}

func isExec(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}

	if runtime.GOOS == "windows" {
		return true
	}

	return info.Mode()&0111 != 0
}
