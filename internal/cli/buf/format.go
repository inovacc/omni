package buf

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/pkg/buf/pkg/bufapi"
)

// RunFormat formats proto files
func RunFormat(w io.Writer, dir string, opts FormatOptions) error {
	// Find proto files
	files, err := FindProtoFiles(dir, nil)
	if err != nil {
		return fmt.Errorf("buf: %w", err)
	}

	if len(files) == 0 {
		_, _ = fmt.Fprintln(w, "No proto files found")

		return nil
	}

	var hasUnformatted bool

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("buf: failed to read %s: %w", file, err)
		}

		formatted := FormatProto(string(content))

		if string(content) == formatted {
			continue
		}

		hasUnformatted = true

		if opts.Diff {
			_, _ = fmt.Fprintf(w, "--- %s\n+++ %s\n", file, file)

			printDiff(w, string(content), formatted)
		}

		if opts.Write {
			if err := os.WriteFile(file, []byte(formatted), 0644); err != nil {
				return fmt.Errorf("buf: failed to write %s: %w", file, err)
			}

			if !opts.Diff {
				_, _ = fmt.Fprintf(w, "Formatted %s\n", file)
			}
		}
	}

	if opts.ExitCode && hasUnformatted {
		return fmt.Errorf("found unformatted files")
	}

	return nil
}

// FormatProto formats a proto file content using the real buf formatter.
// Falls back to a simple cleanup if the protocompile parser fails.
func FormatProto(source string) string {
	formatted, err := bufapi.FormatProto("input.proto", source)
	if err != nil {
		// Fallback: return source with cleaned-up blank lines
		return cleanupBlankLines(source)
	}
	return formatted
}

func cleanupBlankLines(s string) string {
	lines := strings.Split(s, "\n")

	var result []string

	lastWasBlank := false

	for _, line := range lines {
		isBlank := strings.TrimSpace(line) == ""

		if isBlank && lastWasBlank {
			continue
		}

		result = append(result, line)
		lastWasBlank = isBlank
	}

	return strings.Join(result, "\n")
}

func printDiff(w io.Writer, original, formatted string) {
	origLines := strings.Split(original, "\n")
	fmtLines := strings.Split(formatted, "\n")

	// Simple line-by-line diff
	maxLen := max(len(fmtLines), len(origLines))

	for i := range maxLen {
		origLine := ""
		fmtLine := ""

		if i < len(origLines) {
			origLine = origLines[i]
		}

		if i < len(fmtLines) {
			fmtLine = fmtLines[i]
		}

		if origLine != fmtLine {
			if origLine != "" {
				_, _ = fmt.Fprintf(w, "-%s\n", origLine)
			}

			if fmtLine != "" {
				_, _ = fmt.Fprintf(w, "+%s\n", fmtLine)
			}
		} else if origLine != "" {
			_, _ = fmt.Fprintf(w, " %s\n", origLine)
		}
	}
}
