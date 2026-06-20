package sbom_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	clisbom "github.com/inovacc/omni/internal/cli/sbom"
	"github.com/inovacc/omni/pkg/sign"
)

// writeModule creates a minimal Go module directory and returns its path.
func writeModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"),
		[]byte("module github.com/example/app\n\ngo 1.25.0\n\nrequire github.com/spf13/cobra v1.10.2\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	return dir
}

// TestRunSBOMCycloneDX covers the CycloneDX branch of parseFormat + encoding.
func TestRunSBOMCycloneDX(t *testing.T) {
	for _, f := range []string{"cyclonedx", "cdx"} {
		t.Run(f, func(t *testing.T) {
			dir := writeModule(t)
			var buf bytes.Buffer
			err := clisbom.RunSBOM(&buf, []string{dir}, clisbom.SBOMOptions{Format: f, OmniVersion: "v0.1.0"})
			if err != nil {
				t.Fatalf("RunSBOM(%s): %v", f, err)
			}
			var m map[string]any
			if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}
			if m["bomFormat"] != "CycloneDX" {
				t.Errorf("bomFormat = %v, want CycloneDX", m["bomFormat"])
			}
		})
	}
}

// TestRunSBOMArgsAndFromConflicts sweeps the argument-count and --from
// conflict-classification branches.
func TestRunSBOMArgsAndFromConflicts(t *testing.T) {
	dir := writeModule(t)
	file := filepath.Join(t.TempDir(), "binary.bin")
	if err := os.WriteFile(file, []byte("not really a binary"), 0o644); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	tests := []struct {
		name string
		args []string
		opts clisbom.SBOMOptions
	}{
		{"no args", nil, clisbom.SBOMOptions{Format: "spdx"}},
		{"two args", []string{dir, dir}, clisbom.SBOMOptions{Format: "spdx"}},
		{"from module needs dir", []string{file}, clisbom.SBOMOptions{From: "module"}},
		{"from binary needs file", []string{dir}, clisbom.SBOMOptions{From: "binary"}},
		{"unknown from", []string{dir}, clisbom.SBOMOptions{From: "weird"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := clisbom.RunSBOM(&bytes.Buffer{}, tt.args, tt.opts)
			if !cmderr.IsInvalidInput(err) {
				t.Fatalf("err = %v, want ErrInvalidInput", err)
			}
		})
	}
}

// TestRunSBOMToFile covers the --out file-output branch (no signing).
func TestRunSBOMToFile(t *testing.T) {
	dir := writeModule(t)
	out := filepath.Join(t.TempDir(), "sbom.json")

	var buf bytes.Buffer
	err := clisbom.RunSBOM(&buf, []string{dir}, clisbom.SBOMOptions{
		Format: "spdx", OutPath: out, From: "module", OmniVersion: "v0.1.0",
	})
	if err != nil {
		t.Fatalf("RunSBOM(--out): %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("out file not valid JSON: %v", err)
	}
}

// TestRunSBOMSignErrors covers the --sign precondition branches that do not
// require a valid secret key.
func TestRunSBOMSignErrors(t *testing.T) {
	dir := writeModule(t)

	t.Run("sign without out", func(t *testing.T) {
		err := clisbom.RunSBOM(&bytes.Buffer{}, []string{dir}, clisbom.SBOMOptions{
			Format: "spdx", Sign: true,
		})
		if !cmderr.IsInvalidInput(err) {
			t.Fatalf("err = %v, want ErrInvalidInput", err)
		}
	})

	t.Run("sign without key", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "sbom.json")
		err := clisbom.RunSBOM(&bytes.Buffer{}, []string{dir}, clisbom.SBOMOptions{
			Format: "spdx", OutPath: out, Sign: true,
		})
		if !cmderr.IsInvalidInput(err) {
			t.Fatalf("err = %v, want ErrInvalidInput", err)
		}
	})

	t.Run("sign with missing key file", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "sbom.json")
		err := clisbom.RunSBOM(&bytes.Buffer{}, []string{dir}, clisbom.SBOMOptions{
			Format: "spdx", OutPath: out, Sign: true, KeyPath: filepath.Join(dir, "nope.key"),
		})
		if !cmderr.IsNotFound(err) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}

