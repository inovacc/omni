package cli

import (
	"fmt"
)

func RunDirname(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("dirname: missing operand")
	}

	for _, arg := range args {
		fmt.Println(Dirname(arg))
	}
	return nil
}
