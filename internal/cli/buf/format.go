package buf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bufbuild/protocompile/ast"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
)

// RunFormat formats proto files.
func RunFormat(w io.Writer, dir string, opts FormatOptions) error {
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

// FormatProto formats a proto file's source code using AST-based formatting.
// It parses the source into an AST and rewrites it with canonical formatting:
// - 2-space indentation
// - Normalized blank lines (max 1 between declarations)
// - Preserved comments
// Falls back to cleanupBlankLines if parsing fails.
func FormatProto(source string) string {
	handler := reporter.NewHandler(nil)
	fileNode, err := parser.Parse("input.proto", strings.NewReader(source), handler)
	if err != nil || fileNode == nil {
		return cleanupBlankLines(source)
	}

	var buf bytes.Buffer
	if err := formatAST(&buf, fileNode); err != nil {
		return cleanupBlankLines(source)
	}

	result := buf.String()
	if result == "" {
		return cleanupBlankLines(source)
	}

	// Clean up trailing whitespace on each line
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// formatAST walks all tokens in the AST and writes them with canonical
// whitespace. Comments are preserved but whitespace is rewritten.
func formatAST(w io.Writer, file *ast.FileNode) error {
	f := &protoFormatter{
		w:    w,
		file: file,
	}

	tokens := file.Tokens()
	tok, ok := tokens.First()
	for ok {
		info := file.TokenInfo(tok)

		// Write leading comments
		f.emitComments(info.LeadingComments(), true)

		// Determine and write whitespace before this token
		rawText := info.RawText()
		f.emitPreTokenWhitespace(info, rawText)

		// Track depth changes from opening tokens
		if rawText == "{" || rawText == "[" {
			f.depth++
		}

		// Write the token
		f.write(rawText)
		f.lastToken = rawText

		// Write trailing comments
		f.emitComments(info.TrailingComments(), false)

		tok, ok = tokens.Next(tok)
	}

	// Ensure file ends with exactly one newline
	if f.lastChar != '\n' {
		f.write("\n")
	}

	return f.err
}

type protoFormatter struct {
	w         io.Writer
	file      *ast.FileNode
	depth     int
	lastChar  byte
	lastToken string
	atLine    bool // true if we're at the start of a line (after \n)
	err       error
}

func (f *protoFormatter) write(s string) {
	if f.err != nil || s == "" {
		return
	}
	_, f.err = io.WriteString(f.w, s)
	if len(s) > 0 {
		f.lastChar = s[len(s)-1]
		f.atLine = f.lastChar == '\n'
	}
}

func (f *protoFormatter) newline() {
	f.write("\n")
}

func (f *protoFormatter) indent() {
	f.write(strings.Repeat("  ", f.depth))
}

func (f *protoFormatter) emitComments(comments ast.Comments, leading bool) {
	for i := 0; i < comments.Len(); i++ {
		c := comments.Index(i)

		if leading {
			// Preserve blank lines between leading comments
			ws := c.LeadingWhitespace()
			nlCount := strings.Count(ws, "\n")
			if nlCount > 1 && f.lastChar != 0 {
				f.newline() // extra blank line
			}
			if f.atLine || f.lastChar == 0 {
				f.indent()
			} else {
				f.write(" ")
			}
		} else {
			// Trailing comment — single space before
			if f.lastChar != ' ' && f.lastChar != '\n' {
				f.write(" ")
			}
		}

		f.write(c.RawText())

		// Line comments end with a newline
		if strings.HasPrefix(c.RawText(), "//") {
			f.newline()
		}
	}
}

func (f *protoFormatter) emitPreTokenWhitespace(info ast.NodeInfo, rawText string) {
	// Check how many newlines the original source had before this token
	origWS := info.LeadingWhitespace()
	origNewlines := strings.Count(origWS, "\n")

	// Also check leading comments for newline count
	lc := info.LeadingComments()
	if lc.Len() > 0 {
		firstWS := lc.Index(0).LeadingWhitespace()
		if n := strings.Count(firstWS, "\n"); n > origNewlines {
			origNewlines = n
		}
	}

	// Closing tokens: decrease depth first
	if rawText == "}" || rawText == "]" {
		f.depth--
		if f.depth < 0 {
			f.depth = 0
		}
	}

	// Determine if this token should start on a new line
	switch {
	case f.lastChar == 0:
		// Very first token — no whitespace needed

	case rawText == "}" || rawText == "]":
		// Empty body — keep on same line
		if f.lastChar == '{' || f.lastChar == '[' {
			// no whitespace
		} else {
			// Closing brace/bracket on its own line
			if !f.atLine {
				f.newline()
			}
			if origNewlines > 1 {
				f.newline() // preserve blank line before closing brace
			}
			f.indent()
		}

	case rawText == ";", rawText == ",", rawText == ")", rawText == ">":
		// No whitespace before these

	case isBlockOpener(rawText) && f.lastChar != 0:
		// Declarations always start on a new line
		if !f.atLine {
			f.newline()
		}
		// Preserve blank lines between declarations
		if origNewlines > 1 {
			f.newline()
		}
		f.indent()

	case origNewlines > 0:
		// Original source had a newline here — respect it
		if !f.atLine {
			f.newline()
		}
		if origNewlines > 1 {
			f.newline()
		}
		f.indent()

	case f.lastChar == '{' || f.lastChar == '[':
		// After opening brace — newline + indent (unless empty body)
		if rawText != "}" && rawText != "]" {
			f.newline()
			f.indent()
		}

	case f.lastChar == ';':
		// After semicolon — newline
		f.newline()
		f.indent()

	case rawText == "{" || rawText == "=":
		// Space before these
		f.write(" ")

	case rawText == "(":
		// Space before ( after "returns", no space otherwise
		if f.lastChar != '(' && f.lastToken == "returns" {
			f.write(" ")
		}

	case f.lastChar == '(':
		// No space after opening paren

	case f.lastChar == '<' || rawText == "<":
		// No space inside angle brackets

	case f.lastChar == '.', rawText == ".":
		// No space around dots

	default:
		// Default: single space between tokens
		if !f.atLine && f.lastChar != ' ' && f.lastChar != '\t' {
			f.write(" ")
		}
	}
}

// isBlockOpener returns true if this token typically starts a declaration.
func isBlockOpener(text string) bool {
	switch text {
	case "syntax", "edition", "package", "import", "option",
		"message", "enum", "service", "extend", "oneof",
		"rpc", "reserved", "extensions":
		return true
	}
	return false
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
