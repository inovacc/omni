package yq

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/inovacc/omni/pkg/cli/jq"
)

// YqOptions configures the yq command behavior
type YqOptions struct {
	Raw        bool // -r: output raw strings
	Compact    bool // -c: compact JSON output
	OutputJSON bool // -o=json: output as JSON
	OutputYAML bool // -o=yaml: output as YAML (default)
	NullInput  bool // -n: don't read any input
	Indent     int  // indent level (default 2)
}

// RunYq executes yq-like YAML processing
func RunYq(w io.Writer, args []string, opts YqOptions) error {
	filter := "."

	var files []string

	if len(args) > 0 {
		filter = args[0]
		files = args[1:]
	}

	if opts.Indent == 0 {
		opts.Indent = 2
	}

	var inputs []any

	if opts.NullInput {
		inputs = []any{nil}
	} else if len(files) == 0 {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("yq: %w", err)
		}

		docs, err := parseYAML(string(data))
		if err != nil {
			return fmt.Errorf("yq: parse error: %w", err)
		}

		inputs = docs
	} else {
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("yq: %w", err)
			}

			docs, err := parseYAML(string(data))
			if err != nil {
				return fmt.Errorf("yq: %s: parse error: %w", file, err)
			}

			inputs = append(inputs, docs...)
		}
	}

	for _, input := range inputs {
		results, err := jq.ApplyJqFilter(input, filter) // Reuse jq filter logic
		if err != nil {
			return fmt.Errorf("yq: %w", err)
		}

		for _, result := range results {
			if err := outputYqResult(w, result, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseYAML parses YAML content (simplified - handles basic YAML)
func parseYAML(content string) ([]any, error) {
	var docs []any

	// Split by document separator
	docStrings := strings.SplitSeq(content, "\n---")

	for docStr := range docStrings {
		docStr = strings.TrimSpace(docStr)
		if docStr == "" || docStr == "---" {
			continue
		}

		// Remove leading ---
		docStr = strings.TrimPrefix(docStr, "---")
		docStr = strings.TrimSpace(docStr)

		if docStr == "" {
			continue
		}

		doc, err := parseYAMLDocument(docStr)
		if err != nil {
			return nil, err
		}

		docs = append(docs, doc)
	}

	if len(docs) == 0 {
		return []any{nil}, nil
	}

	return docs, nil
}

func parseYAMLDocument(content string) (any, error) {
	lines := strings.Split(content, "\n")
	return parseYAMLLines(lines, 0)
}

func parseYAMLLines(lines []string, baseIndent int) (any, error) {
	if len(lines) == 0 {
		return nil, nil //nolint:nilnil // Valid: empty input returns nil value
	}

	// Check first non-empty line
	firstLine := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			firstLine = line
			break
		}
	}

	if firstLine == "" {
		return nil, nil //nolint:nilnil // Valid: empty content returns nil value
	}

	trimmed := strings.TrimSpace(firstLine)

	// Array item
	if strings.HasPrefix(trimmed, "- ") {
		return parseYAMLArray(lines, baseIndent)
	}

	// Object
	if strings.Contains(trimmed, ":") {
		return parseYAMLObject(lines, baseIndent)
	}

	// Scalar value
	return parseYAMLScalar(trimmed), nil
}

func parseYAMLArray(lines []string, _ int) ([]any, error) {
	var (
		result      []any
		currentItem []string
	)

	itemIndent := -1

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		indent := countIndent(line)
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "- ") {
			// New array item
			if len(currentItem) > 0 {
				item, err := parseArrayItem(currentItem, itemIndent)
				if err != nil {
					return nil, err
				}

				result = append(result, item)
				currentItem = nil
			}

			itemIndent = indent
			// Remove "- " prefix
			content := strings.TrimPrefix(trimmed, "- ")
			if content != "" {
				currentItem = append(currentItem, strings.Repeat(" ", indent+2)+content)
			}
		} else if indent > itemIndent {
			currentItem = append(currentItem, line)
		}
	}

	// Handle last item
	if len(currentItem) > 0 {
		item, err := parseArrayItem(currentItem, itemIndent)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, nil
}

func parseArrayItem(lines []string, baseIndent int) (any, error) {
	if len(lines) == 0 {
		return nil, nil //nolint:nilnil // Valid: empty array item returns nil
	}

	// Single line value
	if len(lines) == 1 {
		return parseYAMLScalar(strings.TrimSpace(lines[0])), nil
	}

	return parseYAMLLines(lines, baseIndent+2)
}

