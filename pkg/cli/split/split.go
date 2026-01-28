package split

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// SplitOptions configures the split command behavior
type SplitOptions struct {
	Lines       int    // -l: put NUMBER lines per output file
	Bytes       string // -b: put SIZE bytes per output file
	Number      int    // -n: generate N output files
	Suffix      int    // -a: generate suffixes of length N (default 2)
	NumericSufx bool   // -d: use numeric suffixes
	Verbose     bool   // --verbose: print diagnostic
}

// RunSplit splits a file into pieces
func RunSplit(w io.Writer, args []string, opts SplitOptions) error {
	// Defaults
	if opts.Suffix == 0 {
		opts.Suffix = 2
	}

	var input io.Reader

	prefix := "x"

	if len(args) == 0 || args[0] == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("split: %w", err)
		}

		defer func() { _ = f.Close() }()

		input = f
	}

	if len(args) > 1 {
		prefix = args[1]
	}

	// Determine split mode
	if opts.Bytes != "" {
		return splitByBytes(w, input, prefix, opts)
	}

	if opts.Lines == 0 {
		opts.Lines = 1000 // default
	}

	return splitByLines(w, input, prefix, opts)
}

func splitByLines(w io.Writer, input io.Reader, prefix string, opts SplitOptions) error {
	scanner := bufio.NewScanner(input)
	fileNum := 0
	lineCount := 0

	var outFile *os.File

	var outWriter *bufio.Writer

	for scanner.Scan() {
		if lineCount%opts.Lines == 0 {
			// Close previous file
			if outFile != nil {
				_ = outWriter.Flush()
				_ = outFile.Close()
			}

			// Open new file
			suffix := generateSuffix(fileNum, opts.Suffix, opts.NumericSufx)
			filename := prefix + suffix

			var err error

			outFile, err = os.Create(filename)
			if err != nil {
				return fmt.Errorf("split: %w", err)
			}

			outWriter = bufio.NewWriter(outFile)

			if opts.Verbose {
				_, _ = fmt.Fprintf(w, "creating file %q\n", filename)
			}

			fileNum++
		}

		_, _ = outWriter.WriteString(scanner.Text())
		_, _ = outWriter.WriteString("\n")
		lineCount++
	}

	if outFile != nil {
		_ = outWriter.Flush()
		_ = outFile.Close()
	}

	return scanner.Err()
}

func splitByBytes(w io.Writer, input io.Reader, prefix string, opts SplitOptions) error {
	size, err := parseByteSize(opts.Bytes)
	if err != nil {
		return fmt.Errorf("split: invalid byte size %q", opts.Bytes)
	}

	buf := make([]byte, size)
	fileNum := 0

	for {
		n, err := io.ReadFull(input, buf)
		if n > 0 {
			suffix := generateSuffix(fileNum, opts.Suffix, opts.NumericSufx)
			filename := prefix + suffix

			outFile, ferr := os.Create(filename)
			if ferr != nil {
				return fmt.Errorf("split: %w", ferr)
			}

			_, _ = outFile.Write(buf[:n])
			_ = outFile.Close()

			if opts.Verbose {
				_, _ = fmt.Fprintf(w, "creating file %q\n", filename)
			}

			fileNum++
		}

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return fmt.Errorf("split: %w", err)
		}
	}

	return nil
}

func generateSuffix(num int, length int, numeric bool) string {
	if numeric {
		format := fmt.Sprintf("%%0%dd", length)
		return fmt.Sprintf(format, num)
	}

	// Alphabetic suffix (aa, ab, ac, ...)
	result := make([]byte, length)

	for i := length - 1; i >= 0; i-- {
		result[i] = 'a' + byte(num%26)
		num /= 26
	}

	return string(result)
}

func parseByteSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}

	multiplier := int64(1)
	suffix := s[len(s)-1]

	switch suffix {
	case 'K', 'k':
		multiplier = 1024
		s = s[:len(s)-1]
	case 'M', 'm':
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	case 'G', 'g':
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return val * multiplier, nil
}
