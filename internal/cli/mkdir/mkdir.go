package mkdir

import (
	"fmt"
	"os"
)

// Options configures the mkdir command behavior
type Options struct {
	Parents bool // -p: make parent directories as needed
}

func RunMkdir(args []string, opts Options) error {
	if len(args) == 0 {
		return fmt.Errorf("mkdir: missing operand")
	}

	for _, path := range args {
		var err error
		if opts.Parents {
			err = os.MkdirAll(path, 0755)
		} else {
			err = os.Mkdir(path, 0755)
		}

		if err != nil {
			return fmt.Errorf("mkdir: %w", err)
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
