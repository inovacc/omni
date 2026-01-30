package buf

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// RunFormat formats proto files
func RunFormat(w io.Writer, dir string, opts FormatOptions) error {
	// Find proto files
	files, err := FindProtoFiles(dir, nil)
	if err != nil {
		return fmt.Errorf("buf: %w", err)
	}

	if len(files) == 0 {
		_, _ = fmt.Fprintln(w, "No proto files found")

		return nil
	}

	var hasUnformatted bool

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("buf: failed to read %s: %w", file, err)
		}

		formatted := FormatProto(string(content))

		if string(content) == formatted {
			continue
		}

		hasUnformatted = true

		if opts.Diff {
			_, _ = fmt.Fprintf(w, "--- %s\n+++ %s\n", file, file)

			printDiff(w, string(content), formatted)
		}

		if opts.Write {
			if err := os.WriteFile(file, []byte(formatted), 0644); err != nil {
				return fmt.Errorf("buf: failed to write %s: %w", file, err)
			}

			if !opts.Diff {
				_, _ = fmt.Fprintf(w, "Formatted %s\n", file)
			}
		}
	}

	if opts.ExitCode && hasUnformatted {
		return fmt.Errorf("found unformatted files")
	}

	return nil
}

// FormatProto formats a proto file content
func FormatProto(source string) string {
	lexer := NewLexer(source)
	tokens := lexer.Tokenize()

	return formatTokens(tokens)
}

func formatTokens(tokens []Token) string {
	var result strings.Builder

	indent := 0
	atLineStart := true
	lastTokenType := TokenEOF
	lastValue := ""

	for i, token := range tokens {
		// Skip EOF
		if token.Type == TokenEOF {
			break
		}

		// Handle indentation changes
		if token.Type == TokenSymbol {
			if token.Value == "{" {
				// Write opening brace
				if !atLineStart && needsSpaceBefore(lastTokenType, lastValue, token) {
					result.WriteString(" ")
				}

				result.WriteString("{")
				result.WriteString("\n")

				indent++
				atLineStart = true
				lastTokenType = token.Type
				lastValue = token.Value

				continue
			}

			if token.Value == "}" {
				// Dedent before closing brace
				indent--

				if indent < 0 {
					indent = 0
				}

				if !atLineStart {
					result.WriteString("\n")
				}

				writeIndent(&result, indent)
				result.WriteString("}")

				// Check if next non-whitespace token is newline or EOF
				nextNonWS := findNextNonWhitespace(tokens, i+1)
				if nextNonWS != nil && nextNonWS.Type != TokenNewline && nextNonWS.Type != TokenEOF {
					result.WriteString("\n")
				}

				atLineStart = false
				lastTokenType = token.Type
				lastValue = token.Value

				continue
			}
		}

		// Handle newlines
		if token.Type == TokenNewline {
			if !atLineStart {
				result.WriteString("\n")

				atLineStart = true
			}

			lastTokenType = token.Type
			lastValue = token.Value

			continue
		}

		// Skip extra whitespace
		if token.Type == TokenWhitespace {
			lastTokenType = token.Type
			lastValue = token.Value

			continue
		}

		// Handle comments
		if token.Type == TokenComment {
			if atLineStart {
				writeIndent(&result, indent)
			} else {
				result.WriteString(" ")
			}

			result.WriteString(token.Value)
			result.WriteString("\n")

			atLineStart = true
			lastTokenType = token.Type
			lastValue = token.Value

			continue
		}

		// Write indentation at line start
		if atLineStart {
			writeIndent(&result, indent)

			atLineStart = false
		} else if needsSpaceBefore(lastTokenType, lastValue, token) {
			result.WriteString(" ")
		}

		// Handle semicolons - add newline after
		if token.Type == TokenSymbol && token.Value == ";" {
			result.WriteString(";")
			result.WriteString("\n")

			atLineStart = true
			lastTokenType = token.Type
			lastValue = token.Value

			continue
		}

		// Write token value
		if token.Type == TokenString {
			result.WriteString("\"")
			result.WriteString(token.Value)
			result.WriteString("\"")
		} else {
			result.WriteString(token.Value)
		}

		lastTokenType = token.Type
		lastValue = token.Value
	}

	// Ensure file ends with newline
	formatted := result.String()
	if !strings.HasSuffix(formatted, "\n") {
		formatted += "\n"
	}

	// Clean up multiple blank lines
	formatted = cleanupBlankLines(formatted)

	return formatted
}

func writeIndent(w *strings.Builder, indent int) {
	for range indent {
		w.WriteString("  ")
	}
}

func needsSpaceBefore(lastType TokenType, lastValue string, current Token) bool {
	// No space after opening symbols
	if lastValue == "(" || lastValue == "[" || lastValue == "<" {
		return false
	}

	// No space before closing symbols
	if current.Value == ")" || current.Value == "]" || current.Value == ">" {
		return false
	}

	// No space before/after dots
	if lastValue == "." || current.Value == "." {
		return false
	}

	// No space before comma, semicolon
	if current.Value == "," || current.Value == ";" {
		return false
	}

	// No space after comma in declarations
	if lastValue == "," {
		return true
	}

	// Space before = and after =
	if current.Value == "=" || lastValue == "=" {
		return true
	}

	// Space between keywords and identifiers
	if lastType == TokenKeyword || lastType == TokenIdent {
		if current.Type == TokenKeyword || current.Type == TokenIdent ||
			current.Type == TokenNumber || current.Type == TokenString {
			return true
		}
	}

	// Space between number and identifier
	if lastType == TokenNumber && current.Type == TokenIdent {
		return true
	}

	return false
}

func findNextNonWhitespace(tokens []Token, start int) *Token {
	for i := start; i < len(tokens); i++ {
		if tokens[i].Type != TokenWhitespace {
			return &tokens[i]
		}
	}

	return nil
}

func cleanupBlankLines(s string) string {
	lines := strings.Split(s, "\n")

	var result []string

	lastWasBlank := false

	for _, line := range lines {
		isBlank := strings.TrimSpace(line) == ""

		if isBlank && lastWasBlank {
			continue
		}

		result = append(result, line)
		lastWasBlank = isBlank
	}

	return strings.Join(result, "\n")
}

func printDiff(w io.Writer, original, formatted string) {
	origLines := strings.Split(original, "\n")
	fmtLines := strings.Split(formatted, "\n")

	// Simple line-by-line diff
	maxLen := max(len(fmtLines), len(origLines))

	for i := range maxLen {
		origLine := ""
		fmtLine := ""

		if i < len(origLines) {
			origLine = origLines[i]
		}

		if i < len(fmtLines) {
			fmtLine = fmtLines[i]
		}

		if origLine != fmtLine {
			if origLine != "" {
				_, _ = fmt.Fprintf(w, "-%s\n", origLine)
			}

			if fmtLine != "" {
				_, _ = fmt.Fprintf(w, "+%s\n", fmtLine)
			}
		} else if origLine != "" {
			_, _ = fmt.Fprintf(w, " %s\n", origLine)
		}
	}
}
