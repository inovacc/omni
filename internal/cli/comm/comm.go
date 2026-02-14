package comm

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/inovacc/omni/internal/cli/output"
)

// CommOptions configures the comm command behavior
type CommOptions struct {
	Suppress1    bool   // -1: suppress column 1 (lines unique to FILE1)
	Suppress2    bool   // -2: suppress column 2 (lines unique to FILE2)
	Suppress3    bool   // -3: suppress column 3 (lines common to both)
	CheckOrder   bool   // --check-order: check input is correctly sorted
	NoCheckOrder bool   // --nocheck-order: do not check input order
	OutputDelim  string // --output-delimiter: use STR as output delimiter
	ZeroTerm     bool   // -z: line delimiter is NUL
	OutputFormat output.Format // output format (text/json/table)
}

// CommResult represents the JSON output for comm
type CommResult struct {
	UniqueToFile1 []string `json:"uniqueToFile1"`
	UniqueToFile2 []string `json:"uniqueToFile2"`
	Common        []string `json:"common"`
}

// RunComm compares two sorted files line by line
func RunComm(w io.Writer, args []string, opts CommOptions) error {
	if len(args) < 2 {
		return fmt.Errorf("comm: missing operand")
	}

	file1, file2 := args[0], args[1]

	// Open files
	var r1, r2 io.Reader

	if file1 == "-" {
		r1 = os.Stdin
	} else {
		f, err := os.Open(file1)
		if err != nil {
			return fmt.Errorf("comm: %w", err)
		}

		defer func() { _ = f.Close() }()

		r1 = f
	}

	if file2 == "-" {
		if file1 == "-" {
			return fmt.Errorf("comm: both files cannot be stdin")
		}

		r2 = os.Stdin
	} else {
		f, err := os.Open(file2)
		if err != nil {
			return fmt.Errorf("comm: %w", err)
		}

		defer func() { _ = f.Close() }()

		r2 = f
	}

	// Set default delimiter
	delim := "\t"
	if opts.OutputDelim != "" {
		delim = opts.OutputDelim
	}

	lineDelim := byte('\n')
	if opts.ZeroTerm {
		lineDelim = 0
	}

	scanner1 := bufio.NewScanner(r1)
	scanner2 := bufio.NewScanner(r2)

	if opts.ZeroTerm {
		scanner1.Split(splitFunc(lineDelim))
		scanner2.Split(splitFunc(lineDelim))
	}

	var line1, line2 string

	has1 := scanner1.Scan()
	has2 := scanner2.Scan()

	if has1 {
		line1 = scanner1.Text()
	}

	if has2 {
		line2 = scanner2.Text()
	}

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var (
		prevLine1, prevLine2 string
		result               CommResult
	)

	for has1 || has2 {
		// Check sort order if requested
		if opts.CheckOrder && !opts.NoCheckOrder {
			if has1 && prevLine1 != "" && line1 < prevLine1 {
				return fmt.Errorf("comm: file 1 is not in sorted order")
			}

			if has2 && prevLine2 != "" && line2 < prevLine2 {
				return fmt.Errorf("comm: file 2 is not in sorted order")
			}
		}

		if !has1 {
			// Only file2 has lines
			if jsonMode {
				result.UniqueToFile2 = append(result.UniqueToFile2, line2)
			} else {
				printCommLine(w, opts, delim, 2, line2)
			}

			prevLine2 = line2
			has2 = scanner2.Scan()

			if has2 {
				line2 = scanner2.Text()
			}
		} else if !has2 {
			// Only file1 has lines
			if jsonMode {
				result.UniqueToFile1 = append(result.UniqueToFile1, line1)
			} else {
				printCommLine(w, opts, delim, 1, line1)
			}

			prevLine1 = line1
			has1 = scanner1.Scan()

			if has1 {
				line1 = scanner1.Text()
			}
		} else if line1 < line2 {
			// Line unique to file1
			if jsonMode {
				result.UniqueToFile1 = append(result.UniqueToFile1, line1)
			} else {
				printCommLine(w, opts, delim, 1, line1)
			}

			prevLine1 = line1
			has1 = scanner1.Scan()

			if has1 {
				line1 = scanner1.Text()
			}
		} else if line1 > line2 {
			// Line unique to file2
			if jsonMode {
				result.UniqueToFile2 = append(result.UniqueToFile2, line2)
			} else {
				printCommLine(w, opts, delim, 2, line2)
			}

			prevLine2 = line2
			has2 = scanner2.Scan()

			if has2 {
				line2 = scanner2.Text()
			}
		} else {
			// Lines are equal
			if jsonMode {
				result.Common = append(result.Common, line1)
			} else {
				printCommLine(w, opts, delim, 3, line1)
			}

			prevLine1 = line1
			prevLine2 = line2
			has1 = scanner1.Scan()
			has2 = scanner2.Scan()

			if has1 {
				line1 = scanner1.Text()
			}

			if has2 {
				line2 = scanner2.Text()
			}
		}
	}

	if err := scanner1.Err(); err != nil {
		return fmt.Errorf("comm: %w", err)
	}

	if err := scanner2.Err(); err != nil {
		return fmt.Errorf("comm: %w", err)
	}

	if jsonMode {
		return f.Print(result)
	}

	return nil
}

// splitFunc returns a split function for the given delimiter
func splitFunc(delim byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		for i, b := range data {
			if b == delim {
				return i + 1, data[:i], nil
			}
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	}
}

func printCommLine(w io.Writer, opts CommOptions, delim string, column int, line string) {
	switch column {
	case 1:
		if !opts.Suppress1 {
			_, _ = fmt.Fprintln(w, line)
		}
	case 2:
		if !opts.Suppress2 {
			prefix := ""
			if !opts.Suppress1 {
				prefix = delim
			}

			_, _ = fmt.Fprintln(w, prefix+line)
		}
	case 3:
		if !opts.Suppress3 {
			prefix := ""
			if !opts.Suppress1 {
				prefix += delim
			}

			if !opts.Suppress2 {
				prefix += delim
			}

			_, _ = fmt.Fprintln(w, prefix+line)
		}
	}
}
