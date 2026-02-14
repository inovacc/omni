package cssfmt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
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
		return fmt.Errorf("css: %w", err)
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

	_, _ = fmt.Fprintln(w, output)

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
		return fmt.Errorf("css: %w", err)
	}

	result := pkgcss.Validate(input)

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(result)
	}

	if result.Valid {
		_, _ = fmt.Fprintln(w, result.Message)
	} else {
		_, _ = fmt.Fprintf(w, "invalid CSS: %s\n", result.Error)
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
