package hexenc

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	clioutput "github.com/inovacc/omni/internal/cli/output"
)

// Options configures the hex encode/decode command behavior
type Options struct {
	OutputFormat clioutput.Format // output format (text, json, table)
	Uppercase    bool          // --upper: use uppercase hex
}

// Result represents the output for JSON mode
type Result struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Mode   string `json:"mode"` // "encode" or "decode"
}

// RunEncode encodes text as hexadecimal
func RunEncode(w io.Writer, args []string, opts Options) error {
	input, err := getInput(args)
	if err != nil {
		return err
	}

	output := hex.EncodeToString([]byte(input))
	if opts.Uppercase {
		output = strings.ToUpper(output)
	}

	f := clioutput.New(w, opts.OutputFormat)

	if f.IsJSON() {
		return f.Print(Result{
			Input:  input,
			Output: output,
			Mode:   "encode",
		})
	}

	_, _ = fmt.Fprintln(w, output)

	return nil
}

// RunDecode decodes hexadecimal to text
func RunDecode(w io.Writer, args []string, opts Options) error {
	input, err := getInput(args)
	if err != nil {
		return err
	}

	// Remove any whitespace or common separators
	cleaned := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == ':' || r == '-' {
			return -1
		}

		return r
	}, input)

	decoded, err := hex.DecodeString(cleaned)
	if err != nil {
		return fmt.Errorf("hex decode: %w", err)
	}

	output := string(decoded)

	f := clioutput.New(w, opts.OutputFormat)

	if f.IsJSON() {
		return f.Print(Result{
			Input:  input,
			Output: output,
			Mode:   "decode",
		})
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
				return "", fmt.Errorf("hex: %w", err)
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
		return "", fmt.Errorf("hex: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}
