package sqlfmt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
	pkgsql "github.com/inovacc/omni/pkg/sqlfmt"
)

// Options configures the SQL formatter
type Options struct {
	Indent    string // Indentation (default: "  ")
	Uppercase bool   // Uppercase keywords (default: true)
	Minify    bool   // Minify output
	Dialect   string // SQL dialect: mysql, postgres, sqlite, generic (default: generic)
}

// ValidateOptions configures SQL validation
type ValidateOptions struct {
	OutputFormat output.Format // Output format
	Dialect      string        // SQL dialect
}

// ValidateResult represents validation output
type ValidateResult struct {
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Run formats SQL input
func Run(w io.Writer, r io.Reader, args []string, opts Options) error {
	input, err := getInput(args, r)
	if err != nil {
		return fmt.Errorf("sql: %w", err)
	}

	var output string
	if opts.Minify {
		output = pkgsql.Minify(input)
	} else {
		var pkgOpts []pkgsql.Option
		if opts.Indent != "" {
			pkgOpts = append(pkgOpts, pkgsql.WithIndent(opts.Indent))
		}

		if opts.Uppercase {
			pkgOpts = append(pkgOpts, pkgsql.WithUppercase())
		}

		output = pkgsql.Format(input, pkgOpts...)
	}

	_, _ = fmt.Fprintln(w, output)

	return nil
}

// RunMinify minifies SQL
func RunMinify(w io.Writer, r io.Reader, args []string, opts Options) error {
	opts.Minify = true

	return Run(w, r, args, opts)
}

// RunValidate validates SQL syntax
func RunValidate(w io.Writer, r io.Reader, args []string, opts ValidateOptions) error {
	input, err := getInput(args, r)
	if err != nil {
		return fmt.Errorf("sql: %w", err)
	}

	pkgResult := pkgsql.Validate(input)
	result := ValidateResult{
		Valid:   pkgResult.Valid,
		Error:   pkgResult.Error,
		Message: pkgResult.Message,
	}

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(result)
	}

	if result.Valid {
		_, _ = fmt.Fprintln(w, result.Message)
	} else {
		_, _ = fmt.Fprintf(w, "invalid SQL: %s\n", result.Error)

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
