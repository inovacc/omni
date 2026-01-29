package realpath

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
)

// RealpathOptions configures the realpath command behavior
type RealpathOptions struct {
	JSON bool // --json: output as JSON
}

// RealpathResult represents realpath output for JSON
type RealpathResult struct {
	Original string `json:"original"`
	Resolved string `json:"resolved"`
}

// RunRealpath prints the resolved absolute path for each argument
func RunRealpath(w io.Writer, args []string, opts RealpathOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("realpath: missing operand")
	}

	var results []RealpathResult

	for _, arg := range args {
		absPath, err := filepath.Abs(arg)
		if err != nil {
			return err
		}

		resolved, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return err
		}

		if opts.JSON {
			results = append(results, RealpathResult{Original: arg, Resolved: resolved})
		} else {
			_, _ = fmt.Fprintln(w, resolved)
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(results)
	}

	return nil
}
