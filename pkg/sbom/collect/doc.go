// Package collect builds a source-format-agnostic model.SBOM from a Go project,
// either from a module directory (go.mod require blocks, via ModuleDir) or from
// a built Go binary (embedded build info, via BinaryFile/Binary).
//
// The module-dir collector uses a small stdlib line scanner over go.mod rather
// than golang.org/x/mod/modfile, keeping the dependency surface to semver only.
// replace/exclude/retract directives are ignored for module-dir SBOMs in this
// phase; replace resolution is performed only for binary SBOMs.
//
// Experimental: this package's API may change until Phase 3 triage promotes it.
package collect
