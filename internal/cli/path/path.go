package path

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/output"
)

func Realpath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return filepath.EvalSymlinks(abs)
}

func Dirname(path string) string {
	return filepath.Dir(path)
}

func Basename(path string) string {
	return filepath.Base(path)
}

func Join(paths ...string) string {
	return filepath.Join(paths...)
}

func Clean(p string) string {
	return filepath.Clean(p)
}

func Abs(p string) (string, error) {
	return filepath.Abs(p)
}

// CleanOptions configures the path clean command behavior
type CleanOptions struct {
	OutputFormat output.Format // output format (text, json, table)
}

// CleanResult represents clean output for JSON
type CleanResult struct {
	Original string `json:"original"`
	Cleaned  string `json:"cleaned"`
}

// RunClean prints the cleaned path for each argument
func RunClean(w io.Writer, args []string, opts CleanOptions) error {
	if len(args) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "path clean: missing operand")
	}

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var results []CleanResult

	for _, arg := range args {
		cleaned := filepath.Clean(arg)
		if jsonMode {
			results = append(results, CleanResult{Original: arg, Cleaned: cleaned})
		} else {
			_, _ = fmt.Fprintln(w, cleaned)
		}
	}

	if jsonMode {
		return f.Print(results)
	}

	return nil
}

// AbsOptions configures the path abs command behavior
type AbsOptions struct {
	OutputFormat output.Format // output format (text, json, table)
}

// AbsResult represents abs output for JSON
type AbsResult struct {
	Original string `json:"original"`
	Absolute string `json:"absolute"`
}

// RunAbs prints the absolute path for each argument
func RunAbs(w io.Writer, args []string, opts AbsOptions) error {
	if len(args) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "path abs: missing operand")
	}

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var results []AbsResult

	for _, arg := range args {
		abs, err := filepath.Abs(arg)
		if err != nil {
			return fmt.Errorf("path abs: %w", err)
		}

		if jsonMode {
			results = append(results, AbsResult{Original: arg, Absolute: abs})
		} else {
			_, _ = fmt.Fprintln(w, abs)
		}
	}

	if jsonMode {
		return f.Print(results)
	}

	return nil
}
