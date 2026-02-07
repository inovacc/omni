package text

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createBenchTextFile(b *testing.B, content string) string {
	b.Helper()
	dir := b.TempDir()
	path := filepath.Join(dir, "bench.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		b.Fatal(err)
	}
	return path
}

func generateSortData(lines int) string {
	var sb strings.Builder
	words := []string{"zebra", "apple", "mango", "banana", "cherry", "date", "elderberry", "fig", "grape", "honeydew"}
	for i := 0; i < lines; i++ {
		_, _ = fmt.Fprintf(&sb, "%s %d\n", words[i%len(words)], i)
	}
	return sb.String()
}

func generateUniqData(lines int) string {
	var sb strings.Builder
	// Already sorted with duplicates
	words := []string{"apple", "apple", "apple", "banana", "banana", "cherry", "date", "date", "date", "date"}
	for i := 0; i < lines; i++ {
		_, _ = fmt.Fprintln(&sb, words[i%len(words)])
	}
	return sb.String()
}

func generateNumericData(lines int) string {
	var sb strings.Builder
	for i := lines; i > 0; i-- {
		_, _ = fmt.Fprintf(&sb, "%d\n", i)
	}
	return sb.String()
}

func BenchmarkRunSort_Small(b *testing.B) {
	path := createBenchTextFile(b, generateSortData(100))
	var buf bytes.Buffer
	opts := SortOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunSort(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunSort_Large(b *testing.B) {
	path := createBenchTextFile(b, generateSortData(100_000))
	var buf bytes.Buffer
	opts := SortOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunSort(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunSort_Numeric(b *testing.B) {
	path := createBenchTextFile(b, generateNumericData(100_000))
	var buf bytes.Buffer
	opts := SortOptions{Numeric: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunSort(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunSort_Reverse(b *testing.B) {
	path := createBenchTextFile(b, generateSortData(100_000))
	var buf bytes.Buffer
	opts := SortOptions{Reverse: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunSort(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunSort_Unique(b *testing.B) {
	path := createBenchTextFile(b, generateSortData(100_000))
	var buf bytes.Buffer
	opts := SortOptions{Unique: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunSort(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunSort_IgnoreCase(b *testing.B) {
	path := createBenchTextFile(b, generateSortData(100_000))
	var buf bytes.Buffer
	opts := SortOptions{IgnoreCase: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunSort(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunUniq_Small(b *testing.B) {
	path := createBenchTextFile(b, generateUniqData(100))
	var buf bytes.Buffer
	opts := UniqOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunUniq(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunUniq_Large(b *testing.B) {
	path := createBenchTextFile(b, generateUniqData(100_000))
	var buf bytes.Buffer
	opts := UniqOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunUniq(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunUniq_Count(b *testing.B) {
	path := createBenchTextFile(b, generateUniqData(100_000))
	var buf bytes.Buffer
	opts := UniqOptions{Count: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunUniq(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunUniq_IgnoreCase(b *testing.B) {
	path := createBenchTextFile(b, generateUniqData(100_000))
	var buf bytes.Buffer
	opts := UniqOptions{IgnoreCase: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunUniq(&buf, nil, []string{path}, opts)
	}
}
