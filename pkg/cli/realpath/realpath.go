package realpath

import (
	"fmt"
	"io"
	"path/filepath"
)

// RunRealpath prints the resolved absolute path for each argument
func RunRealpath(w io.Writer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("realpath: missing operand")
	}

	for _, arg := range args {
		absPath, err := filepath.Abs(arg)
		if err != nil {
			return err
		}

		resolved, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(w, resolved)
	}

	return nil
}
