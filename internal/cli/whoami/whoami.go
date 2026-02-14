package whoami

import (
	"fmt"
	"io"
	"os/user"

	"github.com/inovacc/omni/internal/cli/output"
)

// WhoamiOptions configures the whoami command behavior
type WhoamiOptions struct {
	OutputFormat output.Format // output format (text/json/table)
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

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		result := WhoamiResult{
			Username: u.Username,
			UID:      u.Uid,
			GID:      u.Gid,
			Name:     u.Name,
			HomeDir:  u.HomeDir,
		}

		return f.Print(result)
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
