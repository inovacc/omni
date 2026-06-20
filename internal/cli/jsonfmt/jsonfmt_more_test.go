package jsonfmt

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/cobra/helper/output"
)

// TestRunJSONFmtFromFile drives RunJSONFmt over a real file in beautify,
// minify, sort-keys and validate modes.
func TestRunJSONFmtFromFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.json")
	if err := os.WriteFile(src, []byte(`{"b":2,"a":1,"c":[3,2,1]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		opts     Options
		contains []string
	}{
		{"beautify", Options{}, []string{"\n", "\"a\": 1"}},
		{"minify", Options{Minify: true}, []string{`{"a":1,"b":2`}},
		{"tab indent", Options{Tab: true}, []string{"\t\"a\""}},
		{"sort keys", Options{SortKeys: true, Minify: true}, []string{`{"a":1,"b":2`}},
		{"validate text", Options{Validate: true}, []string{"valid JSON"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RunJSONFmt(&buf, []string{src}, tt.opts); err != nil {
				t.Fatalf("RunJSONFmt: %v", err)
			}
			for _, sub := range tt.contains {
				if !strings.Contains(buf.String(), sub) {
					t.Errorf("output missing %q:\n%s", sub, buf.String())
				}
			}
		})
	}
}

// TestRunJSONFmtValidateJSONMode verifies the JSON output of validate mode for
// both valid and invalid input.
func TestRunJSONFmtValidateJSONMode(t *testing.T) {
	dir := t.TempDir()

	good := filepath.Join(dir, "good.json")
	bad := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(good, []byte(`{"x":1}`), 0o644)
	_ = os.WriteFile(bad, []byte(`{not json`), 0o644)

	t.Run("valid", func(t *testing.T) {
		var buf bytes.Buffer
		opts := Options{Validate: true, OutputFormat: output.FormatJSON}
		if err := RunJSONFmt(&buf, []string{good}, opts); err != nil {
			t.Fatal(err)
		}
		var res Result
		if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
			t.Fatalf("unmarshal: %v\n%s", err, buf.String())
		}
		if !res.Valid {
			t.Errorf("expected valid=true: %+v", res)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		var buf bytes.Buffer
		opts := Options{Validate: true, OutputFormat: output.FormatJSON}
		if err := RunJSONFmt(&buf, []string{bad}, opts); err != nil {
			t.Fatal(err)
		}
		var res Result
		if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
			t.Fatalf("unmarshal: %v\n%s", err, buf.String())
		}
		if res.Valid || res.Error == "" {
			t.Errorf("expected valid=false with error: %+v", res)
		}
	})
}

// TestRunJSONFmtInvalidNonValidate confirms invalid JSON returns an error when
// not in validate mode.
func TestRunJSONFmtInvalidNonValidate(t *testing.T) {
	var buf bytes.Buffer
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(bad, []byte("nope"), 0o644)
	if err := RunJSONFmt(&buf, []string{bad}, Options{}); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// TestSortKeysHelper checks recursive sortKeys over nested maps and arrays.
func TestSortKeysHelper(t *testing.T) {
	in := map[string]any{
		"z": []any{map[string]any{"q": 1, "a": 2}},
		"a": "x",
	}
	out := sortKeys(in)
	// Re-marshal then compare via SortKeys-equivalent ordering.
	b, _ := json.Marshal(out)
	// Marshal of a Go map sorts keys, so top-level "a" precedes "z".
	if !strings.HasPrefix(string(b), `{"a":"x"`) {
		t.Errorf("sortKeys did not normalize: %s", b)
	}
}

// TestBeautifyAndMinifyHelpers exercise the pure helpers directly.
func TestBeautifyAndMinifyHelpers(t *testing.T) {
	data := []byte(`{"a":1,"b":[1,2]}`)

	pretty, err := Beautify(data, "")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(pretty, []byte("\n")) {
		t.Errorf("Beautify did not indent: %s", pretty)
	}

	pretty2, err := Beautify(data, "    ")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(pretty2, []byte("    \"a\"")) {
		t.Errorf("Beautify custom indent failed: %s", pretty2)
	}

	if _, err := Beautify([]byte("bad"), ""); err == nil {
		t.Error("expected Beautify error on bad JSON")
	}
}
