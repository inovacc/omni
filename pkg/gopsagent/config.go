package gopsagent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Config is the on-disk shape of pkg/gopsagent's optional config file.
// JSON-encoded so the agent has zero extra dependencies.
//
// Default path: $GOPS_CONFIG_DIR/config.json,
// or $XDG_CONFIG_HOME/gops/config.json,
// or $HOME/.config/gops/config.json.
//
// Fields are pointers so we can tell "absent" from "explicitly false" —
// callers can layer env-var overrides on top of file-derived values.
type Config struct {
	// NotifyOnStartup makes Listen() print one stderr line when the agent
	// successfully binds — useful for ops dashboards, supervisord-style
	// init systems, or just human verification that the agent came up.
	// Default: false.
	NotifyOnStartup *bool `json:"notify_on_startup,omitempty"`
}

// DefaultConfigPath returns the resolved file path the agent reads on Listen.
// Honours $GOPS_CONFIG_DIR > $XDG_CONFIG_HOME/gops > $HOME/.config/gops.
func DefaultConfigPath() string { return filepath.Join(defaultConfigDir(), "config.json") }

func defaultConfigDir() string {
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

// LoadConfig reads the config file at DefaultConfigPath. Missing file
// returns (zero Config, nil) — absence is not an error since the config is
// entirely optional. A malformed file returns the parse error so callers
// can choose to warn rather than start with surprise defaults.
func LoadConfig() (Config, error) { return LoadConfigFrom(DefaultConfigPath()) }

// LoadConfigFrom reads a config file from path. Use this for tests or
// non-default deployments. Missing path returns (zero, nil).
func LoadConfigFrom(path string) (Config, error) {
	var cfg Config
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("gopsagent: open config %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	if err := json.NewDecoder(f).Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return cfg, fmt.Errorf("gopsagent: parse config %s: %w", path, err)
	}
	return cfg, nil
}

// notifyEnabled returns the final on/off decision for the startup notification.
// Resolution order, highest priority first:
//
//  1. GOPS_AGENT_NOTIFY env var ("1"/"true"/"yes" → enabled; "0"/"false"/"no" → disabled)
//  2. Config file's notify_on_startup field
//  3. Default: false
func notifyEnabled(cfg Config) bool {
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("GOPS_AGENT_NOTIFY"))); v != "" {
		switch v {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	}
	if cfg.NotifyOnStartup != nil {
		return *cfg.NotifyOnStartup
	}
	return false
}

// fireStartupNotification writes the one-line stderr message. Kept as a var
// so tests can swap in a buffer and assert on the rendered text.
var fireStartupNotification = func(addr string, pid int) {
	_, _ = fmt.Fprintf(os.Stderr, "gops agent listening on %s (pid %d)\n", addr, pid)
}
