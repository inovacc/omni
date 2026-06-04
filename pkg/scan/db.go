package scan

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/inovacc/omni/pkg/sign"
)

// ErrStaleDB is returned by CheckFresh when the DB is older than the allowed age.
// The CLI maps it to cmderr.ErrConflict: a stale DB fails LOUDLY, it never
// silently degrades.
var ErrStaleDB = errors.New("vulnerability database is stale")

// ErrInvalidDB is returned when the bundle (zip, manifest, or an OSV entry) is
// malformed. The CLI maps it to cmderr.ErrInvalidInput.
var ErrInvalidDB = errors.New("invalid vulnerability database")

const entriesPrefix = "entries/"

// dbManifest is the required manifest.json carried by every osv-db.zip.
type dbManifest struct {
	SchemaVersion string `json:"schema_version"`
	Generated     string `json:"generated"`
	Ecosystem     string `json:"ecosystem"`
	EntryCount    int    `json:"entry_count"`
}

// DB is a loaded OSV database, indexed by Go module path.
type DB struct {
	generated time.Time
	byPkg     map[string][]osvEntry
}

// osvRequired is the minimal field set every OSV record must carry. omni
// validates only these (id, modified, schema_version) and tolerates all other
// fields, so future OSV minor versions parse without code changes.
type osvRequired struct {
	ID            string `json:"id"`
	Modified      string `json:"modified"`
	SchemaVersion string `json:"schema_version"`
}

// LoadDB parses an osv-db.zip from r (of size n). It fails closed on any
// malformed input: a non-zip, a missing manifest, an unparseable manifest
// timestamp, or an OSV entry missing a required field all return an error
// wrapping ErrInvalidDB. Signature verification happens BEFORE this call (see
// VerifyAndLoadDB); LoadDB itself performs no crypto.
//
// OSV entries are byte-passthrough: each entry/*.json is unmarshalled into the
// subset omni consumes (unknown fields ignored by encoding/json) AND validated
// for the OSV-required id/modified/schema_version only.
func LoadDB(r io.ReaderAt, n int64) (*DB, error) {
	zr, err := zip.NewReader(r, n)
	if err != nil {
		return nil, fmt.Errorf("%w: open osv db zip: %v", ErrInvalidDB, err)
	}
	db := &DB{byPkg: map[string][]osvEntry{}}
	var sawManifest bool
	for _, f := range zr.File {
		if f.Name != "manifest.json" && !strings.HasPrefix(f.Name, entriesPrefix) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("%w: read %s: %v", ErrInvalidDB, f.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, fmt.Errorf("%w: read %s: %v", ErrInvalidDB, f.Name, err)
		}
		switch {
		case f.Name == "manifest.json":
			var m dbManifest
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, fmt.Errorf("%w: parse manifest: %v", ErrInvalidDB, err)
			}
			t, err := time.Parse(time.RFC3339, m.Generated)
			if err != nil {
				return nil, fmt.Errorf("%w: parse manifest.generated: %v", ErrInvalidDB, err)
			}
			db.generated = t
			sawManifest = true
		default: // entries/<id>.json
			// (1) validate the OSV-required fields only.
			var req osvRequired
			if err := json.Unmarshal(data, &req); err != nil {
				return nil, fmt.Errorf("%w: parse %s: %v", ErrInvalidDB, f.Name, err)
			}
			if req.ID == "" || req.Modified == "" || req.SchemaVersion == "" {
				return nil, fmt.Errorf("%w: %s missing required OSV field (id/modified/schema_version)", ErrInvalidDB, f.Name)
			}
			// (2) byte-passthrough into the consumed subset; unknown fields tolerated.
			var e osvEntry
			if err := json.Unmarshal(data, &e); err != nil {
				return nil, fmt.Errorf("%w: parse %s: %v", ErrInvalidDB, f.Name, err)
			}
			for _, a := range e.Affected {
				if a.Package.Ecosystem == "Go" && a.Package.Name != "" {
					db.byPkg[a.Package.Name] = append(db.byPkg[a.Package.Name], e)
				}
			}
		}
	}
	if !sawManifest {
		return nil, fmt.Errorf("%w: osv db missing manifest.json", ErrInvalidDB)
	}
	return db, nil
}

// entriesFor returns the OSV entries indexed under a Go module path.
func (d *DB) entriesFor(pkg string) []osvEntry { return d.byPkg[pkg] }

// Age returns how long ago the DB was generated.
func (d *DB) Age() time.Duration { return time.Since(d.generated) }

// CheckFresh fails loudly (ErrStaleDB) when the DB is older than maxAge.
// maxAge <= 0 disables the gate.
func (d *DB) CheckFresh(maxAge time.Duration) error {
	if maxAge <= 0 {
		return nil
	}
	if d.Age() > maxAge {
		return fmt.Errorf("%w: generated %s ago (max %s)", ErrStaleDB, d.Age().Round(time.Second), maxAge)
	}
	return nil
}

// VerifyAndLoadDB verifies the detached minisig over zipBytes with the public
// key parsed from pubKeyText, then loads the DB. Verification is fail-closed:
// the DB is NEVER loaded on a bad signature. A signature mismatch wraps
// sign.ErrVerification (CLI maps to cmderr.ErrConflict); an unparseable public
// key wraps sign.ErrMalformed (CLI maps to cmderr.ErrInvalidInput).
func VerifyAndLoadDB(zipBytes, minisig, pubKeyText []byte) (*DB, error) {
	pub, err := sign.ParsePublicKey(pubKeyText)
	if err != nil {
		return nil, fmt.Errorf("parse db public key: %w", err)
	}
	if err := sign.Verify(zipBytes, minisig, pub); err != nil {
		return nil, fmt.Errorf("osv db signature verification failed: %w", err)
	}
	return LoadDB(bytes.NewReader(zipBytes), int64(len(zipBytes)))
}
