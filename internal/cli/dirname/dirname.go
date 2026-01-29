package dirname

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
)

// DirnameOptions configures the dirname command behavior
type DirnameOptions struct {
	JSON bool // --json: output as JSON
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

	var results []DirnameResult

	for _, arg := range args {
		dir := filepath.Dir(arg)
		if opts.JSON {
			results = append(results, DirnameResult{Original: arg, Dirname: dir})
		} else {
			_, _ = fmt.Fprintln(w, dir)
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(results)
	}

	return nil
}
