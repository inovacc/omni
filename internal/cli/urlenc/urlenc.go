package urlenc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

// Options configures the url encode/decode command behavior
type Options struct {
	Decode    bool // decode instead of encode
	Component bool // encode/decode as URL component (more aggressive)
	// OutputFormat selects the output format. It is read from the global
	// --json/--table flags via the unified output formatter.
	OutputFormat output.Format
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

	var encoded string
	if opts.Component {
		encoded = url.QueryEscape(input)
	} else {
		encoded = url.PathEscape(input)
	}

	if f := output.New(w, opts.OutputFormat); f.IsJSON() {
		result := Result{
			Input:  input,
			Output: encoded,
			Mode:   "encode",
		}

		return f.Print(result)
	}

	_, _ = fmt.Fprintln(w, encoded)

	return nil
}

// RunDecode decodes URL-encoded text
func RunDecode(w io.Writer, args []string, opts Options) error {
	input, err := getInput(args)
	if err != nil {
		return err
	}

	var decoded string
	if opts.Component {
		decoded, err = url.QueryUnescape(input)
	} else {
		decoded, err = url.PathUnescape(input)
	}

	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("url decode: %s", err))
	}

	if f := output.New(w, opts.OutputFormat); f.IsJSON() {
		result := Result{
			Input:  input,
			Output: decoded,
			Mode:   "decode",
		}

		return f.Print(result)
	}

	_, _ = fmt.Fprintln(w, decoded)

	return nil
}

// getInput reads input from args or stdin
func getInput(args []string) (string, error) {
	if len(args) > 0 {
		// Check if it's a file
		if _, err := os.Stat(args[0]); err == nil {
			content, err := os.ReadFile(args[0])
			if err != nil {
				if errors.Is(err, os.ErrPermission) {
					return "", cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("url: %s", err))
				}
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
		return "", cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("url: %s", err))
	}

	return strings.Join(lines, "\n"), nil
}
