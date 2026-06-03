package sbom_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	clisbom "github.com/inovacc/omni/internal/cli/sbom"
)

func TestRunSBOMModuleSPDX(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"),
		[]byte("module github.com/example/app\n\ngo 1.25.0\n\nrequire github.com/spf13/cobra v1.10.2\n"), 0o644)
	var buf bytes.Buffer
	err := clisbom.RunSBOM(&buf, []string{dir}, clisbom.SBOMOptions{Format: "spdx", OmniVersion: "v0.1.0"})
	if err != nil {
		t.Fatalf("RunSBOM: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["spdxVersion"] != "SPDX-2.3" {
		t.Errorf("spdxVersion = %v", m["spdxVersion"])
	}
}

func TestRunSBOMBadFormat(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.25.0\n"), 0o644)
	err := clisbom.RunSBOM(&bytes.Buffer{}, []string{dir}, clisbom.SBOMOptions{Format: "xml"})
	if !cmderr.IsInvalidInput(err) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestRunSBOMMissingPath(t *testing.T) {
	err := clisbom.RunSBOM(&bytes.Buffer{}, []string{filepath.Join(t.TempDir(), "nope")}, clisbom.SBOMOptions{Format: "spdx"})
	if !cmderr.IsNotFound(err) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
