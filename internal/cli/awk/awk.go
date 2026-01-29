package awk

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// AwkOptions configures the awk command behavior
type AwkOptions struct {
	FieldSeparator string            // -F: input field separator
	Variables      map[string]string // -v: var=value
	Program        string            // program text
}

// RunAwk executes a basic AWK-like pattern scanning
// This is a simplified subset supporting:
// - Field access ($1, $2, etc.)
// - Print statements
// - Pattern matching (BEGIN, END, /regex/)
// - Basic string operations
func RunAwk(w io.Writer, args []string, opts AwkOptions) error {
	if opts.Program == "" && len(args) > 0 {
		opts.Program = args[0]
		args = args[1:]
	}

	if opts.Program == "" {
		return fmt.Errorf("awk: no program text")
	}

	if opts.FieldSeparator == "" {
		opts.FieldSeparator = " "
	}

	program, err := parseAwkProgram(opts.Program)
	if err != nil {
		return fmt.Errorf("awk: %w", err)
	}

	// Execute BEGIN block
	if program.begin != nil {
		if err := executeAwkAction(w, program.begin, nil, 0, opts); err != nil {
			return err
		}
	}

	// Process files or stdin
	if len(args) == 0 {
		if err := awkProcessReader(w, os.Stdin, program, opts); err != nil {
			return err
		}
	} else {
		for _, file := range args {
			if file == "-" {
				if err := awkProcessReader(w, os.Stdin, program, opts); err != nil {
					return err
				}

				continue
			}

			f, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("awk: %w", err)
			}

			err = awkProcessReader(w, f, program, opts)
			if closeErr := f.Close(); closeErr != nil && err == nil {
				err = closeErr
			}

			if err != nil {
				return err
			}
		}
	}

	// Execute END block
	if program.end != nil {
		if err := executeAwkAction(w, program.end, nil, 0, opts); err != nil {
			return err
		}
	}

	return nil
}

type awkProgram struct {
	begin *awkAction
	end   *awkAction
	rules []awkRule
}

type awkRule struct {
	pattern *regexp.Regexp
	action  *awkAction
}

type awkAction struct {
	commands []string
}

func parseAwkProgram(text string) (*awkProgram, error) {
	program := &awkProgram{}
	text = strings.TrimSpace(text)

	// Simple parser for basic AWK syntax
	// Supports: BEGIN{...}, END{...}, /pattern/{...}, {action}

	for len(text) > 0 {
		text = strings.TrimSpace(text)
		if len(text) == 0 {
			break
		}

		if after, ok := strings.CutPrefix(text, "BEGIN"); ok {
			text = after
			text = strings.TrimSpace(text)

			action, rest, err := parseAwkAction(text)
			if err != nil {
				return nil, err
			}

			program.begin = action
			text = rest
		} else if after, ok := strings.CutPrefix(text, "END"); ok {
			text = after
			text = strings.TrimSpace(text)

			action, rest, err := parseAwkAction(text)
			if err != nil {
				return nil, err
			}

			program.end = action
			text = rest
		} else if strings.HasPrefix(text, "/") {
			// Pattern with regex
			end := strings.Index(text[1:], "/")
			if end == -1 {
				return nil, fmt.Errorf("unterminated regex")
			}

			pattern := text[1 : end+1]

			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex: %w", err)
			}

			text = strings.TrimSpace(text[end+2:])

			action, rest, err := parseAwkAction(text)
			if err != nil {
				return nil, err
			}

			program.rules = append(program.rules, awkRule{pattern: re, action: action})
			text = rest
		} else if strings.HasPrefix(text, "{") {
			// Action without a pattern (matches all lines)
			action, rest, err := parseAwkAction(text)
			if err != nil {
				return nil, err
			}

			program.rules = append(program.rules, awkRule{action: action})
			text = rest
		} else {
			// Assume it's a simple print expression
			program.rules = append(program.rules, awkRule{
				action: &awkAction{commands: []string{"print " + text}},
			})

			break
		}
	}

	return program, nil
}

func parseAwkAction(text string) (*awkAction, string, error) {
	if !strings.HasPrefix(text, "{") {
		return nil, text, fmt.Errorf("expected '{'")
	}

	// Find a matching closing brace
	depth := 0
	end := -1

	for i, c := range text {
		if c == '{' {
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}

	if end == -1 {
		return nil, text, fmt.Errorf("unmatched '{'")
	}

	body := text[1:end]
	rest := text[end+1:]

	// Split commands by semicolon or newline
	commands := strings.FieldsFunc(body, func(r rune) bool {
		return r == ';' || r == '\n'
	})

	var trimmed []string

	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd != "" {
			trimmed = append(trimmed, cmd)
		}
	}

	return &awkAction{commands: trimmed}, rest, nil
}

func awkProcessReader(w io.Writer, r io.Reader, program *awkProgram, opts AwkOptions) error {
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Split into fields
		var fields []string
		if opts.FieldSeparator == " " {
			fields = strings.Fields(line)
		} else {
			fields = strings.Split(line, opts.FieldSeparator)
		}

		// Prepend $0 (whole line)
		allFields := append([]string{line}, fields...)

		// Execute matching rules
		for _, rule := range program.rules {
			if rule.pattern != nil && !rule.pattern.MatchString(line) {
				continue
			}

			if err := executeAwkAction(w, rule.action, allFields, lineNum, opts); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

func executeAwkAction(w io.Writer, action *awkAction, fields []string, _ int, _ AwkOptions) error {
	if action == nil {
		return nil
	}

	for _, cmd := range action.commands {
		cmd = strings.TrimSpace(cmd)

		if after, ok := strings.CutPrefix(cmd, "print"); ok {
			args := after
			args = strings.TrimSpace(args)

			if args == "" {
				// print with no args prints $0
				if len(fields) > 0 {
					_, _ = fmt.Fprintln(w, fields[0])
				}
			} else {
				// Expand field references and print
				output := expandAwkFields(args, fields)
				_, _ = fmt.Fprintln(w, output)
			}
		}
	}

	return nil
}

func expandAwkFields(expr string, fields []string) string {
	// Handle comma-separated print arguments
	parts := strings.Split(expr, ",")

	outputs := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		output := expandSingleField(part, fields)
		outputs = append(outputs, output)
	}

	return strings.Join(outputs, " ")
}

func expandSingleField(expr string, fields []string) string {
	// Handle string literals
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return expr[1 : len(expr)-1]
	}

	// Handle field references ($0, $1, $2, etc.)
	if strings.HasPrefix(expr, "$") {
		numStr := expr[1:]
		if num, err := strconv.Atoi(numStr); err == nil {
			if num >= 0 && num < len(fields) {
				return fields[num]
			}

			return ""
		}
	}

	// Handle NF (number of fields)
	if expr == "NF" {
		if len(fields) > 0 {
			return strconv.Itoa(len(fields) - 1) // -1 because $0 is included
		}

		return "0"
	}

	return expr
}
