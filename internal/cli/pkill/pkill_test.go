package pkill

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"syscall"
	"testing"
)

func TestSignalMap(t *testing.T) {
	expectedSignals := map[string]syscall.Signal{
		"HUP":  syscall.SIGHUP,
		"INT":  syscall.SIGINT,
		"QUIT": syscall.SIGQUIT,
		"KILL": syscall.SIGKILL,
		"TERM": syscall.SIGTERM,
		"ABRT": syscall.SIGABRT,
	}

	for name, expected := range expectedSignals {
		t.Run(name, func(t *testing.T) {
			sig, ok := signalMap[name]
			if !ok {
				t.Errorf("signal %q not found in signalMap", name)
				return
			}
			if sig != expected {
				t.Errorf("signalMap[%q] = %v, want %v", name, sig, expected)
			}
		})
	}
}

func TestRun_EmptyPattern(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, "", Options{})
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
	if !strings.Contains(err.Error(), "no pattern specified") {
		t.Errorf("expected 'no pattern specified' error, got: %v", err)
	}
}

func TestRun_InvalidPattern(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, "[invalid", Options{})
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
	if !strings.Contains(err.Error(), "invalid pattern") {
		t.Errorf("expected 'invalid pattern' error, got: %v", err)
	}
}

func TestRun_InvalidSignal(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, "nonexistent_process_xyz_12345", Options{Signal: "INVALID"})
	if err == nil {
		t.Fatal("expected error for invalid signal")
	}
	if !strings.Contains(err.Error(), "invalid signal") {
		t.Errorf("expected 'invalid signal' error, got: %v", err)
	}
}

func TestRun_ListOnly(t *testing.T) {
	var buf bytes.Buffer
	// List processes matching a common pattern (should match at least something)
	err := Run(&buf, ".*", Options{ListOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	// Should have PID and name on each line
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Error("expected at least one process match")
	}
}

func TestRun_ListOnly_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, ".*", Options{ListOnly: true, JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	var results []Result
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Errorf("expected valid JSON output, got error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one result")
	}
	for _, r := range results {
		if !r.Matched {
			t.Errorf("expected all results to be matched, PID %d not matched", r.PID)
		}
	}
}

func TestRun_Count(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, ".*", Options{Count: true})
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(buf.String())
	// Should be a number
	if output == "" {
		t.Error("expected count output")
	}
	if output == "0" {
		t.Error("expected non-zero count")
	}
}

func TestRun_CountJSON(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, ".*", Options{Count: true, JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]int
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("expected valid JSON, got error: %v", err)
	}
	if count, ok := result["count"]; !ok || count == 0 {
		t.Error("expected non-zero count in JSON")
	}
}

func TestRun_ExactMatch(t *testing.T) {
	var buf bytes.Buffer
	// "nonexistent_process_xyz" should not match anything
	err := Run(&buf, "nonexistent_process_xyz_12345", Options{Exact: true, ListOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(buf.String())
	if output != "" {
		t.Errorf("expected no match for nonexistent process, got: %s", output)
	}
}

func TestRun_CaseInsensitive(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, ".*", Options{ListOnly: true, IgnoreCase: true})
	if err != nil {
		t.Fatal(err)
	}
	// Should work without error
}

func TestRun_Newest(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, ".*", Options{Newest: true, ListOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	if len(lines) != 1 {
		t.Errorf("expected exactly 1 line (newest process), got %d", len(lines))
	}
}

func TestRun_Oldest(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, ".*", Options{Oldest: true, ListOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")
	if len(lines) != 1 {
		t.Errorf("expected exactly 1 line (oldest process), got %d", len(lines))
	}
}

func TestRun_NoMatch_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, "nonexistent_process_xyz_12345", Options{ListOnly: true, JSON: true})
	if err != nil {
		t.Fatal(err)
	}
	output := strings.TrimSpace(buf.String())
	if output != "[]" {
		t.Errorf("expected empty JSON array for no matches, got: %s", output)
	}
}

func TestRun_SignalParsing(t *testing.T) {
	tests := []struct {
		name   string
		signal string
		valid  bool
	}{
		{"term", "TERM", true},
		{"kill", "KILL", true},
		{"hup", "HUP", true},
		{"sig_prefix", "SIGTERM", true},
		{"numeric", "15", true},
		{"invalid", "INVALID", false},
		{"invalid_number", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Run(&buf, "nonexistent_process_xyz_12345", Options{Signal: tt.signal, ListOnly: true})
			if tt.valid {
				if err != nil {
					t.Errorf("expected no error for signal %q, got: %v", tt.signal, err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error for invalid signal %q", tt.signal)
				}
			}
		})
	}
}

func TestPatternCompilation(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		exact      bool
		ignoreCase bool
		shouldMatch string
	}{
		{"simple", "test", false, false, "testing"},
		{"exact", "test", true, false, "test"},
		{"case_insensitive", "test", false, true, "Testing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patternStr := tt.pattern
			if tt.exact {
				patternStr = "^" + regexp.QuoteMeta(tt.pattern) + "$"
			}
			if tt.ignoreCase {
				patternStr = "(?i)" + patternStr
			}
			re, err := regexp.Compile(patternStr)
			if err != nil {
				t.Fatal(err)
			}
			if !re.MatchString(tt.shouldMatch) {
				t.Errorf("pattern %q should match %q", patternStr, tt.shouldMatch)
			}
		})
	}
}
