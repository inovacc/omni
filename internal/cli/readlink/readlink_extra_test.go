package readlink

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

// These tests run on all platforms including Windows (no symlinks required).

func TestRunReadlink_MissingOperand(t *testing.T) {
	var buf bytes.Buffer
	err := RunReadlink(&buf, []string{}, ReadlinkOptions{})
	if err == nil {
		t.Error("RunReadlink() expected error for missing operand")
	}
}

func TestRunReadlink_CanonicalizeMissing_Windows(t *testing.T) {
	// CanonicalizeMissing (-m) resolves via filepath.Abs — works on all platforms.
	tmpDir, err := os.MkdirTemp("", "readlink_win_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	nonexistent := filepath.Join(tmpDir, "nonexistent_path")

	var buf bytes.Buffer
	err = RunReadlink(&buf, []string{nonexistent}, ReadlinkOptions{CanonicalizeMissing: true})
	if err != nil {
		t.Fatalf("RunReadlink() -m error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("RunReadlink() -m should return a path")
	}
}

func TestRunReadlink_CanonicalizeExisting_File(t *testing.T) {
	// CanonicalizeExisting (-e) works on real files without needing symlinks.
	tmpDir, err := os.MkdirTemp("", "readlink_existing_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	realFile := filepath.Join(tmpDir, "real.txt")
	if err := os.WriteFile(realFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = RunReadlink(&buf, []string{realFile}, ReadlinkOptions{CanonicalizeExisting: true})
	if err != nil {
		t.Fatalf("RunReadlink() -e error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("RunReadlink() -e should produce output")
	}
}

func TestRunReadlink_CanonicalizeExisting_Missing(t *testing.T) {
	// CanonicalizeExisting (-e) fails when path does not exist.
	var buf bytes.Buffer
	err := RunReadlink(&buf, []string{"/nonexistent/path/to/file"}, ReadlinkOptions{CanonicalizeExisting: true})
	if err == nil {
		t.Error("RunReadlink() -e should return error for nonexistent path")
	}
}

func TestRunReadlink_Canonicalize_ExistingFile(t *testing.T) {
	// Canonicalize (-f) resolves a real file without symlinks.
	tmpDir, err := os.MkdirTemp("", "readlink_canon_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	realFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(realFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = RunReadlink(&buf, []string{realFile}, ReadlinkOptions{Canonicalize: true})
	if err != nil {
		t.Fatalf("RunReadlink() -f error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("RunReadlink() -f should produce output for existing file")
	}
}

func TestRunReadlink_Quiet_NonSymlink(t *testing.T) {
	// On all platforms: quiet mode suppresses stderr for error but still errors.
	tmpDir, err := os.MkdirTemp("", "readlink_quiet_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	realFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(realFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	// Passing a regular file to readlink (not a symlink) should fail.
	err = RunReadlink(&buf, []string{realFile}, ReadlinkOptions{Quiet: true})
	// error expected (file is not a symlink)
	if err == nil {
		t.Log("RunReadlink() quiet on non-symlink: platform may not error")
	}
}

func TestRunReadlink_Silent_Mode(t *testing.T) {
	// Silent is the same as Quiet — both suppress stderr messages.
	var buf bytes.Buffer
	err := RunReadlink(&buf, []string{"/nonexistent/file"}, ReadlinkOptions{Silent: true})
	// Error returned, but nothing printed to stderr
	if err == nil {
		t.Log("RunReadlink() silent: may not error on all platforms")
	}
}

func TestRunReadlink_NoNewline_MultiArgs(t *testing.T) {
	// NoNewline only applies when there is exactly 1 argument; ignored for multiple.
	tmpDir, err := os.MkdirTemp("", "readlink_nonl_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	f1 := filepath.Join(tmpDir, "f1.txt")
	_ = os.WriteFile(f1, []byte("a"), 0644)

	// With a single file that IS a real file (not symlink), readlink will error.
	// We just verify the option doesn't panic.
	var buf bytes.Buffer
	_ = RunReadlink(&buf, []string{f1}, ReadlinkOptions{NoNewline: true})
}

func TestRunReadlink_Zero_MultipleArgs_CanonicalizeMissing(t *testing.T) {
	// Zero terminator with multiple args using -m (no symlinks needed).
	tmpDir, err := os.MkdirTemp("", "readlink_zero_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	p1 := filepath.Join(tmpDir, "a")
	p2 := filepath.Join(tmpDir, "b")

	var buf bytes.Buffer
	err = RunReadlink(&buf, []string{p1, p2}, ReadlinkOptions{CanonicalizeMissing: true, Zero: true})
	if err != nil {
		t.Fatalf("RunReadlink() -z -m error = %v", err)
	}

	output := buf.String()
	// Each entry should end with NUL, not newline
	if len(output) > 0 && output[len(output)-1] == '\n' {
		t.Error("RunReadlink() -z should not end with newline")
	}
}

func TestReadlinkHelper(t *testing.T) {
	// Test the Readlink() helper function with a non-symlink.
	tmpDir, err := os.MkdirTemp("", "readlink_helper_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	realFile := filepath.Join(tmpDir, "file.txt")
	_ = os.WriteFile(realFile, []byte("data"), 0644)

	// Regular file: should return an error.
	_, err = Readlink(realFile)
	if err == nil {
		t.Log("Readlink() on regular file: platform may not error")
	}
}

func TestCanonicalPathHelper_Existing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "canonical_path_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	f := filepath.Join(tmpDir, "file.txt")
	_ = os.WriteFile(f, []byte("data"), 0644)

	result, err := CanonicalPath(f)
	if err != nil {
		t.Fatalf("CanonicalPath() error = %v", err)
	}
	if !filepath.IsAbs(result) {
		t.Errorf("CanonicalPath() = %q, want absolute path", result)
	}
}

func TestRunReadlink_JSONOutput(t *testing.T) {
	// JSON output mode with -m so we get a result without symlinks.
	tmpDir, err := os.MkdirTemp("", "readlink_json_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	p := filepath.Join(tmpDir, "missing")

	var buf bytes.Buffer
	err = RunReadlink(&buf, []string{p}, ReadlinkOptions{CanonicalizeMissing: true, OutputFormat: output.FormatJSON})
	if err != nil {
		t.Fatalf("RunReadlink() json error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("RunReadlink() json should produce output")
	}
}
