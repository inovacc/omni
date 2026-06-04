//go:build ignore

// Command gen_fixtures generates the committed golden-master fixtures for the
// `scan` category: a pkg/sign-signed osv-db.zip (+ detached .minisig + db.pub),
// a tampered copy that fails signature verification, and three SBOM inputs.
//
// Run by hand (never in CI) to (re)materialize the fixtures in this directory:
//
//	go run testing/golden/fixtures/scan/gen_fixtures.go
//
// The bundle uses a FIXED manifest.generated far in the past so the staleness
// gate is exercisable; golden tests therefore never pass --max-db-age except the
// dedicated scan_stale_db case (whose drifting "generated <age> ago" message is
// normalized by strip_db_age). Signing uses a low-cost scrypt cost — NEVER the
// default SENSITIVE cost in automation.
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/inovacc/omni/pkg/sign"
)

const passphrase = "golden-fixture-passphrase"

// fixedModTime keeps zip entry timestamps stable if the bundle is regenerated.
var fixedModTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// One matching OSV entry (high severity, numeric score for byte-stable banding)
// and one non-matching entry, so the matcher reports exactly one finding for
// vuln-sbom.json and none for clean-sbom.json.
const manifestJSON = `{"schema_version":"1.0","generated":"2026-01-01T00:00:00Z","ecosystem":"Go","entry_count":2}`

const entryVuln = `{
  "schema_version": "1.6.0",
  "id": "GO-2026-0001",
  "summary": "Example high-severity vulnerability in github.com/vuln/pkg",
  "modified": "2026-01-01T00:00:00Z",
  "severity": [{"type": "CVSS_V3", "score": "7.5"}],
  "affected": [{
    "package": {"ecosystem": "Go", "name": "github.com/vuln/pkg", "purl": "pkg:golang/github.com/vuln/pkg"},
    "ranges": [{"type": "SEMVER", "events": [{"introduced": "0"}, {"fixed": "1.2.3"}]}]
  }]
}`

const entryOther = `{
  "schema_version": "1.6.0",
  "id": "GO-2026-0002",
  "summary": "Unrelated vulnerability in github.com/unrelated/dep",
  "modified": "2026-01-01T00:00:00Z",
  "severity": [{"type": "CVSS_V3", "score": "5.0"}],
  "affected": [{
    "package": {"ecosystem": "Go", "name": "github.com/unrelated/dep", "purl": "pkg:golang/github.com/unrelated/dep"},
    "ranges": [{"type": "SEMVER", "events": [{"introduced": "0"}, {"fixed": "0.9.0"}]}]
  }]
}`

const vulnSBOM = `{"bomFormat":"CycloneDX","specVersion":"1.5","metadata":{"component":{"name":"example-app"}},"components":[{"type":"library","name":"github.com/vuln/pkg","version":"v1.0.0","purl":"pkg:golang/github.com/vuln/pkg@v1.0.0"}]}`

const cleanSBOM = `{"bomFormat":"CycloneDX","specVersion":"1.5","metadata":{"component":{"name":"example-app"}},"components":[{"type":"library","name":"github.com/safe/lib","version":"v2.0.0","purl":"pkg:golang/github.com/safe/lib@v2.0.0"}]}`

const badSBOM = `{ this is not valid SBOM json `

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(thisFile)

	zipBytes := buildZip()

	lowCost := sign.WithScryptParams(1<<15, 8, 1)
	kp, err := sign.GenerateKeyPair(passphrase, lowCost)
	if err != nil {
		fatal("generate key pair: %v", err)
	}
	sig, err := sign.Sign(zipBytes, kp.SecretKey,
		sign.WithTrustedComment("omni golden scan fixture OSV DB"))
	if err != nil {
		fatal("sign db: %v", err)
	}

	write(filepath.Join(dir, "osv-db.zip"), zipBytes, 0o644)
	write(filepath.Join(dir, "osv-db.zip.minisig"), sig, 0o644)
	write(filepath.Join(dir, "db.pub"), kp.PublicKey.MarshalText(), 0o644)

	// Tampered bundle: flip a byte so the (unchanged) signature no longer
	// verifies -> fail-closed ErrConflict. Reuse the valid signature.
	tampered := append([]byte(nil), zipBytes...)
	tampered[len(tampered)/2] ^= 0xFF
	write(filepath.Join(dir, "osv-db.tampered.zip"), tampered, 0o644)
	write(filepath.Join(dir, "osv-db.tampered.zip.minisig"), sig, 0o644)

	write(filepath.Join(dir, "vuln-sbom.json"), []byte(vulnSBOM+"\n"), 0o644)
	write(filepath.Join(dir, "clean-sbom.json"), []byte(cleanSBOM+"\n"), 0o644)
	write(filepath.Join(dir, "bad-sbom.json"), []byte(badSBOM+"\n"), 0o644)

	fmt.Println("scan fixtures written to", dir)
}

func buildZip() []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name, body string) {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate, Modified: fixedModTime})
		if err != nil {
			fatal("zip create %s: %v", name, err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			fatal("zip write %s: %v", name, err)
		}
	}
	add("manifest.json", manifestJSON)
	add("entries/GO-2026-0001.json", entryVuln)
	add("entries/GO-2026-0002.json", entryOther)
	if err := zw.Close(); err != nil {
		fatal("zip close: %v", err)
	}
	return buf.Bytes()
}

func write(path string, b []byte, mode os.FileMode) {
	if err := os.WriteFile(path, b, mode); err != nil {
		fatal("write %s: %v", path, err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
