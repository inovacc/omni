package dirname

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/inovacc/omni/internal/cli/output"
)

// DirnameOptions configures the dirname command behavior
type DirnameOptions struct {
	OutputFormat output.Format // output format (text, json, table)
}

// DirnameResult represents dirname output for JSON
type DirnameResult struct {
	Original string `json:"original"`
	Dirname  string `json:"dirname"`
}

// RunDirname prints the directory portion of each path
func RunDirname(w io.Writer, args []string, opts DirnameOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("dirname: missing operand")
	}

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var results []DirnameResult

	for _, arg := range args {
		dir := filepath.Dir(arg)
		if jsonMode {
			results = append(results, DirnameResult{Original: arg, Dirname: dir})
		} else {
			_, _ = fmt.Fprintln(w, dir)
		}
	}

	if jsonMode {
		return f.Print(results)
	}

	return nil
}
