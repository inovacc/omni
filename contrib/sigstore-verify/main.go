// Command omni-sigstore-verify verifies a Sigstore bundle against a trusted
// root, optionally pinning the signing certificate identity / OIDC issuer.
//
// It is a SEPARATE, self-contained module on purpose: sigstore-go drags in ~50
// transitive modules (Rekor, go-openapi, CT, TSA, go-tuf, in-toto, ...) plus
// forced golang.org/x/* version bumps. Keeping it out of the main omni module
// preserves omni's lean, pure-Go go.mod. The default `omni verify --bundle`
// returns an "unsupported" error pointing at this binary.
//
// Scope (matches the old omni_sigstore build tag): bundle VERIFICATION only —
// no Rekor upload, no Fulcio issuance, no OCI. Verification is fail-closed: it
// succeeds ONLY when the bundle's signature, transparency-log inclusion, and
// observer timestamps all verify and the policy (artifact + optional identity)
// matches.
//
// Usage:
//
//	omni-sigstore-verify \
//	  --bundle artifact.sigstore.json \
//	  --trusted-root trusted_root.json \
//	  --artifact artifact.tar.gz \
//	  [--certificate-identity user@example.com] \
//	  [--certificate-oidc-issuer https://accounts.google.com]
//
// Instead of --artifact you may supply --artifact-digest sha256:<hex>.
package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/verify"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "omni-sigstore-verify: %v\n", err)
		os.Exit(1)
	}
}

// options collects the parsed command-line flags.
type options struct {
	bundlePath     string
	trustedRoot    string
	artifact       string
	artifactDigest string
	certIdentity   string
	certOIDCIssuer string
}

// run parses flags, performs the verification, and writes the
// VerificationResult as JSON to w on success. It returns a non-nil error
// (suitable for a non-zero exit) on any validation or verification failure.
func run(args []string, w *os.File) error {
	fs := flag.NewFlagSet("omni-sigstore-verify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var opts options
	fs.StringVar(&opts.bundlePath, "bundle", "", "path to the Sigstore bundle (*.sigstore.json) (required)")
	fs.StringVar(&opts.trustedRoot, "trusted-root", "", "path to the Sigstore trusted root JSON (required)")
	fs.StringVar(&opts.artifact, "artifact", "", "path to the signed artifact to hash and verify")
	fs.StringVar(&opts.artifactDigest, "artifact-digest", "", "artifact digest as sha256:<hex> (alternative to --artifact)")
	fs.StringVar(&opts.certIdentity, "certificate-identity", "", "expected signing certificate SAN identity")
	fs.StringVar(&opts.certOIDCIssuer, "certificate-oidc-issuer", "", "expected signing certificate OIDC issuer")

	fs.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: omni-sigstore-verify --bundle PATH --trusted-root PATH (--artifact PATH | --artifact-digest sha256:HEX) [identity flags]\n\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if opts.bundlePath == "" {
		return errors.New("--bundle is required")
	}
	if opts.trustedRoot == "" {
		return errors.New("--trusted-root is required")
	}
	if opts.artifact == "" && opts.artifactDigest == "" {
		return errors.New("one of --artifact or --artifact-digest is required")
	}
	if opts.artifact != "" && opts.artifactDigest != "" {
		return errors.New("--artifact and --artifact-digest are mutually exclusive")
	}

	b, err := bundle.LoadJSONFromPath(opts.bundlePath)
	if err != nil {
		return fmt.Errorf("load bundle %q: %w", opts.bundlePath, err)
	}

	tr, err := root.NewTrustedRootFromPath(opts.trustedRoot)
	if err != nil {
		return fmt.Errorf("load trusted root %q: %w", opts.trustedRoot, err)
	}

	verifier, err := verify.NewVerifier(tr,
		verify.WithTransparencyLog(1),
		verify.WithObserverTimestamps(1),
	)
	if err != nil {
		return fmt.Errorf("configure verifier: %w", err)
	}

	artifactOpt, closeFn, err := artifactPolicy(opts)
	if err != nil {
		return err
	}
	defer closeFn()

	identityOpt, err := identityPolicy(opts)
	if err != nil {
		return err
	}

	policy := verify.NewPolicy(artifactOpt, identityOpt)

	result, err := verifier.Verify(b, policy)
	if err != nil {
		return fmt.Errorf("sigstore bundle verification failed: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("encode verification result: %w", err)
	}
	return nil
}

// artifactPolicy builds the artifact-binding policy option from either
// --artifact (opens and hashes the file) or --artifact-digest (a sha256:<hex>
// value). It returns the option, a close function (no-op for the digest path),
// and any error.
func artifactPolicy(opts options) (verify.ArtifactPolicyOption, func(), error) {
	noop := func() {}

	if opts.artifactDigest != "" {
		algo, hexDigest, ok := strings.Cut(opts.artifactDigest, ":")
		if !ok {
			return nil, noop, fmt.Errorf("invalid --artifact-digest %q: want <algorithm>:<hex>", opts.artifactDigest)
		}
		raw, err := hex.DecodeString(hexDigest)
		if err != nil {
			return nil, noop, fmt.Errorf("invalid --artifact-digest hex: %w", err)
		}
		return verify.WithArtifactDigest(algo, raw), noop, nil
	}

	f, err := os.Open(opts.artifact)
	if err != nil {
		return nil, noop, fmt.Errorf("open artifact %q: %w", opts.artifact, err)
	}
	return verify.WithArtifact(f), func() { _ = f.Close() }, nil
}

// identityPolicy builds the certificate-identity policy option. When neither
// --certificate-identity nor --certificate-oidc-issuer is set, identity
// checking is explicitly skipped (WithoutIdentitiesUnsafe) — the caller is
// verifying a keyed or self-attested bundle and accepts that tradeoff.
func identityPolicy(opts options) (verify.PolicyOption, error) {
	if opts.certIdentity == "" && opts.certOIDCIssuer == "" {
		return verify.WithoutIdentitiesUnsafe(), nil
	}
	id, err := verify.NewShortCertificateIdentity(opts.certOIDCIssuer, "", opts.certIdentity, "")
	if err != nil {
		return nil, fmt.Errorf("invalid certificate identity: %w", err)
	}
	return verify.WithCertificateIdentity(id), nil
}
