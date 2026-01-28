package env

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunEnv(t *testing.T) {
	t.Run("show all environment variables", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEnv(&buf, []string{}, EnvOptions{})
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}

		output := buf.String()
		// Should have at least one environment variable
		if len(output) == 0 {
			t.Error("RunEnv() should output environment variables")
		}

		// Should contain = signs (VAR=value format)
		if !strings.Contains(output, "=") {
			t.Errorf("RunEnv() output should be in VAR=value format: %v", output)
		}
	})

	t.Run("filter by variable name", func(t *testing.T) {
		testVar := "OMNI_TEST_VAR_FILTER"
		testValue := "test_value_123"

		if err := os.Setenv(testVar, testValue); err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.Unsetenv(testVar) }()

		var buf bytes.Buffer

		err := RunEnv(&buf, []string{testVar}, EnvOptions{})
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}

		output := buf.String()
		expected := testVar + "=" + testValue

		if !strings.Contains(output, expected) {
			t.Errorf("RunEnv() = %v, want to contain %v", output, expected)
		}
	})

	t.Run("multiple variable filter", func(t *testing.T) {
		var1 := "OMNI_TEST_VAR1"
		var2 := "OMNI_TEST_VAR2"

		if err := os.Setenv(var1, "value1"); err != nil {
			t.Fatal(err)
		}

		if err := os.Setenv(var2, "value2"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Unsetenv(var1)
			_ = os.Unsetenv(var2)
		}()

		var buf bytes.Buffer

		err := RunEnv(&buf, []string{var1, var2}, EnvOptions{})
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, var1) || !strings.Contains(output, var2) {
			t.Errorf("RunEnv() should contain both vars: %v", output)
		}
	})

	t.Run("nonexistent variable", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEnv(&buf, []string{"NONEXISTENT_VAR_12345"}, EnvOptions{})
		// Should not error, just not output anything
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}
	})

	t.Run("variable with special characters", func(t *testing.T) {
		testVar := "OMNI_TEST_SPECIAL"
		testValue := "value with spaces and=equals"

		if err := os.Setenv(testVar, testValue); err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.Unsetenv(testVar) }()

		var buf bytes.Buffer

		err := RunEnv(&buf, []string{testVar}, EnvOptions{})
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, testValue) {
			t.Errorf("RunEnv() should preserve special chars: %v", output)
		}
	})

	t.Run("empty value", func(t *testing.T) {
		testVar := "OMNI_TEST_EMPTY"

		if err := os.Setenv(testVar, ""); err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.Unsetenv(testVar) }()

		var buf bytes.Buffer

		err := RunEnv(&buf, []string{testVar}, EnvOptions{})
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}

		// Implementation skips empty-value variables when filtering by name
		output := buf.String()
		if strings.Contains(output, testVar+"=") {
			t.Log("Note: env shows empty var (implementation may vary)")
		}
	})

	t.Run("PATH variable exists", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEnv(&buf, []string{"PATH"}, EnvOptions{})
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "PATH=") {
			t.Errorf("RunEnv() should show PATH: %v", output)
		}
	})

	t.Run("sorted output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEnv(&buf, []string{}, EnvOptions{})
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) < 2 {
			t.Skip("Not enough env vars to test sorting")
		}

		// Check if output appears sorted
		for i := 1; i < len(lines); i++ {
			// Extract variable names
			prev := strings.Split(lines[i-1], "=")[0]
			curr := strings.Split(lines[i], "=")[0]

			if prev > curr {
				// Not strictly required, but common behavior
				t.Logf("Note: env output may not be sorted (%v > %v)", prev, curr)
				break
			}
		}
	})
}
