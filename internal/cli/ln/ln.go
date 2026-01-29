package ln

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LnOptions configures the ln command behavior
type LnOptions struct {
	Symbolic    bool // -s: make symbolic links instead of hard links
	Force       bool // -f: remove existing destination files
	Interactive bool // -i: prompt whether to remove destinations
	NoClobber   bool // -n: treat LINK_NAME as a normal file if it is a symlink to a directory
	Verbose     bool // -v: print name of each linked file
	Backup      bool // -b: make a backup of each existing destination file
	Relative    bool // -r: create symbolic links relative to link location
}

// RunLn creates links between files
func RunLn(w io.Writer, args []string, opts LnOptions) error {
	if len(args) < 2 {
		return fmt.Errorf("ln: missing file operand")
	}

	// Handle multiple sources -> directory case
	if len(args) > 2 {
		dest := args[len(args)-1]

		info, err := os.Stat(dest)
		if err != nil || !info.IsDir() {
			return fmt.Errorf("ln: target '%s' is not a directory", dest)
		}

		for _, src := range args[:len(args)-1] {
			linkName := filepath.Join(dest, filepath.Base(src))
			if err := createLink(w, src, linkName, opts); err != nil {
				return err
			}
		}

		return nil
	}

	// Two arguments: source and link name
	return createLink(w, args[0], args[1], opts)
}

func createLink(w io.Writer, target, linkName string, opts LnOptions) error {
	// Check if link already exists
	if _, err := os.Lstat(linkName); err == nil {
		if opts.Force {
			if err := os.Remove(linkName); err != nil {
				return fmt.Errorf("ln: cannot remove '%s': %w", linkName, err)
			}
		} else if opts.Backup {
			backupName := linkName + "~"
			if err := os.Rename(linkName, backupName); err != nil {
				return fmt.Errorf("ln: cannot backup '%s': %w", linkName, err)
			}
		} else {
			return fmt.Errorf("ln: failed to create link '%s': File exists", linkName)
		}
	}

	var err error

	if opts.Symbolic {
		// Handle relative symlinks
		actualTarget := target

		if opts.Relative {
			linkDir := filepath.Dir(linkName)

			relTarget, relErr := filepath.Rel(linkDir, target)
			if relErr == nil {
				actualTarget = relTarget
			}
		}

		err = os.Symlink(actualTarget, linkName)
	} else {
		err = os.Link(target, linkName)
	}

	if err != nil {
		return fmt.Errorf("ln: failed to create link '%s': %w", linkName, err)
	}

	if opts.Verbose {
		if opts.Symbolic {
			_, _ = fmt.Fprintf(w, "'%s' -> '%s'\n", linkName, target)
		} else {
			_, _ = fmt.Fprintf(w, "'%s' => '%s'\n", linkName, target)
		}
	}

	return nil
}

// Symlink creates a symbolic link
func Symlink(target, linkName string) error {
	return os.Symlink(target, linkName)
}

// Link creates a hard link
func Link(oldname, newname string) error {
	return os.Link(oldname, newname)
}
