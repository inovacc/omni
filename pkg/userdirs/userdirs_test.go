package userdirs

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDownloadsDir(t *testing.T) {
	setEnv(t, "HOME", "")
	setEnv(t, "USERPROFILE", "")
	setEnv(t, "HOMEDRIVE", "")
	setEnv(t, "HOMEPATH", "")

	switch runtime.GOOS {
	case "windows":
		setEnv(t, "USERPROFILE", `C:\Users\tester`)

		dir, err := DownloadsDir()
		if err != nil {
			t.Fatalf("DownloadsDir() error = %v", err)
		}

		want := filepath.Join(`C:\Users\tester`, "Downloads")
		if dir != want {
			t.Fatalf("DownloadsDir() = %q, want %q", dir, want)
		}

	default:
		setEnv(t, "HOME", "/tmp/tester")

		dir, err := DownloadsDir()
		if err != nil {
			t.Fatalf("DownloadsDir() error = %v", err)
		}

		want := filepath.Join("/tmp/tester", "Downloads")
		if dir != want {
			t.Fatalf("DownloadsDir() = %q, want %q", dir, want)
		}
	}
}

func TestDocumentsDir(t *testing.T) {
	setEnv(t, "HOME", "")
	setEnv(t, "USERPROFILE", "")
	setEnv(t, "HOMEDRIVE", "")
	setEnv(t, "HOMEPATH", "")

	switch runtime.GOOS {
	case "windows":
		setEnv(t, "USERPROFILE", `C:\Users\tester`)

		dir, err := DocumentsDir()
		if err != nil {
			t.Fatalf("DocumentsDir() error = %v", err)
		}

		want := filepath.Join(`C:\Users\tester`, "Documents")
		if dir != want {
			t.Fatalf("DocumentsDir() = %q, want %q", dir, want)
		}

	default:
		setEnv(t, "HOME", "/tmp/tester")

		dir, err := DocumentsDir()
		if err != nil {
			t.Fatalf("DocumentsDir() error = %v", err)
		}

		want := filepath.Join("/tmp/tester", "Documents")
		if dir != want {
			t.Fatalf("DocumentsDir() = %q, want %q", dir, want)
		}
	}
}

func setEnv(t *testing.T, key, value string) {
	t.Helper()

	oldValue, hadValue := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("setenv %s: %v", key, err)
	}

	t.Cleanup(func() {
		if !hadValue {
			_ = os.Unsetenv(key)
			return
		}

		_ = os.Setenv(key, oldValue)
	})
}
