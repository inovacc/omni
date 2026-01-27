package cli

import (
	"fmt"
	"github.com/inovacc/goshell/pkg/fs"
)

func RunDirname(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("dirname: missing operand")
	}

	for _, arg := range args {
		fmt.Println(fs.Dirname(arg))
	}
	return nil
}
