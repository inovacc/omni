package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// PasteOptions configures the paste command behavior
type PasteOptions struct {
	Delimiters string // -d: use characters from LIST instead of TABs
	Serial     bool   // -s: paste one file at a time instead of in parallel
	Zero       bool   // -z: line delimiter is NUL, not newline
}

// RunPaste merges lines of files
func RunPaste(w io.Writer, args []string, opts PasteOptions) error {
	if len(args) == 0 {
		args = []string{"-"}
	}

	// Default delimiter is TAB
	delimiters := opts.Delimiters
	if delimiters == "" {
		delimiters = "\t"
	}

	// Handle escape sequences in delimiters
	delimiters = expandDelimiters(delimiters)

	lineTerminator := "\n"
	if opts.Zero {
		lineTerminator = "\x00"
	}

	if opts.Serial {
		return pasteSerial(w, args, delimiters, lineTerminator)
	}

	return pasteParallel(w, args, delimiters, lineTerminator)
}

func pasteParallel(w io.Writer, files []string, delimiters, lineTerminator string) error {
	// Open all files
	readers := make([]*bufio.Scanner, len(files))
	for i, file := range files {
		var r io.Reader
		if file == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("paste: %s: %w", file, err)
			}

			defer func() {
				_ = f.Close()
			}()

			r = f
		}

		readers[i] = bufio.NewScanner(r)
	}

	delimIdx := 0

	for {
		var parts []string

		anyMore := false

		for _, scanner := range readers {
			if scanner.Scan() {
				parts = append(parts, scanner.Text())
				anyMore = true
			} else {
				parts = append(parts, "")
			}
		}

		if !anyMore {
			break
		}

		// Join with delimiters (cycling through them)
		var result strings.Builder

		for i, part := range parts {
			if i > 0 {
				delim := string(delimiters[delimIdx%len(delimiters)])
				result.WriteString(delim)

				delimIdx++
			}

			result.WriteString(part)
		}

		delimIdx = 0 // Reset delimiter index for next line

		_, _ = fmt.Fprint(w, result.String()+lineTerminator)
	}

	return nil
}

func pasteSerial(w io.Writer, files []string, delimiters, lineTerminator string) error {
	for _, file := range files {
		var r io.Reader
		if file == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("paste: %s: %w", file, err)
			}

			defer func() {
				_ = f.Close()
			}()

			r = f
		}

		scanner := bufio.NewScanner(r)

		var parts []string
		for scanner.Scan() {
			parts = append(parts, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		// Join with delimiters
		var result strings.Builder

		delimIdx := 0

		for i, part := range parts {
			if i > 0 {
				delim := string(delimiters[delimIdx%len(delimiters)])
				result.WriteString(delim)

				delimIdx++
			}

			result.WriteString(part)
		}

		_, _ = fmt.Fprint(w, result.String()+lineTerminator)
	}

	return nil
}

func expandDelimiters(s string) string {
	var result strings.Builder

	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result.WriteRune('\n')
			case 't':
				result.WriteRune('\t')
			case '\\':
				result.WriteRune('\\')
			case '0':
				result.WriteRune('\x00')
			default:
				result.WriteByte(s[i+1])
			}

			i += 2
		} else {
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String()
}
