package cli

import (
	"fmt"
	"github.com/inovacc/goshell/pkg/fs"
)

func RunLs(args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	entries, err := fs.Ls(dir)
	if err != nil {
		return err
	}

	for _, name := range entries {
		fmt.Println(name)
	}
	return nil
}
