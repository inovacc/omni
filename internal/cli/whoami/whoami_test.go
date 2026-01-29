package whoami

import (
	"bytes"
	"os/user"
	"strings"
	"testing"
)

func TestRunWhoami(t *testing.T) {
	t.Run("returns current user", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunWhoami(&buf, WhoamiOptions{})
		if err != nil {
			t.Fatalf("RunWhoami() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunWhoami() should return a username")
		}
	})

	t.Run("returns non-empty string", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunWhoami(&buf, WhoamiOptions{})
		if err != nil {
			t.Fatalf("RunWhoami() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if len(output) == 0 {
			t.Error("RunWhoami() should return non-empty username")
		}
	})

	t.Run("matches os/user", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunWhoami(&buf, WhoamiOptions{})
		if err != nil {
			t.Fatalf("RunWhoami() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())

		// Get the expected username from os/user
		currentUser, err := user.Current()
		if err != nil {
			t.Skip("Could not get current user for comparison")
		}

		// On Windows, Username may include domain (DOMAIN\user)
		expected := currentUser.Username
		if strings.Contains(expected, "\\") {
			parts := strings.Split(expected, "\\")
			expected = parts[len(parts)-1]
		}

		// The output should match or be contained in the expected
		if output != expected && output != currentUser.Username {
			t.Errorf("RunWhoami() = %v, want %v or %v", output, expected, currentUser.Username)
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunWhoami(&buf, WhoamiOptions{})
		if err != nil {
			t.Fatalf("RunWhoami() error = %v", err)
		}

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunWhoami() output should end with newline")
		}
	})

	t.Run("single line output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunWhoami(&buf, WhoamiOptions{})
		if err != nil {
			t.Fatalf("RunWhoami() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Errorf("RunWhoami() should output exactly one line, got %d", len(lines))
		}
	})

	t.Run("no special characters", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunWhoami(&buf, WhoamiOptions{})
		if err != nil {
			t.Fatalf("RunWhoami() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())

		// Username should not contain control characters
		for _, r := range output {
			if r < 32 && r != '\n' {
				t.Errorf("RunWhoami() contains control character: %d", r)
			}
		}
	})

	t.Run("no leading whitespace", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunWhoami(&buf, WhoamiOptions{})

		output := buf.String()
		if len(output) > 0 && (output[0] == ' ' || output[0] == '\t') {
			t.Error("RunWhoami() should not have leading whitespace")
		}
	})

	t.Run("no trailing whitespace except newline", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunWhoami(&buf, WhoamiOptions{})

		output := buf.String()
		trimmed := strings.TrimRight(output, "\n")

		if strings.HasSuffix(trimmed, " ") || strings.HasSuffix(trimmed, "\t") {
			t.Error("RunWhoami() should not have trailing whitespace except newline")
		}
	})

	t.Run("consistent output", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer

		err := RunWhoami(&buf1, WhoamiOptions{})
		if err != nil {
			t.Fatalf("RunWhoami() error = %v", err)
		}

		err = RunWhoami(&buf2, WhoamiOptions{})
		if err != nil {
			t.Fatalf("RunWhoami() second call error = %v", err)
		}

		if buf1.String() != buf2.String() {
			t.Errorf("RunWhoami() inconsistent: %v vs %v", buf1.String(), buf2.String())
		}
	})

	t.Run("multiple calls consistent", func(t *testing.T) {
		results := make([]string, 5)

		for i := range 5 {
			var buf bytes.Buffer

			_ = RunWhoami(&buf, WhoamiOptions{})
			results[i] = buf.String()
		}

		for i := 1; i < 5; i++ {
			if results[i] != results[0] {
				t.Errorf("RunWhoami() call %d differs: %v vs %v", i, results[i], results[0])
			}
		}
	})

	t.Run("no error", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunWhoami(&buf, WhoamiOptions{})
		if err != nil {
			t.Errorf("RunWhoami() should not error in normal conditions: %v", err)
		}
	})

	t.Run("reasonable length", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunWhoami(&buf, WhoamiOptions{})

		output := strings.TrimSpace(buf.String())
		// Username should be reasonable length (1-256 characters)
		if len(output) < 1 || len(output) > 256 {
			t.Errorf("RunWhoami() username length = %d, seems unreasonable", len(output))
		}
	})

	t.Run("printable characters", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunWhoami(&buf, WhoamiOptions{})

		output := strings.TrimSpace(buf.String())

		for _, r := range output {
			// Allow alphanumeric, underscore, hyphen, and backslash (for Windows domain)
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') || r == '_' || r == '-' ||
				r == '\\' || r == '.' || r == '@' || r == ' ') {
				t.Logf("RunWhoami() contains character: %q (may be valid)", r)
			}
		}
	})
}
