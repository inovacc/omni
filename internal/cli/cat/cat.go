package cat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/inovacc/omni/internal/cli/input"
)

// CatOptions configures the cat command behavior
type CatOptions struct {
	NumberAll      bool // -n: number all output lines
	NumberNonBlank bool // -b: number non-blank output lines
	ShowEnds       bool // -E: display $ at end of each line
	ShowTabs       bool // -T: display TAB characters as ^I
	SqueezeBlank   bool // -s: suppress repeated empty output lines
	ShowNonPrint   bool // -v: use ^ and M- notation, except for LFD and TAB
	JSON           bool // --json: output as JSON array of lines
}

// CatLine represents a line for JSON output
type CatLine struct {
	Number  int    `json:"number,omitempty"`
	Content string `json:"content"`
}

// RunCat executes the cat command with the given options
// r is the default input reader (used when args is empty or contains "-")
func RunCat(w io.Writer, r io.Reader, args []string, opts CatOptions) error {
	sources, err := input.Open(args, r)
	if err != nil {
		return fmt.Errorf("cat: %w", err)
	}
	defer input.CloseAll(sources)

	var allLines []CatLine

	for _, src := range sources {
		if opts.JSON {
			lines, err := catReaderJSON(src.Reader, opts)
			if err != nil {
				return err
			}

			allLines = append(allLines, lines...)
		} else {
			if err := catReader(w, src.Reader, src.Name, opts); err != nil {
				return err
			}
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(allLines)
	}

	return nil
}

func catReaderJSON(r io.Reader, opts CatOptions) ([]CatLine, error) {
	scanner := bufio.NewScanner(r)
	lineNum := 0
	prevBlank := false

	var lines []CatLine

	for scanner.Scan() {
		line := scanner.Text()
		isBlank := len(strings.TrimSpace(line)) == 0

		if opts.SqueezeBlank && isBlank && prevBlank {
			continue
		}

		prevBlank = isBlank

		output := line
		if opts.ShowTabs {
			output = strings.ReplaceAll(output, "\t", "^I")
		}

		if opts.ShowNonPrint {
			output = showNonPrintable(output)
		}

		if opts.ShowEnds {
			output += "$"
		}

		catLine := CatLine{Content: output}

		if opts.NumberAll || (opts.NumberNonBlank && !isBlank) {
			lineNum++
			catLine.Number = lineNum
		}

		lines = append(lines, catLine)
	}

	return lines, scanner.Err()
}

func catReader(w io.Writer, r io.Reader, _ string, opts CatOptions) error {
	scanner := bufio.NewScanner(r)
	lineNum := 0
	prevBlank := false

	for scanner.Scan() {
		line := scanner.Text()
		isBlank := len(strings.TrimSpace(line)) == 0

		// Squeeze blank lines
		if opts.SqueezeBlank && isBlank && prevBlank {
			continue
		}

		prevBlank = isBlank

		// Process line content
		output := line

		// Show tabs as ^I
		if opts.ShowTabs {
			output = strings.ReplaceAll(output, "\t", "^I")
		}

		// Show non-printing characters
		if opts.ShowNonPrint {
			output = showNonPrintable(output)
		}

		// Add line number
		if opts.NumberAll || (opts.NumberNonBlank && !isBlank) {
			lineNum++
			output = fmt.Sprintf("%6d\t%s", lineNum, output)
		}

		// Add $ at end of line
		if opts.ShowEnds {
			output += "$"
		}

		_, err := fmt.Fprintln(w, output)
		if err != nil {
			return fmt.Errorf("cat: write error: %w", err)
		}
	}

	return scanner.Err()
}

func showNonPrintable(s string) string {
	var result strings.Builder

	for _, r := range s {
		switch {
		case r == '\t':
			result.WriteRune(r) // Tabs handled separately with -T
		case r < 32:
			// Control characters
			result.WriteString(fmt.Sprintf("^%c", r+64))
		case r == 127:
			result.WriteString("^?")
		case r > 127:
			// High-bit characters (M- notation)
			if r < 160 {
				result.WriteString(fmt.Sprintf("M-^%c", r-128+64))
			} else {
				result.WriteString(fmt.Sprintf("M-%c", r-128))
			}
		default:
			result.WriteRune(r)
		}
	}

	return result.String()
}

// Cat copies from reader to writer (simple version for compatibility)
func Cat(w io.Writer, r io.Reader) error {
	_, err := io.Copy(w, r)
	return err
}

// CatFiles concatenates multiple files to writer (simple version)
func CatFiles(w io.Writer, paths []string) error {
	return RunCat(w, nil, paths, CatOptions{})
}

// ReadFrom reads all lines from a reader
func ReadFrom(r io.Reader) ([]string, error) {
	sc := bufio.NewScanner(r)

	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	return lines, sc.Err()
}

// WriteTo writes lines to a writer
func WriteTo(w io.Writer, lines []string) error {
	for _, l := range lines {
		if _, err := w.Write([]byte(l + "\n")); err != nil {
			return err
		}
	}

	return nil
}
