package jsonutil

import (
	"strings"
	"testing"
)

func TestQuery(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		filter  string
		want    string
		wantErr bool
	}{
		{
			name:   "identity",
			data:   `{"key":"value"}`,
			filter: ".",
			want:   `{"key":"value"}`,
		},
		{
			name:   "field access",
			data:   `{"name":"John","age":30}`,
			filter: ".name",
			want:   `"John"`,
		},
		{
			name:   "nested field",
			data:   `{"user":{"address":{"city":"NYC"}}}`,
			filter: ".user.address.city",
			want:   `"NYC"`,
		},
		{
			name:   "array index",
			data:   `{"items":["a","b","c"]}`,
			filter: ".items[0]",
			want:   `"a"`,
		},
		{
			name:   "length",
			data:   `[1,2,3]`,
			filter: "length",
			want:   "3",
		},
		{
			name:   "pipe",
			data:   `{"items":[1,2,3]}`,
			filter: ".items | length",
			want:   "3",
		},
		{
			name:   "chained pipes",
			data:   `{"data":{"values":["a","b","c"]}}`,
			filter: ".data | .values | length",
			want:   "3",
		},
		{
			name:   "null value",
			data:   `{"key":null}`,
			filter: ".key",
			want:   "null",
		},
		{
			name:   "boolean",
			data:   `{"yes":true}`,
			filter: ".yes",
			want:   "true",
		},
		{
			name:   "number",
			data:   `{"int":42}`,
			filter: ".int",
			want:   "42",
		},
		{
			name:    "invalid json",
			data:    `{not valid}`,
			filter:  ".",
			wantErr: true,
		},
		{
			name:    "unsupported filter",
			data:    `{"key":"value"}`,
			filter:  "invalid_filter",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Query([]byte(tt.data), tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				gotStr := strings.TrimSpace(string(got))
				if gotStr != tt.want {
					t.Errorf("Query() = %v, want %v", gotStr, tt.want)
				}
			}
		})
	}
}

func TestQueryString(t *testing.T) {
	got, err := QueryString(`{"msg":"hello"}`, ".msg")
	if err != nil {
		t.Fatalf("QueryString() error = %v", err)
	}

	if got != `"hello"` {
		t.Errorf("QueryString() = %v, want %q", got, `"hello"`)
	}
}

func TestQueryReader(t *testing.T) {
	r := strings.NewReader(`{"key":"value"}`)

	got, err := QueryReader(r, ".key")
	if err != nil {
		t.Fatalf("QueryReader() error = %v", err)
	}

	if strings.TrimSpace(string(got)) != `"value"` {
		t.Errorf("QueryReader() = %v", string(got))
	}
}

