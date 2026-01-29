package safepath

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ProtectedPaths contains paths that should be protected from destructive operations
var ProtectedPaths = []string{
	// System directories
	"/",
	"/bin",
	"/boot",
	"/dev",
	"/etc",
	"/lib",
	"/lib64",
	"/opt",
	"/proc",
	"/root",
	"/run",
	"/sbin",
	"/srv",
	"/sys",
	"/usr",
	"/var",

	// Windows system directories
	"C:\\Windows",
	"C:\\Windows\\System32",
	"C:\\Program Files",
	"C:\\Program Files (x86)",
	"C:\\ProgramData",

	// macOS system directories
	"/System",
	"/Library",
	"/Applications",
	"/private",
	"/cores",
}

// ProtectedPatterns contains path patterns that should be protected
var ProtectedPatterns = []string{
	// SSH keys
	".ssh",
	// GPG keys
	".gnupg",
	// AWS credentials
	".aws",
	// Docker
	".docker",
	// Kubernetes
	".kube",
	// Git credentials
	".gitconfig",
	".git-credentials",
	// NPM
	".npmrc",
	// Python
	".pypirc",
	// Cryptocurrency wallets
	".bitcoin",
	".ethereum",
	".monero",
	// Password managers
	".password-store",
	".1password",
	// Browser data
	".mozilla",
	".config/google-chrome",
	".config/chromium",
	// Shell history (contains potentially sensitive commands)
	".bash_history",
	".zsh_history",
	".histfile",
}

// SensitiveFilePatterns matches files that may contain secrets
var SensitiveFilePatterns = []string{
	".env",
	".env.*",
	"*.pem",
	"*.key",
	"*.p12",
	"*.pfx",
	"id_rsa",
	"id_ed25519",
	"id_ecdsa",
	"id_dsa",
	"*.kdbx", // KeePass
	"*.kwallet",
	"secrets.*",
	"credentials.*",
	"*.credentials",
	"*.secret",
	"*.secrets",
}

// CheckResult represents the result of a safety check
type CheckResult struct {
	Path      string
	IsSafe    bool
	Reason    string
	Severity  Severity
	Overrides []string // Flags that can override this check
}

// Severity indicates how dangerous an operation is
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityDanger
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityDanger:
		return "DANGER"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// CheckPath checks if a path is safe for destructive operations
func CheckPath(path string, _ string) CheckResult {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return CheckResult{
			Path:   path,
			IsSafe: false,
			Reason: fmt.Sprintf("cannot resolve path: %v", err),
		}
	}

	// Normalize path separators
	normalizedPath := filepath.Clean(absPath)

	// Check if it's a protected system path
	if result := checkSystemPath(normalizedPath); !result.IsSafe {
		return result
	}

	// Check if it contains sensitive patterns
	if result := checkSensitivePath(normalizedPath); !result.IsSafe {
		return result
	}

	// Check if it matches sensitive file patterns
	if result := checkSensitiveFile(normalizedPath); !result.IsSafe {
		return result
	}

	// Check if it's a home directory root
	if result := checkHomeDirectory(normalizedPath); !result.IsSafe {
		return result
	}

	return CheckResult{
		Path:   path,
		IsSafe: true,
	}
}

// checkSystemPath checks if path is a protected system directory
func checkSystemPath(path string) CheckResult {
	for _, protected := range ProtectedPaths {
		// Normalize the protected path
		protected = filepath.Clean(protected)

		// Exact match or is parent
		if pathEquals(path, protected) {
			return CheckResult{
				Path:      path,
				IsSafe:    false,
				Reason:    fmt.Sprintf("'%s' is a protected system directory", path),
				Severity:  SeverityCritical,
				Overrides: []string{"--force", "--no-preserve-root"},
			}
		}

		// Check if we're trying to operate on a parent of a protected path
		if isParentOf(path, protected) {
			return CheckResult{
				Path:      path,
				IsSafe:    false,
				Reason:    fmt.Sprintf("'%s' contains protected system directories", path),
				Severity:  SeverityCritical,
				Overrides: []string{"--force", "--no-preserve-root"},
			}
		}
	}

	return CheckResult{IsSafe: true}
}

// checkSensitivePath checks if path contains sensitive data patterns
func checkSensitivePath(path string) CheckResult {
	for _, pattern := range ProtectedPatterns {
		// Check if path contains this pattern
		if strings.Contains(path, pattern) || strings.HasSuffix(path, pattern) {
			return CheckResult{
				Path:      path,
				IsSafe:    false,
				Reason:    fmt.Sprintf("'%s' appears to contain sensitive data (%s)", path, pattern),
				Severity:  SeverityDanger,
				Overrides: []string{"--force"},
			}
		}
	}

	return CheckResult{IsSafe: true}
}

// checkSensitiveFile checks if a file matches sensitive patterns
func checkSensitiveFile(path string) CheckResult {
	basename := filepath.Base(path)

	for _, pattern := range SensitiveFilePatterns {
		matched, err := filepath.Match(pattern, basename)
		if err == nil && matched {
			return CheckResult{
				Path:      path,
				IsSafe:    false,
				Reason:    fmt.Sprintf("'%s' appears to be a sensitive file (matches %s)", path, pattern),
				Severity:  SeverityWarning,
				Overrides: []string{"--force"},
			}
		}
	}

	return CheckResult{IsSafe: true}
}

// checkHomeDirectory checks if trying to delete home directory root
func checkHomeDirectory(path string) CheckResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return CheckResult{IsSafe: true}
	}

	home = filepath.Clean(home)
	path = filepath.Clean(path)

	if pathEquals(path, home) {
		return CheckResult{
			Path:      path,
			IsSafe:    false,
			Reason:    "cannot delete home directory",
			Severity:  SeverityCritical,
			Overrides: []string{"--force", "--no-preserve-root"},
		}
	}

	return CheckResult{IsSafe: true}
}

// pathEquals compares two paths for equality (case-insensitive on Windows)
func pathEquals(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}

	return a == b
}

// isParentOf checks if parent is a parent directory of child
func isParentOf(parent, child string) bool {
	parent = filepath.Clean(parent) + string(filepath.Separator)
	child = filepath.Clean(child)

	if runtime.GOOS == "windows" {
		return strings.HasPrefix(strings.ToLower(child), strings.ToLower(parent))
	}

	return strings.HasPrefix(child, parent)
}

// IsSafeForDeletion is a convenience function for rm/rmdir operations
func IsSafeForDeletion(path string) (bool, string) {
	result := CheckPath(path, "delete")
	return result.IsSafe, result.Reason
}

// IsSafeForOverwrite is a convenience function for write operations
func IsSafeForOverwrite(path string) (bool, string) {
	result := CheckPath(path, "overwrite")
	return result.IsSafe, result.Reason
}

// ValidateDeletePaths checks multiple paths and returns all unsafe ones
func ValidateDeletePaths(paths []string) []CheckResult {
	var unsafe []CheckResult

	for _, path := range paths {
		result := CheckPath(path, "delete")
		if !result.IsSafe {
			unsafe = append(unsafe, result)
		}
	}

	return unsafe
}

// FormatError formats a CheckResult as an error message
func FormatError(result CheckResult) string {
	if result.IsSafe {
		return ""
	}

	msg := fmt.Sprintf("[%s] %s", result.Severity, result.Reason)

	if len(result.Overrides) > 0 {
		msg += fmt.Sprintf("\nUse %s to override this check", strings.Join(result.Overrides, " or "))
	}

	return msg
}
