package pwd

import (
	"fmt"
	"io"
	"os"

	"github.com/inovacc/omni/internal/cli/output"
)

// PwdOptions configures the pwd command behavior
type PwdOptions struct {
	OutputFormat output.Format // output format (text, json, table)
}

// PwdResult represents pwd output for JSON
type PwdResult struct {
	Path string `json:"path"`
}

// RunPwd prints the current working directory
func RunPwd(w io.Writer, opts PwdOptions) error {
	wd, err := Pwd()
	if err != nil {
		return err
	}

	f := output.New(w, opts.OutputFormat)

	if f.IsJSON() {
		return f.Print(PwdResult{Path: wd})
	}

	_, _ = fmt.Fprintln(w, wd)

	return nil
}

// Pwd returns the current working directory
func Pwd() (string, error) {
	return os.Getwd()
}
