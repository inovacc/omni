// Package sign implements the I/O glue for the `omni sign` command and its
// `keygen` subcommand. It bridges Cobra to pkg/sign: generating
// passphrase-protected Ed25519 key pairs and producing detached minisign
// signatures. Secret-key material is referenced only by file path and the
// passphrase is read from the OMNI_SIGN_PASSPHRASE environment variable or an
// interactive prompt — NEVER from a command-line flag.
package sign

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/sign"
)

// passphraseEnv is the environment variable that may carry the secret-key
// passphrase (a documented tradeoff). The passphrase is never accepted as a
// command-line flag value.
const passphraseEnv = "OMNI_SIGN_PASSPHRASE"

// secretKeyPerm and publicKeyPerm are the on-disk permissions for generated
// keys: the secret key is owner-only, the public key is world-readable.
const (
	secretKeyPerm os.FileMode = 0o600
	publicKeyPerm os.FileMode = 0o644
)

// KeygenOptions configures `omni sign keygen`.
type KeygenOptions struct {
	// PubPath and KeyPath are the output file paths for the public and secret keys.
	PubPath string
	KeyPath string
	// Comment is an optional untrusted comment written to the key files.
	Comment string
	// ScryptN, ScryptR, ScryptP override the scrypt cost. They are honored only
	// when LowCostOK is true (used by tests); otherwise the SENSITIVE default is
	// used so generated keys are protected at rest.
	ScryptN int
	ScryptR int
	ScryptP int
	// LowCostOK permits a reduced scrypt cost (tests only). NEVER set in production.
	LowCostOK bool
	// Force overwrites existing key files.
	Force bool
}

// SignOptions configures `omni sign`.
type SignOptions struct {
	// KeyPath is the path to the minisign *.key secret key file.
	KeyPath string
	// SigPath is the output signature path. When empty it defaults to
	// "<artifact>.minisig".
	SigPath string
	// TrustedComment is embedded in (and signed by) the signature.
	TrustedComment string
	// UntrustedComment is the first-line comment of the signature file.
	UntrustedComment string
}

// RunKeygen generates a passphrase-protected Ed25519 key pair and writes the
// *.pub (0644) and *.key (0600) files. The passphrase is read from
// OMNI_SIGN_PASSPHRASE or an interactive prompt, never a flag.
func RunKeygen(w io.Writer, opts KeygenOptions) error {
	if opts.PubPath == "" || opts.KeyPath == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "sign keygen: --pub and --key output paths are required")
	}
	if err := rejectInlineKey(opts.PubPath); err != nil {
		return err
	}
	if err := rejectInlineKey(opts.KeyPath); err != nil {
		return err
	}
	if !opts.Force {
		if _, err := os.Stat(opts.KeyPath); err == nil {
			return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("sign keygen: %s already exists (use --force to overwrite)", opts.KeyPath))
		}
		if _, err := os.Stat(opts.PubPath); err == nil {
			return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("sign keygen: %s already exists (use --force to overwrite)", opts.PubPath))
		}
	}

	passphrase, err := readPassphrase(true)
	if err != nil {
		return err
	}

	genOpts := []sign.Option{}
	mOpts := []sign.Option{}
	if opts.LowCostOK && opts.ScryptN > 1 {
		genOpts = append(genOpts, sign.WithScryptParams(opts.ScryptN, opts.ScryptR, opts.ScryptP))
		mOpts = append(mOpts, sign.WithScryptParams(opts.ScryptN, opts.ScryptR, opts.ScryptP))
	}
	if opts.Comment != "" {
		mOpts = append(mOpts, sign.WithUntrustedComment(opts.Comment))
	}

	kp, err := sign.GenerateKeyPair(passphrase, genOpts...)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("sign keygen: %v", err))
	}

	keyText, err := kp.SecretKey.MarshalText(passphrase, mOpts...)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("sign keygen: encode secret key: %v", err))
	}
	pubText := kp.PublicKey.MarshalText()

	if err := os.WriteFile(opts.KeyPath, keyText, secretKeyPerm); err != nil {
		return classifyFileErr(err, opts.KeyPath)
	}
	if err := os.WriteFile(opts.PubPath, pubText, publicKeyPerm); err != nil {
		return classifyFileErr(err, opts.PubPath)
	}
	// Re-assert restrictive perms in case the file pre-existed with looser bits.
	_ = os.Chmod(opts.KeyPath, secretKeyPerm)

	_, _ = fmt.Fprintf(w, "Public key written to %s\n", opts.PubPath)
	_, _ = fmt.Fprintf(w, "Secret key written to %s\n", opts.KeyPath)
	_, _ = fmt.Fprintf(w, "key id: %X\n", kp.PublicKey.KeyID)
	return nil
}

