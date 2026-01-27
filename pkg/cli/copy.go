package cli

import (
	"fmt"
	"path/filepath"
)

func RunCopy(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("cp: missing file operand")
	}

	dest := args[len(args)-1]
	srcs := args[:len(args)-1]

	destStat, err := Stat(dest)
	destIsDir := err == nil && destStat.IsDir()

	if len(srcs) > 1 && !destIsDir {
		return fmt.Errorf("cp: target '%s' is not a directory", dest)
	}

	for _, src := range srcs {
		target := dest
		if destIsDir {
			target = filepath.Join(dest, filepath.Base(src))
		}
		err := Copy(src, target)
		if err != nil {
			return fmt.Errorf("cp: %w", err)
		}
	}
	return nil
}

func RunMove(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("mv: missing file operand")
	}

	dest := args[len(args)-1]
	srcs := args[:len(args)-1]

	destStat, err := Stat(dest)
	destIsDir := err == nil && destStat.IsDir()

	if len(srcs) > 1 && !destIsDir {
		return fmt.Errorf("mv: target '%s' is not a directory", dest)
	}

	for _, src := range srcs {
		target := dest
		if destIsDir {
			target = filepath.Join(dest, filepath.Base(src))
		}
		err := Move(src, target)
		if err != nil {
			return fmt.Errorf("mv: %w", err)
		}
	}
	return nil
}
