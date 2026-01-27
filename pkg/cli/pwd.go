package cli

import (
	"fmt"
	"github.com/inovacc/goshell/pkg/fs"
)

func RunPwd() error {
	wd, err := fs.Pwd()
	if err != nil {
		return err
	}
	fmt.Println(wd)
	return nil
}
