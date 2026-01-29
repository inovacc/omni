// Package caseconv provides text case conversion utilities.
package caseconv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CaseType represents the case convention
type CaseType string

const (
	CaseUpper    CaseType = "upper"    // UPPERCASE
	CaseLower    CaseType = "lower"    // lowercase
	CaseTitle    CaseType = "title"    // Title Case
	CaseSentence CaseType = "sentence" // Sentence case
	CaseCamel    CaseType = "camel"    // camelCase
	CasePascal   CaseType = "pascal"   // PascalCase
	CaseSnake    CaseType = "snake"    // snake_case
	CaseKebab    CaseType = "kebab"    // kebab-case
	CaseConstant CaseType = "constant" // CONSTANT_CASE
	CaseDot      CaseType = "dot"      // dot.case
	CasePath     CaseType = "path"     // path/case
	CaseSwap     CaseType = "swap"     // sWAP cASE
	CaseToggle   CaseType = "toggle"   // Toggle first char
)

// Options configures case conversion
type Options struct {
	Case CaseType // Target case type
	JSON bool     // Output as JSON
}

// Result represents conversion result
type Result struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Case   string `json:"case"`
}

// ListResult represents multiple conversion results
type ListResult struct {
	Case    string   `json:"case"`
	Results []Result `json:"results"`
}

// RunCase executes case conversion
func RunCase(w io.Writer, args []string, opts Options) error {
	var inputs []string

	if len(args) == 0 {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputs = append(inputs, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return err
		}
	} else {
		inputs = args
	}

	if opts.JSON {
		result := ListResult{Case: string(opts.Case)}
		for _, input := range inputs {
			output := Convert(input, opts.Case)
			result.Results = append(result.Results, Result{
				Input:  input,
				Output: output,
				Case:   string(opts.Case),
			})
		}

		return json.NewEncoder(w).Encode(result)
	}

	for _, input := range inputs {
		output := Convert(input, opts.Case)
		_, _ = fmt.Fprintln(w, output)
	}

	return nil
}

// Convert converts a string to the specified case
func Convert(s string, caseType CaseType) string {
	switch caseType {
	case CaseUpper:
		return ToUpper(s)
	case CaseLower:
		return ToLower(s)
	case CaseTitle:
		return ToTitle(s)
	case CaseSentence:
		return ToSentence(s)
	case CaseCamel:
		return ToCamel(s)
	case CasePascal:
		return ToPascal(s)
	case CaseSnake:
		return ToSnake(s)
	case CaseKebab:
		return ToKebab(s)
	case CaseConstant:
		return ToConstant(s)
	case CaseDot:
		return ToDot(s)
	case CasePath:
		return ToPath(s)
	case CaseSwap:
		return ToSwap(s)
	case CaseToggle:
		return ToToggle(s)
	default:
		return s
	}
}

// ToUpper converts to UPPERCASE
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower converts to lowercase
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToTitle converts to Title Case
func ToTitle(s string) string {
	return cases.Title(language.English).String(strings.ToLower(s))
}

// ToSentence converts to Sentence case
func ToSentence(s string) string {
	if len(s) == 0 {
		return s
	}

	lower := strings.ToLower(s)
	runes := []rune(lower)
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

// ToCamel converts to camelCase
func ToCamel(s string) string {
	words := splitIntoWords(s)
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(strings.ToLower(words[0]))

	for _, word := range words[1:] {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])) + strings.ToLower(word[1:]))
		}
	}

	return result.String()
}

// ToPascal converts to PascalCase
func ToPascal(s string) string {
	words := splitIntoWords(s)

	var result strings.Builder

	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])) + strings.ToLower(word[1:]))
		}
	}

	return result.String()
}

// ToSnake converts to snake_case
func ToSnake(s string) string {
	words := splitIntoWords(s)
	result := make([]string, 0, len(words))

	for _, word := range words {
		result = append(result, strings.ToLower(word))
	}

	return strings.Join(result, "_")
}

// ToKebab converts to kebab-case
func ToKebab(s string) string {
	words := splitIntoWords(s)
	result := make([]string, 0, len(words))

	for _, word := range words {
		result = append(result, strings.ToLower(word))
	}

	return strings.Join(result, "-")
}

// ToConstant converts to CONSTANT_CASE
func ToConstant(s string) string {
	words := splitIntoWords(s)
	result := make([]string, 0, len(words))

	for _, word := range words {
		result = append(result, strings.ToUpper(word))
	}

	return strings.Join(result, "_")
}

// ToDot converts to dot.case
func ToDot(s string) string {
	words := splitIntoWords(s)
	result := make([]string, 0, len(words))

	for _, word := range words {
		result = append(result, strings.ToLower(word))
	}

	return strings.Join(result, ".")
}

