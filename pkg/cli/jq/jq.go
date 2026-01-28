package jq

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// JqOptions configures the jq command behavior
type JqOptions struct {
	Raw        bool // -r: output raw strings (no quotes)
	Compact    bool // -c: compact output (no pretty print)
	Slurp      bool // -s: read entire input into array
	NullInput  bool // -n: don't read any input
	Tab        bool // --tab: use tabs for indentation
	Sort       bool // -S: sort object keys
	Color      bool // -C: colorize output (not implemented)
	Monochrome bool // -M: monochrome output
}

// RunJq executes jq-like JSON processing
func RunJq(w io.Writer, args []string, opts JqOptions) error {
	filter := "."

	var files []string

	if len(args) > 0 {
		filter = args[0]
		files = args[1:]
	}

	var inputs []any

	if opts.NullInput {
		inputs = []any{nil}
	} else if len(files) == 0 {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("jq: %w", err)
		}

		if opts.Slurp {
			var items []any

			dec := json.NewDecoder(strings.NewReader(string(data)))

			for {
				var v any
				if err := dec.Decode(&v); err == io.EOF {
					break
				} else if err != nil {
					return fmt.Errorf("jq: parse error: %w", err)
				}

				items = append(items, v)
			}

			inputs = []any{items}
		} else {
			var v any
			if err := json.Unmarshal(data, &v); err != nil {
				return fmt.Errorf("jq: parse error: %w", err)
			}

			inputs = []any{v}
		}
	} else {
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("jq: %w", err)
			}

			var v any
			if err := json.Unmarshal(data, &v); err != nil {
				return fmt.Errorf("jq: %s: parse error: %w", file, err)
			}

			inputs = append(inputs, v)
		}

		if opts.Slurp {
			inputs = []any{inputs}
		}
	}

	for _, input := range inputs {
		results, err := ApplyJqFilter(input, filter)
		if err != nil {
			return fmt.Errorf("jq: %w", err)
		}

		for _, result := range results {
			if err := outputJqResult(w, result, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func ApplyJqFilter(input any, filter string) ([]any, error) {
	filter = strings.TrimSpace(filter)

	// Handle identity
	if filter == "." {
		return []any{input}, nil
	}

	// Handle keys
	if filter == "keys" {
		return jqKeys(input)
	}

	// Handle length
	if filter == "length" {
		return jqLength(input)
	}

	// Handle type
	if filter == "type" {
		return jqType(input)
	}

	// Handle array iteration .[]
	if filter == ".[]" {
		return jqIterateArray(input)
	}

	// Handle pipe (check before field access to support .field | filter)
	if strings.Contains(filter, "|") {
		return jqPipe(input, filter)
	}

	// Handle array index .[n]
	if strings.HasPrefix(filter, ".[") && strings.HasSuffix(filter, "]") {
		indexStr := filter[2 : len(filter)-1]
		if idx, err := strconv.Atoi(indexStr); err == nil {
			return jqArrayIndex(input, idx)
		}
		// Could be a string key like .["key"]
		if strings.HasPrefix(indexStr, "\"") && strings.HasSuffix(indexStr, "\"") {
			key := indexStr[1 : len(indexStr)-1]
			return jqObjectKey(input, key)
		}
	}

	// Handle field access .field or .field.subfield
	if strings.HasPrefix(filter, ".") {
		return jqFieldAccess(input, filter[1:])
	}

	return nil, fmt.Errorf("unsupported filter: %s", filter)
}

func jqKeys(input any) ([]any, error) {
	switch v := input.(type) {
	case map[string]any:
		keys := make([]any, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}

		return []any{keys}, nil
	case []any:
		keys := make([]any, 0, len(v))
		for i := range v {
			keys = append(keys, i)
		}

		return []any{keys}, nil
	default:
		return nil, fmt.Errorf("cannot get keys of %T", input)
	}
}

func jqLength(input any) ([]any, error) {
	switch v := input.(type) {
	case map[string]any:
		return []any{float64(len(v))}, nil
	case []any:
		return []any{float64(len(v))}, nil
	case string:
		return []any{float64(len(v))}, nil
	case nil:
		return []any{float64(0)}, nil
	default:
		return nil, fmt.Errorf("cannot get length of %T", input)
	}
}

func jqType(input any) ([]any, error) {
	switch input.(type) {
	case nil:
		return []any{"null"}, nil
	case bool:
		return []any{"boolean"}, nil
	case float64:
		return []any{"number"}, nil
	case string:
		return []any{"string"}, nil
	case []any:
		return []any{"array"}, nil
	case map[string]any:
		return []any{"object"}, nil
	default:
		return []any{"unknown"}, nil
	}
}

func jqIterateArray(input any) ([]any, error) {
	switch v := input.(type) {
	case []any:
		return v, nil
	case map[string]any:
		var values []any
		for _, val := range v {
			values = append(values, val)
		}

		return values, nil
	default:
		return nil, fmt.Errorf("cannot iterate over %T", input)
	}
}

func jqArrayIndex(input any, idx int) ([]any, error) {
	arr, ok := input.([]any)
	if !ok {
		return nil, fmt.Errorf("cannot index %T with number", input)
	}

	if idx < 0 {
		idx = len(arr) + idx
	}

	if idx < 0 || idx >= len(arr) {
		return []any{nil}, nil
	}

	return []any{arr[idx]}, nil
}

func jqObjectKey(input any, key string) ([]any, error) {
	obj, ok := input.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot index %T with string", input)
	}

	return []any{obj[key]}, nil
}

func jqFieldAccess(input any, path string) ([]any, error) {
	// Handle nested paths like field.subfield
	parts := strings.Split(path, ".")
	current := input

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Check for array access in field like field[0]
		if idx := strings.Index(part, "["); idx != -1 {
			fieldName := part[:idx]
			indexPart := part[idx:]

			// Access the field first
			if fieldName != "" {
				obj, ok := current.(map[string]any)
				if !ok {
					return []any{nil}, nil
				}

				current = obj[fieldName]
			}

			// Then apply array access
			if strings.HasSuffix(indexPart, "]") {
				idxStr := indexPart[1 : len(indexPart)-1]
				if i, err := strconv.Atoi(idxStr); err == nil {
					arr, ok := current.([]any)
					if !ok {
						return []any{nil}, nil
					}

					if i < 0 {
						i = len(arr) + i
					}

					if i < 0 || i >= len(arr) {
						return []any{nil}, nil
					}

					current = arr[i]
				}
			}

			continue
		}

		obj, ok := current.(map[string]any)
		if !ok {
			return []any{nil}, nil
		}

		current = obj[part]
	}

	return []any{current}, nil
}

func jqPipe(input any, filter string) ([]any, error) {
	parts := strings.SplitN(filter, "|", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid pipe")
	}

	// Apply first filter
	intermediate, err := ApplyJqFilter(input, strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, err
	}

	// Apply second filter to each result
	var results []any

	for _, item := range intermediate {
		r, err := ApplyJqFilter(item, strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}

		results = append(results, r...)
	}

	return results, nil
}

func outputJqResult(w io.Writer, result any, opts JqOptions) error {
	// Raw string output
	if opts.Raw {
		if s, ok := result.(string); ok {
			_, _ = fmt.Fprintln(w, s)
			return nil
		}
	}

	// JSON output
	encoder := json.NewEncoder(w)

	if !opts.Compact {
		if opts.Tab {
			encoder.SetIndent("", "\t")
		} else {
			encoder.SetIndent("", "  ")
		}
	}

	return encoder.Encode(result)
}
