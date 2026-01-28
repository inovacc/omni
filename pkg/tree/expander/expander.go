package expander

import (
	"fmt"
	"strings"
)

// Expand takes a path pattern with brace expansion syntax and returns all expanded paths.
// Example: "docs/{guides,apis,gov,arch}" -> ["docs/guides", "docs/apis", "docs/gov", "docs/arch"]
// Supports nested braces: "a/{b,c/{d,e}}" -> ["a/b", "a/c/d", "a/c/e"]
func Expand(pattern string) ([]string, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty pattern")
	}

	// Find the first opening brace
	openIdx := strings.Index(pattern, "{")
	if openIdx == -1 {
		// No braces, return the pattern as-is
		return []string{pattern}, nil
	}

	// Find the matching closing brace
	closeIdx := findMatchingBrace(pattern, openIdx)
	if closeIdx == -1 {
		return nil, fmt.Errorf("unmatched opening brace at position %d", openIdx)
	}

	// Extract prefix, brace content, and suffix
	prefix := pattern[:openIdx]
	braceContent := pattern[openIdx+1 : closeIdx]
	suffix := pattern[closeIdx+1:]

	// Split the brace content by commas (handling nested braces)
	alternatives, err := splitAlternatives(braceContent)
	if err != nil {
		return nil, fmt.Errorf("invalid brace content: %w", err)
	}

	// Generate all combinations
	var results []string

	for _, alt := range alternatives {
		// Recursively expand the combined string
		expanded, err := Expand(prefix + alt + suffix)
		if err != nil {
			return nil, err
		}

		results = append(results, expanded...)
	}

	return results, nil
}

// findMatchingBrace finds the index of the closing brace that matches the opening brace at openIdx.
// Returns -1 if no matching brace is found.
func findMatchingBrace(s string, openIdx int) int {
	if openIdx < 0 || openIdx >= len(s) || s[openIdx] != '{' {
		return -1
	}

	depth := 0

	for i := openIdx; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

// splitAlternatives splits the brace content by commas, respecting nested braces.
// Example: "a,b,c" -> ["a", "b", "c"]
// Example: "a,{b,c},d" -> ["a", "{b,c}", "d"]
func splitAlternatives(content string) ([]string, error) {
	if content == "" {
		return nil, fmt.Errorf("empty brace content")
	}

	var (
		alternatives []string
		current      strings.Builder
	)

	depth := 0

	for i, ch := range content {
		switch ch {
		case '{':
			depth++

			current.WriteRune(ch)
		case '}':
			depth--
			if depth < 0 {
				return nil, fmt.Errorf("unmatched closing brace at position %d", i)
			}

			current.WriteRune(ch)
		case ',':
			if depth == 0 {
				// Top-level comma, this is a separator
				alternatives = append(alternatives, current.String())
				current.Reset()
			} else {
				// Nested comma, part of the content
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	// Add the last alternative
	if current.Len() > 0 {
		alternatives = append(alternatives, current.String())
	}

	if depth != 0 {
		return nil, fmt.Errorf("unmatched braces in content")
	}

	return alternatives, nil
}
