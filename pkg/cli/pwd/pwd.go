package pwd

import (
	"fmt"
	"io"
	"os"
)

// RunPwd prints the current working directory
func RunPwd(w io.Writer) error {
	wd, err := Pwd()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(w, wd)

	return nil
}

// Pwd returns the current working directory
func Pwd() (string, error) {
	return os.Getwd()
}
