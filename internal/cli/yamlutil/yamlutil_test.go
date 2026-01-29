package yamlutil

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    ValidateOptions
		wantErr bool
	}{
		{
			name:    "valid simple yaml",
			input:   "name: test\nvalue: 123",
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "valid nested yaml",
			input:   "parent:\n  child: value\n  list:\n    - item1\n    - item2",
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "valid array",
			input:   "- item1\n- item2\n- item3",
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "valid multi-document",
			input:   "---\nname: doc1\n---\nname: doc2",
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "invalid yaml - bad indentation",
			input:   "name: test\n  invalid: value",
			opts:    ValidateOptions{},
			wantErr: true,
		},
		{
			name:    "invalid yaml - unclosed quote",
			input:   "name: \"unclosed",
			opts:    ValidateOptions{},
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			opts:    ValidateOptions{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunValidate(&buf, []string{tt.input}, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunValidateJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := ValidateOptions{JSON: true}

	err := RunValidate(&buf, []string{"name: test\nvalue: 123"}, opts)
	if err != nil {
		t.Fatalf("RunValidate() error = %v", err)
	}

	var result ValidateResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if !result.Valid {
		t.Errorf("Valid = false, want true")
	}
}

func TestRunValidateJSONInvalid(t *testing.T) {
	var buf bytes.Buffer

	opts := ValidateOptions{JSON: true}

	_ = RunValidate(&buf, []string{"name: \"unclosed"}, opts)

	var result ValidateResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.Valid {
		t.Errorf("Valid = true, want false")
	}

	if result.Error == "" {
		t.Errorf("Error should not be empty")
	}
}

func TestRunFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    FormatOptions
		wantErr bool
	}{
		{
			name:    "format simple yaml",
			input:   "name: test\nvalue: 123",
			opts:    FormatOptions{Indent: 2},
			wantErr: false,
		},
		{
			name:    "format nested yaml",
			input:   "parent:\n  child: value",
			opts:    FormatOptions{Indent: 2},
			wantErr: false,
		},
		{
			name:    "format with custom indent",
			input:   "parent:\n    child: value",
			opts:    FormatOptions{Indent: 4},
			wantErr: false,
		},
		{
			name:    "format to json",
			input:   "name: test\nvalue: 123",
			opts:    FormatOptions{JSON: true, Indent: 2},
			wantErr: false,
		},
		{
			name:    "invalid yaml",
			input:   "name: \"unclosed",
			opts:    FormatOptions{Indent: 2},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunFormat(&buf, []string{tt.input}, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunFormatToJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := FormatOptions{JSON: true, Indent: 2}

	err := RunFormat(&buf, []string{"name: test\nvalue: 123"}, opts)
	if err != nil {
		t.Fatalf("RunFormat() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, `"name"`) {
		t.Errorf("Output should contain JSON key 'name'")
	}

	if !strings.Contains(output, `"test"`) {
		t.Errorf("Output should contain JSON value 'test'")
	}
}

func TestRunFormatPreservesData(t *testing.T) {
	input := "name: test\nlist:\n  - item1\n  - item2\nnested:\n  key: value"

	var buf bytes.Buffer

	err := RunFormat(&buf, []string{input}, FormatOptions{Indent: 2})
	if err != nil {
		t.Fatalf("RunFormat() error = %v", err)
	}

	output := buf.String()

	// Check that key data is preserved
	if !strings.Contains(output, "name:") {
		t.Errorf("Output should contain 'name:'")
	}

	if !strings.Contains(output, "item1") {
		t.Errorf("Output should contain 'item1'")
	}

	if !strings.Contains(output, "nested:") {
		t.Errorf("Output should contain 'nested:'")
	}
}
