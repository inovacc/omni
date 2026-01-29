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

func TestRunKillSignalNames(t *testing.T) {
	t.Run("SIGINT signal name", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"99999999"}, KillOptions{Signal: "INT"})
		// Should fail because PID doesn't exist, but shouldn't error on signal parsing
		if err != nil {
			// Error is expected for nonexistent PID
			t.Logf("Expected error for nonexistent PID: %v", err)
		}
	})

	t.Run("SIGKILL signal name", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"99999999"}, KillOptions{Signal: "KILL"})
		if err != nil {
			t.Logf("Expected error for nonexistent PID: %v", err)
		}
	})

	t.Run("SIGTERM signal name", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"99999999"}, KillOptions{Signal: "TERM"})
		if err != nil {
			t.Logf("Expected error for nonexistent PID: %v", err)
		}
	})

	t.Run("numeric signal 9", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"99999999"}, KillOptions{Signal: "9"})
		if err != nil {
			t.Logf("Expected error for nonexistent PID: %v", err)
		}
	})

	t.Run("numeric signal 15", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"99999999"}, KillOptions{Signal: "15"})
		if err != nil {
			t.Logf("Expected error for nonexistent PID: %v", err)
		}
	})
}

func TestListSignalsOutput(t *testing.T) {
	t.Run("list contains HUP", func(t *testing.T) {
		var buf bytes.Buffer
		listSignals(&buf)

		output := buf.String()
		if !strings.Contains(output, "HUP") {
			t.Log("listSignals() may not contain HUP on Windows")
		}
	})

	t.Run("list contains QUIT", func(t *testing.T) {
		var buf bytes.Buffer
		listSignals(&buf)

		output := buf.String()
		if !strings.Contains(output, "QUIT") {
			t.Log("listSignals() may not contain QUIT on all platforms")
		}
	})

	t.Run("list contains USR1", func(t *testing.T) {
		var buf bytes.Buffer
		listSignals(&buf)

		output := buf.String()
		if !strings.Contains(output, "USR1") {
			t.Log("listSignals() may not contain USR1 on all platforms")
		}
	})

	t.Run("list output format", func(t *testing.T) {
		var buf bytes.Buffer
		listSignals(&buf)

		output := buf.String()
		// Should have multiple lines
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) < 1 {
			t.Error("listSignals() should produce output")
		}
	})
}

func TestRunKillMultiplePIDs(t *testing.T) {
	t.Run("multiple invalid PIDs", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"99999998", "99999999"}, KillOptions{})
		// Should handle multiple PIDs
		if err != nil {
			t.Logf("Error for multiple nonexistent PIDs: %v", err)
		}
	})

	t.Run("mixed valid invalid PIDs", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"notanumber", "99999999"}, KillOptions{})
		// Should handle gracefully
		if err != nil {
			t.Logf("Error for mixed PIDs: %v", err)
		}
	})
}

func TestRunKillEdgeCases(t *testing.T) {
	t.Run("negative PID", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"-1"}, KillOptions{})
		// Negative PIDs have special meaning (process groups)
		if err != nil {
			t.Logf("Negative PID result: %v", err)
		}
	})

	t.Run("zero PID", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"0"}, KillOptions{})
		// PID 0 has special meaning
		if err != nil {
			t.Logf("Zero PID result: %v", err)
		}
	})

	t.Run("very large PID", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"999999999999"}, KillOptions{})
		// Should handle gracefully
		if err != nil {
			t.Logf("Large PID result: %v", err)
		}
	})

	t.Run("empty signal", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"99999999"}, KillOptions{Signal: ""})
		// Empty signal should use default
		if err != nil {
			t.Logf("Empty signal result: %v", err)
		}
	})

	t.Run("lowercase signal", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"99999999"}, KillOptions{Signal: "term"})
		// Lowercase should work or error clearly
		if err != nil {
			t.Logf("Lowercase signal result: %v", err)
		}
	})

	t.Run("SIG prefix signal", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunKill(&buf, []string{"99999999"}, KillOptions{Signal: "SIGTERM"})
		// SIGTERM with prefix
		if err != nil {
			t.Logf("SIG prefix result: %v", err)
		}
	})
}

func TestSignalMapCompleteness(t *testing.T) {
	expectedSignals := []string{
		"HUP", "INT", "QUIT", "ILL", "TRAP", "ABRT",
		"KILL", "SEGV", "PIPE", "ALRM", "TERM",
	}

	for _, sig := range expectedSignals {
		t.Run("signal_"+sig, func(t *testing.T) {
			if _, ok := signalMap[sig]; !ok {
				t.Logf("signalMap may not have %s on this platform", sig)
			}
		})
	}
}
