// Package scan implements the I/O glue for the `omni scan` command and its
// `source` subcommand. It bridges Cobra to pkg/scan: reading an SBOM file (SPDX
// 2.3 / CycloneDX 1.5 JSON) via pkg/sbom/format, loading and signature-verifying
// a signed OSV database bundle via pkg/scan, applying the staleness and
// --fail-on gates, and rendering findings as a text table or JSON.
//
// All failure paths are classified into internal/cli/cmderr sentinels so the
// root command maps them to stable exit codes (see ADR-0008). Reachability
// scanning (`omni scan source`) is deferred from v1.0 and surfaces the
// pkg/scan.ScanSource ErrUnsupported verbatim.
package scan

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/sbom/format"
	"github.com/inovacc/omni/pkg/scan"
	"github.com/inovacc/omni/pkg/sign"
)

// Options configures `omni scan` and `omni scan source`.
type Options struct {
	// DBPath is the path to the signed osv-db.zip bundle.
	DBPath string
	// DBKeyPath is the path to the minisign public key (*.pub) used to verify the
	// DB bundle.
	DBKeyPath string
	// DBSigPath is the detached signature path. When empty it defaults to
	// "<DBPath>.minisig".
	DBSigPath string
	// FailOn is the severity threshold label ("low"/"medium"/"high"/"critical").
	// When non-empty, any finding at or above it trips cmderr.ErrConflict.
	FailOn string
	// JSON selects JSON output instead of the text table.
	JSON bool
	// MaxDBAge gates DB staleness. A DB older than MaxDBAge fails loudly
	// (cmderr.ErrConflict). Zero disables the gate.
	MaxDBAge time.Duration
	// Online enables OSV-API enrichment over net/http (opt-in). It is accepted
	// for signature stability; the default path is fully offline.
	Online bool
	// Reachability is reserved for the future contrib reachability path; it is
	// ignored in v1.0 (scan source returns ErrUnsupported per ADR-0008).
	Reachability bool
}

// RunScan reads the SBOM named by args[0], scans it against the signed OSV DB
// resolved from opts, and renders the report to w. The report is always
// rendered to w BEFORE any --fail-on gate error is returned, so CI sees both the
// findings and a non-zero exit code.
func RunScan(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 || args[0] == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "scan: an SBOM file path is required")
	}

	doc, err := readSBOM(args[0])
	if err != nil {
		return err
	}
	return scanDoc(w, doc, opts)
}

// RunScanStdin is the pipe-stage entry point: it reads an SBOM from r (instead of
// a file) and runs the same matcher/gate path as RunScan. Pipe stages take only
// (w, r, args), so the DB path/key and gate come from opts (typically built by
// OptionsFromEnv). The SBOM parse error classifies as ErrInvalidInput; the report
// is rendered to w BEFORE any --fail-on gate error is returned.
func RunScanStdin(w io.Writer, r io.Reader, _ []string, opts Options) error {
	doc, err := format.Parse(r)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("scan: parse SBOM from stdin: %v", err))
	}
	return scanDoc(w, doc, opts)
}

// scanDoc loads the DB from opts, runs the matcher with the resolved --fail-on
// gate, and renders the report to w before returning the gate error. It is shared
// by RunScan (file input) and RunScanStdin (reader input).
func scanDoc(w io.Writer, doc *format.Document, opts Options) error {
	db, err := loadDB(opts)
	if err != nil {
		return err
	}

	sev, err := resolveFailOn(opts.FailOn)
	if err != nil {
		return err
	}

	report, scanErr := scan.Scan(doc, db, scan.Options{FailOn: sev})
	if rErr := render(w, report, opts.JSON); rErr != nil {
		return rErr
	}
	return scanErr
}

// OptionsFromEnv builds Options from the OMNI_SCAN_* environment variables, used
// by the `omni pipe` stage where flags are unavailable:
//
//	OMNI_SCAN_DB         -> DBPath (path to osv-db.zip)
//	OMNI_SCAN_DB_KEY     -> DBKeyPath (minisign public key)
//	OMNI_SCAN_FAIL_ON    -> FailOn (severity threshold label)
//	OMNI_SCAN_MAX_DB_AGE -> MaxDBAge (a time.ParseDuration string, e.g. "168h")
//
// An unset variable leaves its field at the zero value; an unparseable
// OMNI_SCAN_MAX_DB_AGE is ignored (the staleness gate stays off), matching the
// --max-db-age flag default rather than failing the whole pipe stage.
func OptionsFromEnv() Options {
	opts := Options{
		DBPath:    os.Getenv("OMNI_SCAN_DB"),
		DBKeyPath: os.Getenv("OMNI_SCAN_DB_KEY"),
		FailOn:    os.Getenv("OMNI_SCAN_FAIL_ON"),
	}
	if age := os.Getenv("OMNI_SCAN_MAX_DB_AGE"); age != "" {
		if d, err := time.ParseDuration(age); err == nil {
			opts.MaxDBAge = d
		}
	}
	return opts
}

