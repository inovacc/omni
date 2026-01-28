package cut

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// CutOptions configures the cut command behavior
type CutOptions struct {
	Bytes       string // -b: select only these bytes
	Characters  string // -c: select only these characters
	Fields      string // -f: select only these fields
	Delimiter   string // -d: use DELIM instead of TAB for field delimiter
	OnlyDelim   bool   // -s: do not print lines not containing delimiters
	OutputDelim string // --output-delimiter: use STRING as the output delimiter
	Complement  bool   // --complement: complement the set of selected bytes/chars/fields
}

// RunCut executes the cut command
func RunCut(w io.Writer, args []string, opts CutOptions) error {
	// Validate options - must specify one of -b, -c, or -f
	modes := 0
	if opts.Bytes != "" {
		modes++
	}

	if opts.Characters != "" {
		modes++
	}

	if opts.Fields != "" {
		modes++
	}

	if modes == 0 {
		return fmt.Errorf("you must specify a list of bytes, characters, or fields")
	}

	if modes > 1 {
		return fmt.Errorf("only one type of list may be specified")
	}

	// Default delimiter is TAB
	if opts.Delimiter == "" {
		opts.Delimiter = "\t"
	} else if len(opts.Delimiter) > 1 {
		return fmt.Errorf("the delimiter must be a single character")
	}

	// Output delimiter defaults to input delimiter for fields
	if opts.OutputDelim == "" {
		opts.OutputDelim = opts.Delimiter
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"}
	}

	for _, file := range files {
		var r io.Reader
		if file == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(file)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "cut: %s: %v\n", file, err)
				continue
			}

			defer func() {
				_ = f.Close()
			}()

			r = f
		}

		if err := cutReader(w, r, opts); err != nil {
			return err
		}
	}

	return nil
}

func cutReader(w io.Writer, r io.Reader, opts CutOptions) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		var (
			result string
			err    error
		)

		switch {
		case opts.Fields != "":
			result, err = cutFields(line, opts)
		case opts.Characters != "":
			result, err = cutChars(line, opts.Characters, opts.Complement)
		case opts.Bytes != "":
			result, err = cutBytes(line, opts.Bytes, opts.Complement)
		}

		if err != nil {
			return err
		}

		if result != "" || !opts.OnlyDelim {
			_, _ = fmt.Fprintln(w, result)
		}
	}

	return scanner.Err()
}

func cutFields(line string, opts CutOptions) (string, error) {
	// Check if line contains delimiter
	if !strings.Contains(line, opts.Delimiter) {
		if opts.OnlyDelim {
			return "", nil
		}

		return line, nil
	}

	fields := strings.Split(line, opts.Delimiter)

	ranges, err := parseRanges(opts.Fields, len(fields))
	if err != nil {
		return "", err
	}

	if opts.Complement {
		ranges = complementRanges(ranges, len(fields))
	}

	var selected []string

	for _, idx := range ranges {
		if idx > 0 && idx <= len(fields) {
			selected = append(selected, fields[idx-1])
		}
	}

	return strings.Join(selected, opts.OutputDelim), nil
}

func cutChars(line string, spec string, complement bool) (string, error) {
	runes := []rune(line)

	ranges, err := parseRanges(spec, len(runes))
	if err != nil {
		return "", err
	}

	if complement {
		ranges = complementRanges(ranges, len(runes))
	}

	var result []rune

	for _, idx := range ranges {
		if idx > 0 && idx <= len(runes) {
			result = append(result, runes[idx-1])
		}
	}

	return string(result), nil
}

func cutBytes(line string, spec string, complement bool) (string, error) {
	bytes := []byte(line)

	ranges, err := parseRanges(spec, len(bytes))
	if err != nil {
		return "", err
	}

	if complement {
		ranges = complementRanges(ranges, len(bytes))
	}

	var result []byte

	for _, idx := range ranges {
		if idx > 0 && idx <= len(bytes) {
			result = append(result, bytes[idx-1])
		}
	}

	return string(result), nil
}

// parseRanges parses a range specification like "1,3-5,7-"
func parseRanges(spec string, maxVal int) ([]int, error) {
	var result []int

	seen := make(map[int]bool)

	parts := strings.SplitSeq(spec, ",")
	for part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			// Range: N-M, N-, or -M
			rangeParts := strings.SplitN(part, "-", 2)
			start := 1
			end := maxVal

			if rangeParts[0] != "" {
				var err error

				start, err = strconv.Atoi(rangeParts[0])
				if err != nil {
					return nil, fmt.Errorf("invalid range: %s", part)
				}
			}

			if rangeParts[1] != "" {
				var err error

				end, err = strconv.Atoi(rangeParts[1])
				if err != nil {
					return nil, fmt.Errorf("invalid range: %s", part)
				}
			}

			for i := start; i <= end && i <= maxVal; i++ {
				if !seen[i] {
					result = append(result, i)
					seen[i] = true
				}
			}
		} else {
			// Single number
			n, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid field: %s", part)
			}

			if !seen[n] {
				result = append(result, n)
				seen[n] = true
			}
		}
	}

	return result, nil
}

// complementRanges returns all indices NOT in the given ranges
func complementRanges(ranges []int, maxVal int) []int {
	included := make(map[int]bool)
	for _, r := range ranges {
		included[r] = true
	}

	var result []int

	for i := 1; i <= maxVal; i++ {
		if !included[i] {
			result = append(result, i)
		}
	}

	return result
}
