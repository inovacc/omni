package scan

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/sign"
)

// vulnEntry is an OSV record matching github.com/foo/bar < 1.2.3 at high severity.
const vulnEntry = `{"id":"GO-2026-0001","modified":"2026-01-01T00:00:00Z","schema_version":"1.6.0",` +
	`"summary":"example high vuln","affected":[{"package":{"ecosystem":"Go","name":"github.com/foo/bar"},` +
	`"ranges":[{"type":"SEMVER","events":[{"introduced":"0"},{"fixed":"1.2.3"}]}],` +
	`"severity":[{"type":"CVSS_V3","score":"7.5"}]}],` +
	`"severity":[{"type":"CVSS_V3","score":"7.5"}]}`

// vulnSBOM is a minimal CycloneDX-1.5 SBOM listing the vulnerable component at 1.0.0.
const vulnSBOM = `{"bomFormat":"CycloneDX","specVersion":"1.5","version":1,` +
	`"metadata":{"component":{"name":"app"}},` +
	`"components":[{"type":"library","name":"github.com/foo/bar","purl":"pkg:golang/github.com/foo/bar@v1.0.0"}]}`

// cleanSBOM lists only a non-matching component at a safe version.
const cleanSBOM = `{"bomFormat":"CycloneDX","specVersion":"1.5","version":1,` +
	`"metadata":{"component":{"name":"app"}},` +
	`"components":[{"type":"library","name":"github.com/safe/lib","purl":"pkg:golang/github.com/safe/lib@v2.0.0"}]}`

// buildSignedDB assembles an osv-db.zip (generated `age` in the past), signs it with a
// fresh low-cost key, and writes osv-db.zip / osv-db.zip.minisig / db.pub into dir.
// Returns the db path, key path, and the public key text.
func buildSignedDB(t *testing.T, dir string, age time.Duration) (dbPath, keyPath string) {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	mw, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	gen := time.Now().Add(-age).UTC().Format(time.RFC3339)
	_, _ = mw.Write([]byte(`{"schema_version":"1.0","generated":"` + gen +
		`","ecosystem":"Go","entry_count":` + strconv.Itoa(1) + `}`))
	ew, err := zw.Create("entries/GO-2026-0001.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = ew.Write([]byte(vulnEntry))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zipBytes := buf.Bytes()

	kp, err := sign.GenerateKeyPair("p", sign.WithScryptParams(1<<15, 8, 1))
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	sig, err := sign.Sign(zipBytes, kp.SecretKey)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	dbPath = filepath.Join(dir, "osv-db.zip")
	keyPath = filepath.Join(dir, "db.pub")
	if err := os.WriteFile(dbPath, zipBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dbPath+".minisig", sig, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, kp.PublicKey.MarshalText(), 0o644); err != nil {
		t.Fatal(err)
	}
	return dbPath, keyPath
}

func writeSBOM(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunScanReportNoGate(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)
	sbom := writeSBOM(t, dir, "vuln-sbom.json", vulnSBOM)

	var out bytes.Buffer
	err := RunScan(&out, []string{sbom}, Options{DBPath: dbPath, DBKeyPath: keyPath})
	if err != nil {
		t.Fatalf("RunScan (no gate) = %v, want nil", err)
	}
	if !strings.Contains(out.String(), "GO-2026-0001") {
		t.Errorf("output missing OSV id; got:\n%s", out.String())
	}
}

func TestRunScanFailOnHighTripsConflict(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)
	sbom := writeSBOM(t, dir, "vuln-sbom.json", vulnSBOM)

	var out bytes.Buffer
	err := RunScan(&out, []string{sbom}, Options{DBPath: dbPath, DBKeyPath: keyPath, FailOn: "high"})
	if err == nil || !cmderr.IsConflict(err) {
		t.Fatalf("RunScan(--fail-on high) = %v, want ErrConflict", err)
	}
	// Report must still be rendered before the gate error so CI sees both.
	if !strings.Contains(out.String(), "GO-2026-0001") {
		t.Errorf("report not rendered before gate error; got:\n%s", out.String())
	}
}

func TestRunScanFailOnCriticalBelowThreshold(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)
	sbom := writeSBOM(t, dir, "vuln-sbom.json", vulnSBOM)

	var out bytes.Buffer
	err := RunScan(&out, []string{sbom}, Options{DBPath: dbPath, DBKeyPath: keyPath, FailOn: "critical"})
	if err != nil {
		t.Errorf("RunScan(--fail-on critical) with only a high finding = %v, want nil", err)
	}
}

func TestRunScanClean(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)
	sbom := writeSBOM(t, dir, "clean-sbom.json", cleanSBOM)

	var out bytes.Buffer
	err := RunScan(&out, []string{sbom}, Options{DBPath: dbPath, DBKeyPath: keyPath, FailOn: "low"})
	if err != nil {
		t.Fatalf("RunScan(clean, --fail-on low) = %v, want nil", err)
	}
}