// RunSign produces a detached minisign signature over the artifact named by
// args[0] (or stdin when no path is given) using the secret key at opts.KeyPath.
// The passphrase is read from OMNI_SIGN_PASSPHRASE or an interactive prompt.
func RunSign(w io.Writer, r io.Reader, args []string, opts SignOptions) error {
	if opts.KeyPath == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "sign: --key is required (path to a secret key file)")
	}
	if err := rejectInlineKey(opts.KeyPath); err != nil {
		return err
	}

	var (
		data    []byte
		sigPath = opts.SigPath
		err     error
	)
	if len(args) == 0 || args[0] == "-" {
		if r == nil {
			return cmderr.Wrap(cmderr.ErrInvalidInput, "sign: no input (provide a file path or pipe data)")
		}
		data, err = io.ReadAll(r)
		if err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("sign: read stdin: %v", err))
		}
		if sigPath == "" {
			return cmderr.Wrap(cmderr.ErrInvalidInput, "sign: --sig is required when signing stdin")
		}
	} else {
		artifact := args[0]
		data, err = os.ReadFile(artifact)
		if err != nil {
			return classifyFileErr(err, artifact)
		}
		if sigPath == "" {
			sigPath = artifact + ".minisig"
		}
	}

	keyText, err := os.ReadFile(opts.KeyPath)
	if err != nil {
		return classifyFileErr(err, opts.KeyPath)
	}

	passphrase, err := readPassphrase(false)
	if err != nil {
		return err
	}

	sk, err := sign.ParseSecretKey(keyText, passphrase)
	if err != nil {
		if errors.Is(err, sign.ErrVerification) {
			return cmderr.Wrap(cmderr.ErrInvalidInput, "sign: wrong passphrase or corrupt secret key")
		}
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("sign: parse secret key: %v", err))
	}

	signOpts := []sign.Option{}
	if opts.TrustedComment != "" {
		signOpts = append(signOpts, sign.WithTrustedComment(opts.TrustedComment))
	}
	if opts.UntrustedComment != "" {
		signOpts = append(signOpts, sign.WithUntrustedComment(opts.UntrustedComment))
	}

	sigText, err := sign.Sign(data, sk, signOpts...)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("sign: %v", err))
	}

	if err := os.WriteFile(sigPath, sigText, publicKeyPerm); err != nil {
		return classifyFileErr(err, sigPath)
	}
	_, _ = fmt.Fprintf(w, "Signature written to %s\n", sigPath)
	return nil
}

// readPassphrase returns the secret-key passphrase from OMNI_SIGN_PASSPHRASE or,
// when unset and stdin is a terminal, an interactive (non-echoing) prompt. The
// passphrase is never read from a flag. When confirm is true (keygen) the prompt
// asks twice and rejects a mismatch.
func readPassphrase(confirm bool) (string, error) {
	if env, ok := os.LookupEnv(passphraseEnv); ok {
		return env, nil
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", cmderr.Wrap(cmderr.ErrInvalidInput,
			fmt.Sprintf("sign: no passphrase available (set %s or run interactively)", passphraseEnv))
	}

	_, _ = fmt.Fprint(os.Stderr, "Passphrase: ")
	first, err := term.ReadPassword(fd)
	_, _ = fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("sign: read passphrase: %v", err))
	}
	if confirm {
		_, _ = fmt.Fprint(os.Stderr, "Confirm passphrase: ")
		second, err := term.ReadPassword(fd)
		_, _ = fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("sign: read passphrase: %v", err))
		}
		if string(first) != string(second) {
			return "", cmderr.Wrap(cmderr.ErrInvalidInput, "sign: passphrases do not match")
		}
	}
	return string(first), nil
}

// classifyFileErr maps an os file error to a cmderr sentinel.
func classifyFileErr(err error, path string) error {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("sign: %s", path))
	case errors.Is(err, os.ErrPermission):
		return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("sign: %s", path))
	default:
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("sign: %s: %v", path, err))
	}
}

// rejectInlineKey enforces the key-handling policy: key arguments must be file
// paths, never inline key material. A value containing newlines or a minisign
// comment prefix is treated as inline material and rejected.
func rejectInlineKey(value string) error {
	if strings.ContainsAny(value, "\n\r") || strings.HasPrefix(value, "untrusted comment:") {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "sign: key argument must be a file path, not inline key material")
	}
	return nil
}
