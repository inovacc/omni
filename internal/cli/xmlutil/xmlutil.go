package xmlutil

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// ToXMLOptions configures JSON to XML conversion
type ToXMLOptions struct {
	Root       string // Root element name (default: "root")
	Indent     string // Indentation (default: "  ")
	ItemTag    string // Tag for array items (default: "item")
	AttrPrefix string // Prefix for attributes (default: "-")
}

// FromXMLOptions configures XML to JSON conversion
type FromXMLOptions struct {
	AttrPrefix string // Prefix for attributes in JSON (default: "-")
	TextKey    string // Key for text content (default: "#text")
}

// RunToXML converts JSON to XML
func RunToXML(w io.Writer, r io.Reader, args []string, opts ToXMLOptions) error {
	input, err := getInput(args, r)
	if err != nil {
		return fmt.Errorf("xml: %w", err)
	}

	// Set defaults
	if opts.Root == "" {
		opts.Root = "root"
	}

	if opts.Indent == "" {
		opts.Indent = "  "
	}

	if opts.ItemTag == "" {
		opts.ItemTag = "item"
	}

	if opts.AttrPrefix == "" {
		opts.AttrPrefix = "-"
	}

	// Parse JSON
	var data any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return fmt.Errorf("xml: invalid JSON: %w", err)
	}

	// Convert to XML
	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	if err := writeXMLElement(&buf, opts.Root, data, 0, opts); err != nil {
		return fmt.Errorf("xml: %w", err)
	}

	_, _ = fmt.Fprint(w, buf.String())

	return nil
}

// RunFromXML converts XML to JSON
func RunFromXML(w io.Writer, r io.Reader, args []string, opts FromXMLOptions) error {
	input, err := getInput(args, r)
	if err != nil {
		return fmt.Errorf("json: %w", err)
	}

	// Set defaults
	if opts.AttrPrefix == "" {
		opts.AttrPrefix = "-"
	}

	if opts.TextKey == "" {
		opts.TextKey = "#text"
	}

	// Parse XML
	result, err := parseXML(strings.NewReader(input), opts)
	if err != nil {
		return fmt.Errorf("json: invalid XML: %w", err)
	}

	// Output JSON
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	return encoder.Encode(result)
}

// writeXMLElement writes a JSON value as XML element
func writeXMLElement(buf *bytes.Buffer, name string, value any, depth int, opts ToXMLOptions) error {
	indent := strings.Repeat(opts.Indent, depth)

	switch v := value.(type) {
	case nil:
		buf.WriteString(indent)
		buf.WriteString("<")
		buf.WriteString(name)
		buf.WriteString("/>\n")

	case bool, float64, string:
		buf.WriteString(indent)
		buf.WriteString("<")
		buf.WriteString(name)
		buf.WriteString(">")
		buf.WriteString(escapeXML(fmt.Sprintf("%v", v)))
		buf.WriteString("</")
		buf.WriteString(name)
		buf.WriteString(">\n")

	case []any:
		// Arrays are wrapped in their parent element, items use ItemTag
		buf.WriteString(indent)
		buf.WriteString("<")
		buf.WriteString(name)
		buf.WriteString(">\n")

		for _, item := range v {
			if err := writeXMLElement(buf, opts.ItemTag, item, depth+1, opts); err != nil {
				return err
			}
		}

		buf.WriteString(indent)
		buf.WriteString("</")
		buf.WriteString(name)
		buf.WriteString(">\n")

	case map[string]any:
		attrs, children := separateAttrsAndChildren(v, opts.AttrPrefix)

		buf.WriteString(indent)
		buf.WriteString("<")
		buf.WriteString(name)

		// Write attributes
		if len(attrs) > 0 {
			keys := sortedKeys(attrs)
			for _, k := range keys {
				buf.WriteString(" ")
				buf.WriteString(k)
				buf.WriteString("=\"")
				buf.WriteString(escapeXML(fmt.Sprintf("%v", attrs[k])))
				buf.WriteString("\"")
			}
		}

		if len(children) == 0 {
			buf.WriteString("/>\n")
		} else {
			buf.WriteString(">\n")

			keys := sortedKeys(children)
			for _, k := range keys {
				if err := writeXMLElement(buf, k, children[k], depth+1, opts); err != nil {
					return err
				}
			}

			buf.WriteString(indent)
			buf.WriteString("</")
			buf.WriteString(name)
			buf.WriteString(">\n")
		}

	default:
		return fmt.Errorf("unsupported type: %T", value)
	}

	return nil
}

// separateAttrsAndChildren separates attributes (prefixed keys) from child elements
func separateAttrsAndChildren(obj map[string]any, prefix string) (map[string]any, map[string]any) {
	attrs := make(map[string]any)
	children := make(map[string]any)

	for k, v := range obj {
		if after, ok := strings.CutPrefix(k, prefix); ok {
			attrs[after] = v
		} else {
			children[k] = v
		}
	}

	return attrs, children
}

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")

	return s
}

// sortedKeys returns sorted keys from a map
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// parseXML parses XML into a JSON-compatible structure
func parseXML(r io.Reader, opts FromXMLOptions) (any, error) {
	decoder := xml.NewDecoder(r)

	var (
		root  *xmlNode
		stack []*xmlNode
	)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			node := &xmlNode{
				name:     t.Name.Local,
				attrs:    make(map[string]string),
				children: make([]*xmlNode, 0),
			}

			for _, attr := range t.Attr {
				node.attrs[attr.Name.Local] = attr.Value
			}

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, node)
			} else {
				root = node
			}

			stack = append(stack, node)

		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" && len(stack) > 0 {
				current := stack[len(stack)-1]
				current.text += text
			}
		}
	}

	if root == nil {
		return nil, fmt.Errorf("no root element")
	}

	return root.toJSON(opts), nil
}

// xmlNode represents a parsed XML node
type xmlNode struct {
	name     string
	attrs    map[string]string
	children []*xmlNode
	text     string
}

// toJSON converts an XML node to JSON-compatible map
func (n *xmlNode) toJSON(opts FromXMLOptions) map[string]any {
	result := make(map[string]any)
	content := make(map[string]any)

	// Add attributes with prefix
	for k, v := range n.attrs {
		content[opts.AttrPrefix+k] = v
	}

	// Add text content
	if n.text != "" {
		if len(n.children) == 0 && len(n.attrs) == 0 {
			// Simple text node - just use the text value
			result[n.name] = n.text
			return result
		}

		content[opts.TextKey] = n.text
	}

	// Group children by name
	childGroups := make(map[string][]*xmlNode)
	for _, child := range n.children {
		childGroups[child.name] = append(childGroups[child.name], child)
	}

	// Add children
	for name, children := range childGroups {
		if len(children) == 1 {
			childJSON := children[0].toJSON(opts)
			content[name] = childJSON[children[0].name]
		} else {
			arr := make([]any, len(children))
			for i, child := range children {
				childJSON := child.toJSON(opts)
				arr[i] = childJSON[child.name]
			}

			content[name] = arr
		}
	}

	if len(content) == 0 {
		result[n.name] = nil
	} else {
		result[n.name] = content
	}

	return result
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
