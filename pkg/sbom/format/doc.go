// Package format holds the stable cross-package SBOM boundary and the two
// deterministic emitters that marshal it to SPDX 2.3 JSON and CycloneDX 1.5
// JSON via the standard encoding/json package.
//
// Document is the stable v1.0 boundary type; importers outside pkg/sbom
// (notably pkg/scan in Phase 6) must depend only on this package, never on
// pkg/sbom/model. From resolves a model.SBOM into a Document with purls and
// SPDX-ID slugs computed, slices pre-sorted, and a content hash derived over
// the sorted purls. Encode then writes the requested format.
//
// Determinism contract: identical input bytes produce identical output bytes.
// Output uses struct field order (never maps) and pre-sorted slices, a fixed
// two-space indent, escaping disabled, and a trailing newline. There is no
// UUID, no random serialNumber, and no wall-clock timestamp — the creation
// timestamp defaults to the Unix epoch and is overridable via Options.SourceDate.
package format
