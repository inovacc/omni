package rm

import (
	"errors"
	"fmt"
	"os"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/safepath"
)

// RmOptions configures the rm command behavior
type RmOptions struct {
	Recursive      bool // -r/-R: remove directories and their contents recursively
	Force          bool // -f: ignore nonexistent files, never prompt
	NoPreserveRoot bool // --no-preserve-root: allow deleting protected paths
}

func RunRm(args []string, opts RmOptions) error {
	if len(args) == 0 {
		if opts.Force {
			return nil
		}

		return cmderr.Wrap(cmderr.ErrInvalidInput, "rm: missing operand")
	}

	// Safety check: validate all paths before any deletion
	if !opts.NoPreserveRoot {
		for _, path := range args {
			result := safepath.CheckPath(path, "delete")
			if !result.IsSafe {
				if !opts.Force {
					return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("rm: %s\nUse --force or --no-preserve-root to override", result.Reason))
				}

				// Even with --force, critical paths require --no-preserve-root
				if result.Severity == safepath.SeverityCritical {
					return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("rm: %s\nUse --no-preserve-root to override critical protection", result.Reason))
				}
			}
		}
	}

	for _, path := range args {
		var err error
		if opts.Recursive {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}

		if err != nil {
			if opts.Force && errors.Is(err, os.ErrNotExist) {
				continue
			}
			if errors.Is(err, os.ErrNotExist) {
				return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("rm: %s", err))
			}
			if errors.Is(err, os.ErrPermission) {
				return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("rm: %s", err))
			}
			return fmt.Errorf("rm: %w", err)
		}
	}

	return nil
}

// RmdirOptions configures the rmdir command behavior
type RmdirOptions struct {
	NoPreserveRoot bool // --no-preserve-root: allow deleting protected paths
}

func RunRmdir(args []string, opts RmdirOptions) error {
	if len(args) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "rmdir: missing operand")
	}

	// Safety check: validate all paths before any deletion
	if !opts.NoPreserveRoot {
		for _, path := range args {
			result := safepath.CheckPath(path, "delete")
			if !result.IsSafe {
				return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("rmdir: %s\nUse --no-preserve-root to override", result.Reason))
			}
		}
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
