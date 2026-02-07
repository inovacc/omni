package diff

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	pkgdiff "github.com/inovacc/omni/pkg/textutil/diff"
)

// DiffOptions configures the diff command behavior.
type DiffOptions struct {
	Unified        int  // Number of context lines for unified diff
	Side           bool // Side-by-side output
	Brief          bool // Report only when files differ
	IgnoreCase     bool // Ignore case differences
	IgnoreSpace    bool // Ignore whitespace differences
	IgnoreBlank    bool // Ignore blank lines
	Recursive      bool // Recursively compare directories
	JSON           bool // Compare as JSON
	Color          bool // Colorize output
	Context        int  // Lines of context (old style)
	Width          int  // Output width for side-by-side
	SuppressCommon bool // Suppress common lines in side-by-side
}

// DiffResult represents the result of a diff operation.
type DiffResult struct {
	File1  string
	File2  string
	Differ bool
	Hunks  []DiffHunk
}

// DiffHunk represents a contiguous block of changes.
type DiffHunk struct {
	Start1 int
	Count1 int
	Start2 int
	Count2 int
	Lines  []DiffLine
}

// DiffLine represents a single line in the diff.
type DiffLine struct {
	Type    rune // ' ' for context, '-' for removed, '+' for added
	Content string
}

// RunDiff executes the diff command.
func RunDiff(w io.Writer, args []string, opts DiffOptions) error {
	if len(args) < 2 {
		return fmt.Errorf("diff: missing operand after '%s'", args[0])
	}

	file1, file2 := args[0], args[1]

	// Check if comparing as JSON
	if opts.JSON {
		return diffJSON(w, file1, file2, opts)
	}

	// Check if files are directories
	info1, err := os.Stat(file1)
	if err != nil {
		return fmt.Errorf("diff: %s: %w", file1, err)
	}

	info2, err := os.Stat(file2)
	if err != nil {
		return fmt.Errorf("diff: %s: %w", file2, err)
	}

	if info1.IsDir() && info2.IsDir() {
		if opts.Recursive {
			return diffDirs(w, file1, file2, opts)
		}

		return fmt.Errorf("diff: %s: Is a directory", file1)
	}

	if info1.IsDir() || info2.IsDir() {
		return fmt.Errorf("diff: cannot compare directory to file")
	}

	return diffFiles(w, file1, file2, opts)
}

func diffFiles(w io.Writer, file1, file2 string, opts DiffOptions) error {
	lines1, err := readLines(file1, opts)
	if err != nil {
		return err
	}

	lines2, err := readLines(file2, opts)
	if err != nil {
		return err
	}

	// Compute LCS-based diff using pkg/textutil/diff
	ctx := opts.Unified
	if ctx == 0 {
		ctx = 3
	}

	hunks := pkgdiff.ComputeDiff(lines1, lines2, pkgdiff.WithContext(ctx))

	if len(hunks) == 0 {
		return nil // Files are identical
	}

	if opts.Brief {
		_, _ = fmt.Fprintf(w, "Files %s and %s differ\n", file1, file2)
		return nil
	}

	if opts.Side {
		return printSideBySide(w, lines1, lines2, hunks, opts)
	}

	// Default: unified diff
	return printUnifiedDiff(w, file1, file2, hunks, opts)
}

func readLines(filename string, opts DiffOptions) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("diff: %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if opts.IgnoreCase {
			line = strings.ToLower(line)
		}

		if opts.IgnoreSpace {
			line = strings.Join(strings.Fields(line), " ")
		}

		if opts.IgnoreBlank && strings.TrimSpace(line) == "" {
			continue
		}

		lines = append(lines, line)
	}

	return lines, scanner.Err()
}

func printUnifiedDiff(w io.Writer, file1, file2 string, hunks []pkgdiff.Hunk, opts DiffOptions) error {
	_, _ = fmt.Fprintf(w, "--- %s\n", file1)
	_, _ = fmt.Fprintf(w, "+++ %s\n", file2)

	for _, hunk := range hunks {
		_, _ = fmt.Fprintf(w, "@@ -%d,%d +%d,%d @@\n", hunk.Start1, hunk.Count1, hunk.Start2, hunk.Count2)

		for _, line := range hunk.Lines {
			prefix := string(line.Type)
			if opts.Color {
				switch line.Type {
				case '-':
					_, _ = fmt.Fprintf(w, "\033[31m%s%s\033[0m\n", prefix, line.Content)
				case '+':
					_, _ = fmt.Fprintf(w, "\033[32m%s%s\033[0m\n", prefix, line.Content)
				default:
					_, _ = fmt.Fprintf(w, "%s%s\n", prefix, line.Content)
				}
			} else {
				_, _ = fmt.Fprintf(w, "%s%s\n", prefix, line.Content)
			}
		}
	}

	return nil
}

