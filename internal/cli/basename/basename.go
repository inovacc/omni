package basename

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/inovacc/omni/internal/cli/output"
)

// BasenameOptions configures the basename command behavior
type BasenameOptions struct {
	Suffix       string        // -s: suffix to remove
	OutputFormat output.Format // output format (text, json, table)
}

// BasenameResult represents basename output for JSON
type BasenameResult struct {
	Original string `json:"original"`
	Basename string `json:"basename"`
}

// RunBasename prints the base name of each path, optionally removing a suffix
func RunBasename(w io.Writer, args []string, opts BasenameOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("basename: missing operand")
	}

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var results []BasenameResult

	for _, arg := range args {
		name := filepath.Base(arg)
		if opts.Suffix != "" && len(name) > len(opts.Suffix) && name[len(name)-len(opts.Suffix):] == opts.Suffix {
			name = name[:len(name)-len(opts.Suffix)]
		}

		if jsonMode {
			results = append(results, BasenameResult{Original: arg, Basename: name})
		} else {
			_, _ = fmt.Fprintln(w, name)
		}
	}

	if jsonMode {
		return f.Print(results)
	}

	return nil
}
