package cli

import (
	"fmt"
	"io"
)

// RunBasename prints the base name of each path, optionally removing a suffix
func RunBasename(w io.Writer, args []string, suffix string) error {
	if len(args) == 0 {
		return fmt.Errorf("basename: missing operand")
	}

	for _, arg := range args {
		name := Basename(arg)
		if suffix != "" && len(name) > len(suffix) && name[len(name)-len(suffix):] == suffix {
			name = name[:len(name)-len(suffix)]
		}
		_, _ = fmt.Fprintln(w, name)
	}
	return nil
}
