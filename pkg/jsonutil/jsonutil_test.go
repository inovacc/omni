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
