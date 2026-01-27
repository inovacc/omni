package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
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
}

// RunNl numbers lines of files
func RunNl(w io.Writer, args []string, opts NlOptions) error {
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

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}

	lineNum := opts.StartingNumber

	for _, file := range files {
		var r io.Reader
		if file == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(file)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "nl: %s: %v\n", file, err)
				continue
			}

			defer func() {
				_ = f.Close()
			}()

			r = f
		}

		scanner := bufio.NewScanner(r)
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

			if shouldNumber {
				numStr := formatLineNumber(lineNum, opts.NumberFormat, opts.NumberWidth)
				_, _ = fmt.Fprintf(w, "%s%s%s\n", numStr, opts.NumberSep, line)
				lineNum += opts.Increment
			} else {
				_, _ = fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", opts.NumberWidth), line)
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}
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
