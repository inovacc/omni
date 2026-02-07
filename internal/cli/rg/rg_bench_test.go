package rg

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func generateBenchDir(b *testing.B, files, linesPerFile int) string {
	b.Helper()
	dir := b.TempDir()
	for f := 0; f < files; f++ {
		path := filepath.Join(dir, fmt.Sprintf("file_%d.go", f))
		fh, err := os.Create(path)
		if err != nil {
			b.Fatal(err)
		}
		for i := 0; i < linesPerFile; i++ {
			if i%20 == 0 {
				_, _ = fmt.Fprintf(fh, "func handleError_%d() { // ERROR: something failed\n", i)
			} else {
				_, _ = fmt.Fprintf(fh, "func process_%d() { // INFO: processing data\n", i)
			}
		}
		_ = fh.Close()
	}
	return dir
}

func generateBenchFile(b *testing.B, lines int) string {
	b.Helper()
	dir := b.TempDir()
	path := filepath.Join(dir, "bench.go")
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

func BenchmarkRg_SearchFile_Regex(b *testing.B) {
	path := generateBenchFile(b, 100_000)
	dir := filepath.Dir(path)
	var buf bytes.Buffer
	opts := Options{LineNumber: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Run(&buf, `ERROR .+ wrong`, []string{dir}, opts)
	}
}

func BenchmarkRg_SearchFile_Literal(b *testing.B) {
	path := generateBenchFile(b, 100_000)
	dir := filepath.Dir(path)
	var buf bytes.Buffer
	opts := Options{Fixed: true, LineNumber: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Run(&buf, "ERROR something went wrong", []string{dir}, opts)
	}
}

func BenchmarkRg_SearchDir_Small(b *testing.B) {
	dir := generateBenchDir(b, 10, 100)
	var buf bytes.Buffer
	opts := Options{LineNumber: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Run(&buf, "ERROR", []string{dir}, opts)
	}
}

func BenchmarkRg_SearchDir_Large(b *testing.B) {
	dir := generateBenchDir(b, 50, 1000)
	var buf bytes.Buffer
	opts := Options{LineNumber: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Run(&buf, "ERROR", []string{dir}, opts)
	}
}

func BenchmarkRg_SearchDir_Parallel(b *testing.B) {
	dir := generateBenchDir(b, 50, 1000)
	var buf bytes.Buffer
	opts := Options{LineNumber: true, Threads: 4}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Run(&buf, "ERROR", []string{dir}, opts)
	}
}

func BenchmarkRg_FilesWithMatch(b *testing.B) {
	dir := generateBenchDir(b, 50, 1000)
	var buf bytes.Buffer
	opts := Options{FilesWithMatch: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Run(&buf, "ERROR", []string{dir}, opts)
	}
}

func BenchmarkRg_Count(b *testing.B) {
	dir := generateBenchDir(b, 50, 1000)
	var buf bytes.Buffer
	opts := Options{Count: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Run(&buf, "ERROR", []string{dir}, opts)
	}
}

func BenchmarkRg_TypeFilter(b *testing.B) {
	dir := generateBenchDir(b, 50, 1000)
	var buf bytes.Buffer
	opts := Options{Types: []string{"go"}, LineNumber: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = Run(&buf, "ERROR", []string{dir}, opts)
	}
}
