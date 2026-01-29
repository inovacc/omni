package forloop

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Options configures the for command behavior
type Options struct {
	Delimiter string // delimiter for splitting input
	Variable  string // variable name to use (default: item)
	DryRun    bool   // print commands without executing
	Parallel  int    // number of parallel executions (0 = sequential)
}

// RunRange executes a command for each number in a range
// Usage: for range START END [STEP] -- COMMAND
func RunRange(w io.Writer, start, end, step int, command string, opts Options) error {
	if step == 0 {
		step = 1
		if start > end {
			step = -1
		}
	}

	varName := opts.Variable
	if varName == "" {
		varName = "i"
	}

	for i := start; ; {
		if step > 0 && i > end {
			break
		}

		if step < 0 && i < end {
			break
		}

		cmd := replaceVariable(command, varName, strconv.Itoa(i))

		if opts.DryRun {
			_, _ = fmt.Fprintf(w, "%s\n", cmd)
		} else {
			if err := executeCommand(w, cmd); err != nil {
				return fmt.Errorf("for range: command failed at %d: %w", i, err)
			}
		}

		i += step
	}

	return nil
}

// RunEach executes a command for each item in a list
// Usage: for each ITEM1 ITEM2 ... -- COMMAND
func RunEach(w io.Writer, items []string, command string, opts Options) error {
	varName := opts.Variable
	if varName == "" {
		varName = "item"
	}

	for _, item := range items {
		cmd := replaceVariable(command, varName, item)

		if opts.DryRun {
			_, _ = fmt.Fprintf(w, "%s\n", cmd)
		} else {
			if err := executeCommand(w, cmd); err != nil {
				return fmt.Errorf("for each: command failed for %q: %w", item, err)
			}
		}
	}

	return nil
}

// RunLines executes a command for each line from stdin or file
// Usage: for lines [FILE] -- COMMAND
func RunLines(w io.Writer, r io.Reader, command string, opts Options) error {
	varName := opts.Variable
	if varName == "" {
		varName = "line"
	}

	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		cmd := replaceVariable(command, varName, line)
		cmd = replaceVariable(cmd, "n", strconv.Itoa(lineNum))

		if opts.DryRun {
			_, _ = fmt.Fprintf(w, "%s\n", cmd)
		} else {
			if err := executeCommand(w, cmd); err != nil {
				return fmt.Errorf("for lines: command failed at line %d: %w", lineNum, err)
			}
		}
	}

	return scanner.Err()
}

// RunSplit executes a command for each token split by delimiter
// Usage: for split DELIM INPUT -- COMMAND
func RunSplit(w io.Writer, input, delimiter, command string, opts Options) error {
	varName := opts.Variable
	if varName == "" {
		varName = "item"
	}

	items := strings.Split(input, delimiter)

	for i, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		cmd := replaceVariable(command, varName, item)
		cmd = replaceVariable(cmd, "i", strconv.Itoa(i))

		if opts.DryRun {
			_, _ = fmt.Fprintf(w, "%s\n", cmd)
		} else {
			if err := executeCommand(w, cmd); err != nil {
				return fmt.Errorf("for split: command failed for %q: %w", item, err)
			}
		}
	}

	return nil
}

// RunGlob executes a command for each file matching a glob pattern
// Usage: for glob PATTERN -- COMMAND
func RunGlob(w io.Writer, pattern, command string, opts Options) error {
	varName := opts.Variable
	if varName == "" {
		varName = "file"
	}

	// Use filepath.Glob for pattern matching
	matches, err := globFiles(pattern)
	if err != nil {
		return fmt.Errorf("for glob: %w", err)
	}

	for _, match := range matches {
		cmd := replaceVariable(command, varName, match)

		if opts.DryRun {
			_, _ = fmt.Fprintf(w, "%s\n", cmd)
		} else {
			if err := executeCommand(w, cmd); err != nil {
				return fmt.Errorf("for glob: command failed for %q: %w", match, err)
			}
		}
	}

	return nil
}

// replaceVariable replaces ${var} and $var with value
func replaceVariable(cmd, varName, value string) string {
	// Replace ${var}
	cmd = strings.ReplaceAll(cmd, "${"+varName+"}", value)
	// Replace $var (word boundary aware)
	cmd = replaceWordVariable(cmd, "$"+varName, value)

	return cmd
}

// replaceWordVariable replaces $var considering word boundaries
func replaceWordVariable(s, old, new string) string {
	result := s

	for {
		idx := strings.Index(result, old)
		if idx == -1 {
			break
		}

		// Check if next char is alphanumeric (part of variable name)
		endIdx := idx + len(old)
		if endIdx < len(result) {
			next := result[endIdx]
			if isAlphaNumeric(next) {
				// Skip this, it's part of a longer variable name
				// Find after this occurrence
				result = result[:idx] + "\x00SKIP\x00" + result[idx+1:]
				continue
			}
		}

		result = result[:idx] + new + result[endIdx:]
	}

	// Restore skipped $
	result = strings.ReplaceAll(result, "\x00SKIP\x00", "$")

	return result
}

func isAlphaNumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// executeCommand runs a shell command
func executeCommand(w io.Writer, command string) error {
	var cmd *exec.Cmd

	// Use shell to execute
	if isWindows() {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}

// globFiles returns files matching the pattern
func globFiles(pattern string) ([]string, error) {
	// Simple glob using filepath
	return filepathGlob(pattern)
}
