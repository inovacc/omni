package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunYq(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "yq_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("simple query", func(t *testing.T) {
		file := filepath.Join(tmpDir, "config.yaml")

		content := `name: John
age: 30`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		// args[0] is filter, args[1:] are files
		err := RunYq(&buf, []string{".name", file}, YqOptions{})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "John" {
			t.Errorf("RunYq() = %v, want 'John'", result)
		}
	})

	t.Run("nested object", func(t *testing.T) {
		file := filepath.Join(tmpDir, "nested.yaml")

		content := `user:
  name: Jane
  address:
    city: NYC`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".user.address.city", file}, YqOptions{})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "NYC" {
			t.Errorf("RunYq() = %v, want 'NYC'", result)
		}
	})

	t.Run("array access", func(t *testing.T) {
		file := filepath.Join(tmpDir, "array.yaml")

		content := `items:
  - apple
  - banana
  - cherry`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".items[0]", file}, YqOptions{})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "apple" {
			t.Errorf("RunYq() = %v, want 'apple'", result)
		}
	})

	t.Run("identity query", func(t *testing.T) {
		file := filepath.Join(tmpDir, "identity.yaml")

		content := `key: value
another: data`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".", file}, YqOptions{})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "key") || !strings.Contains(output, "value") {
			t.Errorf("RunYq() missing data in output: %v", output)
		}
	})

	t.Run("json output", func(t *testing.T) {
		file := filepath.Join(tmpDir, "tojson.yaml")

		content := `name: test
value: 123`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".", file}, YqOptions{OutputJSON: true})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
			t.Errorf("RunYq() JSON output invalid: %v", output)
		}
	})

	t.Run("raw output", func(t *testing.T) {
		file := filepath.Join(tmpDir, "raw.yaml")

		content := `message: hello world`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".message", file}, YqOptions{Raw: true})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "hello world" {
			t.Errorf("RunYq() = %v, want 'hello world'", result)
		}
	})
}
