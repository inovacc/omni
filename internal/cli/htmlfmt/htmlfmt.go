package htmlfmt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
	pkghtml "github.com/inovacc/omni/pkg/htmlfmt"
)

// Options configures the HTML formatter
type Options struct {
	Indent    string // Indentation (default: "  ")
	Minify    bool   // Minify output
	SortAttrs bool   // Sort attributes alphabetically
}

// ValidateOptions configures HTML validation
type ValidateOptions struct {
	OutputFormat output.Format // Output format
}

// ValidateResult is an alias for the pkg type
type ValidateResult = pkghtml.ValidateResult

// Run formats HTML input
func Run(w io.Writer, r io.Reader, args []string, opts Options) error {
	input, err := getInput(args, r)
	if err != nil {
		return wrapInputErr("htmlfmt", err)
	}

	var output string
	if opts.Minify {
		output, err = pkghtml.Minify(input)
	} else {
		var pkgOpts []pkghtml.Option
		if opts.Indent != "" {
			pkgOpts = append(pkgOpts, pkghtml.WithIndent(opts.Indent))
		}

		if opts.SortAttrs {
			pkgOpts = append(pkgOpts, pkghtml.WithSortAttrs())
		}

		output, err = pkghtml.Format(input, pkgOpts...)
	}

	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("htmlfmt: parse: %s", err))
	}

	if _, err := fmt.Fprintln(w, output); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("htmlfmt: write: %s", err))
	}

	return nil
}

// RunMinify minifies HTML
func RunMinify(w io.Writer, r io.Reader, args []string, opts Options) error {
	opts.Minify = true
	return Run(w, r, args, opts)
}

// RunValidate validates HTML syntax
func RunValidate(w io.Writer, r io.Reader, args []string, opts ValidateOptions) error {
	input, err := getInput(args, r)
	if err != nil {
		return wrapInputErr("htmlfmt", err)
	}

	result := pkghtml.Validate(input)

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		if err := f.Print(result); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("htmlfmt: write: %s", err))
		}
		if !result.Valid {
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("htmlfmt: parse: %s", result.Error))
		}
		return nil
	}

	if result.Valid {
		if _, err := fmt.Fprintln(w, result.Message); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("htmlfmt: write: %s", err))
		}
		return nil
	}

	_, _ = fmt.Fprintf(w, "invalid HTML: %s\n", result.Error)
	return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("htmlfmt: parse: %s", result.Error))
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
