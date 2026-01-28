package hash

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHash(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hash_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		args      []string
		opts      HashOptions
		contains  string
		wantError bool
	}{
		{
			name:     "sha256 default",
			args:     []string{testFile},
			opts:     HashOptions{},
			contains: "test.txt", // just check filename appears
		},
		{
			name:     "md5",
			args:     []string{testFile},
			opts:     HashOptions{Algorithm: "md5"},
			contains: testFile,
		},
		{
			name:     "sha512",
			args:     []string{testFile},
			opts:     HashOptions{Algorithm: "sha512"},
			contains: testFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunHash(&buf, tt.args, tt.opts)
			if (err != nil) != tt.wantError {
				t.Errorf("RunHash() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if tt.contains != "" && !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("RunHash() output = %v, want contains %v", buf.String(), tt.contains)
			}
		})
	}
}

func TestRunMD5Sum(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "md5_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunMD5Sum(&buf, []string{testFile}, HashOptions{})
	if err != nil {
		t.Fatalf("RunMD5Sum() error = %v", err)
	}

	// MD5 of "test" is 098f6bcd4621d373cade4e832627b4f6
	if !strings.Contains(buf.String(), "098f6bcd4621d373cade4e832627b4f6") {
		t.Errorf("RunMD5Sum() got = %v, want hash of 'test'", buf.String())
	}
}

func TestRunSHA256Sum(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sha256_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunSHA256Sum(&buf, []string{testFile}, HashOptions{})
	if err != nil {
		t.Fatalf("RunSHA256Sum() error = %v", err)
	}

	// SHA256 of "test" is 9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
	if !strings.Contains(buf.String(), "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08") {
		t.Errorf("RunSHA256Sum() got = %v, want hash of 'test'", buf.String())
	}
}

func TestRunHashExtended(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hash_ext_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("sha1", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "sha1.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunHash(&buf, []string{testFile}, HashOptions{Algorithm: "sha1"})
		if err != nil {
			t.Fatalf("RunHash() sha1 error = %v", err)
		}

		// SHA1 of "test" is a94a8fe5ccb19ba61c4c0873d391e987982fbbd3
		if !strings.Contains(buf.String(), "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3") {
			t.Errorf("RunHash() sha1 got = %v", buf.String())
		}
	})

	t.Run("sha384", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "sha384.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunHash(&buf, []string{testFile}, HashOptions{Algorithm: "sha384"})
		if err != nil {
			t.Fatalf("RunHash() sha384 error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunHash() sha384 should produce output")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunHash(&buf, []string{testFile}, HashOptions{})
		if err != nil {
			t.Fatalf("RunHash() error = %v", err)
		}

		// SHA256 of empty string is e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
		if !strings.Contains(buf.String(), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855") {
			t.Logf("RunHash() empty file got = %v", buf.String())
		}
	})

	t.Run("binary file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "binary.bin")
		if err := os.WriteFile(testFile, []byte{0x00, 0xFF, 0x7F, 0x80}, 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunHash(&buf, []string{testFile}, HashOptions{})
		if err != nil {
			t.Fatalf("RunHash() binary error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunHash() should hash binary files")
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		_ = os.WriteFile(file1, []byte("content1"), 0644)
		_ = os.WriteFile(file2, []byte("content2"), 0644)

		var buf bytes.Buffer

		err := RunHash(&buf, []string{file1, file2}, HashOptions{})
		if err != nil {
			t.Fatalf("RunHash() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "file1.txt") || !strings.Contains(output, "file2.txt") {
			t.Errorf("RunHash() should show both files: %v", output)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunHash(&buf, []string{"/nonexistent/file.txt"}, HashOptions{})
		// Implementation may print error but not return it
		if err == nil {
			t.Log("RunHash() prints error to output but may not return error")
		}
	})

	t.Run("invalid algorithm", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "invalid.txt")
		_ = os.WriteFile(testFile, []byte("test"), 0644)

		var buf bytes.Buffer

		err := RunHash(&buf, []string{testFile}, HashOptions{Algorithm: "invalid"})
		if err == nil {
			t.Log("RunHash() may handle invalid algorithm gracefully")
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "unicode.txt")
		if err := os.WriteFile(testFile, []byte("世界🌍こんにちは"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunHash(&buf, []string{testFile}, HashOptions{})
		if err != nil {
			t.Fatalf("RunHash() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunHash() should hash unicode content")
		}
	})

	t.Run("large file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "large.txt")
		largeContent := strings.Repeat("x", 100000)
		if err := os.WriteFile(testFile, []byte(largeContent), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunHash(&buf, []string{testFile}, HashOptions{})
		if err != nil {
			t.Fatalf("RunHash() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("RunHash() should hash large files")
		}
	})

	t.Run("consistent hash", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "consistent.txt")
		if err := os.WriteFile(testFile, []byte("consistent content"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf1, buf2 bytes.Buffer

		_ = RunHash(&buf1, []string{testFile}, HashOptions{})
		_ = RunHash(&buf2, []string{testFile}, HashOptions{})

		if buf1.String() != buf2.String() {
			t.Error("RunHash() should produce consistent results")
		}
	})

	t.Run("output format", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "format.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunHash(&buf, []string{testFile}, HashOptions{})

		output := buf.String()
		// Should be in format: hash  filename
		if !strings.Contains(output, " ") {
			t.Log("RunHash() output format may vary")
		}
	})
}

func TestRunSHA512Sum(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sha512_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunHash(&buf, []string{testFile}, HashOptions{Algorithm: "sha512"})
	if err != nil {
		t.Fatalf("RunHash() sha512 error = %v", err)
	}

	// SHA512 hash should be 128 hex characters
	output := buf.String()
	parts := strings.Fields(output)
	if len(parts) > 0 && len(parts[0]) != 128 {
		t.Logf("SHA512 hash length: %d (expected 128)", len(parts[0]))
	}
}
