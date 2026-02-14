package htmlfmt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
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
		return fmt.Errorf("html: %w", err)
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
		return fmt.Errorf("html: %w", err)
	}

	_, _ = fmt.Fprintln(w, output)

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
		return fmt.Errorf("html: %w", err)
	}

	result := pkghtml.Validate(input)

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(result)
	}

	if result.Valid {
		_, _ = fmt.Fprintln(w, result.Message)
	} else {
		_, _ = fmt.Fprintf(w, "invalid HTML: %s\n", result.Error)
		return fmt.Errorf("validation failed")
	}

	return nil
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
