package tail

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

// TestTailLinesJSONHelper checks the circular-buffer line tail used by JSON mode.
func TestTailLinesJSONHelper(t *testing.T) {
	tests := []struct {
		name string
		in   string
		n    int
		want []string
	}{
		{"all", "a\nb\nc\n", 5, []string{"a", "b", "c"}},
		{"last two", "a\nb\nc\nd\n", 2, []string{"c", "d"}},
		{"zero", "a\nb\n", 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tailLinesJSON(strings.NewReader(tt.in), tt.n)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestRunTailJSON drives RunTail in JSON mode over a real file.
func TestRunTailJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "log.txt")
	if err := os.WriteFile(p, []byte("l1\nl2\nl3\nl4\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	opts := TailOptions{Lines: 2, OutputFormat: output.FormatJSON}
	if err := RunTail(&buf, nil, []string{p}, opts); err != nil {
		t.Fatal(err)
	}

	var results []TailResult
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, buf.String())
	}
	if len(results) != 1 || len(results[0].Lines) != 2 || results[0].Lines[1] != "l4" {
		t.Errorf("unexpected results: %+v", results)
	}
}

// TestTailBytesSeekable exercises the seekable fast path through a file.
func TestTailBytesSeekable(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "data.bin")
	if err := os.WriteFile(p, []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	if err := tailBytes(&buf, f, 3); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "789" {
		t.Errorf("tailBytes seekable = %q, want %q", buf.String(), "789")
	}
}

// TestTailBytesNonSeekable exercises the read-all fallback path.
func TestTailBytesNonSeekable(t *testing.T) {
	var buf bytes.Buffer
	if err := tailBytes(&buf, strings.NewReader("hello world"), 5); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "world" {
		t.Errorf("tailBytes non-seekable = %q, want %q", buf.String(), "world")
	}

	// n larger than data returns whole content.
	var buf2 bytes.Buffer
	if err := tailBytes(&buf2, strings.NewReader("hi"), 100); err != nil {
		t.Fatal(err)
	}
	if buf2.String() != "hi" {
		t.Errorf("tailBytes oversize = %q", buf2.String())
	}
}

// TestTailBytesSeekableOversize covers the seekable path when n exceeds size.
func TestTailBytesSeekableOversize(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "small.bin")
	if err := os.WriteFile(p, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	if err := tailBytes(&buf, f, 100); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "abc" {
		t.Errorf("seekable oversize = %q, want %q", buf.String(), "abc")
	}
}

// TestRunTailBytesMode drives RunTail with -c bytes mode from stdin.
func TestRunTailBytesMode(t *testing.T) {
	var buf bytes.Buffer
	in := strings.NewReader("abcdefghij")
	if err := RunTail(&buf, in, nil, TailOptions{Bytes: 4}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "ghij" {
		t.Errorf("RunTail -c = %q, want %q", buf.String(), "ghij")
	}
}

// TestRunTailLinesMode drives RunTail with default lines from stdin.
func TestRunTailLinesMode(t *testing.T) {
	var buf bytes.Buffer
	in := strings.NewReader("a\nb\nc\nd\ne\n")
	if err := RunTail(&buf, in, nil, TailOptions{Lines: 2}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "d\ne\n" {
		t.Errorf("RunTail -n = %q, want %q", buf.String(), "d\ne\n")
	}
}

// TestRunTailMissingFile verifies not-found classification.
func TestRunTailMissingFile(t *testing.T) {
	var buf bytes.Buffer
	err := RunTail(&buf, nil, []string{filepath.Join(t.TempDir(), "nope")}, TailOptions{})
	if err == nil {
		t.Error("expected error for missing file")
	}
}
