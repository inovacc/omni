package cli

import (
	"fmt"
)

func RunRealpath(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("realpath: missing operand")
	}

	for _, arg := range args {
		abs, err := Realpath(arg)
		if err != nil {
			return err
		}
		fmt.Println(abs)
	}
	return nil
}
