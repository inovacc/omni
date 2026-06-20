package yamlutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

// TestWrapInputErr verifies error classification into cmderr sentinels.
func TestWrapInputErr(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantIs  error
	}{
		{"not exist", os.ErrNotExist, cmderr.ErrNotFound},
		{"permission", os.ErrPermission, cmderr.ErrPermission},
		{"generic", errors.New("boom"), cmderr.ErrIO},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapInputErr("yaml", tt.err)
			if !errors.Is(got, tt.wantIs) {
				t.Errorf("wrapInputErr(%v) not %v: %v", tt.err, tt.wantIs, got)
			}
		})
	}
}

// TestRunValidateMore covers valid and invalid YAML via literal-string args (so
// it never touches os.Stdin) plus JSON output mode.
func TestRunValidateMore(t *testing.T) {
	t.Run("valid literal", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunValidate(&buf, []string{"a: 1\nb: 2\n"}, ValidateOptions{}); err != nil {
			t.Fatalf("expected valid: %v", err)
		}
		if !strings.Contains(buf.String(), "valid YAML") {
			t.Errorf("missing valid message: %s", buf.String())
		}
	})

	t.Run("invalid literal", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunValidate(&buf, []string{"a: : : bad\n  - x\nfoo"}, ValidateOptions{})
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("valid file json mode", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "ok.yaml")
		if err := os.WriteFile(p, []byte("k: v\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		opts := ValidateOptions{OutputFormat: output.FormatJSON}
		if err := RunValidate(&buf, []string{p}, opts); err != nil {
			t.Fatal(err)
		}
		var res ValidateResult
		if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
			t.Fatalf("unmarshal: %v\n%s", err, buf.String())
		}
		if !res.Valid {
			t.Errorf("expected valid: %+v", res)
		}
	})
}

// TestSortKeysHelper exercises recursive map/slice key sorting.
func TestSortKeysHelper(t *testing.T) {
	in := map[string]any{
		"z": 1,
		"a": map[string]any{"y": 2, "b": 3},
		"l": []any{map[string]any{"d": 4}},
	}
	out := sortKeys(in)
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", out)
	}
	// Nested map should also be sortKeys-processed (still a map).
	if _, ok := m["a"].(map[string]any); !ok {
		t.Errorf("nested map not preserved: %T", m["a"])
	}
}

// TestSortStrings verifies the in-place bubble sort.
func TestSortStrings(t *testing.T) {
	tests := []struct {
		in   []string
		want []string
	}{
		{[]string{"c", "a", "b"}, []string{"a", "b", "c"}},
		{[]string{"x"}, []string{"x"}},
		{nil, nil},
		{[]string{"2", "10", "1"}, []string{"1", "10", "2"}},
	}
	for _, tt := range tests {
		got := append([]string(nil), tt.in...)
		sortStrings(got)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("sortStrings(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// TestIsEmpty covers every branch of isEmpty.
func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"nonempty string", "x", false},
		{"empty map", map[string]any{}, true},
		{"nonempty map", map[string]any{"a": 1}, false},
		{"empty slice", []any{}, true},
		{"nonempty slice", []any{1}, false},
		{"int", 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEmpty(tt.v); got != tt.want {
				t.Errorf("isEmpty(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

// TestRemoveEmptyValues drives the recursive nil/empty pruning.
func TestRemoveEmptyValues(t *testing.T) {
	in := map[string]any{
		"keep":      "value",
		"nilval":    nil,
		"emptystr":  "",
		"emptymap":  map[string]any{},
		"emptylist": []any{},
		"nested": map[string]any{
			"a":      1,
			"gone":   nil,
			"subnil": map[string]any{"x": nil},
		},
		"list": []any{"a", nil, ""},
	}

	out := removeEmptyValues(in)
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", out)
	}

	if _, exists := m["nilval"]; exists {
		t.Error("nilval should be removed")
	}
	if _, exists := m["emptystr"]; exists {
		t.Error("emptystr should be removed")
	}
	if _, exists := m["emptymap"]; exists {
		t.Error("emptymap should be removed")
	}
	if m["keep"] != "value" {
		t.Errorf("keep lost: %+v", m)
	}
	// nested.subnil collapses to nil -> removed; nested becomes {a:1}.
	nested, ok := m["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested missing/wrong type: %T", m["nested"])
	}
	if _, exists := nested["subnil"]; exists {
		t.Error("subnil should collapse and be removed")
	}
	// list drops nil and "" leaving just "a".
	list, ok := m["list"].([]any)
	if !ok || len(list) != 1 || list[0] != "a" {
		t.Errorf("list pruning failed: %+v", m["list"])
	}
}

// TestRemoveEmptyAllEmpty confirms a fully-empty map collapses to nil.
func TestRemoveEmptyAllEmpty(t *testing.T) {
	out := removeEmptyValues(map[string]any{"a": nil, "b": ""})
	if out != nil {
		t.Errorf("expected nil, got %+v", out)
	}
	out2 := removeEmptyValues([]any{nil, ""})
	if out2 != nil {
		t.Errorf("expected nil slice, got %+v", out2)
	}
}

// TestGetInputWithFilename covers file, literal-string and (redirected) stdin.
func TestGetInputWithFilename(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.yaml")
	if err := os.WriteFile(p, []byte("a: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("file", func(t *testing.T) {
		content, fn, err := getInputWithFilename([]string{p})
		if err != nil {
			t.Fatal(err)
		}
		if fn != p || !strings.Contains(content, "a: 1") {
			t.Errorf("file branch: fn=%q content=%q", fn, content)
		}
	})

	t.Run("literal", func(t *testing.T) {
		content, fn, err := getInputWithFilename([]string{"x:", "1"})
		if err != nil {
			t.Fatal(err)
		}
		if fn != "" || content != "x: 1" {
			t.Errorf("literal branch: fn=%q content=%q", fn, content)
		}
	})

	t.Run("stdin", func(t *testing.T) {
		tmp, err := os.CreateTemp(dir, "stdin")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tmp.WriteString("from: stdin\n"); err != nil {
			t.Fatal(err)
		}
		if _, err := tmp.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tmp.Close() }()

		old := os.Stdin
		os.Stdin = tmp
		defer func() { os.Stdin = old }()

		content, fn, err := getInputWithFilename(nil)
		if err != nil {
			t.Fatal(err)
		}
		if fn != "" || !strings.Contains(content, "from: stdin") {
			t.Errorf("stdin branch: fn=%q content=%q", fn, content)
		}
	})
}

// TestRunFormatRemoveEmptyAndSort drives RunFormat end-to-end over a file.
func TestRunFormatRemoveEmptyAndSort(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "in.yaml")
	if err := os.WriteFile(p, []byte("b: 2\na: 1\nempty: null\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	opts := FormatOptions{Indent: 2, SortKeys: true, RemoveEmpty: true}
	if err := RunFormat(&buf, []string{p}, opts); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "empty") {
		t.Errorf("null value not removed: %s", out)
	}
	if !strings.Contains(out, "a: 1") || !strings.Contains(out, "b: 2") {
		t.Errorf("keys lost: %s", out)
	}
}
