// Package model holds the source-format-agnostic SBOM representation produced
// by the collectors (module-dir and binary) and consumed by the SPDX 2.3 and
// CycloneDX 1.5 emitters.
//
// An SBOM is a Root Component plus a list of dependency Components (and, for
// binary SBOMs, the Go toolchain). Normalize sorts Components by (Path,Version)
// and drops exact duplicates so downstream emission is byte-deterministic. Slug
// converts a module path into an SPDX-ID-safe token.
//
// Experimental: this package's API may change until Phase 3 triage promotes it.
package model
