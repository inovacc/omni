package comm

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunComm(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "comm_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("basic comparison", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\nc\n"), 0644)
		_ = os.WriteFile(file2, []byte("b\nc\nd\n"), 0644)

		var buf bytes.Buffer

		err := RunComm(&buf, []string{file1, file2}, CommOptions{})
		if err != nil {
			t.Fatalf("RunComm() error = %v", err)
		}

		output := buf.String()
		// Should show: a (col1), b (col3), c (col3), d (col2)
		if !strings.Contains(output, "a") {
			t.Errorf("RunComm() missing unique line from file1")
		}
	})

	t.Run("suppress column 1", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "s1_file1.txt")
		file2 := filepath.Join(tmpDir, "s1_file2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\n"), 0644)
		_ = os.WriteFile(file2, []byte("b\nc\n"), 0644)

		var buf bytes.Buffer

		err := RunComm(&buf, []string{file1, file2}, CommOptions{Suppress1: true})
		if err != nil {
			t.Fatalf("RunComm() error = %v", err)
		}

		output := buf.String()
		// 'a' should not appear (unique to file1)
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "a" {
				t.Errorf("RunComm() -1 should not show 'a'")
			}
		}
	})

	t.Run("suppress column 2", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "s2_file1.txt")
		file2 := filepath.Join(tmpDir, "s2_file2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\n"), 0644)
		_ = os.WriteFile(file2, []byte("b\nc\n"), 0644)

		var buf bytes.Buffer

		err := RunComm(&buf, []string{file1, file2}, CommOptions{Suppress2: true})
		if err != nil {
			t.Fatalf("RunComm() error = %v", err)
		}

		output := buf.String()
		// 'c' should not appear with tab prefix (unique to file2)
		if strings.Contains(output, "\tc") {
			t.Errorf("RunComm() -2 should not show lines unique to file2")
		}
	})

	t.Run("suppress column 3", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "s3_file1.txt")
		file2 := filepath.Join(tmpDir, "s3_file2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\n"), 0644)
		_ = os.WriteFile(file2, []byte("b\nc\n"), 0644)

		var buf bytes.Buffer

		err := RunComm(&buf, []string{file1, file2}, CommOptions{Suppress3: true})
		if err != nil {
			t.Fatalf("RunComm() error = %v", err)
		}

		output := buf.String()
		// 'b' should not appear (common to both)
		if strings.Contains(output, "\t\tb") {
			t.Errorf("RunComm() -3 should not show common lines")
		}
	})

	t.Run("show only common lines", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "common_file1.txt")
		file2 := filepath.Join(tmpDir, "common_file2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\nc\n"), 0644)
		_ = os.WriteFile(file2, []byte("b\nc\nd\n"), 0644)

		var buf bytes.Buffer

		err := RunComm(&buf, []string{file1, file2}, CommOptions{Suppress1: true, Suppress2: true})
		if err != nil {
			t.Fatalf("RunComm() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		lines := strings.Split(output, "\n")
		// Should only show b and c
		if len(lines) != 2 {
			t.Errorf("RunComm() -12 got %d lines, want 2", len(lines))
		}
	})

	t.Run("custom output delimiter", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "delim_file1.txt")
		file2 := filepath.Join(tmpDir, "delim_file2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\n"), 0644)
		_ = os.WriteFile(file2, []byte("b\nc\n"), 0644)

		var buf bytes.Buffer

		err := RunComm(&buf, []string{file1, file2}, CommOptions{OutputDelim: ":"})
		if err != nil {
			t.Fatalf("RunComm() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, ":") {
			t.Errorf("RunComm() should use custom delimiter")
		}
	})

	t.Run("check sorted order", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "unsorted1.txt")
		file2 := filepath.Join(tmpDir, "sorted2.txt")

		_ = os.WriteFile(file1, []byte("b\na\n"), 0644) // unsorted
		_ = os.WriteFile(file2, []byte("a\nb\n"), 0644)

		var buf bytes.Buffer

		err := RunComm(&buf, []string{file1, file2}, CommOptions{CheckOrder: true})
		if err == nil {
			t.Error("RunComm() expected error for unsorted file")
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunComm(&buf, []string{}, CommOptions{})
		if err == nil {
			t.Error("RunComm() expected error for missing operand")
		}
	})

	t.Run("both stdin", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunComm(&buf, []string{"-", "-"}, CommOptions{})
		if err == nil {
			t.Error("RunComm() expected error when both files are stdin")
		}
	})

	t.Run("empty files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "empty1.txt")
		file2 := filepath.Join(tmpDir, "empty2.txt")

		_ = os.WriteFile(file1, []byte(""), 0644)
		_ = os.WriteFile(file2, []byte(""), 0644)

		var buf bytes.Buffer

		err := RunComm(&buf, []string{file1, file2}, CommOptions{})
		if err != nil {
			t.Fatalf("RunComm() error = %v", err)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "exists.txt")
		_ = os.WriteFile(file1, []byte("a\n"), 0644)

		var buf bytes.Buffer

		err := RunComm(&buf, []string{file1, "/nonexistent/file.txt"}, CommOptions{})
		if err == nil {
			t.Error("RunComm() expected error for nonexistent file")
		}
	})
}

func TestSplitFunc(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		delim   byte
		atEOF   bool
		wantAdv int
		wantTok string
		wantNil bool
	}{
		{"newline delim", []byte("hello\nworld"), '\n', false, 6, "hello", false},
		{"null delim", []byte("hello\x00world"), '\x00', false, 6, "hello", false},
		{"at eof", []byte("hello"), '\n', true, 5, "hello", false},
		{"empty at eof", []byte{}, '\n', true, 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := splitFunc(tt.delim)
			adv, tok, _ := fn(tt.data, tt.atEOF)

			if adv != tt.wantAdv {
				t.Errorf("advance = %d, want %d", adv, tt.wantAdv)
			}

			if tt.wantNil {
				if tok != nil {
					t.Errorf("token = %q, want nil", tok)
				}
			} else if string(tok) != tt.wantTok {
				t.Errorf("token = %q, want %q", tok, tt.wantTok)
			}
		})
	}
}
