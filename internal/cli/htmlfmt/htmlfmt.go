package htmlfmt

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

// Options configures the HTML formatter
type Options struct {
	Indent    string // Indentation (default: "  ")
	Minify    bool   // Minify output
	SortAttrs bool   // Sort attributes alphabetically
}

// ValidateOptions configures HTML validation
type ValidateOptions struct {
	JSON bool // Output as JSON
}

// ValidateResult represents validation output
type ValidateResult struct {
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Run formats HTML input
func Run(w io.Writer, r io.Reader, args []string, opts Options) error {
	input, err := getInput(args, r)
	if err != nil {
		return fmt.Errorf("html: %w", err)
	}

	// Set defaults
	if opts.Indent == "" {
		opts.Indent = "  "
	}

	var output string
	if opts.Minify {
		output, err = minifyHTML(input)
	} else {
		output, err = formatHTML(input, opts)
	}

	if err != nil {
		return fmt.Errorf("html: %w", err)
	}

	_, _ = fmt.Fprintln(w, output)

	return nil
}

// RunMinify minifies HTML
func RunMinify(w io.Writer, r io.Reader, args []string, opts Options) error {
	opts.Minify = true
	return Run(w, r, args, opts)
}

// RunValidate validates HTML syntax
func RunValidate(w io.Writer, r io.Reader, args []string, opts ValidateOptions) error {
	input, err := getInput(args, r)
	if err != nil {
		return fmt.Errorf("html: %w", err)
	}

	result := validateHTML(input)

	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(result)
	}

	if result.Valid {
		_, _ = fmt.Fprintln(w, result.Message)
	} else {
		_, _ = fmt.Fprintf(w, "invalid HTML: %s\n", result.Error)
		return fmt.Errorf("validation failed")
	}

	return nil
}

// formatHTML formats HTML with proper indentation
func formatHTML(input string, opts Options) (string, error) {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	formatNode(&buf, doc, 0, opts)

	return strings.TrimSpace(buf.String()), nil
}

// formatNode recursively formats an HTML node
func formatNode(buf *bytes.Buffer, n *html.Node, depth int, opts Options) {
	switch n.Type {
	case html.ErrorNode, html.RawNode:
		// Skip error and raw nodes
		return

	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			formatNode(buf, c, depth, opts)
		}

	case html.DoctypeNode:
		buf.WriteString("<!DOCTYPE ")
		buf.WriteString(n.Data)
		buf.WriteString(">\n")

	case html.ElementNode:
		indent := strings.Repeat(opts.Indent, depth)

		// Self-closing tags
		selfClosing := isSelfClosing(n.Data)

		// Write opening tag
		buf.WriteString(indent)
		buf.WriteString("<")
		buf.WriteString(n.Data)

		// Write attributes
		attrs := n.Attr
		if opts.SortAttrs {
			attrs = sortAttributes(attrs)
		}

		for _, attr := range attrs {
			buf.WriteString(" ")
			buf.WriteString(attr.Key)
			buf.WriteString("=\"")
			buf.WriteString(html.EscapeString(attr.Val))
			buf.WriteString("\"")
		}

		if selfClosing && n.FirstChild == nil {
			buf.WriteString(" />\n")
			return
		}

		buf.WriteString(">")

		// Check if content is inline (single text node without newlines)
		if isInlineContent(n) {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					buf.WriteString(strings.TrimSpace(c.Data))
				}
			}

			buf.WriteString("</")
			buf.WriteString(n.Data)
			buf.WriteString(">\n")

			return
		}

		buf.WriteString("\n")

		// Write children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			formatNode(buf, c, depth+1, opts)
		}

		// Write closing tag
		buf.WriteString(indent)
		buf.WriteString("</")
		buf.WriteString(n.Data)
		buf.WriteString(">\n")

	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" {
			indent := strings.Repeat(opts.Indent, depth)
			buf.WriteString(indent)
			buf.WriteString(text)
			buf.WriteString("\n")
		}

	case html.CommentNode:
		indent := strings.Repeat(opts.Indent, depth)
		buf.WriteString(indent)
		buf.WriteString("<!--")
		buf.WriteString(n.Data)
		buf.WriteString("-->\n")
	}
}

// minifyHTML removes unnecessary whitespace from HTML
func minifyHTML(input string) (string, error) {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	minifyNode(&buf, doc)

	return strings.TrimSpace(buf.String()), nil
}

// minifyNode recursively minifies an HTML node
func minifyNode(buf *bytes.Buffer, n *html.Node) {
	switch n.Type {
	case html.ErrorNode, html.RawNode:
		// Skip error and raw nodes
		return

	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			minifyNode(buf, c)
		}

	case html.DoctypeNode:
		buf.WriteString("<!DOCTYPE ")
		buf.WriteString(n.Data)
		buf.WriteString(">")

	case html.ElementNode:
		selfClosing := isSelfClosing(n.Data)

		buf.WriteString("<")
		buf.WriteString(n.Data)

		for _, attr := range n.Attr {
			buf.WriteString(" ")
			buf.WriteString(attr.Key)
			buf.WriteString("=\"")
			buf.WriteString(html.EscapeString(attr.Val))
			buf.WriteString("\"")
		}

		if selfClosing && n.FirstChild == nil {
			buf.WriteString("/>")
			return
		}

		buf.WriteString(">")

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			minifyNode(buf, c)
		}

		buf.WriteString("</")
		buf.WriteString(n.Data)
		buf.WriteString(">")

	case html.TextNode:
		// Collapse whitespace but preserve at least one space between words
		text := collapseWhitespace(n.Data)
		buf.WriteString(text)

	case html.CommentNode:
		// Skip comments in minified output
	}
}

// validateHTML validates HTML syntax
func validateHTML(input string) ValidateResult {
	input = strings.TrimSpace(input)
	if input == "" {
		return ValidateResult{
			Valid: false,
			Error: "empty input",
		}
	}

	_, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return ValidateResult{
			Valid: false,
			Error: err.Error(),
		}
	}

	return ValidateResult{
		Valid:   true,
		Message: "valid HTML",
	}
}

// isSelfClosing returns true if the tag is self-closing
func isSelfClosing(tag string) bool {
	selfClosingTags := map[string]bool{
		"area": true, "base": true, "br": true, "col": true,
		"embed": true, "hr": true, "img": true, "input": true,
		"link": true, "meta": true, "param": true, "source": true,
		"track": true, "wbr": true,
	}

	return selfClosingTags[strings.ToLower(tag)]
}

// isInlineContent checks if an element has only inline text content
func isInlineContent(n *html.Node) bool {
	// Only one child that is text
	if n.FirstChild == nil || n.FirstChild != n.LastChild {
		return false
	}

	if n.FirstChild.Type != html.TextNode {
		return false
	}

	// No newlines in the text
	return !strings.Contains(n.FirstChild.Data, "\n")
}

// sortAttributes sorts attributes alphabetically
func sortAttributes(attrs []html.Attribute) []html.Attribute {
	sorted := make([]html.Attribute, len(attrs))
	copy(sorted, attrs)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})

	return sorted
}

// collapseWhitespace collapses multiple whitespace characters into single spaces
func collapseWhitespace(s string) string {
	var result strings.Builder

	inWhitespace := false

	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !inWhitespace {
				result.WriteRune(' ')

				inWhitespace = true
			}
		} else {
			result.WriteRune(r)

			inWhitespace = false
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
