package seq

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
)

// SeqOptions configures the seq command behavior
type SeqOptions struct {
	Separator    string        // -s: use STRING to separate numbers
	Format       string        // -f: use printf style FORMAT
	EqualWidth   bool          // -w: equalize width by padding with leading zeros
	OutputFormat output.Format // output format
}

// SeqResult represents seq output for JSON
type SeqResult struct {
	Numbers []float64 `json:"numbers"`
	Count   int       `json:"count"`
}

// RunSeq prints a sequence of numbers
func RunSeq(w io.Writer, args []string, opts SeqOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("seq: missing operand")
	}

	// Set defaults
	if opts.Separator == "" {
		opts.Separator = "\n"
	}

	// Parse arguments: seq [FIRST [INCREMENT]] LAST
	var (
		first, increment, last float64
		err                    error
	)

	switch len(args) {
	case 1:
		first = 1
		increment = 1
		last, err = strconv.ParseFloat(args[0], 64)
	case 2:
		first, err = strconv.ParseFloat(args[0], 64)
		if err == nil {
			last, err = strconv.ParseFloat(args[1], 64)
		}

		increment = 1

		if first > last {
			increment = -1
		}
	case 3:
		first, err = strconv.ParseFloat(args[0], 64)
		if err == nil {
			increment, err = strconv.ParseFloat(args[1], 64)
		}

		if err == nil {
			last, err = strconv.ParseFloat(args[2], 64)
		}
	default:
		return fmt.Errorf("seq: too many arguments")
	}

	if err != nil {
		return fmt.Errorf("seq: invalid argument")
	}

	if increment == 0 {
		return fmt.Errorf("seq: increment must not be zero")
	}

	// Determine format
	format := opts.Format
	if format == "" {
		format = determineSeqFormat(first, increment, last, opts.EqualWidth)
	}

	// Generate sequence
	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		var numbers []float64
		for i := first; (increment > 0 && i <= last) || (increment < 0 && i >= last); i += increment {
			numbers = append(numbers, i)
		}

		return f.Print(SeqResult{Numbers: numbers, Count: len(numbers)})
	}

	isFirst := true

	for i := first; (increment > 0 && i <= last) || (increment < 0 && i >= last); i += increment {
		if !isFirst {
			_, _ = fmt.Fprint(w, opts.Separator)
		}

		_, _ = fmt.Fprintf(w, format, i)
		isFirst = false
	}

	if !isFirst {
		_, _ = fmt.Fprintln(w)
	}

	return nil
}

func determineSeqFormat(first, increment, last float64, equalWidth bool) string {
	// Check if any value has decimals
	hasDecimals := hasDecimalPart(first) || hasDecimalPart(increment) || hasDecimalPart(last)

	if hasDecimals {
		// Determine precision needed
		prec := maxPrecision(first, increment, last)
		if equalWidth {
			width := maxWidth(first, last, prec)
			return fmt.Sprintf("%%0%d.%df", width, prec)
		}

		return fmt.Sprintf("%%.%df", prec)
	}

	if equalWidth {
		width := maxIntWidth(first, last)
		return fmt.Sprintf("%%0%d.0f", width)
	}

	return "%.0f"
}

func hasDecimalPart(f float64) bool {
	return f != math.Trunc(f)
}

func maxPrecision(nums ...float64) int {
	maxPrec := 0

	for _, n := range nums {
		s := strconv.FormatFloat(n, 'f', -1, 64)
		if idx := strings.Index(s, "."); idx >= 0 {
			prec := len(s) - idx - 1
			if prec > maxPrec {
				maxPrec = prec
			}
		}
	}

	return maxPrec
}

func maxWidth(first, last float64, precision int) int {
	w1 := len(fmt.Sprintf("%.*f", precision, first))
	w2 := len(fmt.Sprintf("%.*f", precision, last))

	if w1 > w2 {
		return w1
	}

	return w2
}

func maxIntWidth(first, last float64) int {
	w1 := len(fmt.Sprintf("%.0f", first))
	w2 := len(fmt.Sprintf("%.0f", last))

	if w1 > w2 {
		return w1
	}

	return w2
}
