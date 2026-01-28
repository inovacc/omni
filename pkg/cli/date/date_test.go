package date

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRunDate(t *testing.T) {
	t.Run("default format is RFC3339", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		_, err = time.Parse(time.RFC3339, output)

		if err != nil {
			t.Errorf("RunDate() output not RFC3339: %v, error: %v", output, err)
		}
	})

	t.Run("returns current time", func(t *testing.T) {
		before := time.Now()

		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		after := time.Now()
		output := strings.TrimSpace(buf.String())
		parsed, _ := time.Parse(time.RFC3339, output)

		if parsed.Before(before.Add(-time.Second)) || parsed.After(after.Add(time.Second)) {
			t.Errorf("RunDate() time out of range: got %v, expected between %v and %v", parsed, before, after)
		}
	})

	t.Run("UTC option", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{UTC: true})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())

		// UTC times should end with Z or +00:00
		if !strings.HasSuffix(output, "Z") && !strings.Contains(output, "+00:00") {
			t.Errorf("RunDate() with UTC should be in UTC timezone: %v", output)
		}
	})

	t.Run("custom format", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{Format: "2006-01-02"})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		_, err = time.Parse("2006-01-02", output)

		if err != nil {
			t.Errorf("RunDate() custom format failed: %v, error: %v", output, err)
		}
	})

	t.Run("unix timestamp format", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{Format: "unix"})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())

		// Unix timestamp should be a number
		if len(output) < 10 {
			t.Errorf("RunDate() unix format should return timestamp: %v", output)
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunDate() output should end with newline")
		}
	})
}
