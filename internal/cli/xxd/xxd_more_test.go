package xxd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDumpString verifies the pure DumpString helper produces a standard dump.
func TestDumpString(t *testing.T) {
	out, err := DumpString([]byte("ABC"))
	if err != nil {
		t.Fatal(err)
	}
	// Offset prefix, hex for A(41) B(42) C(43), and ASCII tail.
	if !strings.HasPrefix(out, "00000000: ") {
		t.Errorf("missing offset prefix: %q", out)
	}
	if !strings.Contains(out, "4142 43") {
		t.Errorf("missing hex bytes: %q", out)
	}
	if !strings.Contains(out, "ABC") {
		t.Errorf("missing ASCII: %q", out)
	}
}

// TestDumpStringEmpty confirms empty input dumps to empty output.
func TestDumpStringEmpty(t *testing.T) {
	out, err := DumpString(nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("expected empty dump, got %q", out)
	}
}

// NOTE: Dump (xxd.go:536) is intentionally NOT tested here. It calls
// Run(w, nil, nil, DefaultOptions()) — passing a nil reader and nil args — so
// runDump receives a nil io.Reader and panics on io.ReadFull. It is not
// exercisable offline without modifying production code, which is out of scope.

// TestRunHexDump drives Run in default hex mode over a real file with uppercase.
func TestRunHexDump(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.bin")
	if err := os.WriteFile(p, []byte("Hello!"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := DefaultOptions()
	opts.Uppercase = true
	if err := Run(&buf, nil, []string{p}, opts); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "4865") { // 'H' 'e' uppercase 48 65
		t.Errorf("uppercase hex missing: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "Hello!") {
		t.Errorf("ASCII tail missing: %q", buf.String())
	}
}

// TestRunBitsMode drives Run with -b binary digit dump.
func TestRunBitsMode(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultOptions()
	opts.Bits = true
	opts.Columns = 2
	if err := Run(&buf, strings.NewReader("AB"), nil, opts); err != nil {
		t.Fatal(err)
	}
	// 'A' = 0x41 = 01000001
	if !strings.Contains(buf.String(), "01000001") {
		t.Errorf("binary dump missing: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "AB") {
		t.Errorf("ASCII tail missing: %q", buf.String())
	}
}

// TestRunLengthLimit ensures -l caps output.
func TestRunLengthLimit(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultOptions()
	opts.Length = 3
	if err := Run(&buf, strings.NewReader("ABCDEFGHIJ"), nil, opts); err != nil {
		t.Fatal(err)
	}
	// Only "ABC" dumped -> hex "4142 43" present (group size 2), "DEF" absent.
	if !strings.Contains(buf.String(), "4142 43") {
		t.Errorf("expected first 3 bytes: %q", buf.String())
	}
	if strings.Contains(buf.String(), "DEF") {
		t.Errorf("length limit not honored: %q", buf.String())
	}
}

// TestRunMissingFile verifies not-found handling.
func TestRunMissingFile(t *testing.T) {
	var buf bytes.Buffer
	err := Run(&buf, nil, []string{filepath.Join(t.TempDir(), "nope")}, DefaultOptions())
	if err == nil {
		t.Error("expected error for missing file")
	}
}
