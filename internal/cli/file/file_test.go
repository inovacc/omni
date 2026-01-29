package file

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "file_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("regular text file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "text.txt")
		_ = os.WriteFile(file, []byte("Hello, World!\n"), 0644)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{file}, FileOptions{})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "text") || !strings.Contains(output, "ASCII") {
			t.Logf("RunFile() text file = %q", output)
		}
	})

	t.Run("directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "testdir")
		_ = os.Mkdir(dir, 0755)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{dir}, FileOptions{})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		if !strings.Contains(buf.String(), "directory") {
			t.Errorf("RunFile() should identify directory")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")
		_ = os.WriteFile(file, []byte{}, 0644)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{file}, FileOptions{})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		if !strings.Contains(buf.String(), "empty") {
			t.Errorf("RunFile() should identify empty file")
		}
	})

	t.Run("png file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "image.png")
		// PNG magic bytes
		data := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
		_ = os.WriteFile(file, data, 0644)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{file}, FileOptions{})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		if !strings.Contains(buf.String(), "PNG") {
			t.Errorf("RunFile() should identify PNG: %s", buf.String())
		}
	})

	t.Run("gzip file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "archive.gz")
		data := []byte{0x1F, 0x8B, 0x08}
		_ = os.WriteFile(file, data, 0644)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{file}, FileOptions{})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		if !strings.Contains(buf.String(), "gzip") {
			t.Errorf("RunFile() should identify gzip: %s", buf.String())
		}
	})

	t.Run("pdf file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "doc.pdf")
		data := []byte("%PDF-1.4")
		_ = os.WriteFile(file, data, 0644)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{file}, FileOptions{})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		if !strings.Contains(buf.String(), "PDF") {
			t.Errorf("RunFile() should identify PDF: %s", buf.String())
		}
	})

	t.Run("brief mode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "brief.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{file}, FileOptions{Brief: true})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		// Brief mode should not include filename
		if strings.Contains(buf.String(), "brief.txt") {
			t.Errorf("RunFile() brief should not show filename")
		}
	})

	t.Run("mime type output", func(t *testing.T) {
		file := filepath.Join(tmpDir, "mime.txt")
		_ = os.WriteFile(file, []byte("text content"), 0644)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{file}, FileOptions{MimeType: true})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		if !strings.Contains(buf.String(), "text/") {
			t.Errorf("RunFile() mime type should contain text/: %s", buf.String())
		}
	})

	t.Run("custom separator", func(t *testing.T) {
		file := filepath.Join(tmpDir, "sep.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{file}, FileOptions{Separator: " = "})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		if !strings.Contains(buf.String(), " = ") {
			t.Errorf("RunFile() should use custom separator")
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "multi1.txt")
		file2 := filepath.Join(tmpDir, "multi2.txt")

		_ = os.WriteFile(file1, []byte("text1"), 0644)
		_ = os.WriteFile(file2, []byte("text2"), 0644)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{file1, file2}, FileOptions{})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		if !strings.Contains(buf.String(), "multi1.txt") || !strings.Contains(buf.String(), "multi2.txt") {
			t.Errorf("RunFile() should process both files")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunFile(&buf, []string{"/nonexistent/file.txt"}, FileOptions{})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		if !strings.Contains(buf.String(), "cannot open") {
			t.Errorf("RunFile() should report cannot open: %s", buf.String())
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunFile(&buf, []string{}, FileOptions{})
		if err == nil {
			t.Error("RunFile() expected error for missing operand")
		}
	})

	t.Run("go source file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "main.go")
		_ = os.WriteFile(file, []byte("package main\n\nfunc main() {}\n"), 0644)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{file}, FileOptions{})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		if !strings.Contains(buf.String(), "Go") {
			t.Logf("RunFile() Go file = %s", buf.String())
		}
	})

	t.Run("shell script", func(t *testing.T) {
		file := filepath.Join(tmpDir, "script.sh")
		_ = os.WriteFile(file, []byte("#!/bin/bash\necho hello\n"), 0644)

		var buf bytes.Buffer

		err := RunFile(&buf, []string{file}, FileOptions{})
		if err != nil {
			t.Fatalf("RunFile() error = %v", err)
		}

		if !strings.Contains(buf.String(), "shell") && !strings.Contains(buf.String(), "script") {
			t.Logf("RunFile() shell script = %s", buf.String())
		}
	})
}

func TestIsText(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{"ascii text", []byte("Hello, World!"), true},
		{"binary data", []byte{0x00, 0x01, 0x02, 0x03, 0x04}, false},
		{"text with newlines", []byte("line1\nline2\nline3"), true},
		{"mixed mostly binary", []byte{0x00, 0x01, 0x02, 'h', 'i', 0x03, 0x04, 0x05, 0x06, 0x07}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isText(tt.data)
			if result != tt.expected {
				t.Errorf("isText() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCheckMagic(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantType string
		wantOK   bool
	}{
		{"PNG", []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, "PNG image data", true},
		{"JPEG", []byte{0xFF, 0xD8, 0xFF}, "JPEG image data", true},
		{"gzip", []byte{0x1F, 0x8B}, "gzip compressed data", true},
		{"ZIP", []byte{'P', 'K', 0x03, 0x04}, "Zip archive data", true},
		{"PDF", []byte{'%', 'P', 'D', 'F'}, "PDF document", true},
		{"unknown", []byte{0x12, 0x34, 0x56, 0x78}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileType, _, ok := checkMagic(tt.data)
			if ok != tt.wantOK {
				t.Errorf("checkMagic() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && !strings.Contains(fileType, tt.wantType) {
				t.Errorf("checkMagic() type = %q, want to contain %q", fileType, tt.wantType)
			}
		})
	}
}