func TestRunScanJSON(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)
	sbom := writeSBOM(t, dir, "vuln-sbom.json", vulnSBOM)

	var out bytes.Buffer
	if err := RunScan(&out, []string{sbom}, Options{DBPath: dbPath, DBKeyPath: keyPath, JSON: true}); err != nil {
		t.Fatalf("RunScan(--json) = %v", err)
	}
	s := out.String()
	if !strings.Contains(s, `"findings"`) || !strings.Contains(s, `"GO-2026-0001"`) {
		t.Errorf("JSON output not as expected; got:\n%s", s)
	}
}

func TestRunScanMissingSBOM(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)

	err := RunScan(&bytes.Buffer{}, []string{filepath.Join(dir, "nope.json")},
		Options{DBPath: dbPath, DBKeyPath: keyPath})
	if err == nil || !cmderr.IsNotFound(err) {
		t.Fatalf("RunScan(missing sbom) = %v, want ErrNotFound", err)
	}
}

func TestRunScanBadSBOM(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)
	sbom := writeSBOM(t, dir, "bad-sbom.json", "{ this is : not json")

	err := RunScan(&bytes.Buffer{}, []string{sbom}, Options{DBPath: dbPath, DBKeyPath: keyPath})
	if err == nil || !cmderr.IsInvalidInput(err) {
		t.Fatalf("RunScan(bad sbom) = %v, want ErrInvalidInput", err)
	}
}

func TestRunScanMissingDB(t *testing.T) {
	dir := t.TempDir()
	sbom := writeSBOM(t, dir, "vuln-sbom.json", vulnSBOM)

	err := RunScan(&bytes.Buffer{}, []string{sbom},
		Options{DBPath: filepath.Join(dir, "absent.zip"), DBKeyPath: filepath.Join(dir, "absent.pub")})
	if err == nil || !cmderr.IsNotFound(err) {
		t.Fatalf("RunScan(missing db) = %v, want ErrNotFound", err)
	}
}

func TestRunScanTamperedDBFailsClosed(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)
	sbom := writeSBOM(t, dir, "vuln-sbom.json", vulnSBOM)

	// Flip a byte in the signed zip after signing.
	raw, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)/2] ^= 0xFF
	if err := os.WriteFile(dbPath, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	err = RunScan(&bytes.Buffer{}, []string{sbom}, Options{DBPath: dbPath, DBKeyPath: keyPath})
	if err == nil || !cmderr.IsConflict(err) {
		t.Fatalf("RunScan(tampered db) = %v, want ErrConflict (fail-closed)", err)
	}
}

func TestRunScanStaleDBFailsLoud(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 10*24*time.Hour)
	sbom := writeSBOM(t, dir, "vuln-sbom.json", vulnSBOM)

	err := RunScan(&bytes.Buffer{}, []string{sbom},
		Options{DBPath: dbPath, DBKeyPath: keyPath, MaxDBAge: 7 * 24 * time.Hour})
	if err == nil || !cmderr.IsConflict(err) {
		t.Fatalf("RunScan(stale db) = %v, want ErrConflict (ErrStaleDB)", err)
	}
}

func TestRunScanBadFailOnLabel(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)
	sbom := writeSBOM(t, dir, "vuln-sbom.json", vulnSBOM)

	err := RunScan(&bytes.Buffer{}, []string{sbom},
		Options{DBPath: dbPath, DBKeyPath: keyPath, FailOn: "bogus"})
	if err == nil || !cmderr.IsInvalidInput(err) {
		t.Fatalf("RunScan(bad --fail-on) = %v, want ErrInvalidInput", err)
	}
}

func TestRunScanNoArg(t *testing.T) {
	dir := t.TempDir()
	dbPath, keyPath := buildSignedDB(t, dir, 48*time.Hour)

	err := RunScan(&bytes.Buffer{}, nil, Options{DBPath: dbPath, DBKeyPath: keyPath})
	if err == nil || !cmderr.IsInvalidInput(err) {
		t.Fatalf("RunScan(no sbom arg) = %v, want ErrInvalidInput", err)
	}
}

func TestRunScanNoDB(t *testing.T) {
	dir := t.TempDir()
	sbom := writeSBOM(t, dir, "vuln-sbom.json", vulnSBOM)

	err := RunScan(&bytes.Buffer{}, []string{sbom}, Options{})
	if err == nil || !cmderr.IsInvalidInput(err) {
		t.Fatalf("RunScan(no --db) = %v, want ErrInvalidInput", err)
	}
}

func TestRunScanSourceUnsupported(t *testing.T) {
	// Reachability is deferred (ADR-0008): scan source returns ErrUnsupported
	// for any invocation, even without a DB configured (matches the CLI/golden
	// expectation `scan source ./...` -> exit 6).
	err := RunScanSource(&bytes.Buffer{}, []string{"./..."}, Options{})
	if err == nil || !cmderr.IsUnsupported(err) {
		t.Fatalf("RunScanSource = %v, want ErrUnsupported", err)
	}
}
