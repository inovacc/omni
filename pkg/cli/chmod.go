package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ChmodOptions configures the chmod command behavior
type ChmodOptions struct {
	Recursive bool   // -R: change files and directories recursively
	Verbose   bool   // -v: output a diagnostic for every file processed
	Changes   bool   // -c: like verbose but report only when a change is made
	Silent    bool   // -f: suppress most error messages
	Reference string // --reference: use RFILE's mode instead of MODE values
}

// RunChmod changes file mode bits
func RunChmod(w io.Writer, args []string, opts ChmodOptions) error {
	if len(args) < 2 {
		return fmt.Errorf("chmod: missing operand")
	}

	modeStr := args[0]
	files := args[1:]

	// Parse mode
	var (
		newMode    fs.FileMode
		isSymbolic bool
		symbolicOp func(fs.FileMode) fs.FileMode
	)

	switch {
	case opts.Reference != "":
		// Use reference file's mode
		info, err := os.Stat(opts.Reference)
		if err != nil {
			return fmt.Errorf("chmod: cannot stat '%s': %w", opts.Reference, err)
		}

		newMode = info.Mode().Perm()
	case isOctalMode(modeStr):
		// Octal mode (e.g., 755, 0644)
		mode, err := strconv.ParseUint(modeStr, 8, 32)
		if err != nil {
			return fmt.Errorf("chmod: invalid mode: '%s'", modeStr)
		}

		newMode = fs.FileMode(mode)
	default:
		// Symbolic mode (e.g., u+x, go-w, a=rw)
		isSymbolic = true

		var err error

		symbolicOp, err = parseSymbolicMode(modeStr)
		if err != nil {
			return fmt.Errorf("chmod: invalid mode: '%s'", modeStr)
		}
	}

	for _, file := range files {
		if opts.Recursive {
			err := filepath.WalkDir(file, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					if !opts.Silent {
						_, _ = fmt.Fprintf(os.Stderr, "chmod: cannot access '%s': %v\n", path, err)
					}

					return nil
				}

				return chmodFile(w, path, newMode, isSymbolic, symbolicOp, opts)
			})
			if err != nil {
				return err
			}
		} else {
			if err := chmodFile(w, file, newMode, isSymbolic, symbolicOp, opts); err != nil {
				if !opts.Silent {
					_, _ = fmt.Fprintf(os.Stderr, "chmod: %v\n", err)
				}
			}
		}
	}

	return nil
}

func chmodFile(w io.Writer, path string, newMode fs.FileMode, isSymbolic bool, symbolicOp func(fs.FileMode) fs.FileMode, opts ChmodOptions) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access '%s': %w", path, err)
	}

	oldMode := info.Mode().Perm()

	var targetMode fs.FileMode

	if isSymbolic {
		targetMode = symbolicOp(oldMode)
	} else {
		targetMode = newMode
	}

	if err := os.Chmod(path, targetMode); err != nil {
		return fmt.Errorf("changing permissions of '%s': %w", path, err)
	}

	if opts.Verbose || (opts.Changes && oldMode != targetMode) {
		_, _ = fmt.Fprintf(w, "mode of '%s' changed from %04o to %04o\n", path, oldMode, targetMode)
	}

	return nil
}

func isOctalMode(s string) bool {
	for _, c := range s {
		if c < '0' || c > '7' {
			return false
		}
	}

	return len(s) > 0
}

//nolint:unparam // error kept for API consistency with parseOctalMode
func parseSymbolicMode(mode string) (func(fs.FileMode) fs.FileMode, error) {
	// Parse symbolic mode like u+x, go-w, a=rw, +x, etc.
	return func(current fs.FileMode) fs.FileMode {
		result := current

		// Split by comma for multiple operations
		parts := strings.SplitSeq(mode, ",")

		for part := range parts {
			result = applySymbolicPart(result, part)
		}

		return result
	}, nil
}

func applySymbolicPart(mode fs.FileMode, part string) fs.FileMode {
	// Parse who (u, g, o, a or empty for all)
	who := ""

	i := 0
	for i < len(part) && (part[i] == 'u' || part[i] == 'g' || part[i] == 'o' || part[i] == 'a') {
		who += string(part[i])
		i++
	}

	if who == "" || who == "a" {
		who = "ugo"
	}

	if i >= len(part) {
		return mode
	}

	// Parse operator (+, -, =)
	op := part[i]
	i++

	if i >= len(part) {
		return mode
	}

	// Parse permissions (r, w, x, X, s, t)
	perms := part[i:]

	// Calculate permission bits
	var bits fs.FileMode

	for _, p := range perms {
		switch p {
		case 'r':
			if strings.Contains(who, "u") {
				bits |= 0400
			}

			if strings.Contains(who, "g") {
				bits |= 0040
			}

			if strings.Contains(who, "o") {
				bits |= 0004
			}
		case 'w':
			if strings.Contains(who, "u") {
				bits |= 0200
			}

			if strings.Contains(who, "g") {
				bits |= 0020
			}

			if strings.Contains(who, "o") {
				bits |= 0002
			}
		case 'x':
			if strings.Contains(who, "u") {
				bits |= 0100
			}

			if strings.Contains(who, "g") {
				bits |= 0010
			}

			if strings.Contains(who, "o") {
				bits |= 0001
			}
		}
	}

	switch op {
	case '+':
		mode |= bits
	case '-':
		mode &^= bits
	case '=':
		// Clear existing bits for the specified users, then set new ones
		var clearMask fs.FileMode
		if strings.Contains(who, "u") {
			clearMask |= 0700
		}

		if strings.Contains(who, "g") {
			clearMask |= 0070
		}

		if strings.Contains(who, "o") {
			clearMask |= 0007
		}

		mode = (mode &^ clearMask) | bits
	}

	return mode
}
