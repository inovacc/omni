package rm

import (
	"fmt"
	"os"
)

// RmOptions configures the rm command behavior
type RmOptions struct {
	Recursive bool // -r/-R: remove directories and their contents recursively
	Force     bool // -f: ignore nonexistent files, never prompt
}

func RunRm(args []string, opts RmOptions) error {
	if len(args) == 0 {
		if opts.Force {
			return nil
		}

		return fmt.Errorf("rm: missing operand")
	}

	for _, path := range args {
		var err error
		if opts.Recursive {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}

		if err != nil {
			if opts.Force && os.IsNotExist(err) {
				continue
			}

			return fmt.Errorf("rm: %w", err)
		}
	}

	return nil
}

// RmdirOptions configures the rmdir command behavior
type RmdirOptions struct{}

func RunRmdir(args []string, _ RmdirOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("rmdir: missing operand")
	}

	for _, path := range args {
		err := os.Remove(path)
		if err != nil {
			return fmt.Errorf("rmdir: %w", err)
		}
	}

	return nil
}
