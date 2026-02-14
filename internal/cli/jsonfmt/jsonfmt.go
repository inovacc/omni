package jsonfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/inovacc/omni/internal/cli/output"
)

// Options configures the json format command behavior
type Options struct {
	Minify     bool   // -m: minify (compact) output
	Indent     string // -i: indentation string (default "  ")
	SortKeys   bool   // -s: sort object keys
	Validate   bool   // -v: validate only, don't output
	EscapeHTML bool   // -e: escape HTML characters
	OutputFormat output.Format // output format (for validate mode)
	Tab        bool   // -t: use tabs for indentation
}

// Result represents the JSON output for validate mode
type Result struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
	File  string `json:"file,omitempty"`
}

// RunJSONFmt formats JSON input
func RunJSONFmt(w io.Writer, args []string, opts Options) error {
	// Set default indent
	if opts.Indent == "" {
		if opts.Tab {
			opts.Indent = "\t"
		} else {
			opts.Indent = "  "
		}
	}

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	if len(args) == 0 {
		return processReader(w, os.Stdin, "<stdin>", opts, jsonMode, f)
	}

	for _, file := range args {
		var r io.Reader
		if file == "-" {
			r = os.Stdin
			file = "<stdin>"
		} else {
			opened, err := os.Open(file)
			if err != nil {
				if opts.Validate && jsonMode {
					return f.Print(Result{Valid: false, Error: err.Error(), File: file})
				}

				return fmt.Errorf("json: %w", err)
			}

			defer func() { _ = opened.Close() }()

			r = opened
		}

		if err := processReader(w, r, file, opts, jsonMode, f); err != nil {
			return err
		}
	}

	return nil
}

func processReader(w io.Writer, r io.Reader, filename string, opts Options, jsonMode bool, f *output.Formatter) error {
	data, err := io.ReadAll(r)
	if err != nil {
		if opts.Validate && jsonMode {
			return f.Print(Result{Valid: false, Error: err.Error(), File: filename})
		}

		return fmt.Errorf("json: %w", err)
	}

	// Parse JSON
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		if opts.Validate {
			if jsonMode {
				return f.Print(Result{Valid: false, Error: err.Error(), File: filename})
			}

			return fmt.Errorf("%s: invalid JSON: %w", filename, err)
		}

		return fmt.Errorf("json: invalid JSON: %w", err)
	}

	// Validate only mode
	if opts.Validate {
		if jsonMode {
			return f.Print(Result{Valid: true, File: filename})
		}

		_, _ = fmt.Fprintf(w, "%s: valid JSON\n", filename)

		return nil
	}

	// Sort keys if requested
	if opts.SortKeys {
		v = sortKeys(v)
	}

	// Format output
	var output []byte
	if opts.Minify {
		output, err = json.Marshal(v)
	} else {
		output, err = json.MarshalIndent(v, "", opts.Indent)
	}

	if err != nil {
		return fmt.Errorf("json: %w", err)
	}

	// Escape HTML if not requested (json.Marshal escapes by default)
	if !opts.EscapeHTML && !opts.Minify {
		output = unescapeHTML(output)
	}

	_, _ = w.Write(output)
	_, _ = fmt.Fprintln(w)

	return nil
}

// sortKeys recursively sorts object keys
func sortKeys(v any) any {
	switch val := v.(type) {
	case map[string]any:
		sorted := make(map[string]any, len(val))
		for k, v := range val {
			sorted[k] = sortKeys(v)
		}

		return sorted
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			result[i] = sortKeys(item)
		}

		return result
	default:
		return v
	}
}

// unescapeHTML converts escaped HTML entities back
func unescapeHTML(data []byte) []byte {
	data = bytes.ReplaceAll(data, []byte("\\u003c"), []byte("<"))
	data = bytes.ReplaceAll(data, []byte("\\u003e"), []byte(">"))
	data = bytes.ReplaceAll(data, []byte("\\u0026"), []byte("&"))

	return data
}

// Beautify formats JSON with indentation
func Beautify(data []byte, indent string) ([]byte, error) {
	if indent == "" {
		indent = "  "
	}

	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	return json.MarshalIndent(v, "", indent)
}

