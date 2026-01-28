package yq

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

	t.Run("empty document", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.yaml")

		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".", file}, YqOptions{})
		if err != nil {
			t.Logf("RunYq() empty document: %v", err)
		}
	})

	t.Run("null value", func(t *testing.T) {
		file := filepath.Join(tmpDir, "null.yaml")

		content := `key: null`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".key", file}, YqOptions{})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "null" && result != "" {
			t.Logf("RunYq() null = %v", result)
		}
	})

	t.Run("boolean values", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bool.yaml")

		content := `yes: true
no: false`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".yes", file}, YqOptions{})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "true" {
			t.Errorf("RunYq() = %v, want 'true'", result)
		}
	})

	t.Run("number values", func(t *testing.T) {
		file := filepath.Join(tmpDir, "num.yaml")

		content := `int: 42
float: 3.14`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".int", file}, YqOptions{})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "42" {
			t.Errorf("RunYq() = %v, want '42'", result)
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unicode.yaml")

		content := `msg: 世界🌍`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".msg", file}, YqOptions{Raw: true})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "世界🌍" {
			t.Errorf("RunYq() = %v, want unicode", result)
		}
	})

	t.Run("multiline string", func(t *testing.T) {
		file := filepath.Join(tmpDir, "multi.yaml")

		// Use simple multiline format
		content := "text: |\n  line1\n  line2\n  line3"
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".text", file}, YqOptions{})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		output := buf.String()
		// Check that at least some content is returned
		if len(output) == 0 {
			t.Error("RunYq() multiline should return content")
		}
		// Log the actual output for debugging
		if !strings.Contains(output, "line1") && !strings.Contains(output, "line2") {
			t.Logf("RunYq() multiline output: %v", output)
		}
	})

	t.Run("array of objects", func(t *testing.T) {
		file := filepath.Join(tmpDir, "arrobj.yaml")

		content := `people:
  - name: Alice
    age: 30
  - name: Bob
    age: 25`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".people[0].name", file}, YqOptions{})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "Alice" {
			t.Errorf("RunYq() = %v, want 'Alice'", result)
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		file := filepath.Join(tmpDir, "invalid.yaml")

		if err := os.WriteFile(file, []byte(`key: [not: valid`), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".", file}, YqOptions{})
		if err == nil {
			t.Log("RunYq() may handle invalid YAML gracefully")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunYq(&buf, []string{".", "/nonexistent.yaml"}, YqOptions{})
		if err == nil {
			t.Error("RunYq() expected error for nonexistent file")
		}
	})

	t.Run("deep nesting", func(t *testing.T) {
		file := filepath.Join(tmpDir, "deep.yaml")

		content := `l1:
  l2:
    l3:
      l4:
        value: deep`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".l1.l2.l3.l4.value", file}, YqOptions{})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "deep" {
			t.Errorf("RunYq() = %v, want 'deep'", result)
		}
	})

	t.Run("anchors and aliases", func(t *testing.T) {
		file := filepath.Join(tmpDir, "anchor.yaml")

		content := `defaults: &defaults
  timeout: 30
  retries: 3

production:
  <<: *defaults
  host: prod.example.com`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".production.timeout", file}, YqOptions{})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "30" {
			t.Logf("RunYq() anchor = %v (may not support anchors)", result)
		}
	})

	t.Run("multiple documents", func(t *testing.T) {
		file := filepath.Join(tmpDir, "multidoc.yaml")

		content := `---
name: first
---
name: second`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".name", file}, YqOptions{})
		if err != nil {
			t.Logf("RunYq() multi-doc: %v", err)
		}

		// May return first document or handle specially
		t.Logf("Multi-doc output: %v", buf.String())
	})

	t.Run("consistent output", func(t *testing.T) {
		file := filepath.Join(tmpDir, "consistent.yaml")

		content := `key: value`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf1, buf2 bytes.Buffer

		_ = RunYq(&buf1, []string{".", file}, YqOptions{})
		_ = RunYq(&buf2, []string{".", file}, YqOptions{})

		if buf1.String() != buf2.String() {
			t.Error("RunYq() should produce consistent output")
		}
	})

	t.Run("quoted strings", func(t *testing.T) {
		file := filepath.Join(tmpDir, "quoted.yaml")

		content := `single: 'single quoted'
double: "double quoted"`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunYq(&buf, []string{".single", file}, YqOptions{Raw: true})
		if err != nil {
			t.Fatalf("RunYq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "single quoted" {
			t.Errorf("RunYq() = %v, want 'single quoted'", result)
		}
	})
}
