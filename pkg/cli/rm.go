package cli

import (
	"fmt"
)

func RunRm(args []string, recursive bool, force bool) error {
	if len(args) == 0 {
		if force {
			return nil
		}
		return fmt.Errorf("rm: missing operand")
	}

	for _, path := range args {
		err := Rm(path, recursive)
		if err != nil {
			if force && (IsNotExist(err)) {
				continue
			}
			return fmt.Errorf("rm: %w", err)
		}
	}
	return nil
}
