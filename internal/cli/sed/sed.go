package sed

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/input"
)

// SedOptions configures the sed command behavior
type SedOptions struct {
	Expression []string // -e: add the script to the commands to be executed
	InPlace    bool     // -i: edit files in place
	InPlaceExt string   // -i extension: backup extension for in-place edit
	Quiet      bool     // -n: suppress automatic printing of pattern space
	Extended   bool     // -E/-r: use extended regular expressions
}

// RunSed performs stream editing on input
// r is the default input reader (used when args is empty or contains "-")
func RunSed(w io.Writer, r io.Reader, args []string, opts SedOptions) error {
	if len(opts.Expression) == 0 && len(args) > 0 {
		opts.Expression = []string{args[0]}
		args = args[1:]
	}

	if len(opts.Expression) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "sed: no expression specified")
	}

	// Parse all expressions into commands
	var commands []sedCommand

	for _, expr := range opts.Expression {
		cmd, err := parseSedExpression(expr)
		if err != nil {
			return fmt.Errorf("sed: %w", err)
		}

		commands = append(commands, cmd)
	}

	// Handle in-place editing separately (requires file paths)
	if opts.InPlace {
		if len(args) == 0 {
			return fmt.Errorf("sed: no input files for in-place editing")
		}

		for _, file := range args {
			if file == "-" {
				return fmt.Errorf("sed: cannot do in-place editing on stdin")
			}

			if err := sedProcessInPlace(file, commands, opts); err != nil {
				return err
			}
		}

		return nil
	}

	// Use input package for reading
	sources, err := input.Open(args, r)
	if err != nil {
		return fmt.Errorf("sed: %w", err)
	}
	defer input.CloseAll(sources)

	for _, src := range sources {
		if err := sedProcessReader(w, src.Reader, commands, opts); err != nil {
			return err
		}
	}

	return nil
}

type sedCommand interface {
	execute(line string, lineNum int) (string, bool) // returns modified line and whether to print
}

// substitution command: s/pattern/replacement/flags
type sedSubstitute struct {
	pattern     *regexp.Regexp
	replacement string
	global      bool
	printOnly   bool
	nthMatch    int
}

func (s *sedSubstitute) execute(line string, _ int) (string, bool) {
	var result string
	if s.global {
		result = s.pattern.ReplaceAllString(line, s.replacement)
	} else if s.nthMatch > 0 {
		// Replace nth occurrence
		count := 0
		result = s.pattern.ReplaceAllStringFunc(line, func(match string) string {
			count++
			if count == s.nthMatch {
				return s.pattern.ReplaceAllString(match, s.replacement)
			}

			return match
		})
	} else {
		// Replace first occurrence only
		loc := s.pattern.FindStringIndex(line)
		if loc != nil {
			result = line[:loc[0]] + s.pattern.ReplaceAllString(line[loc[0]:loc[1]], s.replacement) + line[loc[1]:]
		} else {
			result = line
		}
	}

	return result, !s.printOnly
}

// delete command: d
type sedDelete struct {
	addressStart int
	addressEnd   int
	pattern      *regexp.Regexp
}

func (d *sedDelete) execute(line string, lineNum int) (string, bool) {
	if d.pattern != nil {
		if d.pattern.MatchString(line) {
			return "", false
		}

		return line, true
	}

	if d.addressStart > 0 {
		if d.addressEnd > 0 {
			if lineNum >= d.addressStart && lineNum <= d.addressEnd {
				return "", false
			}
		} else if lineNum == d.addressStart {
			return "", false
		}
	}

	return line, true
}

// print command: p
type sedPrint struct {
	pattern *regexp.Regexp
}

func (p *sedPrint) execute(line string, _ int) (string, bool) {
	if p.pattern != nil {
		return line, p.pattern.MatchString(line)
	}

	return line, true
}

// quit command: q
type sedQuit struct{}

func (q *sedQuit) execute(line string, _ int) (string, bool) {
	return line, true
}

