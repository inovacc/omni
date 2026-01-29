package wc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

// WCOptions configures the wc command behavior
type WCOptions struct {
	Lines      bool // -l: print the newline counts
	Words      bool // -w: print the word counts
	Bytes      bool // -c: print the byte counts
	Chars      bool // -m: print the character counts
	MaxLineLen bool // -L: print the maximum display width
	JSON       bool // --json: output in JSON format
}

// WCResult represents the result of a wc operation
type WCResult struct {
	Lines      int    `json:"lines"`
	Words      int    `json:"words"`
	Bytes      int    `json:"bytes"`
	Chars      int    `json:"chars"`
	MaxLineLen int    `json:"maxLineLen"`
	Filename   string `json:"filename,omitempty"`
}

// RunWC executes the wc command
func RunWC(w io.Writer, args []string, opts WCOptions) error {
	// If no flags specified, default to -lwc
	if !opts.Lines && !opts.Words && !opts.Bytes && !opts.Chars && !opts.MaxLineLen {
		opts.Lines = true
		opts.Words = true
		opts.Bytes = true
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"} // stdin
	}

	var totals WCResult
	var results []WCResult

	totals.Filename = "total"

	for _, file := range files {
		var r io.Reader

		filename := file

		if file == "-" {
			r = os.Stdin
			filename = ""
		} else {
			f, err := os.Open(file)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "wc: %s: %v\n", file, err)
				continue
			}

			r = f

			defer func() {
				_ = f.Close()
			}()
		}

		result, err := countReader(r, opts)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "wc: %s: %v\n", file, err)
			continue
		}

		result.Filename = filename

		if opts.JSON {
			results = append(results, result)
		} else {
			printWCResult(w, result, opts)
		}

		// Accumulate totals
		totals.Lines += result.Lines
		totals.Words += result.Words
		totals.Bytes += result.Bytes

		totals.Chars += result.Chars
		if result.MaxLineLen > totals.MaxLineLen {
			totals.MaxLineLen = result.MaxLineLen
		}
	}

	// JSON output
	if opts.JSON {
		if len(files) > 1 {
			results = append(results, totals)
		}

		return json.NewEncoder(w).Encode(results)
	}

	// Print totals if multiple files
	if len(files) > 1 {
		printWCResult(w, totals, opts)
	}

	return nil
}

func countReader(r io.Reader, opts WCOptions) (WCResult, error) {
	var result WCResult

	reader := bufio.NewReader(r)
	inWord := false
	lineLen := 0

	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}

			return result, err
		}

		result.Bytes++

		// Count characters (UTF-8 aware)
		if opts.Chars {
			if (b & 0xC0) != 0x80 { // Not a continuation byte
				result.Chars++
			}
		}

		// Count lines
		if b == '\n' {
			result.Lines++
			if lineLen > result.MaxLineLen {
				result.MaxLineLen = lineLen
			}

			lineLen = 0
		} else {
			lineLen++
		}

		// Count words
		isSpace := b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f' || b == '\v'
		if isSpace {
			inWord = false
		} else if !inWord {
			inWord = true
			result.Words++
		}
	}

	// Handle last line without newline
	if lineLen > result.MaxLineLen {
		result.MaxLineLen = lineLen
	}

	return result, nil
}

func printWCResult(w io.Writer, result WCResult, opts WCOptions) {
	var (
		fields []any
		format string
	)

	if opts.Lines {
		format += "%7d "

		fields = append(fields, result.Lines)
	}

	if opts.Words {
		format += "%7d "

		fields = append(fields, result.Words)
	}

	if opts.Chars {
		format += "%7d "

		fields = append(fields, result.Chars)
	}

	if opts.Bytes {
		format += "%7d "

		fields = append(fields, result.Bytes)
	}

	if opts.MaxLineLen {
		format += "%7d "

		fields = append(fields, result.MaxLineLen)
	}

	if result.Filename != "" {
		format += "%s"

		fields = append(fields, result.Filename)
	}

	format += "\n"
	_, _ = fmt.Fprintf(w, format, fields...)
}

// WC counts lines, words, and bytes in data (for compatibility)
func WC(data []byte) WCResult {
	lines := 0
	words := 0
	chars := utf8.RuneCount(data)
	inWord := false

	for _, b := range data {
		if b == '\n' {
			lines++
		}

		if b == ' ' || b == '\n' || b == '\t' {
			inWord = false
		} else if !inWord {
			inWord = true
			words++
		}
	}

	return WCResult{
		Lines: lines,
		Words: words,
		Bytes: len(data),
		Chars: chars,
	}
}

// WCWithStats is an alias for WC (for compatibility)
func WCWithStats(data []byte) (WCResult, error) {
	return WC(data), nil
}
