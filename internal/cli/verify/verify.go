// Package verify implements the I/O glue for the `omni verify` command. It
// reads a public key and a detached minisign signature from files (NEVER from
// flag values), verifies the signed artifact fail-closed via pkg/sign, and
// classifies every failure into a cmderr sentinel for the root command.
package verify

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/sign"
)

// VerifyOptions configures a verification run. Key material is referenced only
// by file path; no option ever carries raw key bytes.
type VerifyOptions struct {
	// PubPath is the path to the minisign *.pub public key file.
	PubPath string
	// SigPath is the path to the detached *.minisig signature file. When empty,
	// it defaults to "<artifact>.minisig".
	SigPath string
	// BundlePath, when set, selects Sigstore bundle verification (gated behind
	// the omni_sigstore build tag; otherwise ErrUnsupported).
	BundlePath string
	// TrustedRoot is the Sigstore trusted-root path (omni_sigstore only).
	TrustedRoot string
	// CertIdentity / CertOIDCIssuer pin the signing identity (omni_sigstore only).
	CertIdentity   string
	CertOIDCIssuer string
}

// RunVerify verifies the artifact named by args[0] against the signature and
// public key in opts. It returns nil ONLY when verification fully succeeds; any
// failure is mapped to a cmderr sentinel (verification mismatch -> ErrConflict,
// malformed key/signature/flags -> ErrInvalidInput, missing file -> ErrNotFound,
// unreadable/bad-perms -> ErrPermission, sigstore-without-tag -> ErrUnsupported).
func RunVerify(w io.Writer, r io.Reader, args []string, opts VerifyOptions) error {
	if opts.BundlePath != "" {
		return verifyBundle(w, args, opts)
	}

	if len(args) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "verify: missing artifact path")
	}
	artifact := args[0]

	if opts.PubPath == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "verify: --key is required (path to a public key file)")
	}
	if err := rejectInlineKey(opts.PubPath); err != nil {
		return err
	}

	sigPath := opts.SigPath
	if sigPath == "" {
		sigPath = artifact + ".minisig"
	}

	pubText, err := readFile(opts.PubPath)
	if err != nil {
		return err
	}
	pub, err := sign.ParsePublicKey(pubText)
	if err != nil {
		return classifyParse(err, "verify: parse public key")
	}

	sigText, err := readFile(sigPath)
	if err != nil {
		return err
	}

	data, err := readFile(artifact)
	if err != nil {
		return err
	}

	if err := sign.Verify(data, sigText, pub); err != nil {
		// Any fail-closed verification failure maps to ErrConflict (exit 1).
		if errors.Is(err, sign.ErrVerification) {
			return cmderr.Wrap(cmderr.ErrConflict, "signature verification failed")
		}
		return classifyParse(err, "verify")
	}

	_, _ = fmt.Fprintf(w, "Signature and comment signature verified\n")
	return nil
}

// readFile reads a file, classifying not-found and permission errors.
func readFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, classifyFileErr(err, path)
	}
	return b, nil
}

// classifyFileErr maps an os file error to a cmderr sentinel.
func classifyFileErr(err error, path string) error {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("verify: %s", path))
	case errors.Is(err, os.ErrPermission):
		return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("verify: %s", path))
	default:
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("verify: %s: %v", path, err))
	}
}

// classifyParse maps a pkg/sign parse/verify error to a cmderr sentinel.
func classifyParse(err error, ctx string) error {
	switch {
	case errors.Is(err, sign.ErrVerification):
		return cmderr.Wrap(cmderr.ErrConflict, "signature verification failed")
	case errors.Is(err, sign.ErrMalformed):
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("%s: %v", ctx, err))
	default:
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("%s: %v", ctx, err))
	}
}

// rejectInlineKey enforces the key-handling policy: --key must be a file path,
// never inline key material. A value containing newlines or the minisign
// comment prefix is treated as inline material and rejected.
func rejectInlineKey(value string) error {
	if strings.ContainsAny(value, "\n\r") || strings.HasPrefix(value, "untrusted comment:") {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "verify: --key must be a file path, not inline key material")
	}
	return nil
}
