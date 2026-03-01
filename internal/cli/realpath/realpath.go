package realpath

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

// RealpathOptions configures the realpath command behavior
type RealpathOptions struct {
	OutputFormat output.Format // output format (text, json, table)
}

// RealpathResult represents realpath output for JSON
type RealpathResult struct {
	Original string `json:"original"`
	Resolved string `json:"resolved"`
}

// RunRealpath prints the resolved absolute path for each argument
func RunRealpath(w io.Writer, args []string, opts RealpathOptions) error {
	if len(args) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "realpath: missing operand")
	}

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var results []RealpathResult

	for _, arg := range args {
		absPath, err := filepath.Abs(arg)
		if err != nil {
			return fmt.Errorf("realpath: %w", err)
		}

		resolved, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("realpath: %s", err))
			}
			return fmt.Errorf("realpath: %w", err)
		}

		if jsonMode {
			results = append(results, RealpathResult{Original: arg, Resolved: resolved})
		} else {
			_, _ = fmt.Fprintln(w, resolved)
		}
	}

	if jsonMode {
		return f.Print(results)
	}

	return nil
}
