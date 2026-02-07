package utils

import "testing"

func TestTraverseObj(t *testing.T) {
	data := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "deep value",
			},
		},
		"items": []any{
			map[string]any{"name": "first"},
			map[string]any{"name": "second"},
		},
		"count": float64(42),
	}

	tests := []struct {
		name string
		keys []any
		want any
	}{
		{"deep string", []any{"a", "b", "c"}, "deep value"},
		{"array index", []any{"items", 0, "name"}, "first"},
		{"array index 1", []any{"items", 1, "name"}, "second"},
		{"number", []any{"count"}, float64(42)},
		{"missing key", []any{"x"}, nil},
		{"deep missing", []any{"a", "x"}, nil},
		{"out of bounds", []any{"items", 5}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TraverseObj(data, tt.keys...)
			if got != tt.want {
				t.Errorf("TraverseObj(%v) = %v, want %v", tt.keys, got, tt.want)
			}
		})
	}
}

func TestTraverseString(t *testing.T) {
	data := map[string]any{
		"name":  "test",
		"count": float64(42),
	}

	if got := TraverseString(data, "name"); got != "test" {
		t.Errorf("TraverseString(name) = %q", got)
	}

	if got := TraverseString(data, "count"); got != "42" {
		t.Errorf("TraverseString(count) = %q", got)
	}

	if got := TraverseString(data, "missing"); got != "" {
		t.Errorf("TraverseString(missing) = %q", got)
	}
}

func TestParseJSONPath(t *testing.T) {
	keys := ParseJSONPath("a.b.0.c")
	if len(keys) != 4 {
		t.Fatalf("len = %d, want 4", len(keys))
	}

	if keys[0] != "a" || keys[1] != "b" || keys[2] != 0 || keys[3] != "c" {
		t.Errorf("got %v", keys)
	}
}
