//go:build !omni_sbomvalidate

package sbom

import "github.com/inovacc/omni/internal/cli/cmderr"

// validateDocument reports that schema validation is unavailable in the default
// build. The upstream SPDX/CycloneDX validators are heavy dependencies compiled
// only behind the omni_sbomvalidate build tag (see validate_on.go).
func validateDocument(_ []byte, _ string) error {
	return cmderr.Wrap(cmderr.ErrUnsupported, "sbom schema validation requires building with -tags omni_sbomvalidate")
}