// ToPath converts to path/case
func ToPath(s string) string {
	words := splitIntoWords(s)
	result := make([]string, 0, len(words))

	for _, word := range words {
		result = append(result, strings.ToLower(word))
	}

	return strings.Join(result, "/")
}

// ToSwap swaps case of each character
func ToSwap(s string) string {
	var result strings.Builder

	for _, r := range s {
		if unicode.IsUpper(r) {
			result.WriteRune(unicode.ToLower(r))
		} else if unicode.IsLower(r) {
			result.WriteRune(unicode.ToUpper(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ToToggle toggles the first character's case
func ToToggle(s string) string {
	if len(s) == 0 {
		return s
	}

	runes := []rune(s)
	if unicode.IsUpper(runes[0]) {
		runes[0] = unicode.ToLower(runes[0])
	} else {
		runes[0] = unicode.ToUpper(runes[0])
	}

	return string(runes)
}

// splitIntoWords splits a string into words handling various cases
func splitIntoWords(s string) []string {
	// Handle snake_case and kebab-case
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.ReplaceAll(s, "/", " ")

	// Handle PascalCase and camelCase
	var (
		words   []string
		current strings.Builder
	)

	for i, r := range s {
		if r == ' ' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}

			continue
		}

		if unicode.IsUpper(r) && i > 0 {
			prev := rune(s[i-1])
			if !unicode.IsUpper(prev) && prev != ' ' && prev != '_' && prev != '-' && prev != '.' && prev != '/' {
				if current.Len() > 0 {
					words = append(words, current.String())
					current.Reset()
				}
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

// DetectCase detects the case type of a string
func DetectCase(s string) CaseType {
	if len(s) == 0 {
		return CaseLower
	}

	hasUpper := false
	hasLower := false
	hasUnderscore := false
	hasDash := false
	hasDot := false
	hasSlash := false

	for _, r := range s {
		if unicode.IsUpper(r) {
			hasUpper = true
		}

		if unicode.IsLower(r) {
			hasLower = true
		}

		if r == '_' {
			hasUnderscore = true
		}

		if r == '-' {
			hasDash = true
		}

		if r == '.' {
			hasDot = true
		}

		if r == '/' {
			hasSlash = true
		}
	}

	// Check for specific patterns
	if hasUnderscore && hasUpper && !hasLower {
		return CaseConstant
	}

	if hasUnderscore {
		return CaseSnake
	}

	if hasDash {
		return CaseKebab
	}

	if hasDot {
		return CaseDot
	}

	if hasSlash {
		return CasePath
	}

	if hasUpper && !hasLower {
		return CaseUpper
	}

	if hasLower && !hasUpper {
		return CaseLower
	}

	if hasUpper && hasLower {
		// Check first character
		firstRune := []rune(s)[0]
		if unicode.IsUpper(firstRune) {
			return CasePascal
		}

		return CaseCamel
	}

	return CaseLower
}

// ParseCaseType parses a string into a CaseType
func ParseCaseType(s string) (CaseType, error) {
	switch strings.ToLower(s) {
	case "upper", "uppercase":
		return CaseUpper, nil
	case "lower", "lowercase":
		return CaseLower, nil
	case "title", "titlecase":
		return CaseTitle, nil
	case "sentence", "sentencecase":
		return CaseSentence, nil
	case "camel", "camelcase":
		return CaseCamel, nil
	case "pascal", "pascalcase":
		return CasePascal, nil
	case "snake", "snakecase", "snake_case":
		return CaseSnake, nil
	case "kebab", "kebabcase", "kebab-case":
		return CaseKebab, nil
	case "constant", "constantcase", "screaming", "screaming_snake":
		return CaseConstant, nil
	case "dot", "dotcase", "dot.case":
		return CaseDot, nil
	case "path", "pathcase", "path/case":
		return CasePath, nil
	case "swap", "swapcase":
		return CaseSwap, nil
	case "toggle":
		return CaseToggle, nil
	default:
		return "", fmt.Errorf("unknown case type: %s", s)
	}
}

// ValidCaseTypes returns all valid case types
func ValidCaseTypes() []CaseType {
	return []CaseType{
		CaseUpper,
		CaseLower,
		CaseTitle,
		CaseSentence,
		CaseCamel,
		CasePascal,
		CaseSnake,
		CaseKebab,
		CaseConstant,
		CaseDot,
		CasePath,
		CaseSwap,
		CaseToggle,
	}
}

// ConvertAll converts a string to all case types
func ConvertAll(s string) map[CaseType]string {
	result := make(map[CaseType]string)
	for _, ct := range ValidCaseTypes() {
		result[ct] = Convert(s, ct)
	}

	return result
}
