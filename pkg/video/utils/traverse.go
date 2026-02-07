package utils

import (
	"strconv"
	"strings"
)

// TraverseObj navigates nested JSON data (map[string]any / []any) using
// a sequence of keys. This replicates youtube-dl's traverse_obj utility.
//
// Keys can be:
//   - string: map key lookup
//   - int: slice index (negative = from end)
//   - nil: returns the current value
//   - slice of keys: try each and return the first non-nil result
//
// Examples:
//
//	TraverseObj(data, "a", "b")          // data["a"]["b"]
//	TraverseObj(data, "items", 0, "url") // data["items"][0]["url"]
func TraverseObj(obj any, keys ...any) any {
	current := obj
	for _, key := range keys {
		if current == nil {
			return nil
		}

		switch k := key.(type) {
		case string:
			m, ok := current.(map[string]any)
			if !ok {
				return nil
			}

			current = m[k]
		case int:
			arr, ok := current.([]any)
			if !ok {
				return nil
			}

			idx := k
			if idx < 0 {
				idx = len(arr) + idx
			}

			if idx < 0 || idx >= len(arr) {
				return nil
			}

			current = arr[idx]
		case []any:
			// Try each sub-key, return first non-nil.
			for _, subKey := range k {
				result := TraverseObj(current, subKey)
				if result != nil {
					return result
				}
			}

			return nil
		case nil:
			return current
		default:
			return nil
		}
	}

	return current
}

// TraverseString navigates and returns the result as a string.
func TraverseString(obj any, keys ...any) string {
	v := TraverseObj(obj, keys...)
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}

		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return ""
	}
}

// TraverseInt navigates and returns the result as an int64 pointer.
func TraverseInt(obj any, keys ...any) *int64 {
	v := TraverseObj(obj, keys...)
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case float64:
		n := int64(val)
		return &n
	case string:
		return IntOrNone(val)
	default:
		return nil
	}
}

// TraverseFloat navigates and returns the result as a float64 pointer.
func TraverseFloat(obj any, keys ...any) *float64 {
	v := TraverseObj(obj, keys...)
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case float64:
		return &val
	case string:
		return FloatOrNone(val)
	default:
		return nil
	}
}

// TraverseList navigates and returns the result as []any.
func TraverseList(obj any, keys ...any) []any {
	v := TraverseObj(obj, keys...)
	if v == nil {
		return nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil
	}

	return arr
}

// TraverseMap navigates and returns the result as map[string]any.
func TraverseMap(obj any, keys ...any) map[string]any {
	v := TraverseObj(obj, keys...)
	if v == nil {
		return nil
	}

	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}

	return m
}

// ParseJSONPath parses a dot-separated path like "a.b.0.c" into traverse keys.
func ParseJSONPath(path string) []any {
	parts := strings.Split(path, ".")

	keys := make([]any, 0, len(parts))
	for _, p := range parts {
		if n, err := strconv.Atoi(p); err == nil {
			keys = append(keys, n)
		} else {
			keys = append(keys, p)
		}
	}

	return keys
}