// RunScanSource performs a reachability-aware source scan. Reachability is
// deferred from v1.0 (ADR-0008): pkg/scan.ScanSource always returns
// ErrUnsupported, surfaced verbatim here. Because the feature is unavailable
// regardless of inputs, no DB is loaded and no flags are validated first — the
// command reports "unsupported" (exit 6) for any invocation.
func RunScanSource(w io.Writer, args []string, opts Options) error {
	pattern := ""
	if len(args) > 0 {
		pattern = args[0]
	}
	report, scanErr := scan.ScanSource(pattern, nil, scan.Options{Reachability: true})
	if scanErr != nil {
		return scanErr // ErrUnsupported (deferred per ADR-0008)
	}
	return render(w, report, opts.JSON)
}

// dbZipName / dbSigName are the canonical file names of the signed OSV bundle and
// its detached signature, both as served by the release URL and as cached on disk.
const (
	dbZipName = "osv-db.zip"
	dbSigName = "osv-db.zip.minisig"
)

// RunDBUpdate downloads the signed OSV bundle (osv-db.zip + osv-db.zip.minisig)
// from baseURL, verifies the detached signature with the pinned public key at
// keyPath, and writes BOTH files into cacheDir ONLY if verification passes. A
// tampered or unsigned download is fail-closed: nothing is written and a
// classified error (cmderr.ErrConflict for a bad signature) is returned.
//
// Network use is net/http only — pure-Go, no os/exec. The opts argument is
// accepted for signature stability with the other Run* entry points; the update
// path is unconditionally offline-fetch-then-verify and consults no Options
// fields today.
func RunDBUpdate(w io.Writer, _ Options, baseURL, cacheDir, keyPath string) error {
	if baseURL == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "scan db update: a download URL is required")
	}
	if cacheDir == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "scan db update: a cache directory is required")
	}
	if keyPath == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "scan db update: --db-key (minisign public key) is required")
	}

	base := strings.TrimRight(baseURL, "/")
	zipBytes, err := fetch(base + "/" + dbZipName)
	if err != nil {
		return err
	}
	sig, err := fetch(base + "/" + dbSigName)
	if err != nil {
		return err
	}

	pubText, err := os.ReadFile(keyPath)
	if err != nil {
		return classifyFileErr(err, keyPath)
	}

	// Verify BEFORE writing anything: a bad signature must leave the cache untouched.
	db, err := scan.VerifyAndLoadDB(zipBytes, sig, pubText)
	if err != nil {
		return classifyDBErr(err)
	}

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return classifyFileErr(err, cacheDir)
	}
	if err := writeAtomic(filepath.Join(cacheDir, dbZipName), zipBytes); err != nil {
		return err
	}
	if err := writeAtomic(filepath.Join(cacheDir, dbSigName), sig); err != nil {
		return err
	}

	count, generated := manifestSummary(zipBytes)
	if count < 0 {
		// Fall back to the indexed package count if the manifest is unreadable
		// (it can't be: VerifyAndLoadDB already parsed it) — keep the message robust.
		count = 0
	}
	_, err = fmt.Fprintf(w, "db updated: %d entries, generated %s (age %s)\n",
		count, generated, db.Age().Round(time.Second))
	return err
}

// fetch GETs url into memory. A transport error or non-200 status is classified
// as cmderr.ErrIO — the update could not be completed, but nothing is written.
func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec // URL is operator-provided (--url), not attacker-controlled.
	if err != nil {
		return nil, cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("scan db update: fetch %s: %v", url, err))
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("scan db update: fetch %s: status %s", url, resp.Status))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("scan db update: read %s: %v", url, err))
	}
	return data, nil
}

// writeAtomic writes data to path via a temp file in the same directory followed
// by os.Rename, so a partially written cache file is never observable.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".osv-db-*.tmp")
	if err != nil {
		return classifyFileErr(err, path)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("scan db update: write %s: %v", path, err))
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("scan db update: close %s: %v", path, err))
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		_ = os.Remove(tmpName)
		return classifyFileErr(err, path)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("scan db update: rename to %s: %v", path, err))
	}
	return nil
}

