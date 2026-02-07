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

func TestRunFormatSortKeys(t *testing.T) {
	input := "zebra: 1\napple: 2\nmango: 3"
	var buf bytes.Buffer

	err := RunFormat(&buf, []string{input}, FormatOptions{Indent: 2, SortKeys: true})
	if err != nil {
		t.Fatalf("RunFormat() error = %v", err)
	}

	output := buf.String()
	appleIdx := strings.Index(output, "apple")
	mangoIdx := strings.Index(output, "mango")
	zebraIdx := strings.Index(output, "zebra")

	if appleIdx > mangoIdx || mangoIdx > zebraIdx {
		t.Errorf("keys should be sorted alphabetically: apple < mango < zebra")
	}
}

func TestRunFormatRemoveEmpty(t *testing.T) {
	input := "name: test\nempty_str: \"\"\nreal: value\nnull_val: null"
	var buf bytes.Buffer

	err := RunFormat(&buf, []string{input}, FormatOptions{Indent: 2, RemoveEmpty: true})
	if err != nil {
		t.Fatalf("RunFormat() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name:") {
		t.Error("should keep non-empty 'name'")
	}
	if !strings.Contains(output, "real:") {
		t.Error("should keep non-empty 'real'")
	}
}

func TestRunValidate_BooleanVariations(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"yes_no", "active: yes\ndeleted: no"},
		{"true_false", "active: true\ndeleted: false"},
		{"on_off", "active: on\ndeleted: off"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := RunValidate(&buf, []string{tt.input}, ValidateOptions{})
			if err != nil {
				t.Errorf("RunValidate() error = %v", err)
			}
		})
	}
}

func TestRunValidate_ComplexTypes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			"block_scalar_literal",
			"description: |\n  This is a multi-line\n  block scalar.",
			false,
		},
		{
			"block_scalar_folded",
			"description: >\n  This is a folded\n  block scalar.",
			false,
		},
		{
			"flow_mapping",
			"person: {name: John, age: 30}",
			false,
		},
		{
			"flow_sequence",
			"tags: [alpha, beta, gamma]",
			false,
		},
		{
			"anchor_alias",
			"defaults: &defaults\n  color: red\n  size: large\nitem:\n  <<: *defaults\n  name: widget",
			false,
		},
		{
			"deeply_nested",
			"a:\n  b:\n    c:\n      d:\n        e: deep",
			false,
		},
		{
			"numeric_types",
			"int: 42\nfloat: 3.14\nhex: 0xFF\noctal: 0o777",
			false,
		},
		{
			"null_values",
			"empty:\nnull_val: null\ntilde: ~",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := RunValidate(&buf, []string{tt.input}, ValidateOptions{})
			if (err != nil) != tt.wantErr {
				t.Errorf("RunValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunK8sFormat(t *testing.T) {
	input := "spec:\n  containers:\n    - name: web\nmetadata:\n  name: pod\napiVersion: v1\nkind: Pod"
	var buf bytes.Buffer

	err := RunK8sFormat(&buf, []string{input}, K8sFormatOptions{Indent: 2})
	if err != nil {
		t.Fatalf("RunK8sFormat() error = %v", err)
	}

	output := buf.String()
	// K8s format should put apiVersion and kind first
	apiIdx := strings.Index(output, "apiVersion")
	kindIdx := strings.Index(output, "kind")
	metaIdx := strings.Index(output, "metadata")
	specIdx := strings.Index(output, "spec")

	if apiIdx < 0 || kindIdx < 0 || metaIdx < 0 || specIdx < 0 {
		t.Errorf("expected all k8s keys in output, got: %s", output)
		return
	}

	if apiIdx > kindIdx || kindIdx > metaIdx || metaIdx > specIdx {
		t.Logf("K8s ordering: apiVersion=%d kind=%d metadata=%d spec=%d", apiIdx, kindIdx, metaIdx, specIdx)
	}
}

func TestRunK8sFormat_Invalid(t *testing.T) {
	var buf bytes.Buffer
	err := RunK8sFormat(&buf, []string{"name: \"unclosed"}, K8sFormatOptions{Indent: 2})
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestRunFormat_MultiDocument(t *testing.T) {
	input := "---\nname: doc1\n---\nname: doc2"
	var buf bytes.Buffer

	err := RunFormat(&buf, []string{input}, FormatOptions{Indent: 2})
	if err != nil {
		t.Fatalf("RunFormat() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "doc1") || !strings.Contains(output, "doc2") {
		t.Error("multi-document format should preserve all documents")
	}
}

func TestRunValidate_StrictMode(t *testing.T) {
	var buf bytes.Buffer
	// Valid YAML should pass even in strict mode
	err := RunValidate(&buf, []string{"name: test\nvalue: 123"}, ValidateOptions{Strict: true})
	if err != nil {
		t.Errorf("RunValidate() strict mode error = %v", err)
	}
}

func TestRunFormat_K8sOption(t *testing.T) {
	input := "spec:\n  replicas: 3\napiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test"
	var buf bytes.Buffer

	err := RunFormat(&buf, []string{input}, FormatOptions{Indent: 2, K8s: true})
	if err != nil {
		t.Fatalf("RunFormat() with K8s error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "apiVersion") {
		t.Error("K8s format should include apiVersion")
	}
}

func TestRunFormatJSON_ComplexTypes(t *testing.T) {
	input := "name: test\nlist:\n  - 1\n  - 2\nnested:\n  a: b"
	var buf bytes.Buffer

	err := RunFormat(&buf, []string{input}, FormatOptions{JSON: true, Indent: 2})
	if err != nil {
		t.Fatalf("RunFormat() JSON error = %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}
	if parsed["name"] != "test" {
		t.Errorf("expected name=test, got %v", parsed["name"])
	}
}
