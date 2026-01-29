package yaml2struct

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunYAML2Struct(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     Options
		contains []string
	}{
		{
			name: "simple object",
			input: `name: test
count: 42`,
			opts: Options{Name: "Config", Package: "main"},
			contains: []string{
				"package main",
				"type Config struct",
				"Name string",
				"Count int",
			},
		},
		{
			name: "nested object",
			input: `user:
  name: alice
  age: 30`,
			opts: Options{Name: "Response", Package: "models"},
			contains: []string{
				"package models",
				"type Response struct",
				"User User",
				"type User struct",
			},
		},
		{
			name: "array",
			input: `items:
  - id: 1
  - id: 2`,
			opts: Options{Name: "List", Package: "main"},
			contains: []string{
				"type List struct",
				"Items []",
			},
		},
		{
			name: "various types",
			input: `active: true
score: 3.14
label: test`,
			opts: Options{Name: "Mixed", Package: "main"},
			contains: []string{
				"Active bool",
				"Score float64",
				"Label string",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := strings.NewReader(tt.input)

			err := RunYAML2Struct(&buf, r, nil, tt.opts)
			if err != nil {
				t.Fatalf("RunYAML2Struct() error = %v", err)
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

func TestInvalidYAML(t *testing.T) {
	var buf bytes.Buffer
	r := strings.NewReader(":\ninvalid: [yaml")

	err := RunYAML2Struct(&buf, r, nil, Options{})
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
