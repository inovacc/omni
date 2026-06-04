// Package attest is the I/O glue between the Cobra `omni attest` /
// `omni attest verify` commands and the pure-Go pkg/attest library. It bridges
// pkg/sign (the concrete Ed25519 Signer/Verifier) into pkg/attest's
// function-typed hooks, reads/writes files, and classifies every error through
// internal/cli/cmderr so the root command maps it to the correct exit code.
//
// The signing passphrase is read ONLY from the OMNI_SIGN_PASSPHRASE environment
// variable (never a flag), matching the `omni sign` convention. Key material is
// never accepted as a value — only a file path.
package attest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/attest"
	"github.com/inovacc/omni/pkg/sign"
)

// predicateTypeSLSA is the only --predicate-type the generator supports today.
const predicateTypeSLSA = "slsa-provenance"

// GenOptions configures `omni attest` (generate).
type GenOptions struct {
	KeyPath       string // path to the *.key secret key (file path only, never key material)
	ArtifactPath  string // artifact whose sha256 becomes the subject digest
	PredicateType string // only "slsa-provenance" supported
	PredicatePath string // optional pre-built predicate JSON; if empty, build from flags/env
	BuilderID     string // empty -> local builder.id; must be ADR-0009-allowed otherwise
	FromEnv       bool   // populate provenance from GITHUB_* env vars (release path)
	OutPath       string // output envelope path; empty -> stdout
}

// VerifyOptions configures `omni attest verify`.
type VerifyOptions struct {
	KeyPath      string // path to the *.pub public key
	EnvelopePath string // path to the DSSE envelope JSON
	ArtifactPath string // optional: if set, verify subject sha256 matches this artifact
}

// readFile reads path, classifying os errors into cmderr sentinels: a missing
// file -> notFound, a permission error -> perm, anything else -> ErrIO.
func readFile(path string, notFound, perm error) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return nil, cmderr.Wrap(notFound, fmt.Sprintf("read %s", path))
		case os.IsPermission(err):
			return nil, cmderr.Wrap(perm, fmt.Sprintf("read %s", path))
		default:
			return nil, cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("read %s", path))
		}
	}
	return b, nil
}

// RunAttest generates and writes a signed DSSE/SLSA provenance attestation. The
// artifact's sha256 becomes the subject digest; the SLSA Provenance v1 predicate
// is built from flags/env (or a verbatim --predicate file), validated against
// the ADR-0009 builder.id allowlist, marshaled, PAE-signed with the secret key,
// and emitted as an indented DSSE JSON envelope to OutPath (or w if empty).
func RunAttest(w io.Writer, opts GenOptions) error {
	pt := opts.PredicateType
	if pt == "" {
		pt = predicateTypeSLSA
	}
	if pt != predicateTypeSLSA {
		return cmderr.Wrap(cmderr.ErrUnsupported, fmt.Sprintf("only --predicate-type %s is supported", predicateTypeSLSA))
	}

	keyText, err := readFile(opts.KeyPath, cmderr.ErrNotFound, cmderr.ErrPermission)
	if err != nil {
		return err
	}
	sk, err := sign.ParseSecretKey(keyText, os.Getenv("OMNI_SIGN_PASSPHRASE"))
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "parse secret key (check OMNI_SIGN_PASSPHRASE)")
	}

	artifact, err := readFile(opts.ArtifactPath, cmderr.ErrNotFound, cmderr.ErrPermission)
	if err != nil {
		return err
	}

	prov, subjName, err := buildProvenance(opts, opts.ArtifactPath)
	if err != nil {
		return err // already classified
	}
	st := attest.NewStatement(
		[]attest.ResourceDescriptor{attest.SubjectFromBytes(subjName, artifact)},
		prov,
	)

	signer := func(pae []byte) ([]byte, string, error) {
		sig, e := sign.Sign(pae, sk)
		if e != nil {
			return nil, "", e
		}
		return sig, hex.EncodeToString(sk.KeyID[:]), nil
	}
	env, err := attest.SignStatement(st, signer)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, "sign statement")
	}

	out, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, "marshal envelope")
	}
	out = append(out, '\n')

	if opts.OutPath == "" {
		if _, err := w.Write(out); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, "write envelope")
		}
		return nil
	}
	if err := os.WriteFile(opts.OutPath, out, 0o644); err != nil {
		if os.IsPermission(err) {
			return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("write %s", opts.OutPath))
		}
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("write %s", opts.OutPath))
	}
	return nil
}

