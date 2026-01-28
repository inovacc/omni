package copy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyOptions configures the copy command behavior
type CopyOptions struct {
	Recursive bool // -r/-R: copy directories recursively
}

func RunCopy(args []string, _ CopyOptions) error {
	if len(args) < 2 {
		return fmt.Errorf("cp: missing file operand")
	}

	dest := args[len(args)-1]
	srcs := args[:len(args)-1]

	destStat, err := os.Stat(dest)
	destIsDir := err == nil && destStat.IsDir()

	if len(srcs) > 1 && !destIsDir {
		return fmt.Errorf("cp: target '%s' is not a directory", dest)
	}

	for _, src := range srcs {
		target := dest
		if destIsDir {
			target = filepath.Join(dest, filepath.Base(src))
		}

		if err := copyFile(src, target); err != nil {
			return fmt.Errorf("cp: %w", err)
		}
	}

	return nil
}

// MoveOptions configures the move command behavior
type MoveOptions struct{}

func RunMove(args []string, _ MoveOptions) error {
	if len(args) < 2 {
		return fmt.Errorf("mv: missing file operand")
	}

	dest := args[len(args)-1]
	srcs := args[:len(args)-1]

	destStat, err := os.Stat(dest)
	destIsDir := err == nil && destStat.IsDir()

	if len(srcs) > 1 && !destIsDir {
		return fmt.Errorf("mv: target '%s' is not a directory", dest)
	}

	for _, src := range srcs {
		target := dest
		if destIsDir {
			target = filepath.Join(dest, filepath.Base(src))
		}

		if err := os.Rename(src, target); err != nil {
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
