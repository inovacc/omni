// Package validate checks that a value conforms to a well-known data format
// (email address, IP address) and reports the result via exit code.
package validate

import (
	"fmt"
	"io"
	"net/mail"
	"net/netip"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// RunEmail validates s as an RFC 5322 address. A malformed value returns
// ErrConflict (exit 1); a valid value prints "<addr>\tOK" and returns nil.
func RunEmail(w io.Writer, s string) error {
	if _, err := mail.ParseAddress(s); err != nil {
		return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("validate: invalid email: %v", err))
	}

	_, _ = fmt.Fprintf(w, "%s\tOK\n", s)

	return nil
}

// RunIP validates s as an IPv4 or IPv6 address. A malformed value returns
// ErrConflict (exit 1); a valid value prints "<addr>\tOK" and returns nil.
func RunIP(w io.Writer, s string) error {
	if _, err := netip.ParseAddr(s); err != nil {
		return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("validate: invalid ip: %v", err))
	}

	_, _ = fmt.Fprintf(w, "%s\tOK\n", s)

	return nil
}
