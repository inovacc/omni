// Experimental: pkg/scan is a new v1.0 surface; its API may change before it is marked stable.
//
// Package scan is a pure-Go, no-exec, offline-first vulnerability matcher. It
// matches an SBOM (pkg/sbom/format.Document) against a pkg/sign-signed OSV
// database bundle (osv-db.zip) using golang.org/x/mod/semver interval-walk
// matching, and reports Findings whose Severity is a normalized CVSS band.
//
// Severity normalization is dependency-free: a producer-supplied numeric CVSS
// base score is preferred, otherwise a CVSS v3.1 base score is computed in
// closed form from the vector metrics. CVSS v4.0 has no closed-form equation, so
// a v4 record contributes only its numeric score or SeverityUnknown. Findings
// with no usable severity are reported as SeverityUnknown, which sorts below
// SeverityLow for --fail-on gating but is never dropped.
//
// Boundary: pkg/scan imports pkg/sbom/format only, never pkg/sbom/model. The
// signed DB is verified (pkg/sign.Verify) before it is loaded, and a stale DB
// past --max-db-age fails loudly — the scanner is fail-closed throughout.
// Reachability source scanning is deferred from v1.0 per ADR-0008.
package scan