func parseSedExpression(expr string) (sedCommand, error) {
	expr = strings.TrimSpace(expr)

	// Check for address prefix
	if len(expr) > 0 && expr[0] == '/' {
		// Pattern address: /pattern/command
		end := strings.Index(expr[1:], "/")
		if end == -1 {
			return nil, cmderr.Wrap(cmderr.ErrInvalidInput, "sed: unterminated address regex")
		}

		pattern := expr[1 : end+1]

		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("sed: invalid regex: %s", err))
		}

		rest := strings.TrimSpace(expr[end+2:])
		if rest == "d" {
			return &sedDelete{pattern: re}, nil
		}

		if rest == "p" {
			return &sedPrint{pattern: re}, nil
		}
	}

	// Check for line address
	if len(expr) > 0 && (expr[0] >= '0' && expr[0] <= '9') {
		// Find end of address
		i := 0
		for i < len(expr) && expr[i] >= '0' && expr[i] <= '9' {
			i++
		}

		startAddr, _ := strconv.Atoi(expr[:i])
		rest := expr[i:]

		var endAddr int

		if len(rest) > 0 && rest[0] == ',' {
			rest = rest[1:]

			j := 0
			for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
				j++
			}

			if j > 0 {
				endAddr, _ = strconv.Atoi(rest[:j])
				rest = rest[j:]
			}
		}

		rest = strings.TrimSpace(rest)
		if rest == "d" {
			return &sedDelete{addressStart: startAddr, addressEnd: endAddr}, nil
		}
	}

	// Substitution command
	if len(expr) > 0 && expr[0] == 's' {
		return parseSubstitute(expr)
	}

	// Simple commands
	if expr == "d" {
		return &sedDelete{}, nil
	}

	if expr == "p" {
		return &sedPrint{}, nil
	}

	if expr == "q" {
		return &sedQuit{}, nil
	}

	return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("sed: unknown command: %s", expr))
}

func parseSubstitute(expr string) (*sedSubstitute, error) {
	if len(expr) < 4 || expr[0] != 's' {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, "sed: invalid substitution")
	}

	delim := expr[1]

	parts := strings.Split(expr[2:], string(delim))
	if len(parts) < 2 {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, "sed: invalid substitution")
	}

	pattern := parts[0]
	replacement := parts[1]

	flags := ""
	if len(parts) > 2 {
		flags = parts[2]
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("sed: invalid regex: %s", err))
	}

	sub := &sedSubstitute{
		pattern:     re,
		replacement: replacement,
	}

	for _, f := range flags {
		switch f {
		case 'g':
			sub.global = true
		case 'p':
			sub.printOnly = true
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			sub.nthMatch = int(f - '0')
		}
	}

	return sub, nil
}

func sedProcessReader(w io.Writer, r io.Reader, commands []sedCommand, opts SedOptions) error {
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		shouldPrint := !opts.Quiet

		for _, cmd := range commands {
			var doPrint bool

			line, doPrint = cmd.execute(line, lineNum)
			if !doPrint {
				shouldPrint = false
			}

			if _, ok := cmd.(*sedQuit); ok {
				if shouldPrint {
					_, _ = fmt.Fprintln(w, line)
				}

				return nil
			}
		}

		if shouldPrint {
			_, _ = fmt.Fprintln(w, line)
		}
	}

	return scanner.Err()
}

func sedProcessInPlace(path string, commands []sedCommand, opts SedOptions) error {
	// Read entire file
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Create backup if extension specified
	if opts.InPlaceExt != "" {
		backupPath := path + opts.InPlaceExt
		if err := os.WriteFile(backupPath, content, 0644); err != nil {
			return err
		}
	}

	// Process lines
	var output strings.Builder

	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		shouldPrint := !opts.Quiet

		for _, cmd := range commands {
			var doPrint bool

			line, doPrint = cmd.execute(line, lineNum+1)
			if !doPrint {
				shouldPrint = false
			}
		}

		if shouldPrint {
			output.WriteString(line)

			if lineNum < len(lines)-1 {
				output.WriteString("\n")
			}
		}
	}

	// Write back
	return os.WriteFile(path, []byte(output.String()), 0644)
}
