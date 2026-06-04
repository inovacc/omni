package format

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// withMaxSBOMBytes temporarily lowers the read ceiling so the limit can be
// exercised without allocating the full 256 MiB production cap.
func withMaxSBOMBytes(t *testing.T, n int64) {
	t.Helper()
	prev := maxSBOMBytes
	maxSBOMBytes = n
	t.Cleanup(func() { maxSBOMBytes = prev })
}

// TestParseRejectsOverCapInput is the regression for [sbom-parse-unbounded-readall]:
// Parse must reject (not truncate, not OOM) an SBOM larger than maxSBOMBytes.
// Pre-fix, Parse did io.ReadAll(r) with no bound and would happily parse this
// over-cap (but otherwise valid) JSON, so this assertion failed (RED).
func TestParseRejectsOverCapInput(t *testing.T) {
	withMaxSBOMBytes(t, 1024) // tiny cap keeps the test fast

	// Valid CycloneDX JSON padded with whitespace until it clears the cap.
	// Pre-fix this parses fine; post-fix it is rejected before unmarshal.
	pad := strings.Repeat(" ", int(maxSBOMBytes)+512)
	literal := `{"bomFormat":"CycloneDX","specVersion":"1.5",` + pad +
		`"components":[{"purl":"pkg:golang/github.com/spf13/cobra@v1.9.0"}]}`

	if int64(len(literal)) <= maxSBOMBytes {
		t.Fatalf("test setup: input %d <= cap %d", len(literal), maxSBOMBytes)
	}

	_, err := Parse(strings.NewReader(literal))
	if err == nil {
		t.Fatal("Parse accepted an over-cap input; want ErrTooLarge")
	}
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("Parse error = %v; want errors.Is(..., ErrTooLarge)", err)
	}
}

// TestParseAcceptsAtCapInput proves the boundary: input exactly at the cap is
// still accepted (the +1 LimitReader byte is what trips the limit).
func TestParseAcceptsAtCapInput(t *testing.T) {
	literal := `{"bomFormat":"CycloneDX","specVersion":"1.5","components":[]}`
	withMaxSBOMBytes(t, int64(len(literal))) // cap == exact input size

	if _, err := Parse(strings.NewReader(literal)); err != nil {
		t.Fatalf("Parse at exactly the cap should succeed; got %v", err)
	}
}

// TestParseLimitDoesNotDrainGiantReader proves Parse stops reading at the cap
// instead of draining an effectively unbounded stream into memory.
func TestParseLimitDoesNotDrainGiantReader(t *testing.T) {
	withMaxSBOMBytes(t, 4096)

	r := &countingReader{R: bytes.NewReader(bytes.Repeat([]byte("A"), 1<<20))}
	_, err := Parse(r)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("Parse error = %v; want ErrTooLarge", err)
	}
	// LimitReader must have stopped at cap+1, never reading the whole 1 MiB.
	if r.n > maxSBOMBytes+1 {
		t.Fatalf("Parse read %d bytes; want <= %d (cap+1)", r.n, maxSBOMBytes+1)
	}
}

// countingReader records how many bytes were read through it.
type countingReader struct {
	R io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.R.Read(p)
	c.n += int64(n)
	return n, err
}
