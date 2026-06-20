package whoami

import (
	"bytes"
	"encoding/json"
	"os/user"
	"testing"

	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

func TestRunWhoami_JSON(t *testing.T) {
	var buf bytes.Buffer
	if err := RunWhoami(&buf, WhoamiOptions{OutputFormat: output.FormatJSON}); err != nil {
		t.Fatalf("RunWhoami json: %v", err)
	}

	var result WhoamiResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	cur, err := user.Current()
	if err != nil {
		t.Skip("cannot get current user for comparison")
	}

	if result.Username != cur.Username {
		t.Errorf("Username = %q, want %q", result.Username, cur.Username)
	}

	if result.UID != cur.Uid {
		t.Errorf("UID = %q, want %q", result.UID, cur.Uid)
	}

	if result.GID != cur.Gid {
		t.Errorf("GID = %q, want %q", result.GID, cur.Gid)
	}

	if result.HomeDir != cur.HomeDir {
		t.Errorf("HomeDir = %q, want %q", result.HomeDir, cur.HomeDir)
	}
}

func TestWhoami(t *testing.T) {
	got, err := Whoami()
	if err != nil {
		t.Fatalf("Whoami() error = %v", err)
	}

	if got == "" {
		t.Error("Whoami() returned empty username")
	}

	cur, err := user.Current()
	if err != nil {
		t.Skip("cannot get current user for comparison")
	}

	if got != cur.Username {
		t.Errorf("Whoami() = %q, want %q", got, cur.Username)
	}
}

func TestCurrentUser(t *testing.T) {
	u, err := CurrentUser()
	if err != nil {
		t.Fatalf("CurrentUser() error = %v", err)
	}

	if u == nil {
		t.Fatal("CurrentUser() returned nil user")
	}

	if u.Username == "" {
		t.Error("CurrentUser() returned user with empty username")
	}

	cur, err := user.Current()
	if err != nil {
		t.Skip("cannot get current user for comparison")
	}

	if u.Uid != cur.Uid {
		t.Errorf("CurrentUser().Uid = %q, want %q", u.Uid, cur.Uid)
	}
}
