package jsonutil

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Query applies a jq-like filter to raw JSON data and returns the result as JSON bytes.
func Query(data []byte, filter string) ([]byte, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("jsonutil: parse error: %w", err)
	}

	results, err := ApplyFilter(v, filter)
	if err != nil {
		return nil, fmt.Errorf("jsonutil: %w", err)
	}

	if len(results) == 1 {
		return json.Marshal(results[0])
	}

	return json.Marshal(results)
}

// QueryReader applies a jq-like filter to JSON from an io.Reader.
func QueryReader(r io.Reader, filter string) ([]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("jsonutil: %w", err)
	}
	return Query(data, filter)
}

// QueryString applies a jq-like filter to a JSON string and returns the result as a string.
func QueryString(jsonStr, filter string) (string, error) {
	result, err := Query([]byte(jsonStr), filter)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// ApplyFilter applies a jq-like filter expression to parsed JSON data.
// This is the core filter engine that operates on already-parsed Go values.
func ApplyFilter(input any, filter string) ([]any, error) {
	filter = strings.TrimSpace(filter)

	if filter == "." {
		return []any{input}, nil
	}

	if filter == "keys" {
		return filterKeys(input)
	}

	if filter == "length" {
		return filterLength(input)
	}

	if filter == "type" {
		return filterType(input)
	}

	if filter == ".[]" {
		return filterIterate(input)
	}

	if strings.Contains(filter, "|") {
		return filterPipe(input, filter)
	}

	if strings.HasPrefix(filter, ".[") && strings.HasSuffix(filter, "]") {
		indexStr := filter[2 : len(filter)-1]
		if idx, err := strconv.Atoi(indexStr); err == nil {
			return filterArrayIndex(input, idx)
		}
		if strings.HasPrefix(indexStr, "\"") && strings.HasSuffix(indexStr, "\"") {
			key := indexStr[1 : len(indexStr)-1]
			return filterObjectKey(input, key)
		}
	}

	if strings.HasPrefix(filter, ".") {
		return filterFieldAccess(input, filter[1:])
	}

	return nil, fmt.Errorf("unsupported filter: %s", filter)
}

func filterKeys(input any) ([]any, error) {
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

func filterLength(input any) ([]any, error) {
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

func filterType(input any) ([]any, error) {
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

func filterIterate(input any) ([]any, error) {
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

func filterArrayIndex(input any, idx int) ([]any, error) {
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

func filterObjectKey(input any, key string) ([]any, error) {
	obj, ok := input.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot index %T with string", input)
	}
	return []any{obj[key]}, nil
}

func filterFieldAccess(input any, path string) ([]any, error) {
	parts := strings.Split(path, ".")
	current := input

	for _, part := range parts {
		if part == "" {
			continue
		}

		if idx := strings.Index(part, "["); idx != -1 {
			fieldName := part[:idx]
			indexPart := part[idx:]

			if fieldName != "" {
				obj, ok := current.(map[string]any)
				if !ok {
					return []any{nil}, nil
				}
				current = obj[fieldName]
			}

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

func filterPipe(input any, filter string) ([]any, error) {
	parts := strings.SplitN(filter, "|", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid pipe")
	}

	intermediate, err := ApplyFilter(input, strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, err
	}

	var results []any
	for _, item := range intermediate {
		r, err := ApplyFilter(item, strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		results = append(results, r...)
	}

	return results, nil
}
