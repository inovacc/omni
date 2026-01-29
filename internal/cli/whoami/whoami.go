package whoami

import (
	"fmt"
	"io"
	"os/user"
)

// RunWhoami prints the current username
func RunWhoami(w io.Writer) error {
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("cannot get current user: %w", err)
	}

	_, _ = fmt.Fprintln(w, u.Username)

	return nil
}

// Whoami returns the current username
func Whoami() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("cannot get current user: %w", err)
	}

	return u.Username, nil
}

// CurrentUser returns detailed information about the current user
func CurrentUser() (*user.User, error) {
	return user.Current()
}
