//go:build !omni_sigstore

package verify

import (
	"io"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// verifyBundle is the default (no-tag) stub: Sigstore bundle verification pulls
// in a heavy dependency tree and is compiled only with -tags omni_sigstore.
// Without that tag, requesting bundle verification is an unsupported operation.
func verifyBundle(_ io.Writer, _ []string, _ VerifyOptions) error {
	return cmderr.Wrap(cmderr.ErrUnsupported,
		"sigstore bundle verification requires building with -tags omni_sigstore")
}
