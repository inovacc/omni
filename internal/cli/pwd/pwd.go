package pwd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// PwdOptions configures the pwd command behavior
type PwdOptions struct {
	JSON bool // --json: output as JSON
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

	if opts.JSON {
		return json.NewEncoder(w).Encode(PwdResult{Path: wd})
	}

	_, _ = fmt.Fprintln(w, wd)

	return nil
}

// Pwd returns the current working directory
func Pwd() (string, error) {
	return os.Getwd()
}
