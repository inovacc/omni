package verify

import (
	"bytes"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// TestRunVerifyBundleUnsupported asserts that requesting Sigstore bundle
// verification in the default omni binary is an unsupported operation
// classified as cmderr.ErrUnsupported (exit code 6). The capability lives in
// the separate github.com/inovacc/omni/contrib/sigstore-verify module.
func TestRunVerifyBundleUnsupported(t *testing.T) {
	var buf bytes.Buffer
	err := RunVerify(&buf, nil, []string{"artifact.bin"}, VerifyOptions{
		BundlePath: "artifact.sigstore.json",
	})
	if err == nil {
		t.Fatal("RunVerify(--bundle) = nil, want an unsupported error in the default omni binary")
	}
	if !cmderr.IsUnsupported(err) {
		t.Fatalf("RunVerify(--bundle) error = %v, want cmderr.ErrUnsupported", err)
	}
	if code := cmderr.ExitCodeFor(err); code != 6 {
		t.Fatalf("ExitCodeFor(err) = %d, want 6 (ErrUnsupported)", code)
	}
}
