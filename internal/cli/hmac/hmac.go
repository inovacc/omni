// Package hmac computes keyed-hash message authentication codes (HMAC)
// over a message read from an argument or standard input.
package hmac

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// HMACOptions configures HMAC generation.
type HMACOptions struct {
	Algorithm string // sha256 (default), sha1, sha512
	Key       string // shared secret
}

// RunHMAC writes the hex-encoded HMAC of the message using opts.Key.
// The message is args[0] when a non-empty argument is supplied, otherwise it
// is read from r. An empty key or an unknown algorithm yields ErrInvalidInput
// (exit 2).
func RunHMAC(w io.Writer, r io.Reader, args []string, opts HMACOptions) error {
	if opts.Key == "" {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "hmac: missing --key")
	}

	newHash, err := hasherFor(opts.Algorithm)
	if err != nil {
		return err
	}

	mac := hmac.New(newHash, []byte(opts.Key))

	if len(args) > 0 && args[0] != "" {
		_, _ = mac.Write([]byte(args[0]))
	} else if _, err := io.Copy(mac, r); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("hmac: read message: %v", err))
	}

	_, _ = fmt.Fprintf(w, "%s\n", hex.EncodeToString(mac.Sum(nil)))

	return nil
}

func hasherFor(algo string) (func() hash.Hash, error) {
	switch algo {
	case "", "sha256":
		return sha256.New, nil
	case "sha1":
		return sha1.New, nil
	case "sha512":
		return sha512.New, nil
	default:
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("hmac: unknown algorithm %q", algo))
	}
}