func TestApplyFilter(t *testing.T) {
	t.Run("keys", func(t *testing.T) {
		input := map[string]any{"a": 1, "b": 2}

		results, err := ApplyFilter(input, "keys")
		if err != nil {
			t.Fatal(err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		keys, ok := results[0].([]any)
		if !ok {
			t.Fatal("expected []any result")
		}

		if len(keys) != 2 {
			t.Errorf("expected 2 keys, got %d", len(keys))
		}
	})

	t.Run("type", func(t *testing.T) {
		tests := []struct {
			input any
			want  string
		}{
			{nil, "null"},
			{true, "boolean"},
			{float64(1), "number"},
			{"hello", "string"},
			{[]any{}, "array"},
			{map[string]any{}, "object"},
		}
		for _, tt := range tests {
			results, err := ApplyFilter(tt.input, "type")
			if err != nil {
				t.Fatal(err)
			}

			if results[0] != tt.want {
				t.Errorf("type of %v = %v, want %v", tt.input, results[0], tt.want)
			}
		}
	})

	t.Run("iterate array", func(t *testing.T) {
		input := []any{1.0, 2.0, 3.0}

		results, err := ApplyFilter(input, ".[]")
		if err != nil {
			t.Fatal(err)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}
	})

	t.Run("iterate object", func(t *testing.T) {
		input := map[string]any{"a": 1.0, "b": 2.0}

		results, err := ApplyFilter(input, ".[]")
		if err != nil {
			t.Fatal(err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("deep nesting", func(t *testing.T) {
		input := map[string]any{
			"l1": map[string]any{
				"l2": map[string]any{
					"l3": map[string]any{
						"value": "deep",
					},
				},
			},
		}

		results, err := ApplyFilter(input, ".l1.l2.l3.value")
		if err != nil {
			t.Fatal(err)
		}

		if results[0] != "deep" {
			t.Errorf("got %v, want deep", results[0])
		}
	})
}

func TestQueryConsistency(t *testing.T) {
	data := []byte(`{"key":"value"}`)
	r1, _ := Query(data, ".")

	r2, _ := Query(data, ".")
	if string(r1) != string(r2) {
		t.Error("Query should produce consistent results")
	}
}

func TestQueryReader_Error(t *testing.T) {
	// malformed JSON via reader
	r := strings.NewReader(`{bad json}`)
	_, err := QueryReader(r, ".")
	if err == nil {
		t.Error("QueryReader() should error on malformed JSON")
	}
}

func TestQueryString_Error(t *testing.T) {
	_, err := QueryString(`not json`, ".")
	if err == nil {
		t.Error("QueryString() should error on invalid JSON")
	}
}

func TestApplyFilter_Extended(t *testing.T) {
	t.Run("keys of array", func(t *testing.T) {
		input := []any{"a", "b", "c"}
		results, err := ApplyFilter(input, "keys")
		if err != nil {
			t.Fatal(err)
		}
		keys, ok := results[0].([]any)
		if !ok {
			t.Fatal("expected []any")
		}
		if len(keys) != 3 {
			t.Errorf("expected 3 keys, got %d", len(keys))
		}
	})

	t.Run("keys error on non-collection", func(t *testing.T) {
		_, err := ApplyFilter("string", "keys")
		if err == nil {
			t.Error("keys of string should error")
		}
	})

	t.Run("length of string", func(t *testing.T) {
		results, err := ApplyFilter("hello", "length")
		if err != nil {
			t.Fatal(err)
		}
		if results[0] != float64(5) {
			t.Errorf("length of 'hello' = %v, want 5", results[0])
		}
	})

	t.Run("length of nil", func(t *testing.T) {
		results, err := ApplyFilter(nil, "length")
		if err != nil {
			t.Fatal(err)
		}
		if results[0] != float64(0) {
			t.Errorf("length of nil = %v, want 0", results[0])
		}
	})

	t.Run("length error on number", func(t *testing.T) {
		_, err := ApplyFilter(float64(42), "length")
		if err == nil {
			t.Error("length of number should error")
		}
	})

	t.Run("type unknown", func(t *testing.T) {
		// Use a struct type which has no type match
		type custom struct{}
		results, err := ApplyFilter(custom{}, "type")
		if err != nil {
			t.Fatal(err)
		}
		if results[0] != "unknown" {
			t.Errorf("type of custom struct = %v, want unknown", results[0])
		}
	})

	t.Run("iterate error on non-collection", func(t *testing.T) {
		_, err := ApplyFilter("string", ".[]")
		if err == nil {
			t.Error(".[] on string should error")
		}
	})

	t.Run("array index negative", func(t *testing.T) {
		input := []any{"a", "b", "c"}
		results, err := ApplyFilter(input, ".[-1]")
		if err != nil {
			t.Fatal(err)
		}
		if results[0] != "c" {
			t.Errorf(".[−1] = %v, want c", results[0])
		}
	})

	t.Run("array index out of bounds", func(t *testing.T) {
		input := []any{"a", "b"}
		results, err := ApplyFilter(input, ".[10]")
		if err != nil {
			t.Fatal(err)
		}
		if results[0] != nil {
			t.Errorf(".[10] out of bounds = %v, want nil", results[0])
		}
	})

	t.Run("array index on non-array", func(t *testing.T) {
		_, err := ApplyFilter("notarray", ".[0]")
		if err == nil {
			t.Error(".[0] on string should error")
		}
	})

	t.Run("object key via bracket syntax", func(t *testing.T) {
		input := map[string]any{"name": "Alice"}
		results, err := ApplyFilter(input, `.["name"]`)
		if err != nil {
			t.Fatal(err)
		}
		if results[0] != "Alice" {
			t.Errorf(`.["name"] = %v, want Alice`, results[0])
		}
	})

	t.Run("object key on non-object", func(t *testing.T) {
		_, err := ApplyFilter("notobj", `.["key"]`)
		if err == nil {
			t.Error(`.["key"] on non-object should error`)
		}
	})

	t.Run("field access returns nil for missing key", func(t *testing.T) {
		input := map[string]any{"a": 1.0}
		results, err := ApplyFilter(input, ".missing")
		if err != nil {
			t.Fatal(err)
		}
		if results[0] != nil {
			t.Errorf(".missing = %v, want nil", results[0])
		}
	})

	t.Run("field access with array index in path", func(t *testing.T) {
		input := map[string]any{
			"items": []any{"x", "y", "z"},
		}
		results, err := ApplyFilter(input, ".items[1]")
		if err != nil {
			t.Fatal(err)
		}
		if results[0] != "y" {
			t.Errorf(".items[1] = %v, want y", results[0])
		}
	})

	t.Run("field access on non-object returns nil", func(t *testing.T) {
		results, err := ApplyFilter("notobj", ".somekey")
		if err != nil {
			t.Fatal(err)
		}
		if results[0] != nil {
			t.Errorf("field on non-object = %v, want nil", results[0])
		}
	})

	t.Run("pipe with error in left side", func(t *testing.T) {
		_, err := ApplyFilter("string", ".[] | length")
		if err == nil {
			t.Error("pipe with invalid left expression should error")
		}
	})

	t.Run("pipe with error in right side", func(t *testing.T) {
		input := []any{"a", "b"}
		_, err := ApplyFilter(input, ".[] | badfilter")
		if err == nil {
			t.Error("pipe with invalid right expression should error")
		}
	})

	t.Run("empty object identity", func(t *testing.T) {
		results, err := ApplyFilter(map[string]any{}, ".")
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 result for empty object, got %d", len(results))
		}
	})

	t.Run("array root identity", func(t *testing.T) {
		input := []any{1.0, 2.0}
		results, err := ApplyFilter(input, ".")
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 result for array identity, got %d", len(results))
		}
	})

	t.Run("negative index out of bounds returns nil", func(t *testing.T) {
		input := []any{"a"}
		results, err := ApplyFilter(input, ".[-10]")
		if err != nil {
			t.Fatal(err)
		}
		if results[0] != nil {
			t.Errorf(".[-10] = %v, want nil", results[0])
		}
	})
}

func TestQuery_MultiResult(t *testing.T) {
	// .[] on an array produces multiple results → marshalled as array
	got, err := Query([]byte(`[1,2,3]`), ".[]")
	if err != nil {
		t.Fatalf("Query([]) error = %v", err)
	}
	if string(got) != "[1,2,3]" {
		t.Errorf("Query(.[] on array) = %v, want [1,2,3]", string(got))
	}
}

func TestQuery_NullJSON(t *testing.T) {
	got, err := Query([]byte(`null`), "type")
	if err != nil {
		t.Fatalf("Query(null) error = %v", err)
	}
	if string(got) != `"null"` {
		t.Errorf("Query(null type) = %v, want \"null\"", string(got))
	}
}