// TestRunSBOMSignMissingPassphrase asserts that --sign with a valid key file
// but no OMNI_SIGN_PASSPHRASE is rejected as ErrInvalidInput.
func TestRunSBOMSignMissingPassphrase(t *testing.T) {
	dir := writeModule(t)
	keyDir := t.TempDir()
	keyPath := filepath.Join(keyDir, "sk.key")

	kp, err := sign.GenerateKeyPair("pw", sign.WithScryptParams(1<<15, 8, 1))
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	skText, err := kp.SecretKey.MarshalText("pw", sign.WithScryptParams(1<<15, 8, 1))
	if err != nil {
		t.Fatalf("MarshalText sk: %v", err)
	}
	if err := os.WriteFile(keyPath, skText, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	// Ensure the passphrase env is absent for this test.
	t.Setenv("OMNI_SIGN_PASSPHRASE", "")
	_ = os.Unsetenv("OMNI_SIGN_PASSPHRASE")

	out := filepath.Join(t.TempDir(), "sbom.json")
	err = clisbom.RunSBOM(&bytes.Buffer{}, []string{dir}, clisbom.SBOMOptions{
		Format: "spdx", OutPath: out, Sign: true, KeyPath: keyPath, From: "module",
	})
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("err = %v, want ErrInvalidInput (missing passphrase)", err)
	}
}

// TestRunSBOMSignSuccess covers the full happy-path signing flow: a valid key,
// the correct passphrase via env, and a detached signature written to disk.
func TestRunSBOMSignSuccess(t *testing.T) {
	dir := writeModule(t)
	keyPath := filepath.Join(t.TempDir(), "sk.key")

	kp, err := sign.GenerateKeyPair("pw", sign.WithScryptParams(1<<15, 8, 1))
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	skText, err := kp.SecretKey.MarshalText("pw", sign.WithScryptParams(1<<15, 8, 1))
	if err != nil {
		t.Fatalf("MarshalText sk: %v", err)
	}
	if err := os.WriteFile(keyPath, skText, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")

	out := filepath.Join(t.TempDir(), "sbom.json")
	var buf bytes.Buffer
	err = clisbom.RunSBOM(&buf, []string{dir}, clisbom.SBOMOptions{
		Format: "spdx", OutPath: out, Sign: true, KeyPath: keyPath, From: "module",
	})
	if err != nil {
		t.Fatalf("RunSBOM(--sign) error = %v, want nil", err)
	}
	if _, err := os.Stat(out + ".minisig"); err != nil {
		t.Fatalf("signature file not written: %v", err)
	}
	if !strings.Contains(buf.String(), "Signature written") {
		t.Errorf("output = %q, want a 'Signature written' line", buf.String())
	}
}

// TestRunSBOMSignWrongPassphrase asserts a wrong passphrase is classified as
// ErrInvalidInput (wrong passphrase / corrupt key).
func TestRunSBOMSignWrongPassphrase(t *testing.T) {
	dir := writeModule(t)
	keyPath := filepath.Join(t.TempDir(), "sk.key")

	kp, err := sign.GenerateKeyPair("right", sign.WithScryptParams(1<<15, 8, 1))
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	skText, err := kp.SecretKey.MarshalText("right", sign.WithScryptParams(1<<15, 8, 1))
	if err != nil {
		t.Fatalf("MarshalText sk: %v", err)
	}
	if err := os.WriteFile(keyPath, skText, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	t.Setenv("OMNI_SIGN_PASSPHRASE", "wrong")

	out := filepath.Join(t.TempDir(), "sbom.json")
	err = clisbom.RunSBOM(&bytes.Buffer{}, []string{dir}, clisbom.SBOMOptions{
		Format: "spdx", OutPath: out, Sign: true, KeyPath: keyPath, From: "module",
	})
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("err = %v, want ErrInvalidInput (wrong passphrase)", err)
	}
}
