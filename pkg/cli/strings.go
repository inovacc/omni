package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// StringsOptions configures the strings command behavior
type StringsOptions struct {
	MinLength   int    // -n: print sequences of at least N characters
	Offset      string // -t: print offset (d=decimal, o=octal, x=hex)
	AllSections bool   // -a: scan whole file (default)
	Encoding    string // -e: encoding (s=7-bit, S=8-bit, etc)
}

// RunStrings prints printable strings in files
func RunStrings(w io.Writer, args []string, opts StringsOptions) error {
	if opts.MinLength == 0 {
		opts.MinLength = 4 // default minimum string length
	}

	if len(args) == 0 {
		return stringsReader(w, os.Stdin, "<stdin>", opts)
	}

	for _, path := range args {
		if path == "-" {
			if err := stringsReader(w, os.Stdin, "<stdin>", opts); err != nil {
				return err
			}

			continue
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("strings: %w", err)
		}

		err = stringsReader(w, f, path, opts)
		_ = f.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func stringsReader(w io.Writer, r io.Reader, _ string, opts StringsOptions) error {
	buf := make([]byte, 4096)

	var current strings.Builder

	offset := int64(0)
	stringStart := int64(0)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			for i := range n {
				b := buf[i]
				if isPrintableASCII(b) {
					if current.Len() == 0 {
						stringStart = offset + int64(i)
					}

					current.WriteByte(b)
				} else {
					if current.Len() >= opts.MinLength {
						printString(w, current.String(), stringStart, opts)
					}

					current.Reset()
				}
			}

			offset += int64(n)
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("strings: %w", err)
		}
	}

	// Print any remaining string
	if current.Len() >= opts.MinLength {
		printString(w, current.String(), stringStart, opts)
	}

	return nil
}

func isPrintableASCII(b byte) bool {
	return (b >= 32 && b < 127) || b == '\t'
}

func printString(w io.Writer, s string, offset int64, opts StringsOptions) {
	if opts.Offset != "" {
		switch opts.Offset {
		case "d":
			_, _ = fmt.Fprintf(w, "%7d ", offset)
		case "o":
			_, _ = fmt.Fprintf(w, "%7o ", offset)
		case "x":
			_, _ = fmt.Fprintf(w, "%7x ", offset)
		}
	}

	_, _ = fmt.Fprintln(w, s)
}
