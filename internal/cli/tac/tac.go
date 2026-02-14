package tac

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/inovacc/omni/internal/cli/input"
	"github.com/inovacc/omni/internal/cli/output"
)

// TacOptions configures the tac command behavior
type TacOptions struct {
	Before    bool   // -b: attach the separator before instead of after
	Regex     bool   // -r: interpret the separator as a regular expression
	Separator string // -s: use STRING as the separator instead of newline
	OutputFormat output.Format // output format (text/json/table)
}

// TacResult represents tac output for JSON
type TacResult struct {
	Lines []string `json:"lines"`
	Count int      `json:"count"`
}

// RunTac concatenates and prints files in reverse
// r is the default input reader (used when args is empty or contains "-")
func RunTac(w io.Writer, r io.Reader, args []string, opts TacOptions) error {
	sources, err := input.Open(args, r)
	if err != nil {
		return fmt.Errorf("tac: %w", err)
	}
	defer input.CloseAll(sources)

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var allLines []string

	for _, src := range sources {
		if jsonMode {
			lines, err := tacReaderLines(src.Reader, opts)
			if err != nil {
				return err
			}

			allLines = append(allLines, lines...)
		} else {
			if err := tacReader(w, src.Reader, opts); err != nil {
				return err
			}
		}
	}

	if jsonMode {
		return f.Print(TacResult{Lines: allLines, Count: len(allLines)})
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
