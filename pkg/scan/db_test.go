package scan

import (
	"archive/zip"
	"bytes"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/inovacc/omni/pkg/sign"
)

// itoa is a tiny helper so the in-memory manifest can carry an entry_count.
func itoa(n int) string { return strconv.Itoa(n) }

// buildZip assembles an in-memory osv-db.zip with one manifest + the given entry JSONs.
func buildZip(t *testing.T, generated time.Time, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	mw, err := zw.Create("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = mw.Write([]byte(`{"schema_version":"1.0","generated":"` +
		generated.UTC().Format(time.RFC3339) + `","ecosystem":"Go","entry_count":` +
		itoa(len(entries)) + `}`))
	for name, body := range entries {
		ew, err := zw.Create("entries/" + name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = ew.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// validEntry is a minimal OSV record carrying the required id/modified/schema_version.
func validEntry(id, name string) string {
	return `{"id":"` + id + `","modified":"2026-01-01T00:00:00Z","schema_version":"1.6.0",` +
		`"summary":"x","affected":[{"package":{"ecosystem":"Go","name":"` + name + `"},` +
		`"ranges":[{"type":"SEMVER","events":[{"introduced":"0"},{"fixed":"1.2.3"}]}]}]}`
}

func TestLoadDBAndAge(t *testing.T) {
	gen := time.Now().Add(-48 * time.Hour)
	z := buildZip(t, gen, map[string]string{
		"GO-1.json": validEntry("GO-1", "github.com/foo/bar"),
	})
	db, err := LoadDB(bytes.NewReader(z), int64(len(z)))
	if err != nil {
		t.Fatalf("LoadDB: %v", err)
	}
	if got := db.entriesFor("github.com/foo/bar"); len(got) != 1 {
		t.Fatalf("entriesFor = %d, want 1", len(got))
	}
	if db.Age() < 47*time.Hour {
		t.Errorf("Age too small: %v", db.Age())
	}
}

func TestStalenessGate(t *testing.T) {
	z := buildZip(t, time.Now().Add(-10*24*time.Hour), nil)
	db, err := LoadDB(bytes.NewReader(z), int64(len(z)))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CheckFresh(7 * 24 * time.Hour); err == nil {
		t.Fatal("CheckFresh must FAIL for a DB older than max age (fail loud)")
	}
	if err := db.CheckFresh(7 * 24 * time.Hour); !errors.Is(err, ErrStaleDB) {
		t.Errorf("CheckFresh error = %v, want ErrStaleDB", err)
	}
	if err := db.CheckFresh(30 * 24 * time.Hour); err != nil {
		t.Errorf("CheckFresh(30d) = %v, want nil", err)
	}
	if err := db.CheckFresh(0); err != nil {
		t.Errorf("CheckFresh(0) disables gate, want nil, got %v", err)
	}
}

func TestLoadDBBadZipFailsClosed(t *testing.T) {
	if _, err := LoadDB(bytes.NewReader([]byte("not a zip")), 9); err == nil {
		t.Fatal("LoadDB on garbage must fail closed")
	}
}

func TestLoadDBMissingManifestFailsClosed(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	ew, _ := zw.Create("entries/GO-1.json")
	_, _ = ew.Write([]byte(validEntry("GO-1", "github.com/foo/bar")))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	z := buf.Bytes()
	if _, err := LoadDB(bytes.NewReader(z), int64(len(z))); err == nil {
		t.Fatal("LoadDB without manifest.json must fail closed")
	}
}

// TestLoadDBValidatesRequiredFields confirms only id/modified/schema_version are
// required; unknown fields are tolerated (byte-passthrough), but an entry missing
// a required OSV field is rejected as ErrInvalidInput.
func TestLoadDBValidatesRequiredFields(t *testing.T) {
	// missing "id" -> reject
	z := buildZip(t, time.Now(), map[string]string{
		"bad.json": `{"modified":"2026-01-01T00:00:00Z","schema_version":"1.6.0","affected":[]}`,
	})
	if _, err := LoadDB(bytes.NewReader(z), int64(len(z))); err == nil {
		t.Fatal("entry missing id must be rejected")
	}

	// unknown fields tolerated, required fields present -> accepted
	z = buildZip(t, time.Now(), map[string]string{
		"ok.json": `{"id":"GO-9","modified":"2026-01-01T00:00:00Z","schema_version":"1.9.0",` +
			`"some_future_field":{"nested":true},"affected":[{"package":{"ecosystem":"Go","name":"github.com/a/b"}}]}`,
	})
	db, err := LoadDB(bytes.NewReader(z), int64(len(z)))
	if err != nil {
		t.Fatalf("entry with unknown fields must be tolerated: %v", err)
	}
	if len(db.entriesFor("github.com/a/b")) != 1 {
		t.Fatalf("forward-compatible entry not indexed")
	}
}

func TestVerifyAndLoadDBFailsClosed(t *testing.T) {
	kp, err := sign.GenerateKeyPair("p", sign.WithScryptParams(1<<15, 8, 1))
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	pubText := kp.PublicKey.MarshalText()

	z := buildZip(t, time.Now(), map[string]string{
		"GO-1.json": validEntry("GO-1", "github.com/foo/bar"),
	})
	sig, err := sign.Sign(z, kp.SecretKey)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	// Happy path: valid signature loads.
	db, err := VerifyAndLoadDB(z, sig, pubText)
	if err != nil {
		t.Fatalf("VerifyAndLoadDB (valid) = %v, want nil", err)
	}
	if len(db.entriesFor("github.com/foo/bar")) != 1 {
		t.Fatalf("entry not indexed after verified load")
	}

	// Tamper a byte in the zip bytes -> verification must fail, DB never loaded.
	tampered := make([]byte, len(z))
	copy(tampered, z)
	tampered[len(tampered)/2] ^= 0xFF
	if _, err := VerifyAndLoadDB(tampered, sig, pubText); err == nil {
		t.Fatal("VerifyAndLoadDB on tampered zip must fail closed")
	}

	// Bad public key text -> ErrInvalidInput-class parse failure, no load.
	if _, err := VerifyAndLoadDB(z, sig, []byte("not a key")); err == nil {
		t.Fatal("VerifyAndLoadDB with bad public key must fail")
	}
}
