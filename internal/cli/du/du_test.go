package du

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDU(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "du_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test structure
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello world"), 0644)
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content"), 0644)

	t.Run("default output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDU(&buf, []string{tmpDir}, DUOptions{})
		if err != nil {
			t.Fatalf("RunDU() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, tmpDir) {
			t.Errorf("RunDU() should contain path: %s", output)
		}
	})

	t.Run("all files", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDU(&buf, []string{tmpDir}, DUOptions{All: true})
		if err != nil {
			t.Fatalf("RunDU() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "file1.txt") {
			t.Errorf("RunDU() -a should list files: %s", output)
		}
	})

	t.Run("summarize only", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDU(&buf, []string{tmpDir}, DUOptions{SummarizeOnly: true})
		if err != nil {
			t.Fatalf("RunDU() error = %v", err)
		}

		output := buf.String()

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 1 {
			t.Errorf("RunDU() -s should output only one line, got %d", len(lines))
		}
	})

	t.Run("human readable", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDU(&buf, []string{tmpDir}, DUOptions{HumanReadable: true})
		if err != nil {
			t.Fatalf("RunDU() error = %v", err)
		}

		// Output should be present
		if buf.Len() == 0 {
			t.Error("RunDU() -h should produce output")
		}
	})

	t.Run("byte count", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDU(&buf, []string{tmpDir}, DUOptions{ByteCount: true})
		if err != nil {
			t.Fatalf("RunDU() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunDU() -b should produce output")
		}
	})

	t.Run("total", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDU(&buf, []string{tmpDir, tmpDir}, DUOptions{Total: true})
		if err != nil {
			t.Fatalf("RunDU() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "total") {
			t.Errorf("RunDU() -c should show total: %s", output)
		}
	})

	t.Run("null terminator", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDU(&buf, []string{tmpDir}, DUOptions{NullTerminator: true, SummarizeOnly: true})
		if err != nil {
			t.Fatalf("RunDU() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "\x00") {
			t.Error("RunDU() -0 should use null terminator")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		var buf bytes.Buffer

		// Should not error, just print to stderr
		err := RunDU(&buf, []string{"/nonexistent/path/12345"}, DUOptions{})
		if err != nil {
			t.Fatalf("RunDU() error = %v", err)
		}
	})

	t.Run("single file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single.txt")
		_ = os.WriteFile(file, []byte("test content"), 0644)

		var buf bytes.Buffer

		err := RunDU(&buf, []string{file}, DUOptions{All: true})
		if err != nil {
			t.Fatalf("RunDU() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunDU() on single file should produce output")
		}
	})

	t.Run("default path", func(t *testing.T) {
		// Change to temp dir
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)

		defer func() { _ = os.Chdir(origDir) }()

		var buf bytes.Buffer

		err := RunDU(&buf, []string{}, DUOptions{})
		if err != nil {
			t.Fatalf("RunDU() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, ".") {
			t.Errorf("RunDU() default should use current dir: %s", output)
		}
	})
}

func TestFormatHumanSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0"},
		{500, "500"},
		{1023, "1023"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
		{1073741824, "1.0G"},
		{1099511627776, "1.0T"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatHumanSize(tt.bytes)
			if got != tt.expected {
				t.Errorf("FormatHumanSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
			}
		})
	}
}

func TestDiskUsage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "diskusage_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("world"), 0644)

	size, err := DiskUsage(tmpDir)
	if err != nil {
		t.Fatalf("DiskUsage() error = %v", err)
	}

	// Should be at least 10 bytes (5 + 5)
	if size < 10 {
		t.Errorf("DiskUsage() = %d, want >= 10", size)
	}
}

func TestCalculateDirSize(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calcdir_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644)
	subDir := filepath.Join(tmpDir, "sub")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("world"), 0644)

	size := calculateDirSize(tmpDir)
	// Should include both files
	if size < 10 {
		t.Errorf("calculateDirSize() = %d, want >= 10", size)
	}
}
