package bzip2

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Pre-compressed bzip2 data for "hello world"
var helloWorldBz2 = []byte{
	0x42, 0x5a, 0x68, 0x39, 0x31, 0x41, 0x59, 0x26,
	0x53, 0x59, 0x44, 0xf7, 0x13, 0x78, 0x00, 0x00,
	0x01, 0x91, 0x80, 0x40, 0x00, 0x06, 0x44, 0x90,
	0x80, 0x20, 0x00, 0x22, 0x03, 0x34, 0x84, 0x30,
	0x21, 0xb6, 0x81, 0x54, 0x27, 0x8b, 0xb9, 0x22,
	0x9c, 0x28, 0x48, 0x22, 0x7b, 0x89, 0xbc, 0x00,
}

func TestRunBzip2_CompressionNotSupported(t *testing.T) {
	var buf bytes.Buffer

	err := RunBzip2(&buf, []string{"file.txt"}, Bzip2Options{})
	if err == nil {
		t.Error("RunBzip2() should error when compression requested (not supported)")
	}
}

func TestRunBzip2_Decompress(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bzip2_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a bz2 file
	bz2File := filepath.Join(tmpDir, "test.bz2")
	_ = os.WriteFile(bz2File, helloWorldBz2, 0644)

	var buf bytes.Buffer

	err = RunBzip2(&buf, []string{bz2File}, Bzip2Options{
		Decompress: true,
		Stdout:     true,
	})
	if err != nil {
		t.Fatalf("RunBzip2() decompress error = %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("hello")) {
		t.Errorf("RunBzip2() should decompress to 'hello world': got %q", buf.String())
	}
}

func TestRunBzip2_DecompressFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bzip2_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a bz2 file
	bz2File := filepath.Join(tmpDir, "test.bz2")
	_ = os.WriteFile(bz2File, helloWorldBz2, 0644)

	var buf bytes.Buffer

	err = RunBzip2(&buf, []string{bz2File}, Bzip2Options{
		Decompress: true,
		Keep:       true,
	})
	if err != nil {
		t.Fatalf("RunBzip2() decompress file error = %v", err)
	}

	// Check output file exists
	outFile := filepath.Join(tmpDir, "test")
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Error("RunBzip2() should create decompressed file")
	}

	// Original should still exist with -k
	if _, err := os.Stat(bz2File); os.IsNotExist(err) {
		t.Error("RunBzip2() -k should keep original file")
	}
}

func TestRunBzip2_UnknownSuffix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bzip2_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create file without .bz2 suffix
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("not compressed"), 0644)

	var buf bytes.Buffer
	// This will print error to stderr but not fail
	_ = RunBzip2(&buf, []string{testFile}, Bzip2Options{Decompress: true})
}

func TestRunBzip2_NoOverwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bzip2_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create bz2 and output file
	bz2File := filepath.Join(tmpDir, "test.bz2")
	_ = os.WriteFile(bz2File, helloWorldBz2, 0644)

	outFile := filepath.Join(tmpDir, "test")
	_ = os.WriteFile(outFile, []byte("existing"), 0644)

	var buf bytes.Buffer
	// Should print error about existing file
	_ = RunBzip2(&buf, []string{bz2File}, Bzip2Options{Decompress: true})
}

func TestRunBzip2_Force(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bzip2_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create bz2 and output file
	bz2File := filepath.Join(tmpDir, "test.bz2")
	_ = os.WriteFile(bz2File, helloWorldBz2, 0644)

	outFile := filepath.Join(tmpDir, "test")
	_ = os.WriteFile(outFile, []byte("existing"), 0644)

	var buf bytes.Buffer

	err = RunBzip2(&buf, []string{bz2File}, Bzip2Options{
		Decompress: true,
		Force:      true,
		Keep:       true,
	})
	if err != nil {
		t.Fatalf("RunBzip2() force error = %v", err)
	}

	// Output should be overwritten
	data, _ := os.ReadFile(outFile)
	if bytes.Contains(data, []byte("existing")) {
		t.Error("RunBzip2() -f should overwrite existing file")
	}
}

func TestRunBunzip2(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bunzip2_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	bz2File := filepath.Join(tmpDir, "test.bz2")
	_ = os.WriteFile(bz2File, helloWorldBz2, 0644)

	var buf bytes.Buffer

	err = RunBunzip2(&buf, []string{bz2File}, Bzip2Options{Stdout: true})
	if err != nil {
		t.Fatalf("RunBunzip2() error = %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("hello")) {
		t.Errorf("RunBunzip2() should decompress: got %q", buf.String())
	}
}

func TestRunBzcat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bzcat_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	bz2File := filepath.Join(tmpDir, "test.bz2")
	_ = os.WriteFile(bz2File, helloWorldBz2, 0644)

	var buf bytes.Buffer

	err = RunBzcat(&buf, []string{bz2File})
	if err != nil {
		t.Fatalf("RunBzcat() error = %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("hello")) {
		t.Errorf("RunBzcat() should output decompressed: got %q", buf.String())
	}
}

func TestRunBzcat_AddsSuffix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bzcat_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create file with .bz2 suffix
	bz2File := filepath.Join(tmpDir, "test.bz2")
	_ = os.WriteFile(bz2File, helloWorldBz2, 0644)

	var buf bytes.Buffer
	// Call without .bz2 suffix - should try adding it
	_ = RunBzcat(&buf, []string{filepath.Join(tmpDir, "test")})
}

func TestRunBzip2_Verbose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bzip2_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	bz2File := filepath.Join(tmpDir, "test.bz2")
	_ = os.WriteFile(bz2File, helloWorldBz2, 0644)

	var buf bytes.Buffer

	err = RunBzip2(&buf, []string{bz2File}, Bzip2Options{
		Decompress: true,
		Verbose:    true,
		Keep:       true,
	})
	if err != nil {
		t.Fatalf("RunBzip2() verbose error = %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("replaced")) {
		t.Errorf("RunBzip2() verbose should show replacement info: got %q", buf.String())
	}
}
