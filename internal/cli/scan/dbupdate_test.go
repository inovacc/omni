package scan

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/inovacc/omni/pkg/sign"
)

// buildSignedFixtureDB assembles an in-memory osv-db.zip, signs it with a
// fresh low-cost key, and returns the zip bytes, the detached signature, and
// the public-key text. It writes nothing to disk so the bytes can be served
// over httptest.
func buildSignedFixtureDB(t *testing.T) (zipBytes, sig, pub []byte) {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	mw, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	gen := time.Now().UTC().Format(time.RFC3339)
	_, _ = mw.Write([]byte(`{"schema_version":"1.0","generated":"` + gen +
		`","ecosystem":"Go","entry_count":1}`))
	ew, err := zw.Create("entries/GO-2026-0001.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = ew.Write([]byte(vulnEntry))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	zipBytes = buf.Bytes()

	kp, err := sign.GenerateKeyPair("p", sign.WithScryptParams(1<<15, 8, 1))
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	sig, err = sign.Sign(zipBytes, kp.SecretKey)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	return zipBytes, sig, kp.PublicKey.MarshalText()
}

func TestDBUpdateVerifiesBeforeWrite(t *testing.T) {
	zipBytes, sig, pub := buildSignedFixtureDB(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch filepath.Base(r.URL.Path) {
		case "osv-db.zip":
			_, _ = w.Write(zipBytes)
		case "osv-db.zip.minisig":
			_, _ = w.Write(sig)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "db.pub")
	if err := os.WriteFile(keyPath, pub, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := RunDBUpdate(os.Stdout, Options{}, srv.URL, dir, keyPath); err != nil {
		t.Fatalf("RunDBUpdate(good) = %v, want nil", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "osv-db.zip")); err != nil {
		t.Errorf("osv-db.zip not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "osv-db.zip.minisig")); err != nil {
		t.Errorf("osv-db.zip.minisig not written: %v", err)
	}
}

func TestDBUpdateTamperedFailsClosed(t *testing.T) {
	zipBytes, sig, pub := buildSignedFixtureDB(t)
	zipBytes[len(zipBytes)/2] ^= 0xFF // corrupt after signing
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if filepath.Base(r.URL.Path) == "osv-db.zip.minisig" {
			_, _ = w.Write(sig)
			return
		}
		_, _ = w.Write(zipBytes)
	}))
	defer srv.Close()

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "db.pub")
	if err := os.WriteFile(keyPath, pub, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RunDBUpdate(os.Stdout, Options{}, srv.URL, dir, keyPath); err == nil {
		t.Fatal("RunDBUpdate(tampered) must fail closed")
	}
	if _, err := os.Stat(filepath.Join(dir, "osv-db.zip")); !os.IsNotExist(err) {
		t.Error("tampered DB must NOT be written to disk")
	}
	if _, err := os.Stat(filepath.Join(dir, "osv-db.zip.minisig")); !os.IsNotExist(err) {
		t.Error("tampered DB signature must NOT be written to disk")
	}
}
