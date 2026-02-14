package hash

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/internal/cli/output"
)

func createBenchFile(b *testing.B, size int) string {
	b.Helper()
	dir := b.TempDir()
	path := filepath.Join(dir, "bench.dat")
	data := make([]byte, size)

	_, _ = rand.Read(data)
	if err := os.WriteFile(path, data, 0644); err != nil {
		b.Fatal(err)
	}

	return path
}

func BenchmarkRunHash_MD5_Small(b *testing.B) {
	path := createBenchFile(b, 1024) // 1KB

	var buf bytes.Buffer

	opts := HashOptions{Algorithm: "md5"}

	for b.Loop() {
		buf.Reset()
		_ = RunHash(&buf, []string{path}, opts)
	}
}

func BenchmarkRunHash_MD5_Large(b *testing.B) {
	path := createBenchFile(b, 10*1024*1024) // 10MB

	var buf bytes.Buffer

	opts := HashOptions{Algorithm: "md5"}

	for b.Loop() {
		buf.Reset()
		_ = RunHash(&buf, []string{path}, opts)
	}
}

func BenchmarkRunHash_SHA256_Small(b *testing.B) {
	path := createBenchFile(b, 1024) // 1KB

	var buf bytes.Buffer

	opts := HashOptions{Algorithm: "sha256"}

	for b.Loop() {
		buf.Reset()
		_ = RunHash(&buf, []string{path}, opts)
	}
}

func BenchmarkRunHash_SHA256_Large(b *testing.B) {
	path := createBenchFile(b, 10*1024*1024) // 10MB

	var buf bytes.Buffer

	opts := HashOptions{Algorithm: "sha256"}

	for b.Loop() {
		buf.Reset()
		_ = RunHash(&buf, []string{path}, opts)
	}
}

func BenchmarkRunHash_SHA512_Small(b *testing.B) {
	path := createBenchFile(b, 1024) // 1KB

	var buf bytes.Buffer

	opts := HashOptions{Algorithm: "sha512"}

	for b.Loop() {
		buf.Reset()
		_ = RunHash(&buf, []string{path}, opts)
	}
}

func BenchmarkRunHash_SHA512_Large(b *testing.B) {
	path := createBenchFile(b, 10*1024*1024) // 10MB

	var buf bytes.Buffer

	opts := HashOptions{Algorithm: "sha512"}

	for b.Loop() {
		buf.Reset()
		_ = RunHash(&buf, []string{path}, opts)
	}
}

func BenchmarkRunHash_JSON(b *testing.B) {
	path := createBenchFile(b, 1024*1024) // 1MB

	var buf bytes.Buffer

	opts := HashOptions{Algorithm: "sha256", OutputFormat: output.FormatJSON}

	for b.Loop() {
		buf.Reset()
		_ = RunHash(&buf, []string{path}, opts)
	}
}

func BenchmarkRunHash_MultipleFiles(b *testing.B) {
	paths := make([]string, 10)
	for i := range paths {
		paths[i] = createBenchFile(b, 1024*1024) // 1MB each
	}

	var buf bytes.Buffer

	opts := HashOptions{Algorithm: "sha256"}

	for b.Loop() {
		buf.Reset()
		_ = RunHash(&buf, paths, opts)
	}
}
