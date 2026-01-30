package cssfmt

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
)

// Options configures the CSS formatter
type Options struct {
	Indent    string // Indentation (default: "  ")
	Minify    bool   // Minify output
	SortProps bool   // Sort properties alphabetically
	SortRules bool   // Sort selectors alphabetically
}

// ValidateOptions configures CSS validation
type ValidateOptions struct {
	JSON bool // Output as JSON
}

// ValidateResult represents validation output
type ValidateResult struct {
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Run formats CSS input
func Run(w io.Writer, r io.Reader, args []string, opts Options) error {
	input, err := getInput(args, r)
	if err != nil {
		return fmt.Errorf("css: %w", err)
	}

	// Set defaults
	if opts.Indent == "" {
		opts.Indent = "  "
	}

	var output string
	if opts.Minify {
		output = minifyCSS(input)
	} else {
		output = formatCSS(input, opts)
	}

	_, _ = fmt.Fprintln(w, output)

	return nil
}

// RunMinify minifies CSS
func RunMinify(w io.Writer, r io.Reader, args []string, opts Options) error {
	opts.Minify = true
	return Run(w, r, args, opts)
}

// RunValidate validates CSS syntax
func RunValidate(w io.Writer, r io.Reader, args []string, opts ValidateOptions) error {
	input, err := getInput(args, r)
	if err != nil {
		return fmt.Errorf("css: %w", err)
	}

	result := validateCSS(input)

	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(result)
	}

	if result.Valid {
		_, _ = fmt.Fprintln(w, result.Message)
	} else {
		_, _ = fmt.Fprintf(w, "invalid CSS: %s\n", result.Error)
		return fmt.Errorf("validation failed")
	}

	return nil
}

// cssRule represents a CSS rule (selector + declarations)
type cssRule struct {
	selector     string
	declarations []cssDeclaration
	isAtRule     bool
	atRuleBody   string
}

// cssDeclaration represents a CSS property: value pair
type cssDeclaration struct {
	property string
	value    string
}

// formatCSS formats CSS with proper indentation
func formatCSS(input string, opts Options) string {
	rules := parseCSS(input)

	if opts.SortRules {
		sort.Slice(rules, func(i, j int) bool {
			return rules[i].selector < rules[j].selector
		})
	}

	var result strings.Builder

	for i, rule := range rules {
		if i > 0 {
			result.WriteString("\n")
		}

		if rule.isAtRule {
			result.WriteString(rule.selector)

			if rule.atRuleBody != "" {
				result.WriteString(" {\n")
				// Format nested rules
				nestedRules := parseCSS(rule.atRuleBody)
				for j, nested := range nestedRules {
					if j > 0 {
						result.WriteString("\n")
					}

					result.WriteString(opts.Indent)
					result.WriteString(nested.selector)
					result.WriteString(" {\n")
					formatDeclarations(&result, nested.declarations, opts.Indent+opts.Indent, opts)
					result.WriteString(opts.Indent)
					result.WriteString("}\n")
				}

				result.WriteString("}")
			} else {
				result.WriteString(";")
			}
		} else {
			result.WriteString(rule.selector)
			result.WriteString(" {\n")
			formatDeclarations(&result, rule.declarations, opts.Indent, opts)
			result.WriteString("}")
		}
	}

	return strings.TrimSpace(result.String())
}

// formatDeclarations formats CSS declarations
func formatDeclarations(result *strings.Builder, declarations []cssDeclaration, indent string, opts Options) {
	if opts.SortProps {
		sort.Slice(declarations, func(i, j int) bool {
			return declarations[i].property < declarations[j].property
		})
	}

	for _, decl := range declarations {
		result.WriteString(indent)
		result.WriteString(decl.property)
		result.WriteString(": ")
		result.WriteString(decl.value)
		result.WriteString(";\n")
	}
}

// minifyCSS removes unnecessary whitespace from CSS
func minifyCSS(input string) string {
	rules := parseCSS(input)

	var result strings.Builder

	for _, rule := range rules {
		if rule.isAtRule {
			result.WriteString(rule.selector)

			if rule.atRuleBody != "" {
				result.WriteString("{")
				result.WriteString(minifyCSS(rule.atRuleBody))
				result.WriteString("}")
			} else {
				result.WriteString(";")
			}
		} else {
			result.WriteString(rule.selector)
			result.WriteString("{")

			for i, decl := range rule.declarations {
				if i > 0 {
					result.WriteString(";")
				}

				result.WriteString(decl.property)
				result.WriteString(":")
				result.WriteString(decl.value)
			}

			result.WriteString("}")
		}
	}

	return result.String()
}

