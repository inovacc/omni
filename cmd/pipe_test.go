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

// TestPipeRegistryHasScan asserts the scan command is wired into the unified
// pipe registry so `omni pipe` can scan an SBOM read from stdin (scan source and
// scan db update stay Cobra-only — they are not stdin transforms).
func TestPipeRegistryHasScan(t *testing.T) {
	reg := buildPipeRegistry()
	if _, ok := reg.Get("scan"); !ok {
		t.Error("buildPipeRegistry(): \"scan\" not registered")
	}
}

// TestPipeScanDispatchesToGlue confirms the registered scan command reaches the
// real scan.RunScanStdin glue: with no DB configured (no OMNI_SCAN_DB env) it
// returns ErrInvalidInput ("--db ... is required"), proving the wiring is live.
func TestPipeScanDispatchesToGlue(t *testing.T) {
	t.Setenv("OMNI_SCAN_DB", "")
	t.Setenv("OMNI_SCAN_DB_KEY", "")
	reg := buildPipeRegistry()
	cmd, ok := reg.Get("scan")
	if !ok {
		t.Fatal("scan not registered")
	}
	var buf bytes.Buffer
	err := cmd.Run(context.Background(), &buf, strings.NewReader("{}"), nil)
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("scan with no DB: err = %v, want ErrInvalidInput", err)
	}
}

// TestPipeRegistryHasAttestVerify asserts the attest-verify command is wired
// into the unified pipe registry so `omni pipe` can verify a DSSE envelope read
// from stdin (attest generate stays Cobra-only — it needs key/flags and a file
// path, not a stdin transform).
func TestPipeRegistryHasAttestVerify(t *testing.T) {
	reg := buildPipeRegistry()
	if _, ok := reg.Get("attest-verify"); !ok {
		t.Error("buildPipeRegistry(): \"attest-verify\" not registered")
	}
}

// TestPipeAttestVerifyDispatchesToGlue confirms the registered attest-verify
// command reaches the real attest.RunVerifyReader glue: invoked with no pubkey
// path argument it returns an ErrInvalidInput ("missing public key path"),
// proving the wiring is live.
func TestPipeAttestVerifyDispatchesToGlue(t *testing.T) {
	reg := buildPipeRegistry()
	cmd, ok := reg.Get("attest-verify")
	if !ok {
		t.Fatal("attest-verify not registered")
	}
	var buf bytes.Buffer
	err := cmd.Run(context.Background(), &buf, strings.NewReader(""), nil)
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("attest-verify with no args: err = %v, want ErrInvalidInput", err)
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
