package dirname

import (
	"fmt"
	"io"
	"path/filepath"
)

// RunDirname prints the directory portion of each path
func RunDirname(w io.Writer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("dirname: missing operand")
	}

	for _, arg := range args {
		_, _ = fmt.Fprintln(w, filepath.Dir(arg))
	}

	return nil
}
