// Package htmlfmt provides HTML formatting, minification, and validation.
package htmlfmt

import (
	"bytes"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

// Options configures the HTML formatter.
type Options struct {
	Indent    string // Indentation (default: "  ")
	SortAttrs bool   // Sort attributes alphabetically
}

// Option is a functional option for Format.
type Option func(*Options)

// WithIndent sets the indentation string.
func WithIndent(indent string) Option {
	return func(o *Options) { o.Indent = indent }
}

// WithSortAttrs enables alphabetical attribute sorting.
func WithSortAttrs() Option {
	return func(o *Options) { o.SortAttrs = true }
}

// ValidateResult represents HTML validation output.
type ValidateResult struct {
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Format formats HTML input with proper indentation.
func Format(input string, opts ...Option) (string, error) {
	o := Options{Indent: "  "}
	for _, opt := range opts {
		opt(&o)
	}

	return formatHTML(input, o)
}

// Minify removes unnecessary whitespace from HTML.
func Minify(input string) (string, error) {
	return minifyHTML(input)
}

// Validate performs basic HTML syntax validation.
func Validate(input string) ValidateResult {
	return validateHTML(input)
}

// IsSelfClosing returns true if the tag is a self-closing HTML tag.
func IsSelfClosing(tag string) bool {
	return isSelfClosing(tag)
}

// CollapseWhitespace collapses multiple whitespace characters into single spaces.
func CollapseWhitespace(s string) string {
	return collapseWhitespace(s)
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

		selfClosing := isSelfClosing(n.Data)

		buf.WriteString(indent)
		buf.WriteString("<")
		buf.WriteString(n.Data)

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

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			formatNode(buf, c, depth+1, opts)
		}

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
	if n.FirstChild == nil || n.FirstChild != n.LastChild {
		return false
	}

	if n.FirstChild.Type != html.TextNode {
		return false
	}

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
