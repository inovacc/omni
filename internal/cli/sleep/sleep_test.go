package sleep

import (
	"testing"
	"time"
)

func TestParseSleepDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "seconds default",
			input:    "1",
			expected: 1 * time.Second,
		},
		{
			name:     "seconds explicit",
			input:    "2s",
			expected: 2 * time.Second,
		},
		{
			name:     "minutes",
			input:    "1m",
			expected: 1 * time.Minute,
		},
		{
			name:     "hours",
			input:    "1h",
			expected: 1 * time.Hour,
		},
		{
			name:     "days",
			input:    "1d",
			expected: 24 * time.Hour,
		},
		{
			name:     "decimal seconds",
			input:    "0.5",
			expected: 500 * time.Millisecond,
		},
		{
			name:     "decimal with suffix",
			input:    "1.5s",
			expected: 1500 * time.Millisecond,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid",
			input:   "abc",
			wantErr: true,
		},
		{
			name:     "large number",
			input:    "100",
			expected: 100 * time.Second,
		},
		{
			name:     "small decimal",
			input:    "0.001",
			expected: 1 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSleepDuration(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseSleepDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseSleepDuration(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRunSleep(t *testing.T) {
	t.Run("no arguments", func(t *testing.T) {
		err := RunSleep([]string{})
		if err == nil {
			t.Error("RunSleep() expected error with no arguments")
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		err := RunSleep([]string{"invalid"})
		if err == nil {
			t.Error("RunSleep() expected error with invalid argument")
		}
	})

	t.Run("very short sleep", func(t *testing.T) {
		start := time.Now()

		err := RunSleep([]string{"0.001"})
		if err != nil {
			t.Fatalf("RunSleep() error = %v", err)
		}

		elapsed := time.Since(start)
		if elapsed < 1*time.Millisecond {
			t.Errorf("RunSleep() did not wait long enough: %v", elapsed)
		}
	})

	t.Run("multiple arguments", func(t *testing.T) {
		start := time.Now()

		err := RunSleep([]string{"0.001", "0.001"})
		if err != nil {
			t.Fatalf("RunSleep() error = %v", err)
		}

		elapsed := time.Since(start)
		if elapsed < 2*time.Millisecond {
			t.Errorf("RunSleep() should add durations: %v", elapsed)
		}
	})

	t.Run("mixed valid invalid", func(t *testing.T) {
		err := RunSleep([]string{"0.001", "invalid"})
		if err == nil {
			t.Error("RunSleep() expected error with invalid argument")
		}
	})
}
