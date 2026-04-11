package pkill

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"syscall"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
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

	err := Run(&buf, ".*", Options{ListOnly: true, OutputFormat: output.FormatJSON})
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

	err := Run(&buf, ".*", Options{Count: true, OutputFormat: output.FormatJSON})
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
	// "nonexistent_process_xyz" should not match anything → SilentExit(1)
	err := Run(&buf, "nonexistent_process_xyz_12345", Options{Exact: true, ListOnly: true})
	var silent *cmderr.SilentError
	if !errors.As(err, &silent) || silent.Code != 1 {
		t.Fatalf("expected SilentExit(1) for no match, got: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "" {
		t.Errorf("expected no output for nonexistent process, got: %s", output)
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

	// No match → SilentExit(1) even in JSON mode; "[]" is written before the return.
	err := Run(&buf, "nonexistent_process_xyz_12345", Options{ListOnly: true, OutputFormat: output.FormatJSON})
	var silent *cmderr.SilentError
	if !errors.As(err, &silent) || silent.Code != 1 {
		t.Fatalf("expected SilentExit(1) for no match, got: %v", err)
	}

	out := strings.TrimSpace(buf.String())
	if out != "[]" {
		t.Errorf("expected empty JSON array for no matches, got: %s", out)
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
				// Valid signal + no match → SilentExit(1) (classified no-match behavior)
				var silent *cmderr.SilentError
				if err != nil && !errors.As(err, &silent) {
					t.Errorf("expected nil or SilentExit for signal %q, got: %v", tt.signal, err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error for invalid signal %q", tt.signal)
				}
				// Must NOT be a SilentExit — invalid signals should be classified ErrInvalidInput
				var silent *cmderr.SilentError
				if errors.As(err, &silent) {
					t.Errorf("invalid signal %q should not produce SilentExit, got code %d", tt.signal, silent.Code)
				}
			}
		})
	}
}

func TestPatternCompilation(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		exact       bool
		ignoreCase  bool
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
