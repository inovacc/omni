package hash

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

func TestRunHashBinaryMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hash_binary_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.bin")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = RunHash(&buf, []string{testFile}, HashOptions{Binary: true})
	if err != nil {
		t.Fatalf("RunHash() binary mode error = %v", err)
	}

	output := buf.String()
	// Binary mode uses '*' prefix before filename
	if !strings.Contains(output, "*") {
		t.Errorf("RunHash() binary mode output should contain '*', got: %q", output)
	}
}

func TestRunHashRecursiveDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hash_recursive_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create files in subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("recursive hashes directory", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunHash(&buf, []string{tmpDir}, HashOptions{Recursive: true})
		if err != nil {
			t.Fatalf("RunHash() recursive error = %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "file1.txt") || !strings.Contains(output, "file2.txt") {
			t.Errorf("RunHash() recursive should hash files in subdirs, got: %q", output)
		}
	})

	t.Run("non-recursive skips directory", func(t *testing.T) {
		var buf bytes.Buffer
		// Should not error, just print to stderr that it's a directory
		_ = RunHash(&buf, []string{tmpDir}, HashOptions{Recursive: false})
		// output should be empty since it's a directory and non-recursive
		if strings.Contains(buf.String(), "file1.txt") {
			t.Error("RunHash() non-recursive should not hash files in directories")
		}
	})
}

func TestRunHashJSONOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hash_json_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = RunHash(&buf, []string{testFile}, HashOptions{OutputFormat: output.FormatJSON})
	if err != nil {
		t.Fatalf("RunHash() json error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"hash"`) {
		t.Errorf("RunHash() json output should contain 'hash' key, got: %q", output)
	}
	if !strings.Contains(output, `"algorithm"`) {
		t.Errorf("RunHash() json output should contain 'algorithm' key, got: %q", output)
	}
}

func TestVerifyChecksums(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hash_verify_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a file to hash
	testFile := filepath.Join(tmpDir, "data.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	// Compute its hash
	var hashBuf bytes.Buffer
	if err := RunHash(&hashBuf, []string{testFile}, HashOptions{Algorithm: "sha256"}); err != nil {
		t.Fatal(err)
	}
	hashLine := strings.TrimSpace(hashBuf.String())

	// Write checksum file
	checksumFile := filepath.Join(tmpDir, "SHA256SUMS")
	if err := os.WriteFile(checksumFile, []byte(hashLine+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("verify valid checksums", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunHash(&buf, []string{checksumFile}, HashOptions{Algorithm: "sha256", Check: true})
		if err != nil {
			t.Fatalf("RunHash() verify error = %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "OK") {
			t.Errorf("RunHash() verify should print OK, got: %q", output)
		}
	})

	t.Run("verify with quiet mode", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunHash(&buf, []string{checksumFile}, HashOptions{Algorithm: "sha256", Check: true, Quiet: true})
		if err != nil {
			t.Fatalf("RunHash() verify quiet error = %v", err)
		}
		// Quiet mode suppresses OK messages
		if strings.Contains(buf.String(), "OK") {
			t.Error("RunHash() verify quiet should not print OK")
		}
	})

	t.Run("verify with status mode", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunHash(&buf, []string{checksumFile}, HashOptions{Algorithm: "sha256", Check: true, Status: true})
		if err != nil {
			t.Fatalf("RunHash() verify status error = %v", err)
		}
		// Status mode produces no output
		if buf.Len() != 0 {
			t.Errorf("RunHash() verify status should produce no output, got: %q", buf.String())
		}
	})

	t.Run("verify corrupted checksum", func(t *testing.T) {
		badChecksumFile := filepath.Join(tmpDir, "BAD_SHA256SUMS")
		badContent := fmt.Sprintf("0000000000000000000000000000000000000000000000000000000000000000  %s\n", testFile)
		if err := os.WriteFile(badChecksumFile, []byte(badContent), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		err := RunHash(&buf, []string{badChecksumFile}, HashOptions{Algorithm: "sha256", Check: true})
		if err == nil {
			t.Error("RunHash() verify should return error for mismatched checksums")
		}
		if !strings.Contains(buf.String(), "FAILED") {
			t.Errorf("RunHash() verify should print FAILED, got: %q", buf.String())
		}
	})

	t.Run("verify missing file", func(t *testing.T) {
		missingChecksumFile := filepath.Join(tmpDir, "MISSING_SHA256SUMS")
		if err := os.WriteFile(missingChecksumFile, []byte("abcdef  /nonexistent/file.txt\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		err := RunHash(&buf, []string{missingChecksumFile}, HashOptions{Algorithm: "sha256", Check: true})
		if err == nil {
			t.Error("RunHash() verify should return error when file is missing")
		}
	})

	t.Run("verify no checksum file specified", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunHash(&buf, []string{}, HashOptions{Check: true})
		if err == nil {
			t.Error("RunHash() verify with no args should return error")
		}
	})

	t.Run("verify nonexistent checksum file", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunHash(&buf, []string{"/nonexistent/SHA256SUMS"}, HashOptions{Check: true})
		if err == nil {
			t.Error("RunHash() verify nonexistent checksum file should return error")
		}
	})

	t.Run("verify warn malformed lines", func(t *testing.T) {
		malformedFile := filepath.Join(tmpDir, "MALFORMED")
		if err := os.WriteFile(malformedFile, []byte("this-has-no-space-separator\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		// With warn enabled — should not crash
		_ = RunHash(&buf, []string{malformedFile}, HashOptions{Check: true, Warn: true})
	})
}

func TestRunSHA1Sum(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sha1_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = RunSHA1Sum(&buf, []string{testFile}, HashOptions{})
	if err != nil {
		t.Fatalf("RunSHA1Sum() error = %v", err)
	}

	// SHA1 of "test" is a94a8fe5ccb19ba61c4c0873d391e987982fbbd3
	if !strings.Contains(buf.String(), "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3") {
		t.Errorf("RunSHA1Sum() got = %v, want sha1 of 'test'", buf.String())
	}
}

func TestRunHashRecursiveJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hash_recursive_json_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "a.txt"), []byte("aaa"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = RunHash(&buf, []string{tmpDir}, HashOptions{Recursive: true, OutputFormat: output.FormatJSON})
	if err != nil {
		t.Fatalf("RunHash() recursive json error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"hash"`) {
		t.Errorf("RunHash() recursive json should contain hash field, got: %q", output)
	}
}
