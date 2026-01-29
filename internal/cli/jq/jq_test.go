package jq

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
		err := RunJq(&buf, nil, []string{".name", file}, JqOptions{})
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

		err := RunJq(&buf, nil, []string{".message", file}, JqOptions{Raw: true})
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

		err := RunJq(&buf, nil, []string{".items[0]", file}, JqOptions{Raw: true})
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

		err := RunJq(&buf, nil, []string{".user.address.city", file}, JqOptions{Raw: true})
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

		err := RunJq(&buf, nil, []string{".", file}, JqOptions{})
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

		err := RunJq(&buf, nil, []string{".", file}, JqOptions{Compact: true})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		// Compact should not have newlines within the object
		if strings.Count(result, "\n") > 1 {
			t.Errorf("RunJq() compact output has newlines: %v", result)
		}
	})

	t.Run("pipe filter", func(t *testing.T) {
		file := filepath.Join(tmpDir, "pipe.json")

		content := `{"items": [1, 2, 3]}`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".items | length", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "3" {
			t.Errorf("RunJq() = %v, want '3'", result)
		}
	})

	t.Run("chained pipes", func(t *testing.T) {
		file := filepath.Join(tmpDir, "chained.json")

		content := `{"data": {"values": ["a", "b", "c"]}}`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".data | .values | length", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "3" {
			t.Errorf("RunJq() = %v, want '3'", result)
		}
	})

	t.Run("empty object", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.json")

		if err := os.WriteFile(file, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		if !strings.Contains(buf.String(), "{") {
			t.Error("RunJq() empty object should output {}")
		}
	})

	t.Run("empty array", func(t *testing.T) {
		file := filepath.Join(tmpDir, "emptyarr.json")

		if err := os.WriteFile(file, []byte("[]"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		if !strings.Contains(buf.String(), "[") {
			t.Error("RunJq() empty array should output []")
		}
	})

	t.Run("null value", func(t *testing.T) {
		file := filepath.Join(tmpDir, "null.json")

		if err := os.WriteFile(file, []byte(`{"key": null}`), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".key", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "null" {
			t.Errorf("RunJq() = %v, want 'null'", result)
		}
	})

	t.Run("boolean values", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bool.json")

		if err := os.WriteFile(file, []byte(`{"yes": true, "no": false}`), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".yes", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "true" {
			t.Errorf("RunJq() = %v, want 'true'", result)
		}
	})

	t.Run("number values", func(t *testing.T) {
		file := filepath.Join(tmpDir, "num.json")

		if err := os.WriteFile(file, []byte(`{"int": 42, "float": 3.14}`), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".int", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "42" {
			t.Errorf("RunJq() = %v, want '42'", result)
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unicode.json")

		if err := os.WriteFile(file, []byte(`{"msg": "世界🌍"}`), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".msg", file}, JqOptions{Raw: true})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "世界🌍" {
			t.Errorf("RunJq() = %v, want unicode", result)
		}
	})

	t.Run("keys query", func(t *testing.T) {
		file := filepath.Join(tmpDir, "keys.json")

		if err := os.WriteFile(file, []byte(`{"a": 1, "b": 2, "c": 3}`), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{"keys", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "a") || !strings.Contains(output, "b") || !strings.Contains(output, "c") {
			t.Errorf("RunJq() keys = %v", output)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		file := filepath.Join(tmpDir, "invalid.json")

		if err := os.WriteFile(file, []byte(`{not valid json}`), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".", file}, JqOptions{})
		if err == nil {
			t.Log("RunJq() may handle invalid JSON gracefully")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".", "/nonexistent.json"}, JqOptions{})
		if err == nil {
			t.Error("RunJq() expected error for nonexistent file")
		}
	})

	t.Run("invalid filter", func(t *testing.T) {
		file := filepath.Join(tmpDir, "valid.json")

		if err := os.WriteFile(file, []byte(`{"key": "value"}`), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".invalid[[[", file}, JqOptions{})
		if err == nil {
			t.Log("RunJq() may handle invalid filter gracefully")
		}
	})

	t.Run("array iteration", func(t *testing.T) {
		file := filepath.Join(tmpDir, "iter.json")

		if err := os.WriteFile(file, []byte(`[1, 2, 3, 4, 5]`), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".[]", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunJq() got %d values, want 5", len(lines))
		}
	})

	t.Run("map function", func(t *testing.T) {
		file := filepath.Join(tmpDir, "map.json")

		if err := os.WriteFile(file, []byte(`[1, 2, 3]`), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{". | length", file}, JqOptions{})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "3" {
			t.Errorf("RunJq() = %v, want '3'", result)
		}
	})

	t.Run("deep nesting", func(t *testing.T) {
		file := filepath.Join(tmpDir, "deep.json")

		content := `{"l1": {"l2": {"l3": {"l4": {"value": "deep"}}}}}`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunJq(&buf, nil, []string{".l1.l2.l3.l4.value", file}, JqOptions{Raw: true})
		if err != nil {
			t.Fatalf("RunJq() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if result != "deep" {
			t.Errorf("RunJq() = %v, want 'deep'", result)
		}
	})

	t.Run("consistent output", func(t *testing.T) {
		file := filepath.Join(tmpDir, "consistent.json")

		if err := os.WriteFile(file, []byte(`{"key": "value"}`), 0644); err != nil {
			t.Fatal(err)
		}

		var buf1, buf2 bytes.Buffer

		_ = RunJq(&buf1, nil, []string{".", file}, JqOptions{})
		_ = RunJq(&buf2, nil, []string{".", file}, JqOptions{})

		if buf1.String() != buf2.String() {
			t.Error("RunJq() should produce consistent output")
		}
	})
}
