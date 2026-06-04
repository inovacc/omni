// Package reprocheck compares two parallel sets of built artifacts and fails
// closed if any sha256 digest pair differs — the dogfooded reproducible-build
// gate for the v1.0 release. Pure stdlib; runs on every OS.
package reprocheck

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// Options configures a reproducibility check. A and B are equal-length, index-
// aligned lists of file paths from two independent builds.
type Options struct {
	A []string
	B []string
}

// Run compares each (A[i], B[i]) pair by sha256. It writes a per-pair report to
// w and returns cmderr.ErrConflict on the first drift, after reporting all pairs.
func Run(w io.Writer, opts Options) error {
	if len(opts.A) != len(opts.B) || len(opts.A) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "reprocheck requires equal, non-empty A/B file lists")
	}
	drift := false
	for i := range opts.A {
		ha, err := digest(opts.A[i])
		if err != nil {
			return err
		}
		hb, err := digest(opts.B[i])
		if err != nil {
			return err
		}
		if ha == hb {
			_, _ = fmt.Fprintf(w, "reproducible  %s  %s\n", ha[:16], opts.A[i])
		} else {
			drift = true
			_, _ = fmt.Fprintf(w, "DRIFT         %s != %s  %s\n", ha[:16], hb[:16], opts.A[i])
		}
	}
	if drift {
		return cmderr.Wrap(cmderr.ErrConflict, "reproducibility drift detected (binaries differ between builds)")
	}
	return nil
}

func digest(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", cmderr.Wrap(cmderr.ErrNotFound, "open "+path)
		}
		if os.IsPermission(err) {
			return "", cmderr.Wrap(cmderr.ErrPermission, "open "+path)
		}
		return "", cmderr.Wrap(cmderr.ErrIO, "open "+path)
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", cmderr.Wrap(cmderr.ErrIO, "read "+path)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
