package safepath

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCheckPathSystemPaths(t *testing.T) {
	tests := []struct {
		path     string
		wantSafe bool
	}{
		{"/", false},
		{"/bin", false},
		{"/etc", false},
		{"/usr", false},
		{"/home/user/file.txt", true},
		{"/tmp/test", true},
	}

	if runtime.GOOS == "windows" {
		tests = []struct {
			path     string
			wantSafe bool
		}{
			{"C:\\Windows", false},
			{"C:\\Windows\\System32", false},
			{"C:\\Program Files", false},
			{"C:\\Users\\test\\Documents", true},
			{"D:\\data\\file.txt", true},
		}
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := CheckPath(tt.path, "delete")
			if result.IsSafe != tt.wantSafe {
				t.Errorf("CheckPath(%q) safe = %v, want %v; reason: %s",
					tt.path, result.IsSafe, tt.wantSafe, result.Reason)
			}
		})
	}
}

func TestCheckPathSensitivePatterns(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		path     string
		wantSafe bool
	}{
		{filepath.Join(home, ".ssh"), false},
		{filepath.Join(home, ".ssh", "id_rsa"), false},
		{filepath.Join(home, ".aws"), false},
		{filepath.Join(home, ".gnupg"), false},
		{filepath.Join(home, ".kube"), false},
		{filepath.Join(home, "documents", "file.txt"), true},
		{filepath.Join(home, "projects", "code"), true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := CheckPath(tt.path, "delete")
			if result.IsSafe != tt.wantSafe {
				t.Errorf("CheckPath(%q) safe = %v, want %v; reason: %s",
					tt.path, result.IsSafe, tt.wantSafe, result.Reason)
			}
		})
	}
}

func TestCheckPathSensitiveFiles(t *testing.T) {
	tests := []struct {
		path     string
		wantSafe bool
	}{
		{"/tmp/.env", false},
		{"/tmp/.env.local", false},
		{"/tmp/id_rsa", false},
		{"/tmp/server.pem", false},
		{"/tmp/cert.key", false},
		{"/tmp/secrets.json", false},
		{"/tmp/regular.txt", true},
		{"/tmp/config.yaml", true},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.path), func(t *testing.T) {
			result := CheckPath(tt.path, "delete")
			if result.IsSafe != tt.wantSafe {
				t.Errorf("CheckPath(%q) safe = %v, want %v; reason: %s",
					tt.path, result.IsSafe, tt.wantSafe, result.Reason)
			}
		})
	}
}

func TestCheckPathHomeDirectory(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	result := CheckPath(home, "delete")
	if result.IsSafe {
		t.Errorf("Home directory should not be safe for deletion")
	}

	if result.Severity != SeverityCritical {
		t.Errorf("Home directory deletion should be critical severity")
	}
}

func TestIsSafeForDeletion(t *testing.T) {
	// Test with a path that should be safe
	var safePath, unsafePath string
	if runtime.GOOS == "windows" {
		safePath = "C:\\Users\\test\\temp\\test.txt"
		unsafePath = "C:\\Windows"
	} else {
		safePath = "/tmp/test.txt"
		unsafePath = "/etc"
	}

	safe, reason := IsSafeForDeletion(safePath)
	if !safe {
		t.Errorf("IsSafeForDeletion(%s) = false, reason: %s", safePath, reason)
	}

	safe, _ = IsSafeForDeletion(unsafePath)
	if safe {
		t.Errorf("IsSafeForDeletion(%s) = true, should be false", unsafePath)
	}
}

func TestValidateDeletePaths(t *testing.T) {
	var paths []string
	if runtime.GOOS == "windows" {
		paths = []string{
			"C:\\Users\\test\\temp\\safe.txt",
			"C:\\Windows",
			"C:\\Users\\test\\temp\\another.txt",
			"C:\\Program Files",
		}
	} else {
		paths = []string{
			"/tmp/safe.txt",
			"/etc",
			"/tmp/another.txt",
			"/bin",
		}
	}

	unsafe := ValidateDeletePaths(paths)

	if len(unsafe) != 2 {
		t.Errorf("ValidateDeletePaths() returned %d unsafe paths, want 2", len(unsafe))
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityDanger, "DANGER"},
		{SeverityCritical, "CRITICAL"},
	}

	for _, tt := range tests {
		if got := tt.severity.String(); got != tt.want {
			t.Errorf("Severity(%d).String() = %q, want %q", tt.severity, got, tt.want)
		}
	}
}

func TestFormatError(t *testing.T) {
	result := CheckResult{
		Path:      "/etc",
		IsSafe:    false,
		Reason:    "protected system directory",
		Severity:  SeverityCritical,
		Overrides: []string{"--force"},
	}

	msg := FormatError(result)

	if msg == "" {
		t.Error("FormatError() returned empty for unsafe result")
	}

	if !contains(msg, "CRITICAL") {
		t.Error("FormatError() should contain severity")
	}

	if !contains(msg, "--force") {
		t.Error("FormatError() should contain override flag")
	}
}

func TestFormatErrorSafe(t *testing.T) {
	result := CheckResult{
		Path:   "/tmp/file.txt",
		IsSafe: true,
	}

	msg := FormatError(result)
	if msg != "" {
		t.Errorf("FormatError() for safe result = %q, want empty", msg)
	}
}

func TestPathEquals(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"/foo", "/foo", true},
		{"/foo", "/bar", false},
		{"/foo", "/FOO", runtime.GOOS == "windows"},
	}

	for _, tt := range tests {
		got := pathEquals(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("pathEquals(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIsParentOf(t *testing.T) {
	tests := []struct {
		parent, child string
		want          bool
	}{
		{"/usr", "/usr/bin", true},
		{"/usr", "/usr/local/bin", true},
		{"/usr", "/var", false},
		{"/usr/bin", "/usr", false},
	}

	for _, tt := range tests {
		got := isParentOf(tt.parent, tt.child)
		if got != tt.want {
			t.Errorf("isParentOf(%q, %q) = %v, want %v", tt.parent, tt.child, got, tt.want)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
