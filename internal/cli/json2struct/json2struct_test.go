package json2struct

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunJSON2Struct(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     Options
		contains []string
	}{
		{
			name:  "simple object",
			input: `{"name": "test", "count": 42}`,
			opts:  Options{Name: "Config", Package: "main"},
			contains: []string{
				"package main",
				"type Config struct",
				"Name string",
				"Count int",
				`json:"name"`,
				`json:"count"`,
			},
		},
		{
			name:  "nested object",
			input: `{"user": {"name": "alice", "age": 30}}`,
			opts:  Options{Name: "Response", Package: "models"},
			contains: []string{
				"package models",
				"type Response struct",
				"User",
				"type User struct",
				"Name string",
				"Age int",
			},
		},
		{
			name:  "array",
			input: `{"items": [{"id": 1}, {"id": 2}]}`,
			opts:  Options{Name: "List", Package: "main"},
			contains: []string{
				"type List struct",
				"Items []",
				"type Item struct",
				"ID int",
			},
		},
		{
			name:  "with omitempty",
			input: `{"name": "test"}`,
			opts:  Options{Name: "Data", Package: "main", OmitEmpty: true},
			contains: []string{
				`json:"name,omitempty"`,
			},
		},
		{
			name:  "various types",
			input: `{"active": true, "score": 3.14, "label": "test", "data": null}`,
			opts:  Options{Name: "Mixed", Package: "main"},
			contains: []string{
				"Active bool",
				"Score float64",
				"Label string",
				"Data any",
			},
		},
		{
			name:  "empty array",
			input: `{"tags": []}`,
			opts:  Options{Name: "Doc", Package: "main"},
			contains: []string{
				"Tags []any",
			},
		},
		{
			name:  "snake_case keys",
			input: `{"user_name": "alice", "created_at": "2024-01-01"}`,
			opts:  Options{Name: "Record", Package: "main"},
			contains: []string{
				"UserName string",
				"CreatedAt string",
				`json:"user_name"`,
				`json:"created_at"`,
			},
		},
		{
			name:  "acronyms",
			input: `{"id": 1, "url": "http://test.com", "api_key": "secret"}`,
			opts:  Options{Name: "Config", Package: "main"},
			contains: []string{
				"ID int",
				"URL string",
				"APIKey string",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := RunJSON2Struct(&buf, r, nil, tt.opts)
			if err != nil {
				t.Fatalf("RunJSON2Struct() error = %v", err)
			}

			output := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestToGoName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"name", "Name"},
		{"user_name", "UserName"},
		{"userName", "UserName"},
		{"user-name", "UserName"},
		{"id", "ID"},
		{"user_id", "UserID"},
		{"api_url", "APIURL"},
		{"HTMLContent", "Htmlcontent"},
		{"created_at", "CreatedAt"},
		{"", "Field"},
		{"123", "F123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toGoName(tt.input)
			if got != tt.want {
				t.Errorf("toGoName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSingularize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"items", "item"},
		{"users", "user"},
		{"categories", "category"},
		{"boxes", "box"},
		{"data", "dataItem"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := singularize(tt.input)
			if got != tt.want {
				t.Errorf("singularize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestInvalidJSON(t *testing.T) {
	var buf bytes.Buffer

	r := strings.NewReader("not valid json")

	err := RunJSON2Struct(&buf, r, nil, Options{})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
