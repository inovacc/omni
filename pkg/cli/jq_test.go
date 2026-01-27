package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunJq(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jq_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("simple query", func(t *testing.T) {
		file := filepath.Join(tmpDir, "data.json")
		content := `{"name": "John", "age": 30}`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		// args[0] is filter, args[1:] are files
		err := RunJq(&buf, []string{".name", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != `"John"` {
			t.Errorf("RunJq() = %v, want '\"John\"'", result)
		}
	})

	t.Run("raw output", func(t *testing.T) {
		file := filepath.Join(tmpDir, "raw.json")
		content := `{"message": "hello"}`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		err := RunJq(&buf, []string{".message", file}, JqOptions{Raw: true})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "hello" {
			t.Errorf("RunJq() = %v, want 'hello'", result)
		}
	})

	t.Run("array access", func(t *testing.T) {
		file := filepath.Join(tmpDir, "array.json")
		content := `{"items": ["a", "b", "c"]}`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		err := RunJq(&buf, []string{".items[0]", file}, JqOptions{Raw: true})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "a" {
			t.Errorf("RunJq() = %v, want 'a'", result)
		}
	})

	t.Run("nested object", func(t *testing.T) {
		file := filepath.Join(tmpDir, "nested.json")
		content := `{"user": {"name": "Jane", "address": {"city": "NYC"}}}`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		err := RunJq(&buf, []string{".user.address.city", file}, JqOptions{Raw: true})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "NYC" {
			t.Errorf("RunJq() = %v, want 'NYC'", result)
		}
	})

	t.Run("identity query", func(t *testing.T) {
		file := filepath.Join(tmpDir, "identity.json")
		content := `{"key": "value"}`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		err := RunJq(&buf, []string{".", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		if !strings.Contains(buf.String(), "key") {
			t.Errorf("RunJq() missing key in output: %v", buf.String())
		}
	})

	t.Run("compact output", func(t *testing.T) {
		file := filepath.Join(tmpDir, "compact.json")
		content := `{"a": 1, "b": 2}`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		err := RunJq(&buf, []string{".", file}, JqOptions{Compact: true})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		// Compact should not have newlines within the object
		if strings.Count(result, "\n") > 1 {
			t.Errorf("RunJq() compact output has newlines: %v", result)
		}
	})
}
