package yamlutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidateOptions configures the yaml validate command behavior
type ValidateOptions struct {
	JSON   bool // --json: output as JSON
	Strict bool // --strict: fail on unknown fields
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

// RunValidate validates YAML input
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
				return fmt.Errorf("yaml validate: %w", err)
			}

			err = validateReader(w, f, arg, opts)
			_ = f.Close()

			if err != nil {
				hasError = true
			}
		} else {
			// Treat as literal YAML string
			err := validateReader(w, strings.NewReader(arg), "<input>", opts)
			if err != nil {
				hasError = true
			}
		}
	}

	if hasError {
		return fmt.Errorf("yaml validation failed")
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

	decoder := yaml.NewDecoder(strings.NewReader(string(content)))
	if opts.Strict {
		decoder.KnownFields(true)
	}

	// Try to decode all documents in the YAML
	docCount := 0

	for {
		err = decoder.Decode(&data)
		if err == io.EOF {
			break
		}

		if err != nil {
			result := ValidateResult{
				File:  name,
				Valid: false,
				Error: err.Error(),
			}

			// Try to extract line/column info from yaml.TypeError
			if typeErr, ok := err.(*yaml.TypeError); ok {
				result.Message = strings.Join(typeErr.Errors, "; ")
			}

			return outputResult(w, result, opts)
		}

		docCount++
	}

	result := ValidateResult{
		File:    name,
		Valid:   true,
		Message: fmt.Sprintf("valid YAML (%d document(s))", docCount),
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
		_, _ = fmt.Fprintf(w, "%s: invalid YAML - %s\n", result.File, result.Error)
	}

	if !result.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// FormatOptions configures the yaml format command behavior
type FormatOptions struct {
	Indent int  // indentation width
	JSON   bool // output as JSON instead
}

// RunFormat formats YAML input
func RunFormat(w io.Writer, args []string, opts FormatOptions) error {
	input, err := getInput(args)
	if err != nil {
		return err
	}

	var data any

	if err := yaml.Unmarshal([]byte(input), &data); err != nil {
		return fmt.Errorf("yaml format: %w", err)
	}

	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", strings.Repeat(" ", opts.Indent))

		return enc.Encode(data)
	}

	enc := yaml.NewEncoder(w)
	enc.SetIndent(opts.Indent)

	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("yaml format: %w", err)
	}

	return enc.Close()
}

// getInput reads input from args or stdin
func getInput(args []string) (string, error) {
	if len(args) > 0 {
		// Check if it's a file
		if _, err := os.Stat(args[0]); err == nil {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return "", fmt.Errorf("yaml: %w", err)
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
		return "", fmt.Errorf("yaml: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}
