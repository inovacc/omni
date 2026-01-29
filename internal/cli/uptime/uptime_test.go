package uptime

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRunUptime(t *testing.T) {
	t.Run("default output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUptime(&buf, UptimeOptions{})
		if err != nil {
			t.Fatalf("RunUptime() error = %v", err)
		}

		output := buf.String()
		// Should contain time and "up"
		if !strings.Contains(output, "up") {
			t.Errorf("RunUptime() output should contain 'up': %s", output)
		}

		if !strings.Contains(output, "load average") {
			t.Errorf("RunUptime() output should contain 'load average': %s", output)
		}
	})

	t.Run("pretty format", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUptime(&buf, UptimeOptions{Pretty: true})
		if err != nil {
			t.Fatalf("RunUptime() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "up") {
			t.Errorf("RunUptime() pretty output should contain 'up': %s", output)
		}
	})

	t.Run("since format", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUptime(&buf, UptimeOptions{Since: true})
		if err != nil {
			t.Fatalf("RunUptime() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		// Should be a date in format 2006-01-02 15:04:05
		if len(output) < 10 {
			t.Errorf("RunUptime() since output too short: %s", output)
		}
	})
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"1 hour", time.Hour, "1:00"},
		{"2 hours 30 min", 2*time.Hour + 30*time.Minute, "2:30"},
		{"1 day", 24 * time.Hour, "1 day(s), 0:00"},
		{"1 day 5 hours", 29 * time.Hour, "1 day(s), 5:00"},
		{"2 days", 48 * time.Hour, "2 day(s), 0:00"},
		{"5 minutes", 5 * time.Minute, "0:05"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUptime(tt.duration)
			if result != tt.want {
				t.Errorf("formatUptime(%v) = %q, want %q", tt.duration, result, tt.want)
			}
		})
	}
}

func TestFormatPrettyUptime(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		contains []string
	}{
		{"1 minute", time.Minute, []string{"1 minute"}},
		{"5 minutes", 5 * time.Minute, []string{"5 minutes"}},
		{"1 hour", time.Hour, []string{"1 hour"}},
		{"2 hours", 2 * time.Hour, []string{"2 hours"}},
		{"1 day", 24 * time.Hour, []string{"1 day"}},
		{"2 days", 48 * time.Hour, []string{"2 days"}},
		{"1 day 2 hours", 26 * time.Hour, []string{"1 day", "2 hours"}},
		{"1 hour 30 min", time.Hour + 30*time.Minute, []string{"1 hour", "30 minutes"}},
		{"complex", 49*time.Hour + 30*time.Minute, []string{"2 days", "1 hour", "30 minutes"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPrettyUptime(tt.duration)
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("formatPrettyUptime(%v) = %q, should contain %q", tt.duration, result, s)
				}
			}
		})
	}
}

func TestGetUptime(t *testing.T) {
	uptime, err := GetUptime()
	if err != nil {
		t.Fatalf("GetUptime() error = %v", err)
	}

	// Uptime should be positive
	if uptime <= 0 {
		t.Errorf("GetUptime() = %v, want positive duration", uptime)
	}
}
