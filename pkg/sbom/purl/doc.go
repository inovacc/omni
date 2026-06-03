// Package purl builds package-URLs (purls) for Go modules in the form
// pkg:golang/<lowercased-module-path>[@<canonical-version>].
//
// Versions are canonicalized with golang.org/x/mod/semver when valid, while a
// "+incompatible" suffix is preserved and pseudo-versions pass through
// unchanged. Go module paths require no percent-encoding once lowercased, so
// the path is emitted verbatim after lowercasing. The special path "std" with a
// "goX.Y.Z" version represents the Go toolchain.
//
// Experimental: this package's API may change until Phase 3 triage promotes it.
package purl