// RunVerify verifies a DSSE/SLSA envelope fail-closed against a public key. On
// success it prints a one-line summary to w and returns nil; on ANY failure mode
// it returns a classified cmderr error (ErrConflict for a verification/digest
// failure, ErrInvalidInput for malformed inputs, ErrNotFound/ErrPermission for
// file errors).
func RunVerify(w io.Writer, opts VerifyOptions) error {
	pubText, err := readFile(opts.KeyPath, cmderr.ErrNotFound, cmderr.ErrPermission)
	if err != nil {
		return err
	}
	envText, err := readFile(opts.EnvelopePath, cmderr.ErrNotFound, cmderr.ErrPermission)
	if err != nil {
		return err
	}

	var artifact []byte
	if opts.ArtifactPath != "" {
		artifact, err = readFile(opts.ArtifactPath, cmderr.ErrNotFound, cmderr.ErrPermission)
		if err != nil {
			return err
		}
	}

	st, err := verifyEnvelopeBytes(pubText, envText, artifact, opts.ArtifactPath != "")
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(w, "OK: %d subject(s), predicateType %s\n", len(st.Subject), st.PredicateType)
	return nil
}

// RunVerifyReader verifies a DSSE/SLSA envelope read from r against the public
// key whose path is args[0]. It is the pipe-registry entry point (the envelope
// streams in via stdin; the pubkey is a path argument). Artifact binding is not
// available over the pipe interface.
func RunVerifyReader(w io.Writer, r io.Reader, args []string) error {
	if len(args) < 1 || args[0] == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "attest verify: missing public key path argument")
	}
	pubText, err := readFile(args[0], cmderr.ErrNotFound, cmderr.ErrPermission)
	if err != nil {
		return err
	}
	envText, err := io.ReadAll(r)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, "read envelope from input")
	}
	st, err := verifyEnvelopeBytes(pubText, envText, nil, false)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(w, "OK: %d subject(s), predicateType %s\n", len(st.Subject), st.PredicateType)
	return nil
}

// verifyEnvelopeBytes is the shared fail-closed verification core for both the
// file-based RunVerify and the stdin-based RunVerifyReader. It parses the public
// key and envelope JSON, verifies the DSSE signature via pkg/sign, and (when
// bindArtifact is set) requires the artifact's sha256 to match a subject digest.
// Errors are classified per the cmderr table.
func verifyEnvelopeBytes(pubText, envText, artifact []byte, bindArtifact bool) (attest.Statement, error) {
	pub, err := sign.ParsePublicKey(pubText)
	if err != nil {
		return attest.Statement{}, cmderr.Wrap(cmderr.ErrInvalidInput, "parse public key")
	}

	var env attest.Envelope
	if err := json.Unmarshal(envText, &env); err != nil {
		return attest.Statement{}, cmderr.Wrap(cmderr.ErrInvalidInput, "parse envelope JSON")
	}

	verifier := func(pae, sig []byte, _ string) error { return sign.Verify(pae, sig, pub) }
	st, err := attest.VerifyEnvelope(env, verifier)
	if err != nil {
		return attest.Statement{}, cmderr.Wrap(cmderr.ErrConflict, "attestation verification failed")
	}

	if bindArtifact {
		want := attest.SubjectFromBytes("", artifact).Digest["sha256"]
		matched := false
		for _, s := range st.Subject {
			if s.Digest["sha256"] == want {
				matched = true
				break
			}
		}
		if !matched {
			return attest.Statement{}, cmderr.Wrap(cmderr.ErrConflict, "artifact digest does not match any subject")
		}
	}
	return st, nil
}