func printSideBySide(w io.Writer, lines1, lines2 []string, hunks []pkgdiff.Hunk, opts DiffOptions) error {
	width := opts.Width
	if width == 0 {
		width = 130
	}

	colWidth := (width - 3) / 2

	pos1, pos2 := 0, 0
	for _, hunk := range hunks {
		for pos1 < hunk.Start1-1 && pos2 < hunk.Start2-1 {
			if !opts.SuppressCommon {
				left := pkgdiff.TruncateOrPad(lines1[pos1], colWidth)
				right := pkgdiff.TruncateOrPad(lines2[pos2], colWidth)
				_, _ = fmt.Fprintf(w, "%s   %s\n", left, right)
			}

			pos1++
			pos2++
		}

		for _, line := range hunk.Lines {
			switch line.Type {
			case ' ':
				if !opts.SuppressCommon {
					left := pkgdiff.TruncateOrPad(line.Content, colWidth)
					right := pkgdiff.TruncateOrPad(line.Content, colWidth)
					_, _ = fmt.Fprintf(w, "%s   %s\n", left, right)
				}

				pos1++
				pos2++
			case '-':
				left := pkgdiff.TruncateOrPad(line.Content, colWidth)

				right := strings.Repeat(" ", colWidth)
				if opts.Color {
					_, _ = fmt.Fprintf(w, "\033[31m%s\033[0m < %s\n", left, right)
				} else {
					_, _ = fmt.Fprintf(w, "%s < %s\n", left, right)
				}

				pos1++
			case '+':
				left := strings.Repeat(" ", colWidth)

				right := pkgdiff.TruncateOrPad(line.Content, colWidth)
				if opts.Color {
					_, _ = fmt.Fprintf(w, "%s > \033[32m%s\033[0m\n", left, right)
				} else {
					_, _ = fmt.Fprintf(w, "%s > %s\n", left, right)
				}

				pos2++
			}
		}
	}

	for pos1 < len(lines1) && pos2 < len(lines2) {
		if !opts.SuppressCommon {
			left := pkgdiff.TruncateOrPad(lines1[pos1], colWidth)
			right := pkgdiff.TruncateOrPad(lines2[pos2], colWidth)
			_, _ = fmt.Fprintf(w, "%s   %s\n", left, right)
		}

		pos1++
		pos2++
	}

	return nil
}

func diffDirs(w io.Writer, dir1, dir2 string, opts DiffOptions) error {
	entries1, err := os.ReadDir(dir1)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}

	entries2, err := os.ReadDir(dir2)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}

	map1 := make(map[string]os.DirEntry)
	map2 := make(map[string]os.DirEntry)

	for _, e := range entries1 {
		map1[e.Name()] = e
	}

	for _, e := range entries2 {
		map2[e.Name()] = e
	}

	allNames := make(map[string]bool)
	for name := range map1 {
		allNames[name] = true
	}

	for name := range map2 {
		allNames[name] = true
	}

	for name := range allNames {
		path1 := dir1 + "/" + name
		path2 := dir2 + "/" + name

		e1, in1 := map1[name]
		e2, in2 := map2[name]

		if !in1 {
			_, _ = fmt.Fprintf(w, "Only in %s: %s\n", dir2, name)
			continue
		}

		if !in2 {
			_, _ = fmt.Fprintf(w, "Only in %s: %s\n", dir1, name)
			continue
		}

		if e1.IsDir() && e2.IsDir() {
			if err := diffDirs(w, path1, path2, opts); err != nil {
				return err
			}
		} else if !e1.IsDir() && !e2.IsDir() {
			if err := diffFiles(w, path1, path2, opts); err != nil {
				return err
			}
		} else {
			_, _ = fmt.Fprintf(w, "File %s is a %s while file %s is a %s\n",
				path1, fileType(e1), path2, fileType(e2))
		}
	}

	return nil
}

func fileType(e os.DirEntry) string {
	if e.IsDir() {
		return "directory"
	}

	return "regular file"
}

// diffJSON compares two JSON files.
func diffJSON(w io.Writer, file1, file2 string, opts DiffOptions) error {
	data1, err := os.ReadFile(file1)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}

	data2, err := os.ReadFile(file2)
	if err != nil {
		return fmt.Errorf("diff: %w", err)
	}

	var json1, json2 any
	if err := json.Unmarshal(data1, &json1); err != nil {
		return fmt.Errorf("diff: %s: invalid JSON: %w", file1, err)
	}

	if err := json.Unmarshal(data2, &json2); err != nil {
		return fmt.Errorf("diff: %s: invalid JSON: %w", file2, err)
	}

	differences := pkgdiff.CompareJSON(json1, json2)
	if len(differences) == 0 {
		return nil
	}

	if opts.Brief {
		_, _ = fmt.Fprintf(w, "Files %s and %s differ\n", file1, file2)
		return nil
	}

	for _, d := range differences {
		_, _ = fmt.Fprintln(w, d)
	}

	return nil
}
