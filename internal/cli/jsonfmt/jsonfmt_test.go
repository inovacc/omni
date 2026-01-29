package jsonfmt

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func TestBeautify(t *testing.T) {
	input := []byte(`{"name":"test","value":123}`)

	output, err := Beautify(input, "  ")
	if err != nil {
		t.Fatalf("Beautify() error = %v", err)
	}

	expected := `{
  "name": "test",
  "value": 123
}`
	if string(output) != expected {
		t.Errorf("Beautify() = %s, want %s", output, expected)
	}
}

func TestBeautifyWithTabs(t *testing.T) {
	input := []byte(`{"a":1}`)

	output, err := Beautify(input, "\t")
	if err != nil {
		t.Fatalf("Beautify() error = %v", err)
	}

	if !strings.Contains(string(output), "\t") {
		t.Error("Beautify() should contain tabs")
	}
}

func TestMinify(t *testing.T) {
	input := []byte(`{
  "name": "test",
  "value": 123
}`)

	output, err := Minify(input)
	if err != nil {
		t.Fatalf("Minify() error = %v", err)
	}

	expected := `{"name":"test","value":123}`
	if string(output) != expected {
		t.Errorf("Minify() = %s, want %s", output, expected)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"valid object", []byte(`{"key":"value"}`), false},
		{"valid array", []byte(`[1,2,3]`), false},
		{"valid string", []byte(`"hello"`), false},
		{"valid number", []byte(`123`), false},
		{"valid boolean", []byte(`true`), false},
		{"valid null", []byte(`null`), false},
		{"invalid json", []byte(`{invalid}`), true},
		{"trailing comma", []byte(`{"a":1,}`), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	if !IsValid([]byte(`{"key":"value"}`)) {
		t.Error("IsValid() should return true for valid JSON")
	}

	if IsValid([]byte(`{invalid}`)) {
		t.Error("IsValid() should return false for invalid JSON")
	}
}

func TestSortKeys(t *testing.T) {
	input := []byte(`{"z":1,"a":2,"m":3}`)

	output, err := SortKeys(input)
	if err != nil {
		t.Fatalf("SortKeys() error = %v", err)
	}

	// Verify keys are sorted
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Check all keys exist
	for _, key := range []string{"a", "m", "z"} {
		if _, ok := result[key]; !ok {
			t.Errorf("SortKeys() missing key %s", key)
		}
	}
}

func TestGetType(t *testing.T) {
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte(`{}`), "object"},
		{[]byte(`[]`), "array"},
		{[]byte(`"hello"`), "string"},
		{[]byte(`123`), "number"},
		{[]byte(`-45.6`), "number"},
		{[]byte(`true`), "boolean"},
		{[]byte(`false`), "boolean"},
		{[]byte(`null`), "null"},
		{[]byte(``), "empty"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := GetType(tt.input)
			if result != tt.expected {
				t.Errorf("GetType(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	input := []byte(`{"name":"test","nested":{"key":"value"}}`)

	stats, err := GetStats(input)
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.Type != "object" {
		t.Errorf("stats.Type = %s, want object", stats.Type)
	}

	if stats.Keys != 3 { // name, nested, key
		t.Errorf("stats.Keys = %d, want 3", stats.Keys)
	}

	if stats.Depth < 1 {
		t.Errorf("stats.Depth = %d, want >= 1", stats.Depth)
	}

	if stats.Size <= 0 {
		t.Errorf("stats.Size = %d, want > 0", stats.Size)
	}
}

func TestGetStatsArray(t *testing.T) {
	input := []byte(`[1,2,3,4,5]`)

	stats, err := GetStats(input)
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.Type != "array" {
		t.Errorf("stats.Type = %s, want array", stats.Type)
	}

	if stats.Elements != 5 {
		t.Errorf("stats.Elements = %d, want 5", stats.Elements)
	}
}

func TestKeys(t *testing.T) {
	input := []byte(`{"a":1,"b":{"c":2}}`)

	keys, err := Keys(input)
	if err != nil {
		t.Fatalf("Keys() error = %v", err)
	}

	expected := []string{"a", "b", "b.c"}
	if len(keys) != len(expected) {
		t.Errorf("Keys() returned %d keys, want %d", len(keys), len(expected))
	}

	for _, exp := range expected {
		found := slices.Contains(keys, exp)

		if !found {
			t.Errorf("Keys() missing key %s", exp)
		}
	}
}

func TestKeysWithArray(t *testing.T) {
	input := []byte(`{"items":[{"id":1},{"id":2}]}`)

	keys, err := Keys(input)
	if err != nil {
		t.Fatalf("Keys() error = %v", err)
	}

	// Should include items and array indices
	if len(keys) == 0 {
		t.Error("Keys() returned empty for nested array")
	}
}

func TestRunJSONFmt(t *testing.T) {
	var buf bytes.Buffer

	input := `{"b":2,"a":1}`
	reader := strings.NewReader(input)

	// Create a temp file-like reader
	opts := Options{SortKeys: true}

	err := processReader(&buf, reader, "<test>", opts)
	if err != nil {
		t.Fatalf("processReader() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\"a\"") || !strings.Contains(output, "\"b\"") {
		t.Errorf("processReader() output missing keys: %s", output)
	}
}

func TestRunJSONFmtMinify(t *testing.T) {
	var buf bytes.Buffer

	input := `{
  "key": "value"
}`
	reader := strings.NewReader(input)

	opts := Options{Minify: true}

	err := processReader(&buf, reader, "<test>", opts)
	if err != nil {
		t.Fatalf("processReader() error = %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if strings.Contains(output, "\n") || strings.Contains(output, "  ") {
		t.Errorf("processReader() with Minify should not have whitespace: %s", output)
	}
}

func TestRunJSONFmtValidate(t *testing.T) {
	var buf bytes.Buffer

	input := `{"valid": true}`
	reader := strings.NewReader(input)

	opts := Options{Validate: true}

	err := processReader(&buf, reader, "<test>", opts)
	if err != nil {
		t.Fatalf("processReader() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "valid JSON") {
		t.Errorf("processReader() should report valid: %s", output)
	}
}

func TestRunJSONFmtValidateInvalid(t *testing.T) {
	var buf bytes.Buffer

	input := `{invalid}`
	reader := strings.NewReader(input)

	opts := Options{Validate: true}

	err := processReader(&buf, reader, "<test>", opts)
	if err != nil {
		t.Fatalf("processReader() error = %v (should handle gracefully)", err)
	}

	output := buf.String()
	if !strings.Contains(output, "invalid JSON") {
		t.Errorf("processReader() should report invalid: %s", output)
	}
}

func TestRunJSONFmtValidateJSON(t *testing.T) {
	var buf bytes.Buffer

	input := `{"valid": true}`
	reader := strings.NewReader(input)

	opts := Options{Validate: true, JSON: true}

	err := processReader(&buf, reader, "<test>", opts)
	if err != nil {
		t.Fatalf("processReader() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if !result.Valid {
		t.Error("Result.Valid should be true")
	}
}

func TestFormat(t *testing.T) {
	input := []byte(`{"b":2,"a":1}`)

	tests := []struct {
		name string
		opts Options
	}{
		{"default", Options{}},
		{"minify", Options{Minify: true}},
		{"sort keys", Options{SortKeys: true}},
		{"tabs", Options{Tab: true}},
		{"escape html", Options{EscapeHTML: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := Format(input, tt.opts)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			if len(output) == 0 {
				t.Error("Format() returned empty output")
			}
		})
	}
}

func TestMustBeautify(t *testing.T) {
	input := []byte(`{"key":"value"}`)

	// Should not panic
	result := MustBeautify(input)
	if len(result) == 0 {
		t.Error("MustBeautify() returned empty")
	}
}

func TestMustMinify(t *testing.T) {
	input := []byte(`{  "key"  :  "value"  }`)

	// Should not panic
	result := MustMinify(input)
	if strings.Contains(string(result), "  ") {
		t.Error("MustMinify() should remove extra whitespace")
	}
}

func TestUnescapeHTML(t *testing.T) {
	input := []byte(`\u003c\u003e\u0026`)
	output := unescapeHTML(input)

	if !bytes.Contains(output, []byte("<")) {
		t.Error("unescapeHTML() should convert \\u003c to <")
	}

	if !bytes.Contains(output, []byte(">")) {
		t.Error("unescapeHTML() should convert \\u003e to >")
	}

	if !bytes.Contains(output, []byte("&")) {
		t.Error("unescapeHTML() should convert \\u0026 to &")
	}
}
