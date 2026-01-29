package tomlutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// ValidateOptions configures the toml validate command behavior
type ValidateOptions struct {
	JSON bool // --json: output as JSON
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

// RunValidate validates TOML input
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
				return fmt.Errorf("toml validate: %w", err)
			}

			err = validateReader(w, f, arg, opts)
			_ = f.Close()

			if err != nil {
				hasError = true
			}
		} else {
			// Treat as literal TOML string
			err := validateReader(w, strings.NewReader(arg), "<input>", opts)
			if err != nil {
				hasError = true
			}
		}
	}

	if hasError {
		return fmt.Errorf("toml validation failed")
	}

	return nil
}

func validateReader(w io.Writer, r io.Reader, name string, opts ValidateOptions) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return outputResult(w, ValidateResult{
			File:  name,
			Valid: false,
			Error: err.Error(),
		}, opts)
	}

	var data any

	_, err = toml.Decode(string(content), &data)
	if err != nil {
		result := ValidateResult{
			File:  name,
			Valid: false,
			Error: err.Error(),
		}

		// Try to extract line info from parse error
		if parseErr, ok := err.(toml.ParseError); ok {
			result.Line = parseErr.Position.Line
			result.Column = parseErr.Position.Col
			result.Message = parseErr.Message
		}

		return outputResult(w, result, opts)
	}

	result := ValidateResult{
		File:    name,
		Valid:   true,
		Message: "valid TOML",
	}

	return outputResult(w, result, opts)
}

func outputResult(w io.Writer, result ValidateResult, opts ValidateOptions) error {
	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(result)
	}

	if result.Valid {
		_, _ = fmt.Fprintf(w, "%s: %s\n", result.File, result.Message)
	} else {
		if result.Line > 0 {
			_, _ = fmt.Fprintf(w, "%s:%d:%d: invalid TOML - %s\n", result.File, result.Line, result.Column, result.Error)
		} else {
			_, _ = fmt.Fprintf(w, "%s: invalid TOML - %s\n", result.File, result.Error)
		}
	}

	if !result.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// FormatOptions configures the toml format command behavior
type FormatOptions struct {
	Indent int // indentation width (for arrays)
}

// RunFormat formats TOML input
func RunFormat(w io.Writer, args []string, opts FormatOptions) error {
	input, err := getInput(args)
	if err != nil {
		return err
	}

	var data any

	if _, err := toml.Decode(input, &data); err != nil {
		return fmt.Errorf("toml format: %w", err)
	}

	enc := toml.NewEncoder(w)

	if opts.Indent > 0 {
		enc.Indent = strings.Repeat(" ", opts.Indent)
	}

	return enc.Encode(data)
}

// getInput reads input from args or stdin
func getInput(args []string) (string, error) {
	if len(args) > 0 {
		// Check if it's a file
		if _, err := os.Stat(args[0]); err == nil {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return "", fmt.Errorf("toml: %w", err)
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
		return "", fmt.Errorf("toml: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}
