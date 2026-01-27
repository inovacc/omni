package cli

import (
	"fmt"
	"io"
)

// RunRealpath prints the resolved absolute path for each argument
func RunRealpath(w io.Writer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("realpath: missing operand")
	}

	for _, arg := range args {
		abs, err := Realpath(arg)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(w, abs)
	}

	return nil
}
