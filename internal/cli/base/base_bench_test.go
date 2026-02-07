package base

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func createBenchDataFile(b *testing.B, size int) string {
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

func createBenchEncodedFile(b *testing.B, size int, encoding string) string {
	b.Helper()
	dir := b.TempDir()
	raw := make([]byte, size)
	_, _ = rand.Read(raw)
	var encoded string
	switch encoding {
	case "base64":
		encoded = base64.StdEncoding.EncodeToString(raw)
	case "base32":
		encoded = base32.StdEncoding.EncodeToString(raw)
	}
	path := filepath.Join(dir, "bench.enc")
	if err := os.WriteFile(path, []byte(encoded), 0644); err != nil {
		b.Fatal(err)
	}
	return path
}

// Base64 encode benchmarks

func BenchmarkBase64_Encode_Small(b *testing.B) {
	path := createBenchDataFile(b, 1024) // 1KB
	var buf bytes.Buffer
	opts := BaseOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunBase64(&buf, []string{path}, opts)
	}
}

func BenchmarkBase64_Encode_Large(b *testing.B) {
	path := createBenchDataFile(b, 1024*1024) // 1MB
	var buf bytes.Buffer
	opts := BaseOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunBase64(&buf, []string{path}, opts)
	}
}

func BenchmarkBase64_Decode_Small(b *testing.B) {
	path := createBenchEncodedFile(b, 1024, "base64")
	var buf bytes.Buffer
	opts := BaseOptions{Decode: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunBase64(&buf, []string{path}, opts)
	}
}

func BenchmarkBase64_Decode_Large(b *testing.B) {
	path := createBenchEncodedFile(b, 1024*1024, "base64")
	var buf bytes.Buffer
	opts := BaseOptions{Decode: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunBase64(&buf, []string{path}, opts)
	}
}

func BenchmarkBase64_Encode_WithWrap(b *testing.B) {
	path := createBenchDataFile(b, 1024*1024)
	var buf bytes.Buffer
	opts := BaseOptions{Wrap: 76}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunBase64(&buf, []string{path}, opts)
	}
}

// Base32 encode benchmarks

func BenchmarkBase32_Encode_Small(b *testing.B) {
	path := createBenchDataFile(b, 1024)
	var buf bytes.Buffer
	opts := BaseOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunBase32(&buf, []string{path}, opts)
	}
}

func BenchmarkBase32_Encode_Large(b *testing.B) {
	path := createBenchDataFile(b, 1024*1024)
	var buf bytes.Buffer
	opts := BaseOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunBase32(&buf, []string{path}, opts)
	}
}

func BenchmarkBase32_Decode_Small(b *testing.B) {
	path := createBenchEncodedFile(b, 1024, "base32")
	var buf bytes.Buffer
	opts := BaseOptions{Decode: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunBase32(&buf, []string{path}, opts)
	}
}

// Base58 encode benchmarks

func BenchmarkBase58_Encode_Small(b *testing.B) {
	path := createBenchDataFile(b, 1024)
	var buf bytes.Buffer
	opts := BaseOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunBase58(&buf, []string{path}, opts)
	}
}

func BenchmarkBase58_Encode_Large(b *testing.B) {
	path := createBenchDataFile(b, 64*1024) // 64KB - base58 is slow on large data
	var buf bytes.Buffer
	opts := BaseOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunBase58(&buf, []string{path}, opts)
	}
}
