package cli

import (
	"fmt"
)

func RunMkdir(args []string, parents bool) error {
	if len(args) == 0 {
		return fmt.Errorf("mkdir: missing operand")
	}

	for _, path := range args {
		err := Mkdir(path, 0755, parents)
		if err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
	}

	return nil
}

func RunRmdir(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("rmdir: missing operand")
	}

	for _, path := range args {
		err := Rmdir(path)
		if err != nil {
			return fmt.Errorf("rmdir: %w", err)
		}
	}

	return nil
}
