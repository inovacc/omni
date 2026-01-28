package cat

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// CatOptions configures the cat command behavior
type CatOptions struct {
	NumberAll      bool // -n: number all output lines
	NumberNonBlank bool // -b: number non-blank output lines
	ShowEnds       bool // -E: display $ at end of each line
	ShowTabs       bool // -T: display TAB characters as ^I
	SqueezeBlank   bool // -s: suppress repeated empty output lines
	ShowNonPrint   bool // -v: use ^ and M- notation, except for LFD and TAB
}

// RunCat executes the cat command with the given options
func RunCat(w io.Writer, args []string, opts CatOptions) error {
	if len(args) == 0 {
		return catReader(w, os.Stdin, "-", opts)
	}

	for _, path := range args {
		var r io.Reader
		if path == "-" {
			r = os.Stdin
		} else {
			f, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("cat: %s: %w", path, err)
			}

			r = f

			defer func() {
				_ = f.Close()
			}()
		}

		if err := catReader(w, r, path, opts); err != nil {
			return err
		}
	}

	return nil
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
	return RunCat(w, paths, CatOptions{})
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