// validateCSS performs basic CSS syntax validation
func validateCSS(input string) ValidateResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return ValidateResult{
			Valid: false,
			Error: "empty input",
		}
	}

	// Check for balanced braces
	braceCount := 0

	for _, ch := range input {
		switch ch {
		case '{':
			braceCount++
		case '}':
			braceCount--
		}

		if braceCount < 0 {
			return ValidateResult{
				Valid: false,
				Error: "unbalanced braces: unexpected '}'",
			}
		}
	}

	if braceCount > 0 {
		return ValidateResult{
			Valid: false,
			Error: "unbalanced braces: missing '}'",
		}
	}

	// Check for balanced parentheses
	parenCount := 0

	for _, ch := range input {
		switch ch {
		case '(':
			parenCount++
		case ')':
			parenCount--
		}

		if parenCount < 0 {
			return ValidateResult{
				Valid: false,
				Error: "unbalanced parentheses: unexpected ')'",
			}
		}
	}

	if parenCount > 0 {
		return ValidateResult{
			Valid: false,
			Error: "unbalanced parentheses: missing ')'",
		}
	}

	// Check for unclosed strings
	inString := false

	stringChar := rune(0)
	for _, ch := range input {
		if !inString && (ch == '"' || ch == '\'') {
			inString = true
			stringChar = ch
		} else if inString && ch == stringChar {
			inString = false
		}
	}

	if inString {
		return ValidateResult{
			Valid: false,
			Error: "unclosed string",
		}
	}

	return ValidateResult{
		Valid:   true,
		Message: "valid CSS",
	}
}

// parseCSS parses CSS into rules
func parseCSS(input string) []cssRule {
	var rules []cssRule

	input = removeComments(input)
	input = strings.TrimSpace(input)

	i := 0
	for i < len(input) {
		// Skip whitespace
		for i < len(input) && unicode.IsSpace(rune(input[i])) {
			i++
		}

		if i >= len(input) {
			break
		}

		// Check for at-rule
		if input[i] == '@' {
			rule, newI := parseAtRule(input, i)
			rules = append(rules, rule)
			i = newI

			continue
		}

		// Parse selector
		selectorStart := i
		braceDepth := 0

		for i < len(input) {
			if input[i] == '{' {
				if braceDepth == 0 {
					break
				}

				braceDepth++
			} else if input[i] == '}' {
				braceDepth--
			}

			i++
		}

		if i >= len(input) {
			break
		}

		selector := strings.TrimSpace(input[selectorStart:i])
		i++ // skip '{'

		// Parse declarations
		declStart := i

		braceDepth = 1
		for i < len(input) && braceDepth > 0 {
			switch input[i] {
			case '{':
				braceDepth++
			case '}':
				braceDepth--
			}

			i++
		}

		declBody := input[declStart : i-1]
		declarations := parseDeclarations(declBody)

		rules = append(rules, cssRule{
			selector:     selector,
			declarations: declarations,
		})
	}

	return rules
}

// parseAtRule parses an at-rule
func parseAtRule(input string, start int) (cssRule, int) {
	i := start

	// Find the end of the at-rule name
	for i < len(input) && !unicode.IsSpace(rune(input[i])) && input[i] != '{' && input[i] != ';' {
		i++
	}

	// Get the full at-rule including parameters
	for i < len(input) && input[i] != '{' && input[i] != ';' {
		i++
	}

	selector := strings.TrimSpace(input[start:i])

	if i >= len(input) {
		return cssRule{selector: selector, isAtRule: true}, i
	}

	if input[i] == ';' {
		return cssRule{selector: selector, isAtRule: true}, i + 1
	}

	// Has body
	i++ // skip '{'
	bodyStart := i
	braceDepth := 1

	for i < len(input) && braceDepth > 0 {
		switch input[i] {
		case '{':
			braceDepth++
		case '}':
			braceDepth--
		}

		i++
	}

	body := input[bodyStart : i-1]

	return cssRule{
		selector:   selector,
		isAtRule:   true,
		atRuleBody: body,
	}, i
}

// parseDeclarations parses CSS declarations
func parseDeclarations(input string) []cssDeclaration {
	var declarations []cssDeclaration

	input = strings.TrimSpace(input)

	// Split by semicolon, but be careful about values containing semicolons
	parts := splitDeclarations(input)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		before, after, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}

		property := strings.TrimSpace(before)
		value := strings.TrimSpace(after)

		if property != "" && value != "" {
			declarations = append(declarations, cssDeclaration{
				property: property,
				value:    value,
			})
		}
	}

	return declarations
}

// splitDeclarations splits declarations by semicolon
func splitDeclarations(input string) []string {
	var (
		parts   []string
		current strings.Builder
	)

	inString := false
	stringChar := rune(0)
	parenDepth := 0

	for _, ch := range input {
		if !inString && (ch == '"' || ch == '\'') {
			inString = true
			stringChar = ch
			current.WriteRune(ch)
		} else if inString && ch == stringChar {
			inString = false

			current.WriteRune(ch)
		} else if !inString && ch == '(' {
			parenDepth++

			current.WriteRune(ch)
		} else if !inString && ch == ')' {
			parenDepth--

			current.WriteRune(ch)
		} else if !inString && parenDepth == 0 && ch == ';' {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// removeComments removes CSS comments
func removeComments(input string) string {
	var result strings.Builder

	i := 0

	for i < len(input) {
		if i+1 < len(input) && input[i] == '/' && input[i+1] == '*' {
			// Skip until */
			i += 2
			for i+1 < len(input) {
				if input[i] == '*' && input[i+1] == '/' {
					i += 2
					break
				}

				i++
			}
		} else {
			result.WriteByte(input[i])
			i++
		}
	}

	return result.String()
}

// getInput reads input from args (file or literal) or stdin
func getInput(args []string, r io.Reader) (string, error) {
	if len(args) > 0 {
		// Check if it's a file
		if _, err := os.Stat(args[0]); err == nil {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return "", err
			}

			return string(content), nil
		}

		// Treat as literal string
		return strings.Join(args, " "), nil
	}

	// Read from stdin
	scanner := bufio.NewScanner(r)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}
