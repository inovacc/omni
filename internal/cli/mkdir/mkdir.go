package mkdir

import (
	"errors"
	"fmt"
	"os"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// Options configures the mkdir command behavior
type Options struct {
	Parents bool // -p: make parent directories as needed
}

func RunMkdir(args []string, opts Options) error {
	if len(args) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "mkdir: missing operand")
	}

	for _, path := range args {
		var err error
		if opts.Parents {
			err = os.MkdirAll(path, 0755)
		} else {
			err = os.Mkdir(path, 0755)
		}

		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("mkdir: %s", err))
			}
			if errors.Is(err, os.ErrExist) {
				return cmderr.Wrap(cmderr.ErrConflict, fmt.Sprintf("mkdir: %s", err))
			}
			return fmt.Errorf("mkdir: %w", err)
		}
	}

	return nil
}

// RmdirOptions configures the rmdir command behavior
type RmdirOptions struct{}

func RunRmdir(args []string, _ RmdirOptions) error {
	if len(args) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "rmdir: missing operand")
	}

	for _, path := range args {
		err := os.Remove(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("rmdir: %s", err))
			}
			if errors.Is(err, os.ErrPermission) {
				return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("rmdir: %s", err))
			}
			return fmt.Errorf("rmdir: %w", err)
		}
	}

	return nil
}
