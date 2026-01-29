package rev

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// RevOptions configures the rev command behavior
type RevOptions struct {
	JSON bool // --json: output as JSON
}

// RevResult represents rev output for JSON
type RevResult struct {
	Lines []string `json:"lines"`
	Count int      `json:"count"`
}

// RunRev reverses lines character by character
func RunRev(w io.Writer, args []string, opts RevOptions) error {
	var allLines []string

	if len(args) == 0 {
		if opts.JSON {
			lines, err := revReaderLines(os.Stdin)
			if err != nil {
				return err
			}

			allLines = append(allLines, lines...)
		} else {
			return revReader(w, os.Stdin)
		}
	} else {
		for _, path := range args {
			var r io.Reader
			if path == "-" {
				r = os.Stdin
			} else {
				f, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("rev: %w", err)
				}

				defer func() { _ = f.Close() }()

				r = f
			}

			if opts.JSON {
				lines, err := revReaderLines(r)
				if err != nil {
					return err
				}

				allLines = append(allLines, lines...)
			} else {
				if err := revReader(w, r); err != nil {
					return err
				}
			}
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(RevResult{Lines: allLines, Count: len(allLines)})
	}

	return nil
}

func revReaderLines(r io.Reader) ([]string, error) {
	var lines []string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, reverseString(scanner.Text()))
	}

	return lines, scanner.Err()
}

func revReader(w io.Writer, r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		reversed := reverseString(line)
		_, _ = fmt.Fprintln(w, reversed)
	}

	return scanner.Err()
}

func reverseString(s string) string {
	runes := []rune(s)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}
