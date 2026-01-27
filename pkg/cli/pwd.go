package cli

import (
	"fmt"
	"os"
)

func RunPwd() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Println(wd)
	return nil
}
