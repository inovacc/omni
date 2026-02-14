package sqlfmt

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// Options configures the SQL formatter.
type Options struct {
	Indent    string // Indentation string (default: "  ")
	Uppercase bool   // Uppercase keywords
}

// Option is a functional option for SQL formatting.
type Option func(*Options)

// WithIndent sets the indentation string.
func WithIndent(s string) Option {
	return func(o *Options) { o.Indent = s }
}

// WithUppercase enables keyword uppercasing.
func WithUppercase() Option {
	return func(o *Options) { o.Uppercase = true }
}

// ValidateResult represents the result of SQL validation.
type ValidateResult struct {
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// AllKeywords contains all recognized SQL keywords.
var AllKeywords = []string{
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

// Format formats SQL with proper indentation and keyword capitalization.
func Format(input string, opts ...Option) string {
	cfg := Options{Indent: "  "}
	for _, o := range opts {
		o(&cfg)
	}

	return formatSQL(input, cfg)
}

// Minify removes unnecessary whitespace from SQL.
func Minify(input string) string {
	return minifySQL(input)
}

// Validate performs basic SQL syntax validation.
func Validate(input string) ValidateResult {
	return validateSQL(input)
}

// Tokenize splits SQL into tokens.
func Tokenize(input string) []string {
	return tokenizeSQL(input)
}

// IsKeyword checks if a token is a SQL keyword.
func IsKeyword(token string) bool {
	return isKeyword(token)
}

// NeedsSpace determines if a space is needed between two tokens.
func NeedsSpace(prev, curr string) bool {
	return needsSpace(prev, curr)
}

// CheckBalancedQuotes checks if quotes are balanced in a string.
func CheckBalancedQuotes(s string, quote rune) bool {
	return checkBalancedQuotes(s, quote)
}

func formatSQL(input string, opts Options) string {
	input = normalizeWhitespace(input)
	tokens := tokenizeSQL(input)

	var result strings.Builder

	indent := 0
	atLineStart := true
	prevToken := ""

	for i, token := range tokens {
		upper := strings.ToUpper(token)

		switch upper {
		case "(":
			indent++
		case ")":
			indent--
			if indent < 0 {
				indent = 0
			}
		}

		isMajorKeyword := isMajorClause(upper)

		if isMajorKeyword && i > 0 && !atLineStart {
			result.WriteString("\n")

			atLineStart = true
		}

		if atLineStart && indent > 0 {
			result.WriteString(strings.Repeat(opts.Indent, indent))
		}

		if opts.Uppercase && isKeyword(token) {
			token = strings.ToUpper(token)
		}

		if i > 0 && !atLineStart && needsSpace(prevToken, token) {
			result.WriteString(" ")
		}

		result.WriteString(token)
		prevToken = token
		atLineStart = false

		if upper == "," {
			result.WriteString("\n")

			atLineStart = true
		}

		if upper == ";" {
			result.WriteString("\n")

			atLineStart = true
		}
	}

	return strings.TrimSpace(result.String())
}

func isMajorClause(upper string) bool {
	clauses := []string{
		"SELECT", "FROM", "WHERE", "JOIN", "LEFT", "RIGHT", "INNER",
		"OUTER", "FULL", "CROSS", "ON", "AND", "OR", "ORDER", "GROUP",
		"HAVING", "LIMIT", "OFFSET", "UNION", "INTERSECT", "EXCEPT",
		"INSERT", "VALUES", "UPDATE", "SET", "CREATE", "ALTER",
		"DROP", "CASE", "WHEN", "THEN", "ELSE", "END",
	}

	return slices.Contains(clauses, upper)
}

func minifySQL(input string) string {
	input = normalizeWhitespace(input)
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

func validateSQL(input string) ValidateResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return ValidateResult{Valid: false, Error: "empty input"}
	}

	parenCount := 0

	for _, ch := range input {
		switch ch {
		case '(':
			parenCount++
		case ')':
			parenCount--
		}

		if parenCount < 0 {
			return ValidateResult{Valid: false, Error: "unbalanced parentheses: unexpected ')'"}
		}
	}

	if parenCount > 0 {
		return ValidateResult{Valid: false, Error: "unbalanced parentheses: missing ')'"}
	}

	if !checkBalancedQuotes(input, '\'') {
		return ValidateResult{Valid: false, Error: "unbalanced single quotes"}
	}

	if !checkBalancedQuotes(input, '"') {
		return ValidateResult{Valid: false, Error: "unbalanced double quotes"}
	}

	upper := strings.ToUpper(input)
	validStarts := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP", "WITH", "EXPLAIN", "SHOW", "SET", "USE", "BEGIN", "COMMIT", "ROLLBACK", "TRUNCATE", "GRANT", "REVOKE"}
	hasValidStart := false

	for _, start := range validStarts {
		if strings.HasPrefix(upper, start) {
			hasValidStart = true
			break
		}
	}

	if strings.HasPrefix(strings.TrimSpace(input), "--") {
		hasValidStart = true
	}

	if !hasValidStart {
		return ValidateResult{Valid: false, Error: "unrecognized statement type"}
	}

	return ValidateResult{Valid: true, Message: "valid SQL"}
}

func tokenizeSQL(input string) []string {
	var (
		tokens  []string
		current strings.Builder
	)

	inString := false
	stringChar := rune(0)
	inComment := false
	commentType := ""

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]

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

		if ch == '(' || ch == ')' || ch == ',' || ch == ';' || ch == '=' ||
			ch == '<' || ch == '>' || ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}

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

func normalizeWhitespace(s string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

func isKeyword(token string) bool {
	upper := strings.ToUpper(token)
	return slices.Contains(AllKeywords, upper)
}

func needsSpace(prev, curr string) bool {
	if prev == "" || curr == "" {
		return false
	}

	if prev == "(" || curr == ")" {
		return false
	}

	if curr == "(" && isFunction(prev) {
		return false
	}

	if curr == "," || curr == ";" {
		return false
	}

	if prev == "," {
		return false
	}

	if prev == "." || curr == "." {
		return false
	}

	return true
}

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
