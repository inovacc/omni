package timecmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRunTime(t *testing.T) {
	t.Run("successful function", func(t *testing.T) {
		var buf bytes.Buffer

		result, err := RunTime(&buf, func() error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Fatalf("RunTime() error = %v", err)
		}

		if result.ExitCode != 0 {
			t.Errorf("RunTime() exitCode = %d, want 0", result.ExitCode)
		}

		if result.Real < 10*time.Millisecond {
			t.Errorf("RunTime() real = %v, want >= 10ms", result.Real)
		}

		output := buf.String()
		if !strings.Contains(output, "real") {
			t.Errorf("RunTime() should output 'real': %s", output)
		}

		if !strings.Contains(output, "user") {
			t.Errorf("RunTime() should output 'user': %s", output)
		}

		if !strings.Contains(output, "sys") {
			t.Errorf("RunTime() should output 'sys': %s", output)
		}
	})

	t.Run("failing function", func(t *testing.T) {
		var buf bytes.Buffer

		testErr := errors.New("test error")
		result, err := RunTime(&buf, func() error {
			return testErr
		})

		if err != testErr {
			t.Fatalf("RunTime() error = %v, want %v", err, testErr)
		}

		if result.ExitCode != 1 {
			t.Errorf("RunTime() exitCode = %d, want 1", result.ExitCode)
		}
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0, "0m0.000s"},
		{500 * time.Millisecond, "0m0.500s"},
		{1 * time.Second, "0m1.000s"},
		{1*time.Minute + 30*time.Second, "1m30.000s"},
		{2*time.Minute + 15*time.Second + 500*time.Millisecond, "2m15.500s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.expected)
			}
		})
	}
}

func TestStopwatch(t *testing.T) {
	t.Run("basic usage", func(t *testing.T) {
		sw := NewStopwatch()

		time.Sleep(10 * time.Millisecond)

		lap1 := sw.Lap()

		if lap1 < 10*time.Millisecond {
			t.Errorf("Stopwatch.Lap() = %v, want >= 10ms", lap1)
		}

		time.Sleep(10 * time.Millisecond)

		elapsed := sw.Elapsed()

		if elapsed < 20*time.Millisecond {
			t.Errorf("Stopwatch.Elapsed() = %v, want >= 20ms", elapsed)
		}

		laps := sw.Laps()
		if len(laps) != 1 {
			t.Errorf("Stopwatch.Laps() length = %d, want 1", len(laps))
		}
	})

	t.Run("multiple laps", func(t *testing.T) {
		sw := NewStopwatch()

		time.Sleep(5 * time.Millisecond)
		sw.Lap()

		time.Sleep(5 * time.Millisecond)
		sw.Lap()

		time.Sleep(5 * time.Millisecond)
		sw.Lap()

		laps := sw.Laps()
		if len(laps) != 3 {
			t.Errorf("Stopwatch.Laps() length = %d, want 3", len(laps))
		}

		// Each lap should be greater than the previous
		for i := 1; i < len(laps); i++ {
			if laps[i] <= laps[i-1] {
				t.Errorf("Stopwatch.Laps()[%d] = %v, should be > %v", i, laps[i], laps[i-1])
			}
		}
	})

	t.Run("reset", func(t *testing.T) {
		sw := NewStopwatch()

		time.Sleep(10 * time.Millisecond)
		sw.Lap()

		sw.Reset()

		laps := sw.Laps()
		if len(laps) != 0 {
			t.Errorf("Stopwatch.Laps() after reset = %d, want 0", len(laps))
		}

		// Elapsed should be near zero after reset
		elapsed := sw.Elapsed()
		if elapsed > 5*time.Millisecond {
			t.Errorf("Stopwatch.Elapsed() after reset = %v, want < 5ms", elapsed)
		}
	})
}

func TestSleep(t *testing.T) {
	start := time.Now()

	Sleep(10 * time.Millisecond)

	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("Sleep() elapsed = %v, want >= 10ms", elapsed)
	}
}

func TestSleepSeconds(t *testing.T) {
	start := time.Now()

	SleepSeconds(0.01) // 10ms

	elapsed := time.Since(start)

	if elapsed < 10*time.Millisecond {
		t.Errorf("SleepSeconds() elapsed = %v, want >= 10ms", elapsed)
	}
}
