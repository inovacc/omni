package xmlfmt

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// Options configures the xml format command behavior
type Options struct {
	Minify bool   // --minify: remove whitespace
	Indent string // --indent: indentation string (default "  ")
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
