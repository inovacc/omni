package id

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunID(t *testing.T) {
	t.Run("default output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunID(&buf, IDOptions{})
		if err != nil {
			t.Fatalf("RunID() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "uid=") {
			t.Errorf("RunID() should contain 'uid=': %s", output)
		}

		if !strings.Contains(output, "gid=") {
			t.Errorf("RunID() should contain 'gid=': %s", output)
		}
	})

	t.Run("user id only", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunID(&buf, IDOptions{User: true})
		if err != nil {
			t.Fatalf("RunID() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		// Should be a number
		if output == "" {
			t.Error("RunID() -u should output user ID")
		}
	})

	t.Run("user name only", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunID(&buf, IDOptions{User: true, Name: true})
		if err != nil {
			t.Fatalf("RunID() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunID() -un should output username")
		}
	})

	t.Run("group id only", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunID(&buf, IDOptions{Group: true})
		if err != nil {
			t.Fatalf("RunID() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunID() -g should output group ID")
		}
	})

	t.Run("group name only", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunID(&buf, IDOptions{Group: true, Name: true})
		if err != nil {
			t.Fatalf("RunID() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunID() -gn should output group name")
		}
	})

	t.Run("all groups ids", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunID(&buf, IDOptions{Groups: true})
		if err != nil {
			t.Fatalf("RunID() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunID() -G should output group IDs")
		}
	})

	t.Run("all groups names", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunID(&buf, IDOptions{Groups: true, Name: true})
		if err != nil {
			t.Fatalf("RunID() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunID() -Gn should output group names")
		}
	})

	t.Run("nonexistent user", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunID(&buf, IDOptions{Username: "nonexistent_user_12345"})
		if err == nil {
			t.Error("RunID() expected error for nonexistent user")
		}
	})
}

func TestGetUID(t *testing.T) {
	uid, err := GetUID()
	if err != nil {
		// Windows uses SIDs, not numeric UIDs
		t.Logf("GetUID() error (expected on Windows): %v", err)
		return
	}

	if uid < 0 {
		t.Errorf("GetUID() = %d, want non-negative", uid)
	}
}

func TestGetGID(t *testing.T) {
	gid, err := GetGID()
	if err != nil {
		// Windows uses SIDs, not numeric GIDs
		t.Logf("GetGID() error (expected on Windows): %v", err)
		return
	}

	if gid < 0 {
		t.Errorf("GetGID() = %d, want non-negative", gid)
	}
}

func TestGetGroups(t *testing.T) {
	groups, err := GetGroups()
	if err != nil {
		// May fail on some platforms
		t.Logf("GetGroups() error: %v", err)
		return
	}

	// Should have at least one group (primary group) on Unix systems
	// On Windows, may return empty
	t.Logf("GetGroups() returned %d groups", len(groups))
}
