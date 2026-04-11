package cssfmt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
	pkgcss "github.com/inovacc/omni/pkg/cssfmt"
)

// Options configures the CSS formatter
type Options struct {
	Indent    string // Indentation (default: "  ")
	Minify    bool   // Minify output
	SortProps bool   // Sort properties alphabetically
	SortRules bool   // Sort selectors alphabetically
}

// ValidateOptions configures CSS validation
type ValidateOptions struct {
	OutputFormat output.Format // Output format
}

// ValidateResult is an alias for the pkg type
type ValidateResult = pkgcss.ValidateResult

// Run formats CSS input
func Run(w io.Writer, r io.Reader, args []string, opts Options) error {
	input, err := getInput(args, r)
	if err != nil {
		return wrapInputErr("cssfmt", err)
	}

	var output string
	if opts.Minify {
		output = pkgcss.Minify(input)
	} else {
		var pkgOpts []pkgcss.Option
		if opts.Indent != "" {
			pkgOpts = append(pkgOpts, pkgcss.WithIndent(opts.Indent))
		}

		if opts.SortProps {
			pkgOpts = append(pkgOpts, pkgcss.WithSortProps())
		}

		if opts.SortRules {
			pkgOpts = append(pkgOpts, pkgcss.WithSortRules())
		}

		output = pkgcss.Format(input, pkgOpts...)
	}

	if _, err := fmt.Fprintln(w, output); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("cssfmt: write: %s", err))
	}

	return nil
}

// RunMinify minifies CSS
func RunMinify(w io.Writer, r io.Reader, args []string, opts Options) error {
	opts.Minify = true
	return Run(w, r, args, opts)
}

// RunValidate validates CSS syntax
func RunValidate(w io.Writer, r io.Reader, args []string, opts ValidateOptions) error {
	input, err := getInput(args, r)
	if err != nil {
		return wrapInputErr("cssfmt", err)
	}

	result := pkgcss.Validate(input)

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		if err := f.Print(result); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("cssfmt: write: %s", err))
		}
		if !result.Valid {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("cssfmt: parse: %s", result.Error))
		}
		return nil
	}

	if result.Valid {
		if _, err := fmt.Fprintln(w, result.Message); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("cssfmt: write: %s", err))
		}
		return nil
	}

	_, _ = fmt.Fprintf(w, "invalid CSS: %s\n", result.Error)
	return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("cssfmt: parse: %s", result.Error))
}

// wrapInputErr classifies input-reading errors into cmderr sentinels.
func wrapInputErr(cmd string, err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("%s: %s", cmd, err))
	}
	if errors.Is(err, os.ErrPermission) {
		return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("%s: %s", cmd, err))
	}
	return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("%s: %s", cmd, err))
}

// getInput reads input from args (file or literal) or stdin
func getInput(args []string, r io.Reader) (string, error) {
	if len(args) > 0 {
		// Check if it's a file
		if _, err := os.Stat(args[0]); err == nil {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return "", err
			}

			return string(content), nil
		}

		// Treat as literal string
		return strings.Join(args, " "), nil
	}

	// Read from stdin
	scanner := bufio.NewScanner(r)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}
