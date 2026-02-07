package jsonfmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createBenchJSONFile(b *testing.B, content string) string {
	b.Helper()
	dir := b.TempDir()
	path := filepath.Join(dir, "bench.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		b.Fatal(err)
	}
	return path
}

func generateSmallJSONObj() string {
	return `{"name":"test","age":30,"active":true,"tags":["a","b","c"]}`
}

func generateLargeJSONObj() string {
	obj := make(map[string]any)
	for i := 0; i < 500; i++ {
		key := fmt.Sprintf("field_%d", i)
		obj[key] = map[string]any{
			"id":      i,
			"name":    fmt.Sprintf("item_%d", i),
			"value":   i * 10,
			"tags":    []string{"alpha", "beta", "gamma"},
			"nested":  map[string]any{"x": i, "y": i * 2},
			"enabled": i%2 == 0,
		}
	}
	data, _ := json.Marshal(obj)
	return string(data)
}

func generateDeepJSON(depth int) string {
	var sb strings.Builder
	for i := 0; i < depth; i++ {
		_, _ = fmt.Fprintf(&sb, `{"level_%d":`, i)
	}
	sb.WriteString(`"leaf"`)
	for i := 0; i < depth; i++ {
		sb.WriteString("}")
	}
	return sb.String()
}

func BenchmarkBeautify_Small(b *testing.B) {
	path := createBenchJSONFile(b, generateSmallJSONObj())
	var buf bytes.Buffer
	opts := Options{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJSONFmt(&buf, []string{path}, opts)
	}
}

func BenchmarkBeautify_Large(b *testing.B) {
	path := createBenchJSONFile(b, generateLargeJSONObj())
	var buf bytes.Buffer
	opts := Options{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJSONFmt(&buf, []string{path}, opts)
	}
}

func BenchmarkMinify_Small(b *testing.B) {
	// Pre-beautify the JSON so minify has work to do
	pretty, _ := json.MarshalIndent(json.RawMessage(generateSmallJSONObj()), "", "  ")
	path := createBenchJSONFile(b, string(pretty))
	var buf bytes.Buffer
	opts := Options{Minify: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJSONFmt(&buf, []string{path}, opts)
	}
}

func BenchmarkMinify_Large(b *testing.B) {
	pretty, _ := json.MarshalIndent(json.RawMessage(generateLargeJSONObj()), "", "  ")
	path := createBenchJSONFile(b, string(pretty))
	var buf bytes.Buffer
	opts := Options{Minify: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJSONFmt(&buf, []string{path}, opts)
	}
}

func BenchmarkSortKeys_Large(b *testing.B) {
	path := createBenchJSONFile(b, generateLargeJSONObj())
	var buf bytes.Buffer
	opts := Options{SortKeys: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJSONFmt(&buf, []string{path}, opts)
	}
}

func BenchmarkValidate_Small(b *testing.B) {
	path := createBenchJSONFile(b, generateSmallJSONObj())
	var buf bytes.Buffer
	opts := Options{Validate: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJSONFmt(&buf, []string{path}, opts)
	}
}

func BenchmarkValidate_Large(b *testing.B) {
	path := createBenchJSONFile(b, generateLargeJSONObj())
	var buf bytes.Buffer
	opts := Options{Validate: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJSONFmt(&buf, []string{path}, opts)
	}
}

func BenchmarkDeepNesting(b *testing.B) {
	path := createBenchJSONFile(b, generateDeepJSON(50))
	var buf bytes.Buffer
	opts := Options{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJSONFmt(&buf, []string{path}, opts)
	}
}
