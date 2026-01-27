package cli

import (
	"fmt"
)

func RunBasename(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("basename: missing operand")
	}

	name := Basename(args[0])
	if len(args) > 1 {
		suffix := args[1]
		if len(name) > len(suffix) && name[len(name)-len(suffix):] == suffix {
			name = name[:len(name)-len(suffix)]
		}
	}

	fmt.Println(name)
	return nil
}
