package sqlfmt

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// Options configures the SQL formatter
type Options struct {
	Indent    string // Indentation (default: "  ")
	Uppercase bool   // Uppercase keywords (default: true)
	Minify    bool   // Minify output
	Dialect   string // SQL dialect: mysql, postgres, sqlite, generic (default: generic)
}

// ValidateOptions configures SQL validation
type ValidateOptions struct {
	JSON    bool   // Output as JSON
	Dialect string // SQL dialect
}

// ValidateResult represents validation output
type ValidateResult struct {
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// All SQL keywords for case conversion
var allKeywords = []string{
	"SELECT", "FROM", "WHERE", "JOIN", "LEFT", "RIGHT", "INNER", "OUTER", "FULL",
	"CROSS", "ON", "AND", "OR", "NOT", "IN", "EXISTS", "BETWEEN", "LIKE", "IS",
	"NULL", "TRUE", "FALSE", "AS", "DISTINCT", "ALL", "ORDER", "BY", "ASC", "DESC",
	"GROUP", "HAVING", "LIMIT", "OFFSET", "UNION", "INTERSECT", "EXCEPT",
	"INSERT", "INTO", "VALUES", "UPDATE", "SET", "DELETE", "CREATE", "TABLE",
	"ALTER", "DROP", "INDEX", "PRIMARY", "KEY", "FOREIGN", "REFERENCES",
	"CONSTRAINT", "DEFAULT", "CHECK", "UNIQUE", "AUTO_INCREMENT", "AUTOINCREMENT",
	"IF", "CASE", "WHEN", "THEN", "ELSE", "END", "CAST", "COALESCE", "NULLIF",
	"COUNT", "SUM", "AVG", "MIN", "MAX", "WITH", "RECURSIVE", "OVER", "PARTITION",
	"WINDOW", "ROWS", "RANGE", "UNBOUNDED", "PRECEDING", "FOLLOWING", "CURRENT",
	"ROW", "FIRST", "LAST", "NULLS", "RETURNING", "CONFLICT", "DO", "NOTHING",
}

// Run formats SQL input
func Run(w io.Writer, r io.Reader, args []string, opts Options) error {
	input, err := getInput(args, r)
	if err != nil {
		return fmt.Errorf("sql: %w", err)
	}

	// Set defaults
	if opts.Indent == "" {
		opts.Indent = "  "
	}

	var output string
	if opts.Minify {
		output = minifySQL(input)
	} else {
		output = formatSQL(input, opts)
	}

	_, _ = fmt.Fprintln(w, output)

	return nil
}

// RunMinify minifies SQL
func RunMinify(w io.Writer, r io.Reader, args []string, opts Options) error {
	opts.Minify = true

	return Run(w, r, args, opts)
}

// RunValidate validates SQL syntax
func RunValidate(w io.Writer, r io.Reader, args []string, opts ValidateOptions) error {
	input, err := getInput(args, r)
	if err != nil {
		return fmt.Errorf("sql: %w", err)
	}

	result := validateSQL(input)

	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(result)
	}

	if result.Valid {
		_, _ = fmt.Fprintln(w, result.Message)
	} else {
		_, _ = fmt.Fprintf(w, "invalid SQL: %s\n", result.Error)

		return fmt.Errorf("validation failed")
	}

	return nil
}

