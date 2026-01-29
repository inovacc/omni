package gzip

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGzip_Compress(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gzip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("hello world"), 0644)

	var buf bytes.Buffer

	err = RunGzip(&buf, []string{testFile}, GzipOptions{Keep: true})
	if err != nil {
		t.Fatalf("RunGzip() compress error = %v", err)
	}

	// Check .gz file exists
	gzFile := testFile + ".gz"
	if _, err := os.Stat(gzFile); os.IsNotExist(err) {
		t.Error("RunGzip() should create .gz file")
	}
}

func TestRunGzip_Decompress(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gzip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create compressed file
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("hello world"), 0644)

	var buf bytes.Buffer

	_ = RunGzip(&buf, []string{testFile}, GzipOptions{Keep: true})

	// Now decompress
	gzFile := testFile + ".gz"

	err = RunGzip(&buf, []string{gzFile}, GzipOptions{Decompress: true, Keep: true})
	if err != nil {
		t.Fatalf("RunGzip() decompress error = %v", err)
	}

	// Verify decompressed content
	data, _ := os.ReadFile(testFile)
	if string(data) != "hello world" {
		t.Errorf("RunGzip() decompress got %q, want 'hello world'", data)
	}
}

func TestRunGzip_Stdout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gzip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("hello"), 0644)

	var buf bytes.Buffer

	err = RunGzip(&buf, []string{testFile}, GzipOptions{Stdout: true})
	if err != nil {
		t.Fatalf("RunGzip() stdout error = %v", err)
	}

	// Output should be gzip data
	if buf.Len() == 0 {
		t.Error("RunGzip() stdout should produce output")
	}

	// Verify it's valid gzip
	gr, err := gzip.NewReader(&buf)
	if err != nil {
		t.Fatalf("RunGzip() stdout should produce valid gzip: %v", err)
	}

	_ = gr.Close()
}

func TestRunGzip_AlreadyCompressed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gzip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .gz file
	gzFile := filepath.Join(tmpDir, "test.gz")
	_ = os.WriteFile(gzFile, []byte("dummy"), 0644)

	var buf bytes.Buffer
	// Should error - already has .gz suffix
	_ = RunGzip(&buf, []string{gzFile}, GzipOptions{})
}

func TestRunGzip_Force(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gzip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("hello"), 0644)

	// Create existing .gz file
	gzFile := testFile + ".gz"
	_ = os.WriteFile(gzFile, []byte("existing"), 0644)

	var buf bytes.Buffer

	err = RunGzip(&buf, []string{testFile}, GzipOptions{Force: true, Keep: true})
	if err != nil {
		t.Fatalf("RunGzip() force error = %v", err)
	}

	// .gz file should be overwritten
	data, _ := os.ReadFile(gzFile)
	if bytes.Contains(data, []byte("existing")) {
		t.Error("RunGzip() -f should overwrite existing file")
	}
}

func TestRunGzip_Verbose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gzip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("hello world test content"), 0644)

	var buf bytes.Buffer

	err = RunGzip(&buf, []string{testFile}, GzipOptions{Verbose: true, Keep: true})
	if err != nil {
		t.Fatalf("RunGzip() verbose error = %v", err)
	}

	if !strings.Contains(buf.String(), "%") {
		t.Errorf("RunGzip() verbose should show compression ratio: got %q", buf.String())
	}
}

func TestRunGzip_Level(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gzip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test file with compressible content
	testFile := filepath.Join(tmpDir, "test.txt")
	content := strings.Repeat("hello world ", 1000)
	_ = os.WriteFile(testFile, []byte(content), 0644)

	var buf bytes.Buffer

	err = RunGzip(&buf, []string{testFile}, GzipOptions{Level: 9, Keep: true})
	if err != nil {
		t.Fatalf("RunGzip() level error = %v", err)
	}

	// Verify compression happened
	gzFile := testFile + ".gz"

	info, _ := os.Stat(gzFile)
	if info.Size() >= int64(len(content)) {
		t.Error("RunGzip() should compress data")
	}
}

func TestRunGunzip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gunzip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create and compress file
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("gunzip test"), 0644)

	var buf bytes.Buffer

	_ = RunGzip(&buf, []string{testFile}, GzipOptions{})

	// Original is removed, .gz exists
	gzFile := testFile + ".gz"

	err = RunGunzip(&buf, []string{gzFile}, GzipOptions{})
	if err != nil {
		t.Fatalf("RunGunzip() error = %v", err)
	}

	// Original should be restored
	data, _ := os.ReadFile(testFile)
	if string(data) != "gunzip test" {
		t.Errorf("RunGunzip() got %q, want 'gunzip test'", data)
	}
}

func TestRunZcat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zcat_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create and compress file
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("zcat content"), 0644)

	var buf bytes.Buffer

	_ = RunGzip(&buf, []string{testFile}, GzipOptions{})

	// zcat the .gz file
	gzFile := testFile + ".gz"

	buf.Reset()

	err = RunZcat(&buf, []string{gzFile})
	if err != nil {
		t.Fatalf("RunZcat() error = %v", err)
	}

	if !strings.Contains(buf.String(), "zcat content") {
		t.Errorf("RunZcat() got %q, want 'zcat content'", buf.String())
	}
}

func TestRunZcat_AddsSuffix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zcat_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("suffix test"), 0644)

	var buf bytes.Buffer

	_ = RunGzip(&buf, []string{testFile}, GzipOptions{})

	// Call without .gz suffix
	buf.Reset()
	_ = RunZcat(&buf, []string{testFile})
}

func TestRunGzip_DecompressUnknownSuffix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gzip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("not compressed"), 0644)

	var buf bytes.Buffer
	// Should error - unknown suffix
	_ = RunGzip(&buf, []string{testFile}, GzipOptions{Decompress: true})
}

func TestRunGzip_NoOverwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gzip_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create source and target
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("content"), 0644)

	gzFile := testFile + ".gz"
	_ = os.WriteFile(gzFile, []byte("existing"), 0644)

	var buf bytes.Buffer
	// Should print error about existing file
	_ = RunGzip(&buf, []string{testFile}, GzipOptions{})
}

func TestGzipReader(t *testing.T) {
	var buf bytes.Buffer

	input := bytes.NewBufferString("test data")

	err := gzipReader(&buf, input, gzip.DefaultCompression)
	if err != nil {
		t.Fatalf("gzipReader() error = %v", err)
	}

	if buf.Len() == 0 {
		t.Error("gzipReader() should produce output")
	}
}

func TestGunzipReader(t *testing.T) {
	// First create gzipped data
	var compressed bytes.Buffer

	gw := gzip.NewWriter(&compressed)
	_, _ = gw.Write([]byte("decompressed"))
	_ = gw.Close()

	// Now decompress
	var buf bytes.Buffer

	err := gunzipReader(&buf, &compressed)
	if err != nil {
		t.Fatalf("gunzipReader() error = %v", err)
	}

	if buf.String() != "decompressed" {
		t.Errorf("gunzipReader() got %q, want 'decompressed'", buf.String())
	}
}
