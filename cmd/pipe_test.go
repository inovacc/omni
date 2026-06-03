package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// TestPipeRegistryHasSignVerify asserts that the sign and verify commands are
// wired into the unified pipe registry so `omni pipe` can dispatch them
// directly (keygen stays Cobra-only — it has no stdin transform).
func TestPipeRegistryHasSignVerify(t *testing.T) {
	reg := buildPipeRegistry()
	for _, name := range []string{"sign", "verify"} {
		if _, ok := reg.Get(name); !ok {
			t.Errorf("buildPipeRegistry(): %q not registered", name)
		}
	}
}

// TestPipeVerifyDispatchesToGlue confirms the registered verify command reaches
// the real verify.RunVerify glue: invoked with no artifact arg it returns an
// ErrInvalidInput ("missing artifact path"), proving the wiring is live.
func TestPipeVerifyDispatchesToGlue(t *testing.T) {
	reg := buildPipeRegistry()
	cmd, ok := reg.Get("verify")
	if !ok {
		t.Fatal("verify not registered")
	}
	var buf bytes.Buffer
	err := cmd.Run(context.Background(), &buf, strings.NewReader(""), nil)
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("verify with no args: err = %v, want ErrInvalidInput", err)
	}
}

// TestPipeSignDispatchesToGlue confirms the registered sign command reaches the
// real sign.RunSign glue: invoked with no --key it returns an ErrInvalidInput
// ("--key is required"), proving the wiring is live.
func TestPipeSignDispatchesToGlue(t *testing.T) {
	reg := buildPipeRegistry()
	cmd, ok := reg.Get("sign")
	if !ok {
		t.Fatal("sign not registered")
	}
	var buf bytes.Buffer
	err := cmd.Run(context.Background(), &buf, strings.NewReader(""), nil)
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("sign with no key: err = %v, want ErrInvalidInput", err)
	}
}

// TestPipeSbomDispatchesToGlue confirms the registered sbom command reaches the
// real sbom.RunSBOM glue (reader ignored): given a temp module directory it
// writes SPDX JSON to the writer. This proves sbom participates in the unified
// pipe registry like sign/verify.
func TestPipeSbomDispatchesToGlue(t *testing.T) {
	reg := buildPipeRegistry()
	cmd, ok := reg.Get("sbom")
	if !ok {
		t.Fatal("sbom not registered")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"),
		[]byte("module github.com/example/app\n\ngo 1.25.0\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	var buf bytes.Buffer
	if err := cmd.Run(context.Background(), &buf, strings.NewReader(""), []string{dir}); err != nil {
		t.Fatalf("sbom dispatch: %v", err)
	}
	if !strings.Contains(buf.String(), "SPDX-2.3") {
		t.Errorf("sbom output missing SPDX-2.3 marker:\n%s", buf.String())
	}
}
