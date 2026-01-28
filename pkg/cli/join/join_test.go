package join

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunJoin(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "join_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("basic join", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		_ = os.WriteFile(file1, []byte("1 Alice\n2 Bob\n3 Carol\n"), 0644)
		_ = os.WriteFile(file2, []byte("1 100\n2 200\n3 300\n"), 0644)

		var buf bytes.Buffer

		err := RunJoin(&buf, []string{file1, file2}, JoinOptions{})
		if err != nil {
			t.Fatalf("RunJoin() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "1 Alice 100") {
			t.Errorf("RunJoin() missing expected joined line, got: %s", output)
		}
	})

	t.Run("join on different fields", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "f1.txt")
		file2 := filepath.Join(tmpDir, "f2.txt")

		_ = os.WriteFile(file1, []byte("Alice 1\nBob 2\n"), 0644)
		_ = os.WriteFile(file2, []byte("1 100\n2 200\n"), 0644)

		var buf bytes.Buffer

		err := RunJoin(&buf, []string{file1, file2}, JoinOptions{Field1: 2, Field2: 1})
		if err != nil {
			t.Fatalf("RunJoin() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Alice") && !strings.Contains(output, "100") {
			t.Errorf("RunJoin() incorrect output: %s", output)
		}
	})

	t.Run("custom separator", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "sep1.txt")
		file2 := filepath.Join(tmpDir, "sep2.txt")

		_ = os.WriteFile(file1, []byte("1,Alice\n2,Bob\n"), 0644)
		_ = os.WriteFile(file2, []byte("1,100\n2,200\n"), 0644)

		var buf bytes.Buffer

		err := RunJoin(&buf, []string{file1, file2}, JoinOptions{Separator: ","})
		if err != nil {
			t.Fatalf("RunJoin() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, ",") {
			t.Errorf("RunJoin() should use comma separator: %s", output)
		}
	})

	t.Run("ignore case", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "case1.txt")
		file2 := filepath.Join(tmpDir, "case2.txt")

		_ = os.WriteFile(file1, []byte("A data1\nB data2\n"), 0644)
		_ = os.WriteFile(file2, []byte("a value1\nb value2\n"), 0644)

		var buf bytes.Buffer

		err := RunJoin(&buf, []string{file1, file2}, JoinOptions{IgnoreCase: true})
		if err != nil {
			t.Fatalf("RunJoin() error = %v", err)
		}

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) < 2 {
			t.Errorf("RunJoin() with ignore case should match, got: %s", output)
		}
	})

	t.Run("print unpairable from file1", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "u1.txt")
		file2 := filepath.Join(tmpDir, "u2.txt")

		_ = os.WriteFile(file1, []byte("1 data1\n2 data2\n3 data3\n"), 0644)
		_ = os.WriteFile(file2, []byte("1 value1\n"), 0644)

		var buf bytes.Buffer

		err := RunJoin(&buf, []string{file1, file2}, JoinOptions{Unpaired1: true})
		if err != nil {
			t.Fatalf("RunJoin() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "data2") || !strings.Contains(output, "data3") {
			t.Errorf("RunJoin() -a1 should show unpairable lines: %s", output)
		}
	})

	t.Run("print unpairable from file2", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "ua1.txt")
		file2 := filepath.Join(tmpDir, "ua2.txt")

		_ = os.WriteFile(file1, []byte("1 data1\n"), 0644)
		_ = os.WriteFile(file2, []byte("1 value1\n2 value2\n"), 0644)

		var buf bytes.Buffer

		err := RunJoin(&buf, []string{file1, file2}, JoinOptions{Unpaired2: true})
		if err != nil {
			t.Fatalf("RunJoin() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "value2") {
			t.Errorf("RunJoin() -a2 should show unpairable lines: %s", output)
		}
	})

	t.Run("only unpairable from file1", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "v1.txt")
		file2 := filepath.Join(tmpDir, "v2.txt")

		_ = os.WriteFile(file1, []byte("1 data1\n2 data2\n"), 0644)
		_ = os.WriteFile(file2, []byte("1 value1\n"), 0644)

		var buf bytes.Buffer

		err := RunJoin(&buf, []string{file1, file2}, JoinOptions{OnlyUnpaired1: true})
		if err != nil {
			t.Fatalf("RunJoin() error = %v", err)
		}

		output := buf.String()
		// Should not contain the matched line
		if strings.Contains(output, "value1") {
			t.Errorf("RunJoin() -v1 should not show joined lines: %s", output)
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunJoin(&buf, []string{}, JoinOptions{})
		if err == nil {
			t.Error("RunJoin() expected error for missing operand")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "exists.txt")
		_ = os.WriteFile(file1, []byte("1 data\n"), 0644)

		var buf bytes.Buffer

		err := RunJoin(&buf, []string{file1, "/nonexistent/file.txt"}, JoinOptions{})
		if err == nil {
			t.Error("RunJoin() expected error for nonexistent file")
		}
	})

	t.Run("empty files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "empty1.txt")
		file2 := filepath.Join(tmpDir, "empty2.txt")

		_ = os.WriteFile(file1, []byte(""), 0644)
		_ = os.WriteFile(file2, []byte(""), 0644)

		var buf bytes.Buffer

		err := RunJoin(&buf, []string{file1, file2}, JoinOptions{})
		if err != nil {
			t.Fatalf("RunJoin() error = %v", err)
		}
	})
}

func TestJoinLineKey(t *testing.T) {
	tests := []struct {
		name     string
		fields   []string
		fieldIdx int
		expected string
	}{
		{"first field", []string{"a", "b", "c"}, 0, "a"},
		{"middle field", []string{"a", "b", "c"}, 1, "b"},
		{"last field", []string{"a", "b", "c"}, 2, "c"},
		{"out of bounds", []string{"a", "b"}, 5, ""},
		{"negative index", []string{"a", "b"}, -1, ""},
		{"empty fields", []string{}, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jl := joinLine{fields: tt.fields}
			result := jl.key(tt.fieldIdx)
			if result != tt.expected {
				t.Errorf("key(%d) = %q, want %q", tt.fieldIdx, result, tt.expected)
			}
		})
	}
}
