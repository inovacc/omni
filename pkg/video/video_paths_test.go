package video

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBuildFilename_DefaultsToDownloadsDir(t *testing.T) {
	setVideoEnv(t, "HOME", "")
	setVideoEnv(t, "USERPROFILE", "")

	switch runtime.GOOS {
	case "windows":
		setVideoEnv(t, "USERPROFILE", `C:\Users\tester`)
	default:
		setVideoEnv(t, "HOME", "/tmp/tester")
	}

	c := &Client{}
	info := &VideoInfo{
		ID:    "abc123",
		Title: "my video",
	}
	f := &Format{Ext: "mp4"}

	got := c.buildFilename(info, f)

	var home string
	if runtime.GOOS == "windows" {
		home = `C:\Users\tester`
	} else {
		home = "/tmp/tester"
	}

	want := filepath.Join(home, "Downloads", "my video.mp4")
	if got != want {
		t.Fatalf("buildFilename() = %q, want %q", got, want)
	}
}

func TestBuildFilename_UsesCustomOutputTemplate(t *testing.T) {
	c := &Client{
		opts: Options{
			Output: "%(title)s.%(ext)s",
		},
	}

	info := &VideoInfo{
		ID:    "abc123",
		Title: "my video",
	}
	f := &Format{Ext: "webm"}

	got := c.buildFilename(info, f)
	if got != "my video.webm" {
		t.Fatalf("buildFilename() = %q, want %q", got, "my video.webm")
	}
}

func TestEnsureOutputDir(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a", "b", "file.mp4")

	if err := ensureOutputDir(path); err != nil {
		t.Fatalf("ensureOutputDir() error = %v", err)
	}

	wantDir := filepath.Join(root, "a", "b")
	if _, err := os.Stat(wantDir); err != nil {
		t.Fatalf("expected output dir to exist: %v", err)
	}
}

func setVideoEnv(t *testing.T, key, value string) {
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
