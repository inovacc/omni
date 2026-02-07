package grep

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func generateTestFile(b *testing.B, lines int) string {
	b.Helper()
	dir := b.TempDir()
	path := filepath.Join(dir, "bench.txt")
	f, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < lines; i++ {
		if i%10 == 0 {
			_, _ = fmt.Fprintf(f, "line %d: ERROR something went wrong\n", i)
		} else {
			_, _ = fmt.Fprintf(f, "line %d: INFO everything is fine\n", i)
		}
	}
	_ = f.Close()
	return path
}

func BenchmarkRunGrep_SmallFile(b *testing.B) {
	path := generateTestFile(b, 100)
	var buf bytes.Buffer
	opts := GrepOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunGrep(&buf, nil, "ERROR", []string{path}, opts)
	}
}

func BenchmarkRunGrep_LargeFile(b *testing.B) {
	path := generateTestFile(b, 100_000)
	var buf bytes.Buffer
	opts := GrepOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunGrep(&buf, nil, "ERROR", []string{path}, opts)
	}
}

func BenchmarkRunGrep_CaseInsensitive(b *testing.B) {
	path := generateTestFile(b, 100_000)
	var buf bytes.Buffer
	opts := GrepOptions{IgnoreCase: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunGrep(&buf, nil, "error", []string{path}, opts)
	}
}

func BenchmarkRunGrep_Regex(b *testing.B) {
	path := generateTestFile(b, 100_000)
	var buf bytes.Buffer
	opts := GrepOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunGrep(&buf, nil, `line \d+: ERROR .+ wrong`, []string{path}, opts)
	}
}

func BenchmarkRunGrep_FixedString(b *testing.B) {
	path := generateTestFile(b, 100_000)
	var buf bytes.Buffer
	opts := GrepOptions{FixedStrings: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunGrep(&buf, nil, "ERROR something went wrong", []string{path}, opts)
	}
}

func BenchmarkRunGrep_Count(b *testing.B) {
	path := generateTestFile(b, 100_000)
	var buf bytes.Buffer
	opts := GrepOptions{Count: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunGrep(&buf, nil, "ERROR", []string{path}, opts)
	}
}

func BenchmarkRunGrep_Context(b *testing.B) {
	path := generateTestFile(b, 100_000)
	var buf bytes.Buffer
	opts := GrepOptions{Context: 3}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunGrep(&buf, nil, "ERROR", []string{path}, opts)
	}
}
