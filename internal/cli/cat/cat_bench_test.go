package cat

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createBenchCatFile(b *testing.B, lines int) string {
	b.Helper()
	dir := b.TempDir()
	path := filepath.Join(dir, "bench.txt")
	f, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < lines; i++ {
		_, _ = fmt.Fprintf(f, "line %d: %s\n", i, strings.Repeat("x", 80))
	}
	_ = f.Close()
	return path
}

func BenchmarkRunCat_Small(b *testing.B) {
	path := createBenchCatFile(b, 100)
	var buf bytes.Buffer
	opts := CatOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunCat(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunCat_Large(b *testing.B) {
	path := createBenchCatFile(b, 100_000)
	var buf bytes.Buffer
	opts := CatOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunCat(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunCat_NumberAll(b *testing.B) {
	path := createBenchCatFile(b, 100_000)
	var buf bytes.Buffer
	opts := CatOptions{NumberAll: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunCat(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunCat_NumberNonBlank(b *testing.B) {
	// Create file with blank lines interspersed
	dir := b.TempDir()
	path := filepath.Join(dir, "bench_blank.txt")
	f, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 100_000; i++ {
		if i%5 == 0 {
			_, _ = fmt.Fprintln(f)
		} else {
			_, _ = fmt.Fprintf(f, "line %d: content\n", i)
		}
	}
	_ = f.Close()

	var buf bytes.Buffer
	opts := CatOptions{NumberNonBlank: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunCat(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunCat_ShowEnds(b *testing.B) {
	path := createBenchCatFile(b, 100_000)
	var buf bytes.Buffer
	opts := CatOptions{ShowEnds: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunCat(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunCat_SqueezeBlank(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "bench_squeeze.txt")
	f, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 100_000; i++ {
		if i%3 == 0 {
			_, _ = fmt.Fprintln(f)
		} else {
			_, _ = fmt.Fprintf(f, "line %d\n", i)
		}
	}
	_ = f.Close()

	var buf bytes.Buffer
	opts := CatOptions{SqueezeBlank: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunCat(&buf, nil, []string{path}, opts)
	}
}

func BenchmarkRunCat_MultipleFiles(b *testing.B) {
	paths := make([]string, 10)
	for i := range paths {
		paths[i] = createBenchCatFile(b, 10_000)
	}
	var buf bytes.Buffer
	opts := CatOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunCat(&buf, nil, paths, opts)
	}
}
