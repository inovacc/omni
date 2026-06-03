package purl

import (
	"strings"

	"golang.org/x/mod/semver"
)

// ForModule returns the package-url for a Go module given its path and version.
// Format: pkg:golang/<lowercased-module-path>[@<canonical-version>].
// Pseudo-versions pass through; a "+incompatible" suffix is preserved; an empty
// version yields a purl with no "@" suffix. Special path "std" + a "goX.Y.Z"
// version represents the Go toolchain.
func ForModule(modulePath, version string) string {
	p := "pkg:golang/" + strings.ToLower(modulePath)
	v := normalizeVersion(version)
	if v == "" {
		return p
	}
	return p + "@" + v
}

// normalizeVersion canonicalizes a semver while preserving "+incompatible";
// non-semver values (e.g. "go1.25.0", "(devel)") pass through unchanged.
func normalizeVersion(version string) string {
	if version == "" {
		return ""
	}
	base, incompat := version, ""
	if strings.HasSuffix(version, "+incompatible") {
		base = strings.TrimSuffix(version, "+incompatible")
		incompat = "+incompatible"
	}
	if semver.IsValid(base) {
		return semver.Canonical(base) + incompat
	}
	return version
}
