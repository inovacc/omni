package kill

import (
	"bytes"
	"strings"
	"testing"
)

func TestListSignals(t *testing.T) {
	var buf bytes.Buffer
	listSignals(&buf)

	output := buf.String()

	// Should contain common signals
	if !strings.Contains(output, "INT") {
		t.Error("listSignals() missing INT signal")
	}

	if !strings.Contains(output, "KILL") {
		t.Error("listSignals() missing KILL signal")
	}

	if !strings.Contains(output, "TERM") {
		t.Error("listSignals() missing TERM signal")
	}
}

func TestRunKillList(t *testing.T) {
	var buf bytes.Buffer

	err := RunKill(&buf, []string{}, KillOptions{List: true})
	if err != nil {
		t.Fatalf("RunKill() with List error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "SIG") {
		t.Errorf("RunKill() list output missing signals: %v", output)
	}
}

func TestRunKillNoArgs(t *testing.T) {
	var buf bytes.Buffer

	err := RunKill(&buf, []string{}, KillOptions{})
	if err == nil {
		t.Error("RunKill() expected error with no arguments")
	}
}

func TestRunKillInvalidPID(t *testing.T) {
	var buf bytes.Buffer

	err := RunKill(&buf, []string{"notanumber"}, KillOptions{})
	// Should not panic, should handle gracefully
	if err == nil {
		t.Log("RunKill() handled invalid PID gracefully")
	}
}

func TestRunKillInvalidSignal(t *testing.T) {
	var buf bytes.Buffer

	err := RunKill(&buf, []string{"12345"}, KillOptions{Signal: "INVALID"})
	if err == nil {
		t.Error("RunKill() expected error with invalid signal")
	}
}

func TestDefaultSignal(t *testing.T) {
	sig := defaultSignal()
	// SIGTERM should be 15
	if sig != 15 {
		t.Errorf("defaultSignal() = %d, want 15 (SIGTERM)", sig)
	}
}

func TestSignalMap(t *testing.T) {
	// Check common signals exist in the map
	signals := []string{"INT", "KILL", "TERM"}
	for _, name := range signals {
		if _, ok := signalMap[name]; !ok {
			t.Errorf("signalMap missing %s", name)
		}
	}

	// Check numeric aliases
	numericAliases := []string{"2", "9", "15"}
	for _, num := range numericAliases {
		if _, ok := signalMap[num]; !ok {
			t.Errorf("signalMap missing numeric alias %s", num)
		}
	}
}
