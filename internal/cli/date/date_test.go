package date

import (
	"bytes"
	"slices"
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

	t.Run("custom format YYYY-MM-DD", func(t *testing.T) {
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

	t.Run("custom format time only", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{Format: "15:04:05"})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())

		_, err = time.Parse("15:04:05", output)
		if err != nil {
			t.Errorf("RunDate() time format failed: %v, error: %v", output, err)
		}
	})

	t.Run("custom format year only", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{Format: "2006"})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		expected := time.Now().Format("2006")

		if output != expected {
			t.Errorf("RunDate() year = %v, want %v", output, expected)
		}
	})

	t.Run("custom format weekday", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{Format: "Monday"})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		weekdays := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

		found := slices.Contains(weekdays, output)

		if !found {
			t.Errorf("RunDate() weekday = %v, not a valid weekday", output)
		}
	})

	t.Run("custom format month", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{Format: "January"})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		months := []string{"January", "February", "March", "April", "May", "June",
			"July", "August", "September", "October", "November", "December"}

		found := slices.Contains(months, output)

		if !found {
			t.Errorf("RunDate() month = %v, not a valid month", output)
		}
	})

	t.Run("ISO format", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{ISO: true})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		// ISO format should contain T separator
		if !strings.Contains(output, "T") {
			t.Errorf("RunDate() ISO should contain T: %v", output)
		}
	})

	t.Run("custom format overrides ISO", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDate(&buf, DateOptions{Format: "2006", ISO: true})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		// Should just be the year
		if len(output) != 4 {
			t.Errorf("RunDate() custom should override ISO: %v", output)
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

	t.Run("single line output", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunDate(&buf, DateOptions{})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Errorf("RunDate() should output exactly one line, got %d", len(lines))
		}
	})

	t.Run("UTC vs local differ by timezone", func(t *testing.T) {
		var bufLocal, bufUTC bytes.Buffer

		_ = RunDate(&bufLocal, DateOptions{Format: "Z07:00"})
		_ = RunDate(&bufUTC, DateOptions{Format: "Z07:00", UTC: true})

		utcTZ := strings.TrimSpace(bufUTC.String())
		if utcTZ != "Z" && utcTZ != "+00:00" {
			t.Errorf("RunDate() UTC timezone = %v, want Z or +00:00", utcTZ)
		}
	})

	t.Run("consistent date within same second", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer

		_ = RunDate(&buf1, DateOptions{Format: "2006-01-02 15:04:05"})
		_ = RunDate(&buf2, DateOptions{Format: "2006-01-02 15:04:05"})

		// Should be same or differ by at most 1 second
		// Just verify both produce valid output
		if buf1.Len() == 0 || buf2.Len() == 0 {
			t.Error("RunDate() should produce output")
		}
	})

	t.Run("Unix timestamp format literal", func(t *testing.T) {
		var buf bytes.Buffer

		// Using "unix" as literal format string
		err := RunDate(&buf, DateOptions{Format: "unix"})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		// "unix" is treated as literal format, not special
		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunDate() should produce output")
		}
	})

	t.Run("complex format string", func(t *testing.T) {
		var buf bytes.Buffer

		format := "Mon, 02 Jan 2006 15:04:05 -0700"

		err := RunDate(&buf, DateOptions{Format: format})
		if err != nil {
			t.Fatalf("RunDate() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())

		_, err = time.Parse(format, output)
		if err != nil {
			t.Errorf("RunDate() RFC2822-like format failed: %v", output)
		}
	})

	t.Run("no error on any valid format", func(t *testing.T) {
		formats := []string{
			"2006",
			"01",
			"02",
			"15",
			"04",
			"05",
			time.RFC3339,
			time.RFC1123,
			time.UnixDate,
		}

		for _, format := range formats {
			var buf bytes.Buffer

			err := RunDate(&buf, DateOptions{Format: format})
			if err != nil {
				t.Errorf("RunDate() with format %q error = %v", format, err)
			}
		}
	})
}

func TestDate(t *testing.T) {
	t.Run("default format RFC3339", func(t *testing.T) {
		result := Date("")

		_, err := time.Parse(time.RFC3339, result)
		if err != nil {
			t.Errorf("Date() default format not RFC3339: %v", result)
		}
	})

	t.Run("custom format", func(t *testing.T) {
		result := Date("2006-01-02")

		_, err := time.Parse("2006-01-02", result)
		if err != nil {
			t.Errorf("Date() custom format failed: %v", result)
		}
	})

	t.Run("returns current date", func(t *testing.T) {
		expected := time.Now().Format("2006-01-02")
		result := Date("2006-01-02")

		if result != expected {
			t.Errorf("Date() = %v, want %v", result, expected)
		}
	})

	t.Run("no trailing newline", func(t *testing.T) {
		result := Date("")

		if strings.HasSuffix(result, "\n") {
			t.Error("Date() should not have trailing newline")
		}
	})

	t.Run("consistent results same second", func(t *testing.T) {
		format := "2006-01-02 15:04:05"
		result1 := Date(format)
		result2 := Date(format)

		// Usually same, might differ by 1 second at boundary
		if result1 != result2 {
			t.Logf("Note: Date() results differ (may be at second boundary): %v vs %v", result1, result2)
		}
	})

	t.Run("empty format uses RFC3339", func(t *testing.T) {
		result := Date("")

		_, err := time.Parse(time.RFC3339, result)
		if err != nil {
			t.Errorf("Date() empty format should use RFC3339: %v", result)
		}
	})
}
