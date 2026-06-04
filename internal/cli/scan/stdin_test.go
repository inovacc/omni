package scan

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// TestRunScanStdinReportNoGate scans an SBOM read from an io.Reader (the pipe
// path) and, with no --fail-on, renders the report without returning an error.
func TestRunScanStdinReportNoGate(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)

	var out bytes.Buffer
	err := RunScanStdin(&out, strings.NewReader(vulnSBOM), nil,
		Options{DBPath: dbPath, DBKeyPath: keyPath})
	if err != nil {
		t.Fatalf("RunScanStdin (no gate) = %v, want nil", err)
	}
	if !strings.Contains(out.String(), "GO-2026-0001") {
		t.Errorf("output missing OSV id; got:\n%s", out.String())
	}
}

// TestRunScanStdinFailOnHighTripsConflict confirms the --fail-on gate trips
// ErrConflict on the stdin path too, and the report is rendered first.
func TestRunScanStdinFailOnHighTripsConflict(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)

	var out bytes.Buffer
	err := RunScanStdin(&out, strings.NewReader(vulnSBOM), nil,
		Options{DBPath: dbPath, DBKeyPath: keyPath, FailOn: "high"})
	if err == nil || !cmderr.IsConflict(err) {
		t.Fatalf("RunScanStdin(--fail-on high) = %v, want ErrConflict", err)
	}
	if !strings.Contains(out.String(), "GO-2026-0001") {
		t.Errorf("report not rendered before gate error; got:\n%s", out.String())
	}
}

// TestRunScanStdinBadSBOM confirms a malformed SBOM on stdin classifies as
// ErrInvalidInput, matching the file path.
func TestRunScanStdinBadSBOM(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)

	err := RunScanStdin(&bytes.Buffer{}, strings.NewReader("{ not json"), nil,
		Options{DBPath: dbPath, DBKeyPath: keyPath})
	if err == nil || !cmderr.IsInvalidInput(err) {
		t.Fatalf("RunScanStdin(bad sbom) = %v, want ErrInvalidInput", err)
	}
}

// TestRunScanStdinNoDB confirms a missing --db (no env, empty Options) is
// ErrInvalidInput.
func TestRunScanStdinNoDB(t *testing.T) {
	err := RunScanStdin(&bytes.Buffer{}, strings.NewReader(vulnSBOM), nil, Options{})
	if err == nil || !cmderr.IsInvalidInput(err) {
		t.Fatalf("RunScanStdin(no --db) = %v, want ErrInvalidInput", err)
	}
}

// TestOptionsFromEnv reads the OMNI_SCAN_* environment variables into Options.
func TestOptionsFromEnv(t *testing.T) {
	t.Setenv("OMNI_SCAN_DB", "/tmp/osv-db.zip")
	t.Setenv("OMNI_SCAN_DB_KEY", "/tmp/db.pub")
	t.Setenv("OMNI_SCAN_FAIL_ON", "high")
	t.Setenv("OMNI_SCAN_MAX_DB_AGE", "168h")

	opts := OptionsFromEnv()
	if opts.DBPath != "/tmp/osv-db.zip" {
		t.Errorf("DBPath = %q, want /tmp/osv-db.zip", opts.DBPath)
	}
	if opts.DBKeyPath != "/tmp/db.pub" {
		t.Errorf("DBKeyPath = %q, want /tmp/db.pub", opts.DBKeyPath)
	}
	if opts.FailOn != "high" {
		t.Errorf("FailOn = %q, want high", opts.FailOn)
	}
	if opts.MaxDBAge != 168*time.Hour {
		t.Errorf("MaxDBAge = %v, want 168h", opts.MaxDBAge)
	}
}

// TestOptionsFromEnvEmpty confirms an unset environment yields a zero Options.
func TestOptionsFromEnvEmpty(t *testing.T) {
	t.Setenv("OMNI_SCAN_DB", "")
	t.Setenv("OMNI_SCAN_DB_KEY", "")
	t.Setenv("OMNI_SCAN_FAIL_ON", "")
	t.Setenv("OMNI_SCAN_MAX_DB_AGE", "")

	opts := OptionsFromEnv()
	if opts.DBPath != "" || opts.DBKeyPath != "" || opts.FailOn != "" || opts.MaxDBAge != 0 {
		t.Errorf("OptionsFromEnv() with empty env = %+v, want zero Options", opts)
	}
}

// TestOptionsFromEnvBadDuration confirms an unparseable OMNI_SCAN_MAX_DB_AGE is
// ignored (left at zero) rather than panicking — the staleness gate simply stays
// off, matching the duration flag default.
func TestOptionsFromEnvBadDuration(t *testing.T) {
	t.Setenv("OMNI_SCAN_MAX_DB_AGE", "not-a-duration")
	opts := OptionsFromEnv()
	if opts.MaxDBAge != 0 {
		t.Errorf("MaxDBAge with bad env = %v, want 0", opts.MaxDBAge)
	}
}
