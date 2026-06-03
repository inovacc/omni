//go:build !omni_sigstore

package verify

import (
	"bytes"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// TestRunVerifyBundleUnsupportedWithoutTag asserts that, in a default build
// (without -tags omni_sigstore), requesting Sigstore bundle verification is an
// unsupported operation classified as cmderr.ErrUnsupported (exit code 6).
func TestRunVerifyBundleUnsupportedWithoutTag(t *testing.T) {
	var buf bytes.Buffer
	err := RunVerify(&buf, nil, []string{"artifact.bin"}, VerifyOptions{
		BundlePath: "artifact.sigstore.json",
	})
	if err == nil {
		t.Fatal("RunVerify(--bundle) = nil, want an unsupported error in a default (no-tag) build")
	}
	if !cmderr.IsUnsupported(err) {
		t.Fatalf("RunVerify(--bundle) error = %v, want cmderr.ErrUnsupported", err)
	}
	if code := cmderr.ExitCodeFor(err); code != 6 {
		t.Fatalf("ExitCodeFor(err) = %d, want 6 (ErrUnsupported)", code)
	}
}
