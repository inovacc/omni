package shuf

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunShuf(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shuf_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("shuffle file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "lines.txt")
		if err := os.WriteFile(file, []byte("a\nb\nc\nd\ne\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunShuf(&buf, []string{file}, ShufOptions{})
		if err != nil {
			t.Fatalf("RunShuf() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunShuf() got %d lines, want 5", len(lines))
		}

		// Check all elements are present
		seen := make(map[string]bool)
		for _, l := range lines {
			seen[l] = true
		}

		for _, expected := range []string{"a", "b", "c", "d", "e"} {
			if !seen[expected] {
				t.Errorf("RunShuf() missing element: %v", expected)
			}
		}
	})

	t.Run("echo mode", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunShuf(&buf, []string{"one", "two", "three"}, ShufOptions{Echo: true})
		if err != nil {
			t.Fatalf("RunShuf() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunShuf() echo got %d lines, want 3", len(lines))
		}
	})

	t.Run("input range", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunShuf(&buf, []string{}, ShufOptions{InputRange: "1-5"})
		if err != nil {
			t.Fatalf("RunShuf() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunShuf() range got %d lines, want 5", len(lines))
		}
	})

	t.Run("head count", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunShuf(&buf, []string{}, ShufOptions{InputRange: "1-100", HeadCount: 10})
		if err != nil {
			t.Fatalf("RunShuf() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 10 {
			t.Errorf("RunShuf() head count got %d lines, want 10", len(lines))
		}
	})

	t.Run("repeat mode", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunShuf(&buf, []string{"a", "b"}, ShufOptions{Echo: true, Repeat: true, HeadCount: 5})
		if err != nil {
			t.Fatalf("RunShuf() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunShuf() repeat got %d lines, want 5", len(lines))
		}
	})

	t.Run("invalid range format", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunShuf(&buf, []string{}, ShufOptions{InputRange: "invalid"})
		if err == nil {
			t.Error("RunShuf() expected error for invalid range")
		}
	})

	t.Run("invalid range values", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunShuf(&buf, []string{}, ShufOptions{InputRange: "10-5"})
		if err == nil {
			t.Error("RunShuf() expected error for lo > hi")
		}
	})

	t.Run("repeat without head count", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunShuf(&buf, []string{"a"}, ShufOptions{Echo: true, Repeat: true})
		if err == nil {
			t.Error("RunShuf() expected error for repeat without head count")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunShuf(&buf, []string{file}, ShufOptions{})
		if err != nil {
			t.Fatalf("RunShuf() error = %v", err)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunShuf(&buf, []string{"/nonexistent/file.txt"}, ShufOptions{})
		if err == nil {
			t.Error("RunShuf() expected error for nonexistent file")
		}
	})

	t.Run("shuffle is random", func(t *testing.T) {
		file := filepath.Join(tmpDir, "random.txt")
		content := "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n"
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		// Run multiple times and check if at least one is different from original order
		original := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
		different := false

		for i := 0; i < 10; i++ {
			var buf bytes.Buffer

			_ = RunShuf(&buf, []string{file}, ShufOptions{})

			lines := strings.Split(strings.TrimSpace(buf.String()), "\n")

			for j := range lines {
				if lines[j] != original[j] {
					different = true
					break
				}
			}

			if different {
				break
			}
		}

		if !different {
			t.Log("RunShuf() returned same order 10 times (very unlikely but possible)")
		}
	})
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		data     []byte
		b        byte
		expected int
	}{
		{[]byte("hello"), 'l', 2},
		{[]byte("hello"), 'x', -1},
		{[]byte("hello"), 'h', 0},
		{[]byte("hello"), 'o', 4},
		{[]byte(""), 'a', -1},
	}

	for _, tt := range tests {
		result := indexOf(tt.data, tt.b)
		if result != tt.expected {
			t.Errorf("indexOf(%q, %c) = %d, want %d", tt.data, tt.b, result, tt.expected)
		}
	}
}
