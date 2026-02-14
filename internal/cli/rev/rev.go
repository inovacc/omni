package rev

import (
	"bufio"
	"fmt"
	"io"

	"github.com/inovacc/omni/internal/cli/input"
	"github.com/inovacc/omni/internal/cli/output"
)

// RevOptions configures the rev command behavior
type RevOptions struct {
	OutputFormat output.Format // output format (text/json/table)
}

// RevResult represents rev output for JSON
type RevResult struct {
	Lines []string `json:"lines"`
	Count int      `json:"count"`
}

// RunRev reverses lines character by character
// r is the default input reader (used when args is empty or contains "-")
func RunRev(w io.Writer, r io.Reader, args []string, opts RevOptions) error {
	sources, err := input.Open(args, r)
	if err != nil {
		return fmt.Errorf("rev: %w", err)
	}
	defer input.CloseAll(sources)

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var allLines []string

	for _, src := range sources {
		if jsonMode {
			lines, err := revReaderLines(src.Reader)
			if err != nil {
				return err
			}

			allLines = append(allLines, lines...)
		} else {
			if err := revReader(w, src.Reader); err != nil {
				return err
			}
		}
	}

	if jsonMode {
		return f.Print(RevResult{Lines: allLines, Count: len(allLines)})
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
