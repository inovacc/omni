//go:build omni_sigstore

// This file is compiled ONLY with -tags omni_sigstore. It pulls in the heavy
// sigstore-go dependency tree (Rekor, go-openapi, CT, TSA, go-tuf) that the
// default omni binary intentionally excludes. v1.0 scope is bundle
// VERIFICATION only: no Rekor upload, no Fulcio issuance, no OCI.
package verify

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	sgverify "github.com/sigstore/sigstore-go/pkg/verify"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// verifyBundle verifies a Sigstore bundle (--bundle) against a trusted root
// (--trusted-root), enforcing the optional certificate identity / OIDC issuer
// pins. It returns nil ONLY when the bundle's signature, transparency-log
// inclusion, and observer timestamps all verify and the policy matches.
//
// Failure classification:
//   - missing/unreadable bundle, trusted root, or artifact -> ErrNotFound/ErrPermission/ErrIO
//   - missing --trusted-root, malformed bundle/root, bad identity flags -> ErrInvalidInput
//   - any cryptographic / policy mismatch -> ErrConflict
func verifyBundle(w io.Writer, args []string, opts VerifyOptions) error {
	if len(args) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "verify: missing artifact path")
	}
	artifact := args[0]

	if opts.TrustedRoot == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput,
			"verify: --trusted-root is required for sigstore bundle verification")
	}

	b, err := bundle.LoadJSONFromPath(opts.BundlePath)
	if err != nil {
		return classifyBundleFileErr(err, opts.BundlePath, "load bundle")
	}

	tr, err := root.NewTrustedRootFromPath(opts.TrustedRoot)
	if err != nil {
		return classifyBundleFileErr(err, opts.TrustedRoot, "load trusted root")
	}

	verifier, err := sgverify.NewVerifier(tr,
		sgverify.WithTransparencyLog(1),
		sgverify.WithObserverTimestamps(1),
	)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("verify: configure verifier: %v", err))
	}

	identityOpt, err := bundleIdentityPolicy(opts)
	if err != nil {
		return err
	}

	f, err := os.Open(artifact)
	if err != nil {
		return classifyFileErr(err, artifact)
	}
	defer func() { _ = f.Close() }()

	policy := sgverify.NewPolicy(sgverify.WithArtifact(f), identityOpt)

	if _, err := verifier.Verify(b, policy); err != nil {
		// Every cryptographic / policy failure is a fail-closed mismatch.
		return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("sigstore bundle verification failed: %v", err))
	}

	_, _ = fmt.Fprintf(w, "Sigstore bundle verified\n")
	return nil
}

// bundleIdentityPolicy builds the certificate-identity policy option from the
// CertIdentity / CertOIDCIssuer flags. When neither is set, identity checking
// is explicitly skipped (WithoutIdentitiesUnsafe) — the caller is verifying a
// keyed or self-attested bundle and accepts that tradeoff.
func bundleIdentityPolicy(opts VerifyOptions) (sgverify.PolicyOption, error) {
	if opts.CertIdentity == "" && opts.CertOIDCIssuer == "" {
		return sgverify.WithoutIdentitiesUnsafe(), nil
	}

	id, err := sgverify.NewShortCertificateIdentity(opts.CertOIDCIssuer, "", opts.CertIdentity, "")
	if err != nil {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput,
			fmt.Sprintf("verify: invalid certificate identity: %v", err))
	}
	return sgverify.WithCertificateIdentity(id), nil
}

// classifyBundleFileErr maps a sigstore loader error to a cmderr sentinel. A
// missing/unreadable file is classified by its os error; anything else is a
// malformed input.
func classifyBundleFileErr(err error, path, ctx string) error {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("verify: %s: %s", ctx, path))
	case errors.Is(err, os.ErrPermission):
		return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("verify: %s: %s", ctx, path))
	default:
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("verify: %s: %s: %v", ctx, path, err))
	}
}
