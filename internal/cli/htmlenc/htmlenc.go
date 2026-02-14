package htmlenc

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
)

// Options configures the html encode/decode command behavior
type Options struct {
	OutputFormat output.Format // Output format
}

// Result represents the output for JSON mode
type Result struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Mode   string `json:"mode"` // "encode" or "decode"
}

// RunEncode encodes text as HTML entities
func RunEncode(w io.Writer, args []string, opts Options) error {
	input, err := getInput(args)
	if err != nil {
		return err
	}

	encOutput := html.EscapeString(input)

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(Result{
			Input:  input,
			Output: encOutput,
			Mode:   "encode",
		})
	}

	_, _ = fmt.Fprintln(w, encOutput)

	return nil
}

// RunDecode decodes HTML entities
func RunDecode(w io.Writer, args []string, opts Options) error {
	input, err := getInput(args)
	if err != nil {
		return err
	}

	decOutput := html.UnescapeString(input)

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(Result{
			Input:  input,
			Output: decOutput,
			Mode:   "decode",
		})
	}

	_, _ = fmt.Fprintln(w, decOutput)

	return nil
}

// getInput reads input from args or stdin
func getInput(args []string) (string, error) {
	if len(args) > 0 {
		// Check if it's a file
		if _, err := os.Stat(args[0]); err == nil {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return "", fmt.Errorf("html: %w", err)
			}

			return strings.TrimSpace(string(content)), nil
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
		return "", fmt.Errorf("html: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}