// formatSQL formats SQL with proper indentation and keyword capitalization
func formatSQL(input string, opts Options) string {
	// Normalize whitespace
	input = normalizeWhitespace(input)

	// Tokenize
	tokens := tokenizeSQL(input)

	// Format
	var result strings.Builder

	indent := 0
	atLineStart := true
	prevToken := ""

	for i, token := range tokens {
		upper := strings.ToUpper(token)

		// Handle indentation changes
		switch upper {
		case "(":
			indent++
		case ")":
			indent--
			if indent < 0 {
				indent = 0
			}
		}

		// Check if this is a major keyword that should start on new line
		isMajorKeyword := isMajorClause(upper)

		// Add newline before major keywords (except at start)
		if isMajorKeyword && i > 0 && !atLineStart {
			result.WriteString("\n")

			atLineStart = true
		}

		// Add indent at line start
		if atLineStart && indent > 0 {
			result.WriteString(strings.Repeat(opts.Indent, indent))
		}

		// Convert keywords to uppercase if enabled
		if opts.Uppercase && isKeyword(token) {
			token = strings.ToUpper(token)
		}

		// Add space between tokens if needed
		if i > 0 && !atLineStart && needsSpace(prevToken, token) {
			result.WriteString(" ")
		}

		result.WriteString(token)
		prevToken = token
		atLineStart = false

		// Add newline after comma (except within function calls)
		if upper == "," {
			result.WriteString("\n")

			atLineStart = true
		}

		// Add newline after semicolon
		if upper == ";" {
			result.WriteString("\n")

			atLineStart = true
		}
	}

	return strings.TrimSpace(result.String())
}

// isMajorClause checks if a token is a major SQL clause keyword
func isMajorClause(upper string) bool {
	clauses := []string{
		"SELECT", "FROM", "WHERE", "JOIN", "LEFT", "RIGHT", "INNER",
		"OUTER", "FULL", "CROSS", "ON", "AND", "OR", "ORDER", "GROUP",
		"HAVING", "LIMIT", "OFFSET", "UNION", "INTERSECT", "EXCEPT",
		"INSERT", "VALUES", "UPDATE", "SET", "CREATE", "ALTER",
		"DROP", "CASE", "WHEN", "THEN", "ELSE", "END",
	}
	// Note: DELETE is not included here because "DELETE FROM" should stay together

	return slices.Contains(clauses, upper)
}

// minifySQL removes unnecessary whitespace
func minifySQL(input string) string {
	// Normalize whitespace
	input = normalizeWhitespace(input)

	// Tokenize and rejoin with minimal spacing
	tokens := tokenizeSQL(input)

	var result strings.Builder

	for i, token := range tokens {
		if i > 0 && needsSpace(tokens[i-1], token) {
			result.WriteString(" ")
		}

		result.WriteString(token)
	}

	return result.String()
}

// validateSQL performs basic SQL syntax validation
func validateSQL(input string) ValidateResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return ValidateResult{
			Valid:   false,
			Error:   "empty input",
			Message: "",
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
				Valid:   false,
				Error:   "unbalanced parentheses: unexpected ')'",
				Message: "",
			}
		}
	}

	if parenCount > 0 {
		return ValidateResult{
			Valid:   false,
			Error:   "unbalanced parentheses: missing ')'",
			Message: "",
		}
	}

	// Check for balanced quotes
	if !checkBalancedQuotes(input, '\'') {
		return ValidateResult{
			Valid:   false,
			Error:   "unbalanced single quotes",
			Message: "",
		}
	}

	if !checkBalancedQuotes(input, '"') {
		return ValidateResult{
			Valid:   false,
			Error:   "unbalanced double quotes",
			Message: "",
		}
	}

	// Check for basic statement structure
	upper := strings.ToUpper(input)
	validStarts := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP", "WITH", "EXPLAIN", "SHOW", "SET", "USE", "BEGIN", "COMMIT", "ROLLBACK", "TRUNCATE", "GRANT", "REVOKE"}
	hasValidStart := false

	for _, start := range validStarts {
		if strings.HasPrefix(upper, start) {
			hasValidStart = true
			break
		}
	}

	// Also check for comments starting with --
	if strings.HasPrefix(strings.TrimSpace(input), "--") {
		hasValidStart = true
	}

	if !hasValidStart {
		return ValidateResult{
			Valid:   false,
			Error:   "unrecognized statement type",
			Message: "",
		}
	}

	return ValidateResult{
		Valid:   true,
		Error:   "",
		Message: "valid SQL",
	}
}

