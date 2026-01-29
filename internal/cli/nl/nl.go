package nl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/inovacc/omni/internal/cli/input"
)

// NlOptions configures the nl command behavior
type NlOptions struct {
	BodyNumbering   string // -b: body line numbering style (a=all, t=nonempty, n=none)
	HeaderNumbering string // -h: header line numbering style
	FooterNumbering string // -f: footer line numbering style
	NumberFormat    string // -n: line number format (ln=left, rn=right, rz=right-zeros)
	NumberWidth     int    // -w: line number width
	NumberSep       string // -s: separator between number and text
	StartingNumber  int    // -v: starting line number
	Increment       int    // -i: line number increment
	NoRenumber      bool   // -p: do not reset line numbers at sections
	JSON            bool   // --json: output as JSON
}

// NlLine represents a numbered line for JSON output
type NlLine struct {
	Number int    `json:"number,omitempty"`
	Text   string `json:"text"`
}

// NlResult represents nl output for JSON
type NlResult struct {
	Lines []NlLine `json:"lines"`
	Count int      `json:"count"`
}

// RunNl numbers lines of files
// r is the default input reader (used when args is empty or contains "-")
func RunNl(w io.Writer, r io.Reader, args []string, opts NlOptions) error {
	// Set defaults
	if opts.BodyNumbering == "" {
		opts.BodyNumbering = "t" // number non-empty lines by default
	}

	if opts.HeaderNumbering == "" {
		opts.HeaderNumbering = "n"
	}

	if opts.FooterNumbering == "" {
		opts.FooterNumbering = "n"
	}

	if opts.NumberFormat == "" {
		opts.NumberFormat = "rn" // right justified
	}

	if opts.NumberWidth == 0 {
		opts.NumberWidth = 6
	}

	if opts.NumberSep == "" {
		opts.NumberSep = "\t"
	}

	if opts.StartingNumber == 0 {
		opts.StartingNumber = 1
	}

	if opts.Increment == 0 {
		opts.Increment = 1
	}

	sources, err := input.Open(args, r)
	if err != nil {
		return fmt.Errorf("nl: %w", err)
	}
	defer input.CloseAll(sources)

	lineNum := opts.StartingNumber

	var jsonLines []NlLine

	for _, src := range sources {
		scanner := bufio.NewScanner(src.Reader)
		for scanner.Scan() {
			line := scanner.Text()

			// Determine if line should be numbered
			shouldNumber := false

			switch opts.BodyNumbering {
			case "a":
				shouldNumber = true
			case "t":
				shouldNumber = strings.TrimSpace(line) != ""
			case "n":
				shouldNumber = false
			}

			if opts.JSON {
				if shouldNumber {
					jsonLines = append(jsonLines, NlLine{Number: lineNum, Text: line})
					lineNum += opts.Increment
				} else {
					jsonLines = append(jsonLines, NlLine{Number: 0, Text: line})
				}
			} else {
				if shouldNumber {
					numStr := formatLineNumber(lineNum, opts.NumberFormat, opts.NumberWidth)
					_, _ = fmt.Fprintf(w, "%s%s%s\n", numStr, opts.NumberSep, line)
					lineNum += opts.Increment
				} else {
					_, _ = fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", opts.NumberWidth), line)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(NlResult{Lines: jsonLines, Count: len(jsonLines)})
	}

	return nil
}

func formatLineNumber(num int, format string, width int) string {
	switch format {
	case "ln": // left justified, no leading zeros
		return fmt.Sprintf("%-*d", width, num)
	case "rn": // right justified, no leading zeros
		return fmt.Sprintf("%*d", width, num)
	case "rz": // right justified, leading zeros
		return fmt.Sprintf("%0*d", width, num)
	default:
		return fmt.Sprintf("%*d", width, num)
	}
}
