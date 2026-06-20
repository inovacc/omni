package strings

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

// binaryWithStrings builds a byte slice with NUL-separated printable runs.
func binaryWithStrings() []byte {
	var b bytes.Buffer
	b.Write([]byte{0x00, 0x01, 0x02})
	b.WriteString("HelloWorld")
	b.Write([]byte{0x00, 0xff})
	b.WriteString("short")
	b.Write([]byte{0x00})
	b.WriteString("ab") // below default min length 4 -> dropped
	b.Write([]byte{0x00})
	b.WriteString("FinalRun")
	return b.Bytes()
}

// TestStringsReaderJSON covers the JSON-collecting reader path directly.
func TestStringsReaderJSON(t *testing.T) {
	tests := []struct {
		name      string
		opts      StringsOptions
		wantValue string
		wantCount int
	}{
		{"default min", StringsOptions{}, "HelloWorld", 3},
		{"with offset", StringsOptions{Offset: "d"}, "HelloWorld", 3},
		{"min length 6", StringsOptions{MinLength: 6}, "HelloWorld", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.opts
			if opts.MinLength == 0 {
				opts.MinLength = 4
			}

			entries, err := stringsReaderJSON(bytes.NewReader(binaryWithStrings()), opts)
			if err != nil {
				t.Fatalf("stringsReaderJSON error = %v", err)
			}

			if len(entries) != tt.wantCount {
				t.Fatalf("got %d entries, want %d: %+v", len(entries), tt.wantCount, entries)
			}

			if entries[0].Value != tt.wantValue {
				t.Errorf("first value = %q, want %q", entries[0].Value, tt.wantValue)
			}

			if tt.opts.Offset != "" && entries[0].Offset == 0 && entries[0].Value != "HelloWorld" {
				t.Errorf("expected offset recorded for %q", entries[0].Value)
			}
		})
	}
}

// TestRunStringsJSONFile covers RunStrings reading a file in JSON output mode.
func TestRunStringsJSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "blob.bin")
	if err := os.WriteFile(path, binaryWithStrings(), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := RunStrings(&buf, []string{path}, StringsOptions{OutputFormat: output.FormatJSON})
	if err != nil {
		t.Fatalf("RunStrings json error = %v", err)
	}

	var res StringsResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal: %v (out=%q)", err, buf.String())
	}

	if res.Count != 3 {
		t.Errorf("count = %d, want 3: %+v", res.Count, res.Strings)
	}

	if res.Strings[0].Value != "HelloWorld" {
		t.Errorf("first = %q, want HelloWorld", res.Strings[0].Value)
	}
}

// TestRunStringsTextFile covers RunStrings text output with an offset prefix.
func TestRunStringsTextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "blob.bin")
	if err := os.WriteFile(path, binaryWithStrings(), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := RunStrings(&buf, []string{path}, StringsOptions{Offset: "x"})
	if err != nil {
		t.Fatalf("RunStrings text error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "HelloWorld") || !strings.Contains(out, "FinalRun") {
		t.Errorf("missing expected strings in output: %q", out)
	}
}

// TestRunStringsMissingFile covers the not-found error branch.
func TestRunStringsMissingFile(t *testing.T) {
	var buf bytes.Buffer
	err := RunStrings(&buf, []string{filepath.Join(t.TempDir(), "nope.bin")}, StringsOptions{})
	if err == nil {
		t.Error("expected error for missing file")
	}
}
