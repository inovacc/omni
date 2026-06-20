package userdirs

import (
	"path/filepath"
	"runtime"
	"testing"
)

// clearHomeEnv blanks every environment variable homeDir() consults so each
// table case starts from a known, empty baseline and only sets what it needs.
// t.Setenv registers automatic cleanup, so the original environment is
// restored when the (sub)test finishes.
func clearHomeEnv(t *testing.T) {
	t.Helper()

	for _, k := range []string{"HOME", "USERPROFILE", "HOMEDRIVE", "HOMEPATH"} {
		t.Setenv(k, "")
	}
}

// TestHomeDir_Branches drives homeDir() through every reachable fallback
// branch using only environment variables (no network, no exec). Because
// os.UserHomeDir() reads USERPROFILE on Windows and HOME on Unix, the set of
// reachable branches differs per OS; each case is guarded by runtime.GOOS so
// the assertions stay deterministic on whatever platform the suite runs on.
func TestHomeDir_Branches(t *testing.T) {
	type envSet struct {
		home      string
		userProf  string
		homeDrive string
		homePath  string
	}

	tests := []struct {
		name    string
		onlyOS  string // "" = all; otherwise restrict to this GOOS
		env     envSet
		want    string
		wantErr bool
	}{
		{
			// Windows fast path: USERPROFILE wins before anything else.
			name:   "windows USERPROFILE wins",
			onlyOS: "windows",
			env:    envSet{userProf: `C:\Users\alice`},
			want:   `C:\Users\alice`,
		},
		{
			// Windows: USERPROFILE empty -> os.UserHomeDir() fails ->
			// HOME env fallback is taken.
			name:   "windows HOME fallback",
			onlyOS: "windows",
			env:    envSet{home: `D:\home\bob`},
			want:   `D:\home\bob`,
		},
		{
			// Windows: USERPROFILE + HOME empty, os.UserHomeDir() fails ->
			// HOMEDRIVE+HOMEPATH fallback is taken.
			name:   "windows HOMEDRIVE+HOMEPATH fallback",
			onlyOS: "windows",
			env:    envSet{homeDrive: `E:`, homePath: `\users\carol`},
			want:   `E:\users\carol`,
		},
		{
			// Windows: every source empty -> final error branch.
			name:    "windows no source -> error",
			onlyOS:  "windows",
			env:     envSet{},
			wantErr: true,
		},
		{
			// Unix: os.UserHomeDir() reads HOME and succeeds, so the
			// UserHomeDir success branch is taken (not the raw HOME getenv).
			name:   "unix UserHomeDir via HOME",
			onlyOS: "unix",
			env:    envSet{home: "/home/dave"},
			want:   "/home/dave",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.onlyOS == "windows" && runtime.GOOS != "windows" {
				t.Skipf("case is windows-specific; GOOS=%s", runtime.GOOS)
			}
			if tt.onlyOS == "unix" && runtime.GOOS == "windows" {
				t.Skipf("case is unix-specific; GOOS=%s", runtime.GOOS)
			}

			clearHomeEnv(t)
			if tt.env.home != "" {
				t.Setenv("HOME", tt.env.home)
			}
			if tt.env.userProf != "" {
				t.Setenv("USERPROFILE", tt.env.userProf)
			}
			if tt.env.homeDrive != "" {
				t.Setenv("HOMEDRIVE", tt.env.homeDrive)
			}
			if tt.env.homePath != "" {
				t.Setenv("HOMEPATH", tt.env.homePath)
			}

			got, err := homeDir()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("homeDir() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("homeDir() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("homeDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestHomeDir_HomeFallbackUnix exercises the explicit os.Getenv("HOME")
// fallback on Unix. os.UserHomeDir() already returns HOME on Unix, so to reach
// the raw-getenv branch we would need UserHomeDir to fail while HOME is set —
// which does not happen on Unix. This test therefore only asserts the common
// success contract and is skipped on Windows where the path differs.
func TestHomeDir_HomeFallbackUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("HOME-driven path verified by TestHomeDir_Branches on Windows")
	}

	clearHomeEnv(t)
	t.Setenv("HOME", "/var/empty/eve")

	got, err := homeDir()
	if err != nil {
		t.Fatalf("homeDir() error = %v", err)
	}
	if got != "/var/empty/eve" {
		t.Fatalf("homeDir() = %q, want /var/empty/eve", got)
	}
}

// TestPublicDirs_DeriveFromHome confirms DownloadsDir/DocumentsDir join the
// resolved home with the expected leaf and that the same home source feeds
// both, covering the success branches of both exported functions across OSes.
func TestPublicDirs_DeriveFromHome(t *testing.T) {
	clearHomeEnv(t)

	var home string
	switch runtime.GOOS {
	case "windows":
		home = `C:\Users\frank`
		t.Setenv("USERPROFILE", home)
	default:
		home = "/home/frank"
		t.Setenv("HOME", home)
	}

	tests := []struct {
		name string
		fn   func() (string, error)
		leaf string
	}{
		{"downloads", DownloadsDir, "Downloads"},
		{"documents", DocumentsDir, "Documents"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn()
			if err != nil {
				t.Fatalf("%s() error = %v", tt.name, err)
			}
			want := filepath.Join(home, tt.leaf)
			if got != want {
				t.Fatalf("%s() = %q, want %q", tt.name, got, want)
			}
		})
	}
}

// TestPublicDirs_PropagateHomeError verifies that when no home source is
// available, both DownloadsDir and DocumentsDir surface the homeDir() error
// instead of returning a bogus path. Reachable deterministically on Windows
// (os.UserHomeDir() fails without USERPROFILE); on Unix os.UserHomeDir() has
// OS fallbacks, so the case is skipped there.
func TestPublicDirs_PropagateHomeError(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("error branch only deterministically reachable on Windows")
	}

	tests := []struct {
		name string
		fn   func() (string, error)
	}{
		{"downloads", DownloadsDir},
		{"documents", DocumentsDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearHomeEnv(t)

			got, err := tt.fn()
			if err == nil {
				t.Fatalf("%s() = %q, want error when no home source", tt.name, got)
			}
		})
	}
}
