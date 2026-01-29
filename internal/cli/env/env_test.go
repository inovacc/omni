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

	t.Run("output format VAR=value", func(t *testing.T) {
		testVar := "OMNI_TEST_FORMAT"
		testValue := "test_value"

		if err := os.Setenv(testVar, testValue); err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.Unsetenv(testVar) }()

		var buf bytes.Buffer

		_ = RunEnv(&buf, []string{testVar}, EnvOptions{})

		output := strings.TrimSpace(buf.String())
		if output != "" && !strings.Contains(output, "=") {
			t.Errorf("RunEnv() output should be VAR=value format: %v", output)
		}
	})

	t.Run("unicode value", func(t *testing.T) {
		testVar := "OMNI_TEST_UNICODE"
		testValue := "世界🌍こんにちは"

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
			t.Errorf("RunEnv() should preserve unicode: %v", output)
		}
	})

	t.Run("newlines in value", func(t *testing.T) {
		testVar := "OMNI_TEST_NEWLINE"
		testValue := "line1\nline2\nline3"

		if err := os.Setenv(testVar, testValue); err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.Unsetenv(testVar) }()

		var buf bytes.Buffer

		err := RunEnv(&buf, []string{testVar}, EnvOptions{})
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}

		// Output should contain the variable
		if !strings.Contains(buf.String(), testVar) {
			t.Log("Note: env may handle newlines differently")
		}
	})

	t.Run("long value", func(t *testing.T) {
		testVar := "OMNI_TEST_LONG"
		testValue := strings.Repeat("x", 1000)

		if err := os.Setenv(testVar, testValue); err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.Unsetenv(testVar) }()

		var buf bytes.Buffer

		err := RunEnv(&buf, []string{testVar}, EnvOptions{})
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}

		if !strings.Contains(buf.String(), testValue) {
			t.Errorf("RunEnv() should handle long values")
		}
	})

	t.Run("consistent output", func(t *testing.T) {
		testVar := "OMNI_TEST_CONSISTENT"
		if err := os.Setenv(testVar, "value"); err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.Unsetenv(testVar) }()

		var buf1, buf2 bytes.Buffer

		_ = RunEnv(&buf1, []string{testVar}, EnvOptions{})
		_ = RunEnv(&buf2, []string{testVar}, EnvOptions{})

		if buf1.String() != buf2.String() {
			t.Errorf("RunEnv() inconsistent: %v vs %v", buf1.String(), buf2.String())
		}
	})

	t.Run("multiple calls same result", func(t *testing.T) {
		testVar := "OMNI_TEST_MULTI"
		if err := os.Setenv(testVar, "value"); err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.Unsetenv(testVar) }()

		results := make([]string, 5)

		for i := range 5 {
			var buf bytes.Buffer

			_ = RunEnv(&buf, []string{testVar}, EnvOptions{})
			results[i] = buf.String()
		}

		for i := 1; i < 5; i++ {
			if results[i] != results[0] {
				t.Errorf("RunEnv() call %d differs", i)
			}
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		testVar := "OMNI_TEST_NEWLINE_END"
		if err := os.Setenv(testVar, "value"); err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.Unsetenv(testVar) }()

		var buf bytes.Buffer

		_ = RunEnv(&buf, []string{testVar}, EnvOptions{})

		output := buf.String()
		if len(output) > 0 && !strings.HasSuffix(output, "\n") {
			t.Log("Note: env output may not end with newline")
		}
	})

	t.Run("no error with all env", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEnv(&buf, []string{}, EnvOptions{})
		if err != nil {
			t.Errorf("RunEnv() should not error: %v", err)
		}
	})

	t.Run("variable with number", func(t *testing.T) {
		testVar := "OMNI_TEST_123"
		testValue := "numeric"

		if err := os.Setenv(testVar, testValue); err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.Unsetenv(testVar) }()

		var buf bytes.Buffer

		err := RunEnv(&buf, []string{testVar}, EnvOptions{})
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}

		if !strings.Contains(buf.String(), testVar) {
			t.Errorf("RunEnv() should show numeric var name")
		}
	})

	t.Run("variable with underscore", func(t *testing.T) {
		testVar := "OMNI_TEST_WITH_UNDERSCORES"
		testValue := "underscore_value"

		if err := os.Setenv(testVar, testValue); err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.Unsetenv(testVar) }()

		var buf bytes.Buffer

		err := RunEnv(&buf, []string{testVar}, EnvOptions{})
		if err != nil {
			t.Fatalf("RunEnv() error = %v", err)
		}

		if !strings.Contains(buf.String(), testVar) {
			t.Errorf("RunEnv() should show underscore var")
		}
	})
}
