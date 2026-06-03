//go:build !omni_sbomvalidate

package sbom_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	clisbom "github.com/inovacc/omni/internal/cli/sbom"
)

// TestRunSBOMValidateNoTag pins the default-build behavior of --validate: with
// no omni_sbomvalidate build tag the schema validator is unavailable, so a
// Validate request must fail closed with ErrUnsupported (exit 6) rather than
// silently passing. The tagged positive counterpart lives in validate_on_test.go.
func TestRunSBOMValidateNoTag(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.25.0\n"), 0o644)
	err := clisbom.RunSBOM(&bytes.Buffer{}, []string{dir}, clisbom.SBOMOptions{Format: "spdx", Validate: true, OmniVersion: "v0.1.0"})
	if !cmderr.IsUnsupported(err) {
		t.Errorf("err = %v, want ErrUnsupported", err)
	}
}
