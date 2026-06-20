package fold

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/input"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

// FoldOptions configures the fold command behavior
type FoldOptions struct {
	Width        int           // -w: use WIDTH columns instead of 80
	Bytes        bool          // -b: count bytes rather than columns
	Spaces       bool          // -s: break at spaces
	OutputFormat output.Format // output format (text, json, table) — honors global --json
}

// FoldResult represents fold output for JSON mode.
type FoldResult struct {
	Lines []string `json:"lines"`
	Count int      `json:"count"`
}

// RunFold wraps input lines to fit in specified width
// r is the default input reader (used when args is empty or contains "-")
func RunFold(w io.Writer, r io.Reader, args []string, opts FoldOptions) error {
	if opts.Width <= 0 {
		opts.Width = 80
	}

	sources, err := input.Open(args, r)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("fold: %s", err))
		}
		return fmt.Errorf("fold: %w", err)
	}
	defer input.CloseAll(sources)

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	var allLines []string

	for _, src := range sources {
		lines, err := foldReader(w, src.Reader, opts, jsonMode)
		if err != nil {
			return err
		}

		if jsonMode {
			allLines = append(allLines, lines...)
		}
	}

	if jsonMode {
		return f.Print(FoldResult{Lines: allLines, Count: len(allLines)})
	}

	return nil
}

func foldReader(w io.Writer, r io.Reader, opts FoldOptions, jsonMode bool) ([]string, error) {
	scanner := bufio.NewScanner(r)

	var lines []string

	for scanner.Scan() {
		line := scanner.Text()

		folded := foldLine(line, opts)
		for _, part := range folded {
			if jsonMode {
				lines = append(lines, part)
			} else if _, err := fmt.Fprintln(w, part); err != nil {
				return nil, err
			}
		}
	}

	return lines, scanner.Err()
}

func foldLine(line string, opts FoldOptions) []string {
	if len(line) == 0 {
		return []string{""}
	}

	var result []string

	width := opts.Width

	for len(line) > 0 {
		var cutPoint int

		if opts.Bytes {
			if len(line) <= width {
				result = append(result, line)
				break
			}

			cutPoint = width
		} else {
			// Count by runes (columns)
			if utf8.RuneCountInString(line) <= width {
				result = append(result, line)
				break
			}
			// Find byte position for width runes
			cutPoint = 0

			runeCount := 0
			for i := range line {
				if runeCount >= width {
					cutPoint = i
					break
				}

				runeCount++
			}

			if cutPoint == 0 {
				cutPoint = len(line)
			}
		}

		// If -s flag, try to break at last space
		if opts.Spaces && cutPoint < len(line) {
			lastSpace := strings.LastIndex(line[:cutPoint], " ")
			if lastSpace > 0 {
				cutPoint = lastSpace + 1
			}
		}

		result = append(result, line[:cutPoint])
		line = line[cutPoint:]
	}

	return result
}
