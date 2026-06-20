package paste

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/input"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

// PasteOptions configures the paste command behavior
type PasteOptions struct {
	Delimiters   string        // -d: use characters from LIST instead of TABs
	Serial       bool          // -s: paste one file at a time instead of in parallel
	Zero         bool          // -z: line delimiter is NUL, not newline
	OutputFormat output.Format // output format (text, json, table) — honors global --json
}

// PasteResult represents paste output for JSON mode. Each row holds the
// per-file fields that were merged for that output line.
type PasteResult struct {
	Rows  [][]string `json:"rows"`
	Count int        `json:"count"`
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

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var (
		rows [][]string
		err  error
	)

	if opts.Serial {
		rows, err = pasteSerial(w, r, args, delimiters, lineTerminator, jsonMode)
	} else {
		rows, err = pasteParallel(w, r, args, delimiters, lineTerminator, jsonMode)
	}

	if err != nil {
		return err
	}

	if jsonMode {
		return f.Print(PasteResult{Rows: rows, Count: len(rows)})
	}

	return nil
}

func pasteParallel(w io.Writer, defaultReader io.Reader, files []string, delimiters, lineTerminator string, jsonMode bool) ([][]string, error) {
	// Open all files using input package
	sources, err := input.Open(files, defaultReader)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("paste: %s", err))
		}
		return nil, fmt.Errorf("paste: %w", err)
	}
	defer input.CloseAll(sources)

	readers := make([]*bufio.Scanner, len(sources))
	for i, src := range sources {
		readers[i] = bufio.NewScanner(src.Reader)
	}

	delimIdx := 0

	var rows [][]string

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

		if jsonMode {
			row := make([]string, len(parts))
			copy(row, parts)
			rows = append(rows, row)

			continue
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

	return rows, nil
}

func pasteSerial(w io.Writer, defaultReader io.Reader, files []string, delimiters, lineTerminator string, jsonMode bool) ([][]string, error) {
	sources, err := input.Open(files, defaultReader)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("paste: %s", err))
		}
		return nil, fmt.Errorf("paste: %w", err)
	}
	defer input.CloseAll(sources)

	var rows [][]string

	for _, src := range sources {
		scanner := bufio.NewScanner(src.Reader)

		var parts []string
		for scanner.Scan() {
			parts = append(parts, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		if jsonMode {
			row := make([]string, len(parts))
			copy(row, parts)
			rows = append(rows, row)

			continue
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

	return rows, nil
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
