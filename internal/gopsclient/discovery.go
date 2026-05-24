// Package gopsclient is the omni-side client for the embeddable runtime
// agent shipped in pkg/gopsagent. It locates a target agent by reading the
// pid-keyed address file in $HOME/.config/gops (overridable via env vars),
// then performs an opcode roundtrip per call.
//
// Adapted from github.com/inovacc/gops (MIT) — see THIRD_PARTY_LICENSES/gops-MIT.txt.
package gopsclient

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigDir returns the directory where agent address files live.
// Order of precedence: $GOPS_CONFIG_DIR, $XDG_CONFIG_HOME/gops, $HOME/.config/gops.
func ConfigDir() string {
	if d := os.Getenv("GOPS_CONFIG_DIR"); d != "" {
		return d
	}
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "gops")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "gops")
}

// AddrForPID reads "host:port" written by an embedded agent for pid.
func AddrForPID(pid int) (string, error) {
	p := filepath.Join(ConfigDir(), fmt.Sprintf("%d", pid))
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// HasAgent is a non-erroring convenience for "does PID have an agent?".
// Returns true iff AddrForPID succeeds with a non-empty value.
func HasAgent(pid int) bool {
	addr, err := AddrForPID(pid)
	return err == nil && addr != ""
}
