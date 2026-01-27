package cli

import (
	"fmt"
	"github.com/inovacc/goshell/pkg/fs"
)

func RunRealpath(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("realpath: missing operand")
	}

	for _, arg := range args {
		abs, err := fs.Realpath(arg)
		if err != nil {
			return err
		}
		fmt.Println(abs)
	}
	return nil
}
