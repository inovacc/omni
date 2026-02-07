package jq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func generateSmallJSON() string {
	return `{"name":"test","age":30,"active":true}`
}

func generateNestedJSON() string {
	return `{"user":{"name":"test","address":{"city":"NYC","zip":"10001"},"tags":["a","b","c"]}}`
}

func generateArrayJSON(size int) string {
	items := make([]string, size)
	for i := 0; i < size; i++ {
		items[i] = fmt.Sprintf(`{"id":%d,"name":"item_%d","value":%d}`, i, i, i*10)
	}
	return "[" + strings.Join(items, ",") + "]"
}

func generateLargeJSON() string {
	obj := make(map[string]any)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("field_%d", i)
		obj[key] = map[string]any{
			"id":    i,
			"name":  fmt.Sprintf("item_%d", i),
			"value": i * 10,
			"tags":  []string{"a", "b", "c"},
		}
	}
	data, _ := json.Marshal(obj)
	return string(data)
}

func BenchmarkApplyJqFilter_Identity(b *testing.B) {
	input := generateSmallJSON()
	var parsed any
	_ = json.Unmarshal([]byte(input), &parsed)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ApplyJqFilter(parsed, ".")
	}
}

func BenchmarkApplyJqFilter_SimpleField(b *testing.B) {
	input := generateSmallJSON()
	var parsed any
	_ = json.Unmarshal([]byte(input), &parsed)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ApplyJqFilter(parsed, ".name")
	}
}

func BenchmarkApplyJqFilter_NestedField(b *testing.B) {
	input := generateNestedJSON()
	var parsed any
	_ = json.Unmarshal([]byte(input), &parsed)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ApplyJqFilter(parsed, ".user.address.city")
	}
}

func BenchmarkApplyJqFilter_ArrayIteration(b *testing.B) {
	input := generateArrayJSON(1000)
	var parsed any
	_ = json.Unmarshal([]byte(input), &parsed)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ApplyJqFilter(parsed, ".[]")
	}
}

func BenchmarkApplyJqFilter_ArrayIndex(b *testing.B) {
	input := generateArrayJSON(1000)
	var parsed any
	_ = json.Unmarshal([]byte(input), &parsed)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ApplyJqFilter(parsed, ".[500]")
	}
}

func BenchmarkApplyJqFilter_Pipe(b *testing.B) {
	input := generateArrayJSON(100)
	var parsed any
	_ = json.Unmarshal([]byte(input), &parsed)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ApplyJqFilter(parsed, ".[0] | .name")
	}
}

func BenchmarkApplyJqFilter_Keys(b *testing.B) {
	input := generateLargeJSON()
	var parsed any
	_ = json.Unmarshal([]byte(input), &parsed)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ApplyJqFilter(parsed, "keys")
	}
}

func BenchmarkApplyJqFilter_Length(b *testing.B) {
	input := generateArrayJSON(10000)
	var parsed any
	_ = json.Unmarshal([]byte(input), &parsed)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ApplyJqFilter(parsed, "length")
	}
}

func BenchmarkRunJq_SmallJSON(b *testing.B) {
	input := generateSmallJSON()
	var buf bytes.Buffer
	opts := JqOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJq(&buf, strings.NewReader(input), []string{".name"}, opts)
	}
}

func BenchmarkRunJq_LargeJSON(b *testing.B) {
	input := generateLargeJSON()
	var buf bytes.Buffer
	opts := JqOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJq(&buf, strings.NewReader(input), []string{"."}, opts)
	}
}

func BenchmarkRunJq_Compact(b *testing.B) {
	input := generateLargeJSON()
	var buf bytes.Buffer
	opts := JqOptions{Compact: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJq(&buf, strings.NewReader(input), []string{"."}, opts)
	}
}

func BenchmarkRunJq_SortKeys(b *testing.B) {
	input := generateLargeJSON()
	var buf bytes.Buffer
	opts := JqOptions{Sort: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunJq(&buf, strings.NewReader(input), []string{"."}, opts)
	}
}
