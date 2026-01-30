package csvutil

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// ToCSVOptions configures the JSON to CSV conversion
type ToCSVOptions struct {
	Header    bool   // Include header row (default: true)
	Delimiter string // Field delimiter (default: ",")
	NoQuotes  bool   // Don't quote fields
}

// FromCSVOptions configures the CSV to JSON conversion
type FromCSVOptions struct {
	Header    bool   // First row is header (default: true)
	Delimiter string // Field delimiter (default: ",")
	Array     bool   // Output as array even for single row
}

// RunToCSV converts JSON to CSV
func RunToCSV(w io.Writer, r io.Reader, args []string, opts ToCSVOptions) error {
	input, err := getInput(args, r)
	if err != nil {
		return fmt.Errorf("csv: %w", err)
	}

	// Parse JSON
	var data any

	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return fmt.Errorf("csv: invalid JSON: %w", err)
	}

	// Convert to array if single object
	var arr []any

	switch v := data.(type) {
	case []any:
		arr = v
	case map[string]any:
		arr = []any{v}
	default:
		return fmt.Errorf("csv: JSON must be an array or object")
	}

	if len(arr) == 0 {
		return nil
	}

	// Get headers from first object
	headers, err := extractHeaders(arr)
	if err != nil {
		return fmt.Errorf("csv: %w", err)
	}

	// Create CSV writer
	csvWriter := csv.NewWriter(w)
	if opts.Delimiter != "" && len(opts.Delimiter) > 0 {
		csvWriter.Comma = rune(opts.Delimiter[0])
	}

	// Write header row
	if opts.Header {
		if err := csvWriter.Write(headers); err != nil {
			return fmt.Errorf("csv: %w", err)
		}
	}

	// Write data rows
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}

		row := make([]string, len(headers))
		for i, h := range headers {
			row[i] = getNestedValue(obj, h)
		}

		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("csv: %w", err)
		}
	}

	csvWriter.Flush()

	return csvWriter.Error()
}

// RunFromCSV converts CSV to JSON
func RunFromCSV(w io.Writer, r io.Reader, args []string, opts FromCSVOptions) error {
	input, err := getInputReader(args, r)
	if err != nil {
		return fmt.Errorf("csv: %w", err)
	}

	defer func() {
		if closer, ok := input.(io.Closer); ok {
			_ = closer.Close()
		}
	}()

	csvReader := csv.NewReader(input)
	if opts.Delimiter != "" && len(opts.Delimiter) > 0 {
		csvReader.Comma = rune(opts.Delimiter[0])
	}

	csvReader.FieldsPerRecord = -1 // Allow variable fields

	records, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("csv: %w", err)
	}

	if len(records) == 0 {
		_, _ = fmt.Fprintln(w, "[]")
		return nil
	}

	var headers []string

	var dataStart int

	if opts.Header {
		headers = records[0]
		dataStart = 1
	} else {
		// Generate column names
		if len(records) > 0 {
			headers = make([]string, len(records[0]))
			for i := range headers {
				headers[i] = fmt.Sprintf("col%d", i+1)
			}
		}

		dataStart = 0
	}

	// Convert records to JSON objects
	result := make([]map[string]any, 0, len(records)-dataStart)
	for i := dataStart; i < len(records); i++ {
		row := records[i]
		obj := make(map[string]any)

		for j, header := range headers {
			if j < len(row) {
				obj[header] = row[j]
			} else {
				obj[header] = ""
			}
		}

		result = append(result, obj)
	}

	// Output JSON
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if len(result) == 1 && !opts.Array {
		return encoder.Encode(result[0])
	}

	return encoder.Encode(result)
}

// extractHeaders gets all unique field names from the JSON array, flattening nested objects
func extractHeaders(arr []any) ([]string, error) {
	headerSet := make(map[string]struct{})

	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("array elements must be objects")
		}

		flattenKeys(obj, "", headerSet)
	}

	// Sort headers for consistent output
	headers := make([]string, 0, len(headerSet))
	for h := range headerSet {
		headers = append(headers, h)
	}

	sort.Strings(headers)

	return headers, nil
}

// flattenKeys recursively collects keys from nested objects using dot notation
func flattenKeys(obj map[string]any, prefix string, result map[string]struct{}) {
	for key, value := range obj {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			flattenKeys(v, fullKey, result)
		default:
			result[fullKey] = struct{}{}
		}
	}
}

// getNestedValue retrieves a value from a nested object using dot notation
func getNestedValue(obj map[string]any, path string) string {
	parts := strings.Split(path, ".")
	current := any(obj)

	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}

		current = m[part]
	}

	return formatValue(current)
}

// formatValue converts a value to string for CSV output
func formatValue(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case float64:
		// Format without trailing zeros for integers
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}

		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case []any, map[string]any:
		// Serialize complex values as JSON
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}

		return string(b)
	default:
		return fmt.Sprintf("%v", val)
	}
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

// getInputReader returns a reader for args (file) or stdin
func getInputReader(args []string, r io.Reader) (io.Reader, error) {
	if len(args) > 0 {
		// Check if it's a file
		if _, err := os.Stat(args[0]); err == nil {
			f, err := os.Open(args[0])
			if err != nil {
				return nil, err
			}

			return f, nil
		}

		// Treat as literal CSV string
		return strings.NewReader(strings.Join(args, "\n")), nil
	}

	return r, nil
}
