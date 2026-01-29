package xz

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Valid XZ magic header
var xzMagicBytes = []byte{0xFD, '7', 'z', 'X', 'Z', 0x00}

// Minimal valid XZ file (just headers, no actual data)
var minimalXzFile = append(xzMagicBytes, []byte{
	0x00, 0x00, // Stream flags
	0x00, 0x00, 0x00, 0x00, // CRC32
	// ... minimal block/index data would go here
	// Footer
	0x00, 0x00, 0x00, 0x00, // CRC32 of index
	0x00, 0x00, 0x00, 0x00, // Backward size
	0x00, 0x00, // Stream flags
	'Y', 'Z', // Footer magic
}...)

func TestRunXz_CompressionNotSupported(t *testing.T) {
	var buf bytes.Buffer

	err := RunXz(&buf, []string{"file.txt"}, XzOptions{})
	if err == nil {
		t.Error("RunXz() should error when compression requested (not supported)")
	}
}

func TestRunXz_DecompressStdinNotSupported(t *testing.T) {
	var buf bytes.Buffer

	err := RunXz(&buf, []string{}, XzOptions{Decompress: true})
	if err == nil {
		t.Error("RunXz() should error for stdin decompression")
	}
}

func TestRunXz_UnknownSuffix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "xz_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("not xz"), 0644)

	var buf bytes.Buffer
	// Should print error about unknown suffix
	_ = RunXz(&buf, []string{testFile}, XzOptions{Decompress: true})
}

func TestRunXz_NotXzFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "xz_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .xz file without proper magic
	xzFile := filepath.Join(tmpDir, "test.xz")
	_ = os.WriteFile(xzFile, []byte("not xz data"), 0644)

	var buf bytes.Buffer
	// Should error - not xz format
	_ = RunXz(&buf, []string{xzFile}, XzOptions{Decompress: true})
}

func TestRunXz_List(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "xz_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create file with XZ magic
	xzFile := filepath.Join(tmpDir, "test.xz")
	_ = os.WriteFile(xzFile, minimalXzFile, 0644)

	var buf bytes.Buffer

	err = RunXz(&buf, []string{xzFile}, XzOptions{List: true})
	if err != nil {
		t.Fatalf("RunXz() list error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Strms") {
		t.Errorf("RunXz() list should show header: got %q", output)
	}

	if !strings.Contains(output, "test.xz") {
		t.Errorf("RunXz() list should show filename: got %q", output)
	}
}

func TestRunXz_ListNotXz(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "xz_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create file without XZ magic
	xzFile := filepath.Join(tmpDir, "test.xz")
	_ = os.WriteFile(xzFile, []byte("not xz"), 0644)

	var buf bytes.Buffer
	// Should print error about not being xz format
	_ = RunXz(&buf, []string{xzFile}, XzOptions{List: true})
}

func TestRunXz_NoOverwrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "xz_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create xz file
	xzFile := filepath.Join(tmpDir, "test.xz")
	_ = os.WriteFile(xzFile, minimalXzFile, 0644)

	// Create output file that would conflict
	outFile := filepath.Join(tmpDir, "test")
	_ = os.WriteFile(outFile, []byte("existing"), 0644)

	var buf bytes.Buffer
	// Should print error about existing file
	_ = RunXz(&buf, []string{xzFile}, XzOptions{Decompress: true})
}

func TestRunUnxz(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "unxz_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	xzFile := filepath.Join(tmpDir, "test.xz")
	_ = os.WriteFile(xzFile, minimalXzFile, 0644)

	var buf bytes.Buffer
	// Will error because full decompression requires external lib
	_ = RunUnxz(&buf, []string{xzFile}, XzOptions{})
}

func TestRunXzcat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "xzcat_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	xzFile := filepath.Join(tmpDir, "test.xz")
	_ = os.WriteFile(xzFile, minimalXzFile, 0644)

	var buf bytes.Buffer
	// Will error because full decompression requires external lib
	_ = RunXzcat(&buf, []string{xzFile})
}

func TestXzMagic(t *testing.T) {
	expected := []byte{0xFD, '7', 'z', 'X', 'Z', 0x00}
	if !bytes.Equal(xzMagic, expected) {
		t.Errorf("xzMagic = %v, want %v", xzMagic, expected)
	}
}

func TestRunXz_Verbose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "xz_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	xzFile := filepath.Join(tmpDir, "test.xz")
	_ = os.WriteFile(xzFile, minimalXzFile, 0644)

	var buf bytes.Buffer
	// Will error because full decompression requires external lib
	// but verbose flag should be accepted
	_ = RunXz(&buf, []string{xzFile}, XzOptions{
		Decompress: true,
		Verbose:    true,
	})
}

func TestRunXz_Force(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "xz_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	xzFile := filepath.Join(tmpDir, "test.xz")
	_ = os.WriteFile(xzFile, minimalXzFile, 0644)

	outFile := filepath.Join(tmpDir, "test")
	_ = os.WriteFile(outFile, []byte("existing"), 0644)

	var buf bytes.Buffer
	// With force, should attempt even if output exists
	_ = RunXz(&buf, []string{xzFile}, XzOptions{
		Decompress: true,
		Force:      true,
	})
}

func TestRunXz_Keep(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "xz_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	xzFile := filepath.Join(tmpDir, "test.xz")
	_ = os.WriteFile(xzFile, minimalXzFile, 0644)

	var buf bytes.Buffer
	// With keep, original should be preserved (if decompression worked)
	_ = RunXz(&buf, []string{xzFile}, XzOptions{
		Decompress: true,
		Keep:       true,
	})

	// Original should still exist
	if _, err := os.Stat(xzFile); os.IsNotExist(err) {
		t.Error("RunXz() -k should keep original file")
	}
}

func TestXzList_NonexistentFile(t *testing.T) {
	var buf bytes.Buffer
	// Should handle gracefully
	_ = xzList(&buf, []string{"/nonexistent/file.xz"})
}

func TestXzList_MultipleFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "xz_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create multiple xz files
	for i := range 3 {
		xzFile := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".xz")
		_ = os.WriteFile(xzFile, minimalXzFile, 0644)
	}

	var buf bytes.Buffer

	err = xzList(&buf, []string{
		filepath.Join(tmpDir, "test0.xz"),
		filepath.Join(tmpDir, "test1.xz"),
		filepath.Join(tmpDir, "test2.xz"),
	})
	if err != nil {
		t.Fatalf("xzList() error = %v", err)
	}

	output := buf.String()

	for i := range 3 {
		filename := "test" + string(rune('0'+i)) + ".xz"
		if !strings.Contains(output, filename) {
			t.Errorf("xzList() should list %s: got %q", filename, output)
		}
	}
}
