package yq

import (
	"bytes"
	"strings"
	"testing"
)

// TestRunYqQuery drives RunYq over sample YAML in YAML and JSON output modes.
func TestRunYqQuery(t *testing.T) {
	doc := "name: alice\nage: 30\nnested:\n  key: value\nlist:\n  - a\n  - b\n"

	tests := []struct {
		name     string
		filter   string
		opts     YqOptions
		contains []string
	}{
		{"identity yaml", ".", YqOptions{}, []string{"name: alice", "age: 30"}},
		{"field scalar", ".name", YqOptions{}, []string{"alice"}},
		{"field json", ".age", YqOptions{OutputJSON: true}, []string{"30"}},
		{"nested map", ".nested", YqOptions{}, []string{"key: value"}},
		{"list", ".list", YqOptions{}, []string{"- a", "- b"}},
		{"raw string", ".name", YqOptions{Raw: true}, []string{"alice"}},
		{"compact json", ".nested", YqOptions{OutputJSON: true, Compact: true}, []string{`{"key":"value"}`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			args := append([]string{tt.filter}, "-")
			in := strings.NewReader(doc)
			if err := RunYq(&buf, in, args, tt.opts); err != nil {
				t.Fatalf("RunYq: %v", err)
			}
			for _, sub := range tt.contains {
				if !strings.Contains(buf.String(), sub) {
					t.Errorf("output missing %q:\n%s", sub, buf.String())
				}
			}
		})
	}
}

// TestRunYqNullInput exercises the -n path.
func TestRunYqNullInput(t *testing.T) {
	var buf bytes.Buffer
	if err := RunYq(&buf, strings.NewReader(""), []string{"."}, YqOptions{NullInput: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "null") {
		t.Errorf("expected null output: %q", buf.String())
	}
}

// TestRunYqFromStdinReader confirms reading via the provided reader (no files).
func TestRunYqFromStdinReader(t *testing.T) {
	var buf bytes.Buffer
	in := strings.NewReader("a: 1\nb: two\n")
	if err := RunYq(&buf, in, []string{"."}, YqOptions{}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "a: 1") || !strings.Contains(out, "b: two") {
		t.Errorf("unexpected output: %q", out)
	}
}

// TestRunYqMissingFile verifies not-found classification.
func TestRunYqMissingFile(t *testing.T) {
	var buf bytes.Buffer
	err := RunYq(&buf, strings.NewReader(""), []string{".", "/no/such/file.yaml"}, YqOptions{})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// TestWriteYAML covers the scalar/collection branches of the YAML emitter.
func TestWriteYAML(t *testing.T) {
	tests := []struct {
		name     string
		v        any
		contains string
	}{
		{"nil", nil, "null"},
		{"bool", true, "true"},
		{"int float", float64(42), "42"},
		{"real float", 3.5, "3.5"},
		{"plain string", "hello", "hello"},
		{"quoted string", "a: b", `"a: b"`},
		{"empty list", []any{}, "[]"},
		{"empty map", map[string]any{}, "{}"},
		{"scalar list", []any{"x", "y"}, "- x"},
		{"scalar map", map[string]any{"k": "v"}, "k: v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := writeYAML(&buf, tt.v, 0, 2); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("writeYAML(%v) = %q, want contains %q", tt.v, buf.String(), tt.contains)
			}
		})
	}
}

// TestWriteYAMLNested covers nested collections (non-scalar recursion).
func TestWriteYAMLNested(t *testing.T) {
	var buf bytes.Buffer
	v := map[string]any{
		"outer": map[string]any{"inner": "val"},
		"arr":   []any{map[string]any{"a": float64(1)}},
	}
	if err := writeYAML(&buf, v, 0, 2); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "inner: val") {
		t.Errorf("nested map not rendered: %q", out)
	}
}

// TestIsScalar covers the scalar type predicate.
func TestIsScalar(t *testing.T) {
	scalars := []any{nil, true, float64(1), "s"}
	for _, s := range scalars {
		if !isScalar(s) {
			t.Errorf("isScalar(%v) = false, want true", s)
		}
	}
	nonScalars := []any{[]any{1}, map[string]any{"a": 1}}
	for _, s := range nonScalars {
		if isScalar(s) {
			t.Errorf("isScalar(%v) = true, want false", s)
		}
	}
}

// TestParseYAMLValue exercises the value parser via the public parser path.
func TestParseYAMLValue(t *testing.T) {
	// empty value -> nil
	v, err := parseYAMLValue(nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Errorf("expected nil for empty, got %v", v)
	}

	// non-empty nested lines parse into a map.
	v, err = parseYAMLValue([]string{"  k: v"}, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := v.(map[string]any)
	if !ok || m["k"] != "v" {
		t.Errorf("parseYAMLValue nested = %#v", v)
	}
}

// TestParseYAMLScalarTypes covers scalar coercion (bool/null/number/quoted).
func TestParseYAMLScalarTypes(t *testing.T) {
	tests := []struct {
		in   string
		want any
	}{
		{`"quoted"`, "quoted"},
		{"'single'", "single"},
		{"true", true},
		{"no", false},
		{"null", nil},
		{"~", nil},
		{"42", float64(42)},
		{"3.14", 3.14},
		{"plain", "plain"},
	}
	for _, tt := range tests {
		if got := parseYAMLScalar(tt.in); got != tt.want {
			t.Errorf("parseYAMLScalar(%q) = %v (%T), want %v", tt.in, got, got, tt.want)
		}
	}
}
