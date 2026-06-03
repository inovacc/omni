//go:build omni_sbomvalidate

package sbom_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	clisbom "github.com/inovacc/omni/internal/cli/sbom"
)

// TestRunSBOMValidateTagged confirms that with the omni_sbomvalidate build tag,
// --validate runs the in-process validator and a well-formed module SBOM passes
// (no error, output still produced). The default (no-tag) counterpart in
// sbom_test.go asserts the same call returns ErrUnsupported.
func TestRunSBOMValidateTagged(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"),
		[]byte("module github.com/example/app\n\ngo 1.25.0\n\nrequire github.com/spf13/cobra v1.10.2\n"), 0o644)
	var buf bytes.Buffer
	err := clisbom.RunSBOM(&buf, []string{dir}, clisbom.SBOMOptions{Format: "spdx", Validate: true, OmniVersion: "v0.1.0"})
	if err != nil {
		t.Fatalf("RunSBOM validate (tagged): %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected SBOM output after validation")
	}
}