// tokenizeSQL splits SQL into tokens
func tokenizeSQL(input string) []string {
	var tokens []string

	var current strings.Builder

	inString := false
	stringChar := rune(0)
	inComment := false
	commentType := ""

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		// Handle string literals
		if !inComment && (ch == '\'' || ch == '"') {
			if !inString {
				inString = true
				stringChar = ch

				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}

				current.WriteRune(ch)
			} else if ch == stringChar {
				current.WriteRune(ch)
				// Check for escaped quote
				if i+1 < len(runes) && runes[i+1] == stringChar {
					i++
					current.WriteRune(runes[i])
				} else {
					inString = false

					tokens = append(tokens, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(ch)
			}

			continue
		}

		if inString {
			current.WriteRune(ch)

			continue
		}

		// Handle comments
		if !inComment {
			if ch == '-' && i+1 < len(runes) && runes[i+1] == '-' {
				inComment = true
				commentType = "--"

				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}

				current.WriteString("--")

				i++

				continue
			}

			if ch == '/' && i+1 < len(runes) && runes[i+1] == '*' {
				inComment = true
				commentType = "/*"

				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}

				current.WriteString("/*")

				i++

				continue
			}
		} else {
			current.WriteRune(ch)

			if commentType == "--" && ch == '\n' {
				inComment = false

				tokens = append(tokens, current.String())
				current.Reset()
			} else if commentType == "/*" && ch == '*' && i+1 < len(runes) && runes[i+1] == '/' {
				current.WriteRune(runes[i+1])
				i++
				inComment = false

				tokens = append(tokens, current.String())
				current.Reset()
			}

			continue
		}

		// Handle operators and punctuation
		if ch == '(' || ch == ')' || ch == ',' || ch == ';' || ch == '=' ||
			ch == '<' || ch == '>' || ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}

			// Handle multi-character operators
			if i+1 < len(runes) {
				next := runes[i+1]
				if (ch == '<' && next == '=') || (ch == '>' && next == '=') ||
					(ch == '<' && next == '>') || (ch == '!' && next == '=') ||
					(ch == '|' && next == '|') {
					tokens = append(tokens, string([]rune{ch, next}))
					i++

					continue
				}
			}

			tokens = append(tokens, string(ch))

			continue
		}

		// Handle whitespace
		if unicode.IsSpace(ch) {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}

			continue
		}

		current.WriteRune(ch)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// normalizeWhitespace collapses multiple whitespace into single space
func normalizeWhitespace(s string) string {
	re := regexp.MustCompile(`\s+`)

	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// isKeyword checks if a token is a SQL keyword
func isKeyword(token string) bool {
	upper := strings.ToUpper(token)

	return slices.Contains(allKeywords, upper)
}

// needsSpace determines if a space is needed between two tokens
func needsSpace(prev, curr string) bool {
	if prev == "" || curr == "" {
		return false
	}

	// No space after opening paren or before closing paren
	if prev == "(" || curr == ")" {
		return false
	}

	// No space before opening paren after function names
	if curr == "(" && isFunction(prev) {
		return false
	}

	// No space before comma, semicolon
	if curr == "," || curr == ";" {
		return false
	}

	// No space after commas (newline will be added instead)
	if prev == "," {
		return false
	}

	// No space around dots (for qualified names)
	if prev == "." || curr == "." {
		return false
	}

	return true
}

// isFunction checks if a token is a SQL function name
func isFunction(token string) bool {
	upper := strings.ToUpper(token)

	functions := []string{
		"COUNT", "SUM", "AVG", "MIN", "MAX", "COALESCE", "NULLIF",
		"CAST", "CONVERT", "SUBSTRING", "CONCAT", "LENGTH", "TRIM",
		"UPPER", "LOWER", "ROUND", "ABS", "NOW", "DATE", "TIME",
		"YEAR", "MONTH", "DAY", "HOUR", "MINUTE", "SECOND",
		"IF", "IFNULL", "NVL", "IIF", "CASE",
	}

	return slices.Contains(functions, upper)
}

// checkBalancedQuotes checks if quotes are balanced
func checkBalancedQuotes(s string, quote rune) bool {
	count := 0
	escaped := false

	for _, ch := range s {
		if ch == '\\' {
			escaped = !escaped

			continue
		}

		if ch == quote && !escaped {
			count++
		}

		escaped = false
	}

	return count%2 == 0
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
