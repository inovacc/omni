package urlenc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

// Options configures the url encode/decode command behavior
type Options struct {
	Decode    bool // decode instead of encode
	Component bool // encode/decode as URL component (more aggressive)
	JSON      bool // --json: output as JSON
}

// Result represents the output for JSON mode
type Result struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Mode   string `json:"mode"` // "encode" or "decode"
}

// RunEncode encodes text as URL-safe
func RunEncode(w io.Writer, args []string, opts Options) error {
	input, err := getInput(args)
	if err != nil {
		return err
	}

	var output string
	if opts.Component {
		output = url.QueryEscape(input)
	} else {
		output = url.PathEscape(input)
	}

	if opts.JSON {
		result := Result{
			Input:  input,
			Output: output,
			Mode:   "encode",
		}

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(result)
	}

	_, _ = fmt.Fprintln(w, output)

	return nil
}

// RunDecode decodes URL-encoded text
func RunDecode(w io.Writer, args []string, opts Options) error {
	input, err := getInput(args)
	if err != nil {
		return err
	}

	var output string
	if opts.Component {
		output, err = url.QueryUnescape(input)
	} else {
		output, err = url.PathUnescape(input)
	}

	if err != nil {
		return fmt.Errorf("url decode: %w", err)
	}

	if opts.JSON {
		result := Result{
			Input:  input,
			Output: output,
			Mode:   "decode",
		}

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(result)
	}

	_, _ = fmt.Fprintln(w, output)

	return nil
}

// getInput reads input from args or stdin
func getInput(args []string) (string, error) {
	if len(args) > 0 {
		// Check if it's a file
		if _, err := os.Stat(args[0]); err == nil {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return "", fmt.Errorf("url: %w", err)
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
		return "", fmt.Errorf("url: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}
