package scan

import (
	"archive/zip"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
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

// allowLoopback enables the test-only loopback exception for httptest servers
// (which bind to 127.0.0.1) and restores the prior state on cleanup.
func allowLoopback(t *testing.T) {
	t.Helper()
	prev := allowLoopbackFetch
	allowLoopbackFetch = true
	t.Cleanup(func() { allowLoopbackFetch = prev })
}

func TestDBUpdateVerifiesBeforeWrite(t *testing.T) {
	allowLoopback(t)
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
	allowLoopback(t)
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

// TestDBUpdateBodyTooLargeFailsClosed proves the download is bounded: a server
// that streams more than the (test-lowered) byte cap is rejected with
// cmderr.ErrIO and nothing is written to the cache. Without the io.LimitReader
// guard, fetch's bare io.ReadAll would buffer the whole body and the call would
// not error here (RED).
func TestDBUpdateBodyTooLargeFailsClosed(t *testing.T) {
	allowLoopback(t)

	// Lower the body cap for the duration of the test so we never allocate
	// hundreds of MiB.
	prev := maxFetchBytes
	maxFetchBytes = 1 << 10 // 1 KiB
	t.Cleanup(func() { maxFetchBytes = prev })

	oversized := bytes.Repeat([]byte("A"), int(maxFetchBytes)+4096)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(oversized)
	}))
	defer srv.Close()

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "db.pub")
	if err := os.WriteFile(keyPath, []byte("unused"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := RunDBUpdate(os.Stdout, Options{}, srv.URL, dir, keyPath)
	if err == nil {
		t.Fatal("RunDBUpdate(oversized body) must fail closed")
	}
	if !errors.Is(err, cmderr.ErrIO) {
		t.Errorf("oversized body error = %v, want cmderr.ErrIO", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "osv-db.zip")); !os.IsNotExist(statErr) {
		t.Error("oversized download must NOT be written to disk")
	}
}

// TestDBUpdateRedirectToLoopbackRefused proves redirect targets are validated:
// when loopback is NOT allowed, a 302 to a loopback host is refused before the
// body is read, classified as cmderr.ErrIO. Best-effort SSRF guard. Without the
// CheckRedirect host check, DefaultClient would silently follow the redirect.
func TestDBUpdateRedirectToLoopbackRefused(t *testing.T) {
	// Intentionally do NOT allow loopback here: the redirect target is a
	// loopback host and must be refused.
	loopback := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("secret-internal-data"))
	}))
	defer loopback.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, loopback.URL, http.StatusFound)
	}))
	defer redirector.Close()

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "db.pub")
	if err := os.WriteFile(keyPath, []byte("unused"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := RunDBUpdate(os.Stdout, Options{}, redirector.URL, dir, keyPath)
	if err == nil {
		t.Fatal("RunDBUpdate(redirect to loopback) must be refused")
	}
	if !errors.Is(err, cmderr.ErrIO) {
		t.Errorf("redirect-refused error = %v, want cmderr.ErrIO", err)
	}
	// The error chain must mention the SSRF/redirect refusal, not a generic
	// transport hiccup, so we know the guard fired.
	if !strings.Contains(err.Error(), "non-public") && !strings.Contains(err.Error(), "redirect") {
		t.Errorf("redirect-refused error %q should mention the redirect guard", err)
	}
}
