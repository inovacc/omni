package gopsagent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_MissingFileIsZeroValue(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfigFrom(filepath.Join(dir, "nope.json"))
	if err != nil {
		t.Fatalf("LoadConfigFrom on missing file should be (zero, nil); got err=%v", err)
	}
	if cfg.NotifyOnStartup != nil {
		t.Errorf("missing file should leave NotifyOnStartup nil; got %v", *cfg.NotifyOnStartup)
	}
}

func TestLoadConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	want := true
	src := Config{NotifyOnStartup: &want}
	b, _ := json.Marshal(src)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}
	if got.NotifyOnStartup == nil || *got.NotifyOnStartup != true {
		t.Errorf("NotifyOnStartup = %v, want true", got.NotifyOnStartup)
	}
}

func TestLoadConfig_MalformedReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadConfigFrom(path)
	if err == nil {
		t.Error("malformed config should return parse error")
	}
}

func TestNotifyEnabled_Resolution(t *testing.T) {
	// Helper to materialize a *bool inline.
	bp := func(b bool) *bool { return &b }

	cases := []struct {
		name string
		env  string // value for GOPS_AGENT_NOTIFY; "<unset>" means clear
		cfg  Config
		want bool
	}{
		{"unset env, no config → default off", "<unset>", Config{}, false},
		{"unset env, config true", "<unset>", Config{NotifyOnStartup: bp(true)}, true},
		{"unset env, config explicit false", "<unset>", Config{NotifyOnStartup: bp(false)}, false},
		{"env=1 trumps config false", "1", Config{NotifyOnStartup: bp(false)}, true},
		{"env=true trumps config false", "true", Config{NotifyOnStartup: bp(false)}, true},
		{"env=0 trumps config true", "0", Config{NotifyOnStartup: bp(true)}, false},
		{"env=off trumps config true", "off", Config{NotifyOnStartup: bp(true)}, false},
		{"env=garbage falls through to config", "wat", Config{NotifyOnStartup: bp(true)}, true},
		{"env=garbage falls through to default", "wat", Config{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.env == "<unset>" {
				t.Setenv("GOPS_AGENT_NOTIFY", "")
			} else {
				t.Setenv("GOPS_AGENT_NOTIFY", tc.env)
			}
			if got := notifyEnabled(tc.cfg); got != tc.want {
				t.Errorf("notifyEnabled = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDefaultConfigPath_HonoursGOPSConfigDir(t *testing.T) {
	t.Setenv("GOPS_CONFIG_DIR", "/custom/path")
	if got := DefaultConfigPath(); got != filepath.Join("/custom/path", "config.json") {
		t.Errorf("DefaultConfigPath = %s, want /custom/path/config.json", got)
	}
}
