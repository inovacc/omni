package csvutil

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunToCSV(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    ToCSVOptions
		want    string
		wantErr bool
	}{
		{
			name:  "simple array",
			input: `[{"name":"John","age":30},{"name":"Jane","age":25}]`,
			opts:  ToCSVOptions{Header: true},
			want:  "age,name\n30,John\n25,Jane\n",
		},
		{
			name:  "single object",
			input: `{"name":"John","age":30}`,
			opts:  ToCSVOptions{Header: true},
			want:  "age,name\n30,John\n",
		},
		{
			name:  "nested objects",
			input: `[{"user":{"name":"John"},"id":1}]`,
			opts:  ToCSVOptions{Header: true},
			want:  "id,user.name\n1,John\n",
		},
		{
			name:  "no header",
			input: `[{"name":"John","age":30}]`,
			opts:  ToCSVOptions{Header: false},
			want:  "30,John\n",
		},
		{
			name:  "custom delimiter",
			input: `[{"name":"John","age":30}]`,
			opts:  ToCSVOptions{Header: true, Delimiter: ";"},
			want:  "age;name\n30;John\n",
		},
		{
			name:  "null values",
			input: `[{"name":"John","age":null}]`,
			opts:  ToCSVOptions{Header: true},
			want:  "age,name\n,John\n",
		},
		{
			name:  "boolean values",
			input: `[{"active":true,"name":"John"}]`,
			opts:  ToCSVOptions{Header: true},
			want:  "active,name\ntrue,John\n",
		},
		{
			name:  "array values",
			input: `[{"tags":["a","b"],"name":"John"}]`,
			opts:  ToCSVOptions{Header: true},
			want:  "name,tags\nJohn,\"[\"\"a\"\",\"\"b\"\"]\"\n",
		},
		{
			name:  "empty array",
			input: `[]`,
			opts:  ToCSVOptions{Header: true},
			want:  "",
		},
		{
			name:    "invalid JSON",
			input:   `not json`,
			opts:    ToCSVOptions{Header: true},
			wantErr: true,
		},
		{
			name:    "primitive value",
			input:   `"string"`,
			opts:    ToCSVOptions{Header: true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := RunToCSV(&buf, r, nil, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunToCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := buf.String()
				if got != tt.want {
					t.Errorf("RunToCSV() =\n%q\nwant\n%q", got, tt.want)
				}
			}
		})
	}
}

func TestRunToCSVFromArgs(t *testing.T) {
	var buf bytes.Buffer

	err := RunToCSV(&buf, nil, []string{`[{"name":"John"}]`}, ToCSVOptions{Header: true})
	if err != nil {
		t.Errorf("RunToCSV() error = %v", err)
	}

	want := "name\nJohn\n"
	if buf.String() != want {
		t.Errorf("RunToCSV() = %q, want %q", buf.String(), want)
	}
}

func TestRunFromCSV(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    FromCSVOptions
		want    string
		wantErr bool
	}{
		{
			name:  "simple csv with header",
			input: "name,age\nJohn,30\nJane,25",
			opts:  FromCSVOptions{Header: true},
			want: `[
  {
    "age": "30",
    "name": "John"
  },
  {
    "age": "25",
    "name": "Jane"
  }
]
`,
		},
		{
			name:  "single row returns object",
			input: "name,age\nJohn,30",
			opts:  FromCSVOptions{Header: true},
			want: `{
  "age": "30",
  "name": "John"
}
`,
		},
		{
			name:  "single row with array flag",
			input: "name,age\nJohn,30",
			opts:  FromCSVOptions{Header: true, Array: true},
			want: `[
  {
    "age": "30",
    "name": "John"
  }
]
`,
		},
		{
			name:  "no header",
			input: "John,30\nJane,25",
			opts:  FromCSVOptions{Header: false},
			want: `[
  {
    "col1": "John",
    "col2": "30"
  },
  {
    "col1": "Jane",
    "col2": "25"
  }
]
`,
		},
		{
			name:  "custom delimiter",
			input: "name;age\nJohn;30",
			opts:  FromCSVOptions{Header: true, Delimiter: ";"},
			want: `{
  "age": "30",
  "name": "John"
}
`,
		},
		{
			name:  "empty input",
			input: "",
			opts:  FromCSVOptions{Header: true},
			want:  "[]\n",
		},
		{
			name:  "variable field count",
			input: "a,b,c\n1,2\n3,4,5",
			opts:  FromCSVOptions{Header: true},
			want: `[
  {
    "a": "1",
    "b": "2",
    "c": ""
  },
  {
    "a": "3",
    "b": "4",
    "c": "5"
  }
]
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := RunFromCSV(&buf, r, nil, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunFromCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := buf.String()
				if got != tt.want {
					t.Errorf("RunFromCSV() =\n%q\nwant\n%q", got, tt.want)
				}
			}
		})
	}
}

func TestGetNestedValue(t *testing.T) {
	tests := []struct {
		name string
		obj  map[string]any
		key  string
		want string
	}{
		{
			name: "simple key",
			obj:  map[string]any{"name": "John"},
			key:  "name",
			want: "John",
		},
		{
			name: "nested key",
			obj:  map[string]any{"user": map[string]any{"name": "John"}},
			key:  "user.name",
			want: "John",
		},
		{
			name: "missing key",
			obj:  map[string]any{"name": "John"},
			key:  "age",
			want: "",
		},
		{
			name: "missing nested key",
			obj:  map[string]any{"user": map[string]any{"name": "John"}},
			key:  "user.age",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getNestedValue(tt.obj, tt.key)
			if got != tt.want {
				t.Errorf("getNestedValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"integer float", float64(42), "42"},
		{"decimal float", 3.14, "3.14"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"array", []any{"a", "b"}, `["a","b"]`},
		{"object", map[string]any{"key": "value"}, `{"key":"value"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.value)
			if got != tt.want {
				t.Errorf("formatValue(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestExtractHeaders(t *testing.T) {
	tests := []struct {
		name    string
		array   []any
		want    int // expected number of headers
		wantErr bool
	}{
		{
			name:  "simple objects",
			array: []any{map[string]any{"a": 1, "b": 2}},
			want:  2,
		},
		{
			name:  "nested objects",
			array: []any{map[string]any{"user": map[string]any{"name": "John"}}},
			want:  1, // user.name
		},
		{
			name:    "non-object element",
			array:   []any{"string"},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractHeaders(tt.array)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractHeaders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("extractHeaders() returned %d headers, want %d", len(got), tt.want)
			}
		})
	}
}
