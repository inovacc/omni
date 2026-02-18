package copy

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// CopyOptions configures the copy command behavior
type CopyOptions struct {
	Recursive bool // -r/-R: copy directories recursively
}

func RunCopy(args []string, _ CopyOptions) error {
	if len(args) < 2 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "cp: missing file operand")
	}

	dest := args[len(args)-1]
	srcs := args[:len(args)-1]

	destStat, err := os.Stat(dest)
	destIsDir := err == nil && destStat.IsDir()

	if len(srcs) > 1 && !destIsDir {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("cp: target '%s' is not a directory", dest))
	}

	for _, src := range srcs {
		target := dest
		if destIsDir {
			target = filepath.Join(dest, filepath.Base(src))
		}

		if err := copyFile(src, target); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("cp: %s", err))
			}
			if errors.Is(err, os.ErrPermission) {
				return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("cp: %s", err))
			}
			return fmt.Errorf("cp: %w", err)
		}
	}

	return nil
}

// MoveOptions configures the move command behavior
type MoveOptions struct{}

func RunMove(args []string, _ MoveOptions) error {
	if len(args) < 2 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "mv: missing file operand")
	}

	dest := args[len(args)-1]
	srcs := args[:len(args)-1]

	destStat, err := os.Stat(dest)
	destIsDir := err == nil && destStat.IsDir()

	if len(srcs) > 1 && !destIsDir {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("mv: target '%s' is not a directory", dest))
	}

	for _, src := range srcs {
		target := dest
		if destIsDir {
			target = filepath.Join(dest, filepath.Base(src))
		}

		if err := os.Rename(src, target); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("mv: %s", err))
			}
			if errors.Is(err, os.ErrPermission) {
				return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("mv: %s", err))
			}
			return fmt.Errorf("mv: %w", err)
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}

	defer func() {
		_ = source.Close()
	}()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer func() {
		_ = destination.Close()
	}()

	_, err = io.Copy(destination, source)

	return err
}
