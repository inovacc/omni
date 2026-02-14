package xmlfmt

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
)

// Options configures the xml format command behavior
type Options struct {
	Minify bool   // --minify: remove whitespace
	Indent string // --indent: indentation string (default "  ")
}

// ValidateOptions configures the xml validate command behavior
type ValidateOptions struct {
	OutputFormat output.Format // Output format
}

// ValidateResult represents the output for JSON mode
type ValidateResult struct {
	File    string `json:"file,omitempty"`
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message,omitempty"`
}

// Run formats XML input
func Run(w io.Writer, args []string, opts Options) error {
	input, err := getInput(args)
	if err != nil {
		return err
	}

	if opts.Indent == "" {
		opts.Indent = "  "
	}

	var output string
	if opts.Minify {
		output, err = minifyXML(input)
	} else {
		output, err = formatXML(input, opts.Indent)
	}

	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(w, output)

	return nil
}

func formatXML(input string, indent string) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(input))

	var buf bytes.Buffer

	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", indent)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", fmt.Errorf("xml: %w", err)
		}

		if err := encoder.EncodeToken(token); err != nil {
			return "", fmt.Errorf("xml: %w", err)
		}
	}

	if err := encoder.Flush(); err != nil {
		return "", fmt.Errorf("xml: %w", err)
	}

	return buf.String(), nil
}

func minifyXML(input string) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(input))

	var buf bytes.Buffer

	encoder := xml.NewEncoder(&buf)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", fmt.Errorf("xml: %w", err)
		}

		// Skip whitespace-only character data
		if charData, ok := token.(xml.CharData); ok {
			trimmed := strings.TrimSpace(string(charData))
			if trimmed == "" {
				continue
			}

			token = xml.CharData(trimmed)
		}

		if err := encoder.EncodeToken(token); err != nil {
			return "", fmt.Errorf("xml: %w", err)
		}
	}

	if err := encoder.Flush(); err != nil {
		return "", fmt.Errorf("xml: %w", err)
	}

	return buf.String(), nil
}

// getInput reads input from args or stdin
func getInput(args []string) (string, error) {
	if len(args) > 0 {
		// Check if it's a file
		if _, err := os.Stat(args[0]); err == nil {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return "", fmt.Errorf("xml: %w", err)
			}

			return string(content), nil
		}

		// Treat as literal string
		return strings.Join(args, " "), nil
	}

	// Read from stdin
	scanner := bufio.NewScanner(os.Stdin)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("xml: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}

// RunValidate validates XML input
func RunValidate(w io.Writer, args []string, opts ValidateOptions) error {
	if len(args) == 0 {
		// Read from stdin
		return validateReader(w, os.Stdin, "<stdin>", opts)
	}

	var hasError bool

	for _, arg := range args {
		// Check if it's a file
		if info, err := os.Stat(arg); err == nil && !info.IsDir() {
			f, err := os.Open(arg)
			if err != nil {
				return fmt.Errorf("xml validate: %w", err)
			}

			err = validateReader(w, f, arg, opts)
			_ = f.Close()

			if err != nil {
				hasError = true
			}
		} else {
			// Treat as literal XML string
			err := validateReader(w, strings.NewReader(arg), "<input>", opts)
			if err != nil {
				hasError = true
			}
		}
	}

	if hasError {
		return fmt.Errorf("xml validation failed")
	}

	return nil
}

func validateReader(w io.Writer, r io.Reader, name string, opts ValidateOptions) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return outputValidateResult(w, ValidateResult{
			File:  name,
			Valid: false,
			Error: err.Error(),
		}, opts)
	}

	decoder := xml.NewDecoder(bytes.NewReader(content))

	for {
		_, err := decoder.Token()
		if err == io.EOF {
			break
		}

		if err != nil {
			result := ValidateResult{
				File:  name,
				Valid: false,
				Error: err.Error(),
			}

			// Try to extract position from syntax error
			if syntaxErr, ok := err.(*xml.SyntaxError); ok {
				result.Line = syntaxErr.Line
				result.Message = syntaxErr.Msg
			}

			return outputValidateResult(w, result, opts)
		}
	}

	result := ValidateResult{
		File:    name,
		Valid:   true,
		Message: "valid XML",
	}

	return outputValidateResult(w, result, opts)
}

func outputValidateResult(w io.Writer, result ValidateResult, opts ValidateOptions) error {
	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(result)
	}

	if result.Valid {
		_, _ = fmt.Fprintf(w, "%s: %s\n", result.File, result.Message)
	} else {
		if result.Line > 0 {
			_, _ = fmt.Fprintf(w, "%s:%d: invalid XML - %s\n", result.File, result.Line, result.Error)
		} else {
			_, _ = fmt.Fprintf(w, "%s: invalid XML - %s\n", result.File, result.Error)
		}
	}

	if !result.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}
