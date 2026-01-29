package whoami

import (
	"encoding/json"
	"fmt"
	"io"
	"os/user"
)

// WhoamiOptions configures the whoami command behavior
type WhoamiOptions struct {
	JSON bool // --json: output as JSON
}

// WhoamiResult represents whoami output for JSON
type WhoamiResult struct {
	Username string `json:"username"`
	UID      string `json:"uid"`
	GID      string `json:"gid"`
	Name     string `json:"name,omitempty"`
	HomeDir  string `json:"home_dir"`
}

// RunWhoami prints the current username
func RunWhoami(w io.Writer, opts WhoamiOptions) error {
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("cannot get current user: %w", err)
	}

	if opts.JSON {
		result := WhoamiResult{
			Username: u.Username,
			UID:      u.Uid,
			GID:      u.Gid,
			Name:     u.Name,
			HomeDir:  u.HomeDir,
		}

		return json.NewEncoder(w).Encode(result)
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
