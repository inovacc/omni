package paste

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/inovacc/omni/internal/cli/input"
)

// PasteOptions configures the paste command behavior
type PasteOptions struct {
	Delimiters string // -d: use characters from LIST instead of TABs
	Serial     bool   // -s: paste one file at a time instead of in parallel
	Zero       bool   // -z: line delimiter is NUL, not newline
}

// RunPaste merges lines of files
// r is the default input reader (used when args is empty or contains "-")
func RunPaste(w io.Writer, r io.Reader, args []string, opts PasteOptions) error {
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
		return pasteSerial(w, r, args, delimiters, lineTerminator)
	}

	return pasteParallel(w, r, args, delimiters, lineTerminator)
}

func pasteParallel(w io.Writer, defaultReader io.Reader, files []string, delimiters, lineTerminator string) error {
	// Open all files using input package
	sources, err := input.Open(files, defaultReader)
	if err != nil {
		return fmt.Errorf("paste: %w", err)
	}
	defer input.CloseAll(sources)

	readers := make([]*bufio.Scanner, len(sources))
	for i, src := range sources {
		readers[i] = bufio.NewScanner(src.Reader)
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

func pasteSerial(w io.Writer, defaultReader io.Reader, files []string, delimiters, lineTerminator string) error {
	sources, err := input.Open(files, defaultReader)
	if err != nil {
		return fmt.Errorf("paste: %w", err)
	}
	defer input.CloseAll(sources)

	for _, src := range sources {
		scanner := bufio.NewScanner(src.Reader)

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
