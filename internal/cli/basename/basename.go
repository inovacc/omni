package basename

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
)

// BasenameOptions configures the basename command behavior
type BasenameOptions struct {
	Suffix string // -s: suffix to remove
	JSON   bool   // --json: output as JSON
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

	var results []BasenameResult

	for _, arg := range args {
		name := filepath.Base(arg)
		if opts.Suffix != "" && len(name) > len(opts.Suffix) && name[len(name)-len(opts.Suffix):] == opts.Suffix {
			name = name[:len(name)-len(opts.Suffix)]
		}

		if opts.JSON {
			results = append(results, BasenameResult{Original: arg, Basename: name})
		} else {
			_, _ = fmt.Fprintln(w, name)
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(results)
	}

	return nil
}
