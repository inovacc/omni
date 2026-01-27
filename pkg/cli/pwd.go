package cli

import (
	"fmt"
	"os"
)

func RunPwd() error {
	wd, err := Pwd()
	if err != nil {
		return err
	}
	fmt.Println(wd)
	return nil
}

func Pwd() (string, error) {
	return os.Getwd()
}
