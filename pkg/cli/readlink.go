package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ReadlinkOptions configures the readlink command behavior
type ReadlinkOptions struct {
	Canonicalize         bool // -f: canonicalize by following every symlink, all components must exist
	CanonicalizeExisting bool // -e: like -f, but fail if any component doesn't exist
	CanonicalizeMissing  bool // -m: like -f, but allow missing components
	NoNewline            bool // -n: do not output trailing newline
	Quiet                bool // -q: quiet mode
	Silent               bool // -s: silent mode (same as -q)
	Verbose              bool // -v: verbose mode
	Zero                 bool // -z: end each output line with NUL, not newline
}

// RunReadlink prints symbolic link targets or canonical file names
func RunReadlink(w io.Writer, args []string, opts ReadlinkOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("readlink: missing operand")
	}

	terminator := "\n"
	if opts.Zero {
		terminator = "\x00"
	}

	if opts.NoNewline && len(args) == 1 {
		terminator = ""
	}

	hasError := false

	for _, path := range args {
		var (
			result string
			err    error
		)

		if opts.Canonicalize || opts.CanonicalizeExisting || opts.CanonicalizeMissing {
			result, err = canonicalize(path, opts)
		} else {
			// Just read the symlink target
			result, err = os.Readlink(path)
		}

		if err != nil {
			if !opts.Quiet && !opts.Silent {
				_, _ = fmt.Fprintf(os.Stderr, "readlink: %s: %v\n", path, err)
			}

			hasError = true

			continue
		}

		_, _ = fmt.Fprint(w, result+terminator)
	}

	if hasError {
		return fmt.Errorf("readlink: some operations failed")
	}

	return nil
}

func canonicalize(path string, opts ReadlinkOptions) (string, error) {
	if opts.CanonicalizeMissing {
		// Allow missing components - resolve what exists
		return filepath.Abs(path)
	}

	if opts.CanonicalizeExisting {
		// All components must exist
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		// Verify path exists
		if _, err := os.Stat(absPath); err != nil {
			return "", err
		}

		return filepath.EvalSymlinks(absPath)
	}

	// Default -f: follow symlinks, all components must exist
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return filepath.EvalSymlinks(absPath)
}

// Readlink reads the target of a symbolic link
func Readlink(path string) (string, error) {
	return os.Readlink(path)
}

// CanonicalPath returns the canonical absolute path
func CanonicalPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return filepath.EvalSymlinks(absPath)
}
