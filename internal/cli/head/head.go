package head

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// HeadOptions configures the head command behavior
type HeadOptions struct {
	Lines   int  // -n: number of lines to print
	Bytes   int  // -c: number of bytes to print
	Quiet   bool // -q: never print headers
	Verbose bool // -v: always print headers
	JSON    bool // --json: output as JSON
}

// HeadResult represents head output for JSON
type HeadResult struct {
	File  string   `json:"file,omitempty"`
	Lines []string `json:"lines"`
}

// DefaultHeadLines is the default number of lines for head
const DefaultHeadLines = 10

// RunHead executes the head command
func RunHead(w io.Writer, args []string, opts HeadOptions) error {
	if opts.Lines == 0 && opts.Bytes == 0 {
		opts.Lines = DefaultHeadLines
	}

	files := args
	if len(files) == 0 {
		files = []string{"-"} // stdin
	}

	showHeaders := len(files) > 1 || opts.Verbose
	if opts.Quiet {
		showHeaders = false
	}

	var results []HeadResult

	for i, file := range files {
		var r io.Reader

		filename := file

		if file == "-" {
			r = os.Stdin
			filename = "standard input"
		} else {
			f, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("head: cannot open '%s' for reading: %w", file, err)
			}

			r = f

			defer func() {
				_ = f.Close()
			}()
		}

		if opts.JSON {
			lines, err := headLinesJSON(r, opts.Lines)
			if err != nil {
				return err
			}

			results = append(results, HeadResult{File: filename, Lines: lines})

			continue
		}

		if showHeaders {
			if i > 0 {
				_, _ = fmt.Fprintln(w)
			}

			_, _ = fmt.Fprintf(w, "==> %s <==\n", filename)
		}

		if opts.Bytes > 0 {
			if err := headBytes(w, r, opts.Bytes); err != nil {
				return err
			}
		} else {
			if err := headLines(w, r, opts.Lines); err != nil {
				return err
			}
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(results)
	}

	return nil
}

func headLinesJSON(r io.Reader, n int) ([]string, error) {
	scanner := bufio.NewScanner(r)

	var lines []string

	count := 0

	for scanner.Scan() {
		if count >= n {
			break
		}

		lines = append(lines, scanner.Text())
		count++
	}

	return lines, scanner.Err()
}

func headLines(w io.Writer, r io.Reader, n int) error {
	scanner := bufio.NewScanner(r)

	count := 0
	for scanner.Scan() {
		if count >= n {
			break
		}

		_, _ = fmt.Fprintln(w, scanner.Text())
		count++
	}

	return scanner.Err()
}

func headBytes(w io.Writer, r io.Reader, n int) error {
	buf := make([]byte, n)

	read, err := io.ReadFull(r, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return err
	}

	_, err = w.Write(buf[:read])

	return err
}

// Head returns the first n lines from a slice (for compatibility)
func Head(lines []string, n int) []string {
	if n > len(lines) {
		n = len(lines)
	}

	if n < 0 {
		n = 0
	}

	return lines[:n]
}
