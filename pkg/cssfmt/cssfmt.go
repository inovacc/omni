// Package cssfmt provides CSS formatting, minification, and validation.
package cssfmt

import (
	"sort"
	"strings"
	"unicode"
)

// Options configures the CSS formatter.
type Options struct {
	Indent    string // Indentation (default: "  ")
	SortProps bool   // Sort properties alphabetically
	SortRules bool   // Sort selectors alphabetically
}

// Option is a functional option for Format.
type Option func(*Options)

// WithIndent sets the indentation string.
func WithIndent(indent string) Option {
	return func(o *Options) { o.Indent = indent }
}

// WithSortProps enables alphabetical property sorting.
func WithSortProps() Option {
	return func(o *Options) { o.SortProps = true }
}

// WithSortRules enables alphabetical selector sorting.
func WithSortRules() Option {
	return func(o *Options) { o.SortRules = true }
}

// ValidateResult represents CSS validation output.
type ValidateResult struct {
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Rule represents a CSS rule (selector + declarations).
type Rule struct {
	Selector     string
	Declarations []Declaration
	IsAtRule     bool
	AtRuleBody   string
}

// Declaration represents a CSS property: value pair.
type Declaration struct {
	Property string
	Value    string
}

// Format formats CSS input with proper indentation.
func Format(input string, opts ...Option) string {
	o := Options{Indent: "  "}
	for _, opt := range opts {
		opt(&o)
	}

	return formatCSS(input, o)
}

// Minify removes unnecessary whitespace from CSS.
func Minify(input string) string {
	return minifyCSS(input)
}

// Validate performs basic CSS syntax validation.
func Validate(input string) ValidateResult {
	return validateCSS(input)
}

// Parse parses CSS input into rules.
func Parse(input string) []Rule {
	internal := parseCSS(input)

	rules := make([]Rule, len(internal))
	for i, r := range internal {
		decls := make([]Declaration, len(r.declarations))
		for j, d := range r.declarations {
			decls[j] = Declaration{Property: d.property, Value: d.value}
		}

		rules[i] = Rule{
			Selector:     r.selector,
			Declarations: decls,
			IsAtRule:     r.isAtRule,
			AtRuleBody:   r.atRuleBody,
		}
	}

	return rules
}

// RemoveComments removes CSS comments from input.
func RemoveComments(input string) string {
	return removeComments(input)
}

// ParseDeclarations parses CSS declarations from a string.
func ParseDeclarations(input string) []Declaration {
	internal := parseDeclarations(input)

	decls := make([]Declaration, len(internal))
	for i, d := range internal {
		decls[i] = Declaration{Property: d.property, Value: d.value}
	}

	return decls
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
