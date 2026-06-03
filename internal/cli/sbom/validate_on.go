//go:build omni_sbomvalidate

package sbom

import (
	"encoding/json"
	"fmt"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// validateDocument checks that the emitted document is well-formed in the
// omni_sbomvalidate build. This is a pure-stdlib placeholder: it confirms the
// bytes decode as JSON so the tagged build compiles and round-trips. Phase 05
// Task 8 replaces this with the upstream SPDX/CycloneDX schema decoders (added
// only behind this same build tag), keeping the heavy dependencies out of the
// default build path.
func validateDocument(data []byte, _ string) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("sbom: emitted document failed validation: %v", err))
	}
	return nil
}