func parseYAMLObject(lines []string, _ int) (map[string]any, error) {
	result := make(map[string]any)

	var (
		currentKey   string
		currentValue []string
	)

	keyIndent := -1

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		indent := countIndent(line)
		trimmed := strings.TrimSpace(line)

		// Check if this is a key line
		colonIdx := strings.Index(trimmed, ":")
		if colonIdx > 0 && (keyIndent == -1 || indent == keyIndent) {
			// Save previous key-value
			if currentKey != "" {
				value, err := parseYAMLValue(currentValue, keyIndent)
				if err != nil {
					return nil, err
				}

				result[currentKey] = value
				currentValue = nil
			}

			keyIndent = indent
			currentKey = trimmed[:colonIdx]

			// Check for inline value
			valueStr := strings.TrimSpace(trimmed[colonIdx+1:])
			if valueStr != "" {
				result[currentKey] = parseYAMLScalar(valueStr)
				currentKey = ""
			}
		} else if indent > keyIndent && currentKey != "" {
			currentValue = append(currentValue, line)
		}
	}

	// Handle last key
	if currentKey != "" {
		value, err := parseYAMLValue(currentValue, keyIndent)
		if err != nil {
			return nil, err
		}

		result[currentKey] = value
	}

	return result, nil
}

func parseYAMLValue(lines []string, baseIndent int) (any, error) {
	if len(lines) == 0 {
		return nil, nil //nolint:nilnil // Valid: empty value returns nil
	}

	return parseYAMLLines(lines, baseIndent)
}

func parseYAMLScalar(s string) any {
	s = strings.TrimSpace(s)

	// Remove quotes
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return s[1 : len(s)-1]
	}

	// Boolean
	lower := strings.ToLower(s)
	if lower == "true" || lower == "yes" || lower == "on" {
		return true
	}

	if lower == "false" || lower == "no" || lower == "off" {
		return false
	}

	// Null
	if lower == "null" || lower == "~" || s == "" {
		return nil
	}

	// Number
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return float64(i)
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	return s
}

func countIndent(line string) int {
	count := 0

loop:
	for _, c := range line {
		switch c {
		case ' ':
			count++
		case '\t':
			count += 2
		default:
			break loop
		}
	}

	return count
}

func outputYqResult(w io.Writer, result any, opts YqOptions) error {
	if opts.Raw {
		if s, ok := result.(string); ok {
			_, _ = fmt.Fprintln(w, s)
			return nil
		}
	}

	if opts.OutputJSON {
		encoder := json.NewEncoder(w)
		if !opts.Compact {
			encoder.SetIndent("", "  ")
		}

		return encoder.Encode(result)
	}

	// Output as YAML
	return writeYAML(w, result, 0, opts.Indent)
}

func writeYAML(w io.Writer, v any, depth int, indentSize int) error {
	indent := strings.Repeat(" ", depth*indentSize)

	switch val := v.(type) {
	case nil:
		_, _ = fmt.Fprintln(w, "null")
	case bool:
		_, _ = fmt.Fprintln(w, val)
	case float64:
		if val == float64(int64(val)) {
			_, _ = fmt.Fprintf(w, "%d\n", int64(val))
		} else {
			_, _ = fmt.Fprintf(w, "%g\n", val)
		}
	case string:
		if needsQuoting(val) {
			_, _ = fmt.Fprintf(w, "%q\n", val)
		} else {
			_, _ = fmt.Fprintln(w, val)
		}
	case []any:
		if len(val) == 0 {
			_, _ = fmt.Fprintln(w, "[]")
			return nil
		}

		for i, item := range val {
			if i == 0 && depth > 0 {
				_, _ = fmt.Fprint(w, "\n")
			}

			_, _ = fmt.Fprintf(w, "%s- ", indent)
			if isScalar(item) {
				if err := writeYAML(w, item, 0, indentSize); err != nil {
					return err
				}
			} else {
				if err := writeYAML(w, item, depth+1, indentSize); err != nil {
					return err
				}
			}
		}
	case map[string]any:
		if len(val) == 0 {
			_, _ = fmt.Fprintln(w, "{}")
			return nil
		}

		first := true
		for k, item := range val {
			if first && depth > 0 {
				_, _ = fmt.Fprint(w, "\n")
				first = false
			}

			_, _ = fmt.Fprintf(w, "%s%s: ", indent, k)
			if isScalar(item) {
				if err := writeYAML(w, item, 0, indentSize); err != nil {
					return err
				}
			} else {
				if err := writeYAML(w, item, depth+1, indentSize); err != nil {
					return err
				}
			}
		}
	default:
		_, _ = fmt.Fprintf(w, "%v\n", val)
	}

	return nil
}

func isScalar(v any) bool {
	switch v.(type) {
	case nil, bool, float64, string:
		return true
	default:
		return false
	}
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	// Check for special characters
	for _, c := range s {
		if c == ':' || c == '#' || c == '\n' || c == '\t' {
			return true
		}
	}
	// Check for special values
	lower := strings.ToLower(s)
	if lower == "true" || lower == "false" || lower == "null" || lower == "yes" || lower == "no" {
		return true
	}

	return false
}