// manifestSummary best-effort reads entry_count and generated from the bundle's
// manifest.json for the success message. It returns (-1, "") if the manifest is
// unreadable; callers fall back gracefully.
func manifestSummary(zipBytes []byte) (entryCount int, generated string) {
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return -1, ""
	}
	for _, f := range zr.File {
		if f.Name != "manifest.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return -1, ""
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return -1, ""
		}
		var m struct {
			Generated  string `json:"generated"`
			EntryCount int    `json:"entry_count"`
		}
		if err := json.Unmarshal(data, &m); err != nil {
			return -1, ""
		}
		return m.EntryCount, m.Generated
	}
	return -1, ""
}

// readSBOM reads and parses an SBOM file into a format.Document, classifying
// file and parse errors into cmderr sentinels.
func readSBOM(path string) (*format.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, classifyFileErr(err, path)
	}
	defer func() { _ = f.Close() }()

	doc, err := format.Parse(f)
	if err != nil {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("scan: parse SBOM %s: %v", path, err))
	}
	return doc, nil
}

// loadDB resolves the DB zip, its detached signature, and the public key, then
// verifies and loads the bundle, applying the staleness gate. Every failure is
// classified into a cmderr sentinel.
func loadDB(opts Options) (*scan.DB, error) {
	if opts.DBPath == "" {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, "scan: --db (path to osv-db.zip) is required")
	}
	if opts.DBKeyPath == "" {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, "scan: --db-key (minisign public key) is required")
	}

	sigPath := opts.DBSigPath
	if sigPath == "" {
		sigPath = opts.DBPath + ".minisig"
	}

	zipBytes, err := os.ReadFile(opts.DBPath)
	if err != nil {
		return nil, classifyFileErr(err, opts.DBPath)
	}
	sig, err := os.ReadFile(sigPath)
	if err != nil {
		return nil, classifyFileErr(err, sigPath)
	}
	pubText, err := os.ReadFile(opts.DBKeyPath)
	if err != nil {
		return nil, classifyFileErr(err, opts.DBKeyPath)
	}

	db, err := scan.VerifyAndLoadDB(zipBytes, sig, pubText)
	if err != nil {
		return nil, classifyDBErr(err)
	}

	if err := db.CheckFresh(opts.MaxDBAge); err != nil {
		return nil, cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("scan: %v", err))
	}
	return db, nil
}

// resolveFailOn maps the --fail-on label to a scan.Severity. An empty label
// disables the gate (SeverityUnknown); an unrecognized label is ErrInvalidInput.
func resolveFailOn(label string) (scan.Severity, error) {
	if label == "" {
		return scan.SeverityUnknown, nil
	}
	sev, ok := scan.ParseSeverity(label)
	if !ok {
		return scan.SeverityUnknown, cmderr.Wrap(cmderr.ErrInvalidInput,
			fmt.Sprintf("scan: unknown --fail-on severity %q (want none/low/medium/high/critical)", label))
	}
	return sev, nil
}

// render writes the report as JSON (opts.JSON) or a stable text table.
func render(w io.Writer, report scan.Report, asJSON bool) error {
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("scan: encode JSON: %v", err))
		}
		return nil
	}
	return renderText(w, report)
}

// renderText prints a stable, path-free, age-free text table. DBAge is
// intentionally omitted from text output so the rendering is byte-deterministic
// for golden tests; it is available only via --json.
func renderText(w io.Writer, report scan.Report) error {
	if len(report.Findings) == 0 {
		_, err := fmt.Fprintf(w, "no vulnerabilities found (%d components scanned)\n", report.Scanned)
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tPACKAGE\tVERSION\tFIXED\tSEVERITY\tSUMMARY"); err != nil {
		return err
	}
	for _, f := range report.Findings {
		fixed := f.FixedVersion
		if fixed == "" {
			fixed = "-"
		}
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			f.ID, f.Package, f.Version, fixed, f.Severity, f.Summary); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "%d vulnerabilit%s found in %d components scanned\n",
		len(report.Findings), plural(len(report.Findings)), report.Scanned)
	return err
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}

// classifyFileErr maps an os file error to a cmderr sentinel.
func classifyFileErr(err error, path string) error {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("scan: %s", path))
	case errors.Is(err, os.ErrPermission):
		return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("scan: %s", path))
	default:
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("scan: %s: %v", path, err))
	}
}

// classifyDBErr maps a pkg/scan or pkg/sign DB error to a cmderr sentinel:
// a signature mismatch fails closed as ErrConflict; a malformed key or bundle
// is ErrInvalidInput.
func classifyDBErr(err error) error {
	switch {
	case errors.Is(err, sign.ErrVerification):
		return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("scan: %v", err))
	case errors.Is(err, sign.ErrMalformed):
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("scan: %v", err))
	case errors.Is(err, scan.ErrInvalidDB):
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("scan: %v", err))
	case errors.Is(err, scan.ErrStaleDB):
		return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("scan: %v", err))
	default:
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("scan: load OSV database: %v", err))
	}
}
