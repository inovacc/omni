package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// ColumnOptions configures the column command behavior
type ColumnOptions struct {
	Table         bool   // -t: determine column count based on input
	Separator     string // -s: column delimiter characters for -t option
	OutputSep     string // -o: output separator for table mode
	Columns       int    // -c: output is formatted for display width of N (default 80)
	FillRows      bool   // -x: fill rows before columns
	NoMerge       bool   // -n: do not merge multiple adjacent delimiters
	Right         bool   // -R: right-align columns
	JSON          bool   // -J: output as JSON
	TableName     string // -N: table name for JSON output
	ColumnHeaders string // -H: specify column headers
}

// RunColumn columnates lists
func RunColumn(w io.Writer, args []string, opts ColumnOptions) error {
	if opts.Columns == 0 {
		opts.Columns = 80
	}

	if opts.Separator == "" {
		opts.Separator = " \t"
	}

	if opts.OutputSep == "" {
		opts.OutputSep = "  "
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}

	// Read all lines
	var lines []string
	for _, file := range files {
		var r io.Reader
		if file == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(file)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "column: %s: %v\n", file, err)
				continue
			}
			defer func() {
				_ = f.Close()
			}()
			r = f
		}

		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
	}

	if opts.Table {
		return columnTable(w, lines, opts)
	}

	return columnFill(w, lines, opts)
}

func columnTable(w io.Writer, lines []string, opts ColumnOptions) error {
	if len(lines) == 0 {
		return nil
	}

	// Split lines into fields
	var rows [][]string
	var maxCols int

	for _, line := range lines {
		var fields []string
		if opts.NoMerge {
			fields = strings.Split(line, opts.Separator[:1])
		} else {
			fields = strings.FieldsFunc(line, func(r rune) bool {
				return strings.ContainsRune(opts.Separator, r)
			})
		}
		rows = append(rows, fields)
		if len(fields) > maxCols {
			maxCols = len(fields)
		}
	}

	// Calculate column widths
	colWidths := make([]int, maxCols)
	for _, row := range rows {
		for i, field := range row {
			if len(field) > colWidths[i] {
				colWidths[i] = len(field)
			}
		}
	}

	// Handle custom headers
	if opts.ColumnHeaders != "" {
		headers := strings.Split(opts.ColumnHeaders, ",")
		for i, header := range headers {
			if i < maxCols && len(header) > colWidths[i] {
				colWidths[i] = len(header)
			}
		}
		// Print headers
		for i, header := range headers {
			if i > 0 {
				_, _ = fmt.Fprint(w, opts.OutputSep)
			}
			if i < maxCols {
				if opts.Right {
					_, _ = fmt.Fprintf(w, "%*s", colWidths[i], header)
				} else {
					_, _ = fmt.Fprintf(w, "%-*s", colWidths[i], header)
				}
			}
		}
		_, _ = fmt.Fprintln(w)
	}

	// Print rows
	for _, row := range rows {
		for i := 0; i < maxCols; i++ {
			if i > 0 {
				_, _ = fmt.Fprint(w, opts.OutputSep)
			}
			field := ""
			if i < len(row) {
				field = row[i]
			}
			if opts.Right {
				_, _ = fmt.Fprintf(w, "%*s", colWidths[i], field)
			} else {
				// Don't pad the last column
				if i == maxCols-1 {
					_, _ = fmt.Fprint(w, field)
				} else {
					_, _ = fmt.Fprintf(w, "%-*s", colWidths[i], field)
				}
			}
		}
		_, _ = fmt.Fprintln(w)
	}

	return nil
}

func columnFill(w io.Writer, lines []string, opts ColumnOptions) error {
	if len(lines) == 0 {
		return nil
	}

	// Find maximum width
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	// Add padding
	colWidth := maxWidth + 2
	numCols := opts.Columns / colWidth
	if numCols < 1 {
		numCols = 1
	}

	numRows := (len(lines) + numCols - 1) / numCols

	if opts.FillRows {
		// Fill rows first
		for i := 0; i < len(lines); i++ {
			if i > 0 && i%numCols == 0 {
				_, _ = fmt.Fprintln(w)
			}
			_, _ = fmt.Fprintf(w, "%-*s", colWidth, lines[i])
		}
		_, _ = fmt.Fprintln(w)
	} else {
		// Fill columns first (default)
		for row := 0; row < numRows; row++ {
			for col := 0; col < numCols; col++ {
				idx := col*numRows + row
				if idx < len(lines) {
					if col == numCols-1 || idx+numRows >= len(lines) {
						_, _ = fmt.Fprint(w, lines[idx])
					} else {
						_, _ = fmt.Fprintf(w, "%-*s", colWidth, lines[idx])
					}
				}
			}
			_, _ = fmt.Fprintln(w)
		}
	}

	return nil
}
