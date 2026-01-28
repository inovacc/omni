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

		err := RunWhoami(&buf)
		if err != nil {
			t.Fatalf("RunWhoami() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunWhoami() should return a username")
		}
	})

	t.Run("matches os/user", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunWhoami(&buf)
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

		err := RunWhoami(&buf)
		if err != nil {
			t.Fatalf("RunWhoami() error = %v", err)
		}

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunWhoami() output should end with newline")
		}
	})

	t.Run("no special characters", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunWhoami(&buf)
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

	t.Run("consistent output", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer

		err := RunWhoami(&buf1)
		if err != nil {
			t.Fatalf("RunWhoami() error = %v", err)
		}

		err = RunWhoami(&buf2)
		if err != nil {
			t.Fatalf("RunWhoami() second call error = %v", err)
		}

		if buf1.String() != buf2.String() {
			t.Errorf("RunWhoami() inconsistent: %v vs %v", buf1.String(), buf2.String())
		}
	})
}
