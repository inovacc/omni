package tac

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// TacOptions configures the tac command behavior
type TacOptions struct {
	Before    bool   // -b: attach the separator before instead of after
	Regex     bool   // -r: interpret the separator as a regular expression
	Separator string // -s: use STRING as the separator instead of newline
	JSON      bool   // --json: output as JSON
}

// TacResult represents tac output for JSON
type TacResult struct {
	Lines []string `json:"lines"`
	Count int      `json:"count"`
}

// RunTac concatenates and prints files in reverse
func RunTac(w io.Writer, args []string, opts TacOptions) error {
	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}

	var allLines []string

	for _, file := range files {
		var r io.Reader
		if file == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(file)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "tac: %s: %v\n", file, err)
				continue
			}

			defer func() {
				_ = f.Close()
			}()

			r = f
		}

		if opts.JSON {
			lines, err := tacReaderLines(r, opts)
			if err != nil {
				return err
			}

			allLines = append(allLines, lines...)
		} else {
			if err := tacReader(w, r, opts); err != nil {
				return err
			}
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(TacResult{Lines: allLines, Count: len(allLines)})
	}

	return nil
}

func readLines(r io.Reader, opts TacOptions) ([]string, error) {
	var lines []string

	scanner := bufio.NewScanner(r)

	if opts.Separator != "" && opts.Separator != "\n" {
		// Custom separator - read entire content
		var content strings.Builder
		for scanner.Scan() {
			if content.Len() > 0 {
				content.WriteString("\n")
			}

			content.WriteString(scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		lines = strings.Split(content.String(), opts.Separator)
	} else {
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return lines, nil
}

func tacReaderLines(r io.Reader, opts TacOptions) ([]string, error) {
	lines, err := readLines(r, opts)
	if err != nil {
		return nil, err
	}
	// Reverse the lines
	reversed := make([]string, len(lines))
	for i, line := range lines {
		reversed[len(lines)-1-i] = line
	}

	return reversed, nil
}

func tacReader(w io.Writer, r io.Reader, opts TacOptions) error {
	lines, err := readLines(r, opts)
	if err != nil {
		return err
	}

	// Print in reverse order
	sep := "\n"
	if opts.Separator != "" {
		sep = opts.Separator
	}

	for i := len(lines) - 1; i >= 0; i-- {
		if opts.Before && i < len(lines)-1 {
			_, _ = fmt.Fprint(w, sep)
		}

		_, _ = fmt.Fprint(w, lines[i])
		if !opts.Before && i > 0 {
			_, _ = fmt.Fprint(w, sep)
		}
	}

	// Add final newline if using default separator
	if opts.Separator == "" || opts.Separator == "\n" {
		_, _ = fmt.Fprintln(w)
	}

	return nil
}

// Reverse reverses a slice of strings
func Reverse(lines []string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[len(lines)-1-i] = line
	}

	return result
}
