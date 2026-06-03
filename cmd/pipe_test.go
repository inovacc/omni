package cmd

import (
	"bytes"
	"context"
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
