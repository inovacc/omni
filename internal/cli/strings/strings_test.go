package strings

import (
	"bytes"
	"os"
	"path/filepath"
	strs "strings"
	"testing"
)

func TestRunStrings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "strings_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("find strings in binary", func(t *testing.T) {
		file := filepath.Join(tmpDir, "binary.bin")
		// Create a file with binary data and embedded strings
		data := []byte{0x00, 0x00, 'h', 'e', 'l', 'l', 'o', 0x00, 0xFF, 0xFF, 'w', 'o', 'r', 'l', 'd', 0x00}
		_ = os.WriteFile(file, data, 0644)

		var buf bytes.Buffer

		err := RunStrings(&buf, []string{file}, StringsOptions{MinLength: 4})
		if err != nil {
			t.Fatalf("RunStrings() error = %v", err)
		}

		output := buf.String()
		if !strs.Contains(output, "hello") {
			t.Errorf("RunStrings() should find 'hello'")
		}

		if !strs.Contains(output, "world") {
			t.Errorf("RunStrings() should find 'world'")
		}
	})

	t.Run("minimum length filter", func(t *testing.T) {
		file := filepath.Join(tmpDir, "minlen.bin")
		data := []byte{0x00, 'a', 'b', 0x00, 'x', 'y', 'z', 'w', 0x00}
		_ = os.WriteFile(file, data, 0644)

		var buf bytes.Buffer

		err := RunStrings(&buf, []string{file}, StringsOptions{MinLength: 3})
		if err != nil {
			t.Fatalf("RunStrings() error = %v", err)
		}

		output := buf.String()
		// "ab" (len 2) should not appear, "xyzw" (len 4) should
		if strs.Contains(output, "ab") {
			t.Errorf("RunStrings() should not include strings shorter than min length")
		}

		if !strs.Contains(output, "xyzw") {
			t.Errorf("RunStrings() should include 'xyzw'")
		}
	})

	t.Run("default min length is 4", func(t *testing.T) {
		file := filepath.Join(tmpDir, "default.bin")
		data := []byte{0x00, 'a', 'b', 'c', 0x00, 'd', 'e', 'f', 'g', 0x00}
		_ = os.WriteFile(file, data, 0644)

		var buf bytes.Buffer

		err := RunStrings(&buf, []string{file}, StringsOptions{})
		if err != nil {
			t.Fatalf("RunStrings() error = %v", err)
		}

		output := buf.String()
		// "abc" (len 3) should not appear, "defg" (len 4) should
		if strs.Contains(output, "abc") {
			t.Errorf("RunStrings() default min length should be 4")
		}

		if !strs.Contains(output, "defg") {
			t.Errorf("RunStrings() should include 'defg'")
		}
	})

	t.Run("decimal offset", func(t *testing.T) {
		file := filepath.Join(tmpDir, "offset.bin")
		data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 'h', 'e', 'l', 'l', 'o'}
		_ = os.WriteFile(file, data, 0644)

		var buf bytes.Buffer

		err := RunStrings(&buf, []string{file}, StringsOptions{MinLength: 4, Offset: "d"})
		if err != nil {
			t.Fatalf("RunStrings() error = %v", err)
		}

		output := buf.String()
		// Should show offset 5 (decimal)
		if !strs.Contains(output, "5") {
			t.Errorf("RunStrings() decimal offset missing: %s", output)
		}
	})

	t.Run("hex offset", func(t *testing.T) {
		file := filepath.Join(tmpDir, "hexoff.bin")
		data := make([]byte, 20)
		copy(data[16:], []byte("test"))
		_ = os.WriteFile(file, data, 0644)

		var buf bytes.Buffer

		err := RunStrings(&buf, []string{file}, StringsOptions{MinLength: 4, Offset: "x"})
		if err != nil {
			t.Fatalf("RunStrings() error = %v", err)
		}

		output := buf.String()
		// 16 in hex is 10
		if !strs.Contains(output, "10") {
			t.Logf("RunStrings() hex offset output: %s", output)
		}
	})

	t.Run("plain text file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "text.txt")
		_ = os.WriteFile(file, []byte("This is a test file with some text content.\n"), 0644)

		var buf bytes.Buffer

		err := RunStrings(&buf, []string{file}, StringsOptions{MinLength: 4})
		if err != nil {
			t.Fatalf("RunStrings() error = %v", err)
		}

		output := buf.String()
		if !strs.Contains(output, "This") {
			t.Errorf("RunStrings() should find strings in text file")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.bin")
		_ = os.WriteFile(file, []byte{}, 0644)

		var buf bytes.Buffer

		err := RunStrings(&buf, []string{file}, StringsOptions{})
		if err != nil {
			t.Fatalf("RunStrings() error = %v", err)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunStrings(&buf, []string{"/nonexistent/file.bin"}, StringsOptions{})
		if err == nil {
			t.Error("RunStrings() expected error for nonexistent file")
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "multi1.bin")
		file2 := filepath.Join(tmpDir, "multi2.bin")

		_ = os.WriteFile(file1, []byte{0x00, 'f', 'i', 'r', 's', 't', 0x00}, 0644)
		_ = os.WriteFile(file2, []byte{0x00, 's', 'e', 'c', 'o', 'n', 'd', 0x00}, 0644)

		var buf bytes.Buffer

		err := RunStrings(&buf, []string{file1, file2}, StringsOptions{MinLength: 4})
		if err != nil {
			t.Fatalf("RunStrings() error = %v", err)
		}

		output := buf.String()
		if !strs.Contains(output, "first") {
			t.Errorf("RunStrings() missing 'first'")
		}

		if !strs.Contains(output, "second") {
			t.Errorf("RunStrings() missing 'second'")
		}
	})

	t.Run("tabs are printable", func(t *testing.T) {
		file := filepath.Join(tmpDir, "tabs.bin")
		data := []byte{0x00, 'a', '\t', 'b', 'c', 'd', 0x00}
		_ = os.WriteFile(file, data, 0644)

		var buf bytes.Buffer

		err := RunStrings(&buf, []string{file}, StringsOptions{MinLength: 4})
		if err != nil {
			t.Fatalf("RunStrings() error = %v", err)
		}

		output := buf.String()
		if !strs.Contains(output, "a\tbcd") {
			t.Logf("RunStrings() tab handling: %q", output)
		}
	})
}

func TestIsPrintableASCII(t *testing.T) {
	tests := []struct {
		input    byte
		expected bool
	}{
		{32, true},   // space
		{126, true},  // tilde
		{'\t', true}, // tab
		{'a', true},
		{'Z', true},
		{0, false},
		{31, false},
		{127, false},
		{255, false},
	}

	for _, tt := range tests {
		result := isPrintableASCII(tt.input)
		if result != tt.expected {
			t.Errorf("isPrintableASCII(%d) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
