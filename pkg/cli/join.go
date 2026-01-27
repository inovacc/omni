package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// JoinOptions configures the join command behavior
type JoinOptions struct {
	Field1        int    // -1: join on this FIELD of file 1
	Field2        int    // -2: join on this FIELD of file 2
	Separator     string // -t: use CHAR as input and output field separator
	OutputFields  string // -o: output format specification
	IgnoreCase    bool   // -i: ignore case when comparing fields
	CheckOrder    bool   // --check-order: check input is sorted
	NoCheckOrder  bool   // --nocheck-order: do not check input is sorted
	Empty         string // -e: replace missing fields with EMPTY
	Unpaired1     bool   // -a 1: print unpairable lines from file 1
	Unpaired2     bool   // -a 2: print unpairable lines from file 2
	OnlyUnpaired1 bool   // -v 1: print only unpairable lines from file 1
	OnlyUnpaired2 bool   // -v 2: print only unpairable lines from file 2
}

// RunJoin joins lines of two files on a common field
func RunJoin(w io.Writer, args []string, opts JoinOptions) error {
	if len(args) != 2 {
		return fmt.Errorf("join: missing operand")
	}

	// Default to field 1 (1-indexed)
	if opts.Field1 <= 0 {
		opts.Field1 = 1
	}
	if opts.Field2 <= 0 {
		opts.Field2 = 1
	}

	// Default separator is whitespace
	sep := opts.Separator
	if sep == "" {
		sep = " "
	}

	// Read file 1
	lines1, err := readJoinFile(args[0], sep)
	if err != nil {
		return err
	}

	// Read file 2
	lines2, err := readJoinFile(args[1], sep)
	if err != nil {
		return err
	}

	// Build index for file 2
	index2 := make(map[string][]joinLine)
	for _, line := range lines2 {
		key := line.key(opts.Field2 - 1)
		if opts.IgnoreCase {
			key = strings.ToLower(key)
		}
		index2[key] = append(index2[key], line)
	}

	// Track matched lines from file 2
	matched2 := make(map[int]bool)

	// Process file 1
	for _, line1 := range lines1 {
		key := line1.key(opts.Field1 - 1)
		if opts.IgnoreCase {
			key = strings.ToLower(key)
		}

		matches, found := index2[key]
		if found {
			for _, line2 := range matches {
				matched2[line2.index] = true
				if !opts.OnlyUnpaired1 && !opts.OnlyUnpaired2 {
					outputJoinedLine(w, line1, line2, opts, sep)
				}
			}
		} else if opts.Unpaired1 || opts.OnlyUnpaired1 {
			outputUnpairedLine(w, line1, opts, sep)
		}
	}

	// Output unmatched lines from file 2
	if opts.Unpaired2 || opts.OnlyUnpaired2 {
		for _, line2 := range lines2 {
			if !matched2[line2.index] {
				outputUnpairedLine(w, line2, opts, sep)
			}
		}
	}

	return nil
}

type joinLine struct {
	fields []string
	index  int
}

func (j joinLine) key(fieldIdx int) string {
	if fieldIdx < 0 || fieldIdx >= len(j.fields) {
		return ""
	}
	return j.fields[fieldIdx]
}

func readJoinFile(path string, sep string) ([]joinLine, error) {
	var r io.Reader
	if path == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("join: %w", err)
		}
		defer func() {
			_ = f.Close()
		}()
		r = f
	}

	var lines []joinLine
	scanner := bufio.NewScanner(r)
	idx := 0
	for scanner.Scan() {
		text := scanner.Text()
		var fields []string
		if sep == " " {
			fields = strings.Fields(text)
		} else {
			fields = strings.Split(text, sep)
		}
		lines = append(lines, joinLine{fields: fields, index: idx})
		idx++
	}

	return lines, scanner.Err()
}

func outputJoinedLine(w io.Writer, line1, line2 joinLine, opts JoinOptions, sep string) {
	outSep := sep
	if outSep == " " {
		outSep = " "
	}

	// Default: output join field, then remaining fields from both files
	var parts []string

	// Add join field
	joinField := line1.key(opts.Field1 - 1)
	parts = append(parts, joinField)

	// Add remaining fields from file 1
	for i, f := range line1.fields {
		if i != opts.Field1-1 {
			parts = append(parts, f)
		}
	}

	// Add remaining fields from file 2
	for i, f := range line2.fields {
		if i != opts.Field2-1 {
			parts = append(parts, f)
		}
	}

	_, _ = fmt.Fprintln(w, strings.Join(parts, outSep))
}

func outputUnpairedLine(w io.Writer, line joinLine, opts JoinOptions, sep string) {
	outSep := sep
	if outSep == " " {
		outSep = " "
	}
	_, _ = fmt.Fprintln(w, strings.Join(line.fields, outSep))
}