// Minify compacts JSON by removing whitespace
func Minify(data []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	return json.Marshal(v)
}

// Validate checks if data is valid JSON
func Validate(data []byte) error {
	var v any
	return json.Unmarshal(data, &v)
}

// IsValid returns true if data is valid JSON
func IsValid(data []byte) bool {
	return json.Valid(data)
}

// SortKeys returns JSON with sorted object keys
func SortKeys(data []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	sorted := sortKeys(v)

	return json.MarshalIndent(sorted, "", "  ")
}

// Format formats JSON with custom options
func Format(data []byte, opts Options) ([]byte, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	if opts.SortKeys {
		v = sortKeys(v)
	}

	if opts.Indent == "" {
		if opts.Tab {
			opts.Indent = "\t"
		} else {
			opts.Indent = "  "
		}
	}

	if opts.Minify {
		return json.Marshal(v)
	}

	output, err := json.MarshalIndent(v, "", opts.Indent)
	if err != nil {
		return nil, err
	}

	if !opts.EscapeHTML {
		output = unescapeHTML(output)
	}

	return output, nil
}

// MustBeautify formats JSON and panics on error
func MustBeautify(data []byte) []byte {
	result, err := Beautify(data, "  ")
	if err != nil {
		panic(err)
	}

	return result
}

// MustMinify compacts JSON and panics on error
func MustMinify(data []byte) []byte {
	result, err := Minify(data)
	if err != nil {
		panic(err)
	}

	return result
}

// GetType returns the JSON type of the root element
func GetType(data []byte) string {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return "empty"
	}

	switch data[0] {
	case '{':
		return "object"
	case '[':
		return "array"
	case '"':
		return "string"
	case 't', 'f':
		return "boolean"
	case 'n':
		return "null"
	default:
		if (data[0] >= '0' && data[0] <= '9') || data[0] == '-' {
			return "number"
		}

		return "unknown"
	}
}

// Stats returns statistics about JSON data
type Stats struct {
	Type        string `json:"type"`
	Keys        int    `json:"keys,omitempty"`
	Elements    int    `json:"elements,omitempty"`
	Depth       int    `json:"depth"`
	Size        int    `json:"size"`
	MinifiedLen int    `json:"minifiedLen"`
}

// GetStats returns statistics about JSON data
func GetStats(data []byte) (*Stats, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	minified, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal for stats: %w", err)
	}

	stats := &Stats{
		Type:        GetType(data),
		Depth:       getDepth(v, 0),
		Size:        len(data),
		MinifiedLen: len(minified),
	}

	switch val := v.(type) {
	case map[string]any:
		stats.Keys = countKeys(val)
	case []any:
		stats.Elements = len(val)
	}

	return stats, nil
}

func getDepth(v any, current int) int {
	maxDepth := current

	switch val := v.(type) {
	case map[string]any:
		for _, item := range val {
			d := getDepth(item, current+1)
			if d > maxDepth {
				maxDepth = d
			}
		}
	case []any:
		for _, item := range val {
			d := getDepth(item, current+1)
			if d > maxDepth {
				maxDepth = d
			}
		}
	}

	return maxDepth
}

func countKeys(obj map[string]any) int {
	count := len(obj)
	for _, v := range obj {
		if nested, ok := v.(map[string]any); ok {
			count += countKeys(nested)
		}
	}

	return count
}

// Keys returns all keys from a JSON object (recursively)
func Keys(data []byte) ([]string, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	keys := collectKeys(v, "")
	sort.Strings(keys)

	return keys, nil
}

func collectKeys(v any, prefix string) []string {
	var keys []string

	switch val := v.(type) {
	case map[string]any:
		for k, item := range val {
			path := k
			if prefix != "" {
				path = prefix + "." + k
			}

			keys = append(keys, path)
			keys = append(keys, collectKeys(item, path)...)
		}
	case []any:
		for i, item := range val {
			path := fmt.Sprintf("%s[%d]", prefix, i)
			keys = append(keys, collectKeys(item, path)...)
		}
	}

	return keys
}
