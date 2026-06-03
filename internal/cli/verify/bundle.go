package verify

import (
	"io"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// verifyBundle always reports that Sigstore bundle verification is unsupported
// in the default omni binary.
//
// Sigstore bundle verification drags in the sigstore-go dependency tree (~50
// transitive modules: Rekor, go-openapi, CT, TSA, go-tuf, in-toto, ...), which
// would pollute the lean pure-Go omni go.mod via MVS. A build tag isolates
// compilation but NOT the module graph, so the capability is delivered as a
// SEPARATE, self-contained module instead. Build/install that module, or use
// `omni verify` for minisign signatures.
func verifyBundle(_ io.Writer, _ []string, _ VerifyOptions) error {
	return cmderr.Wrap(cmderr.ErrUnsupported,
		"sigstore bundle verification is provided by the separate module github.com/inovacc/omni/contrib/sigstore-verify — build/install that, or use 'omni verify' for minisign signatures")
}
