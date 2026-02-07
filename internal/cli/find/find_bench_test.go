package find

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func createBenchDirTree(b *testing.B, dirs, filesPerDir int) string {
	b.Helper()
	root := b.TempDir()
	for d := 0; d < dirs; d++ {
		dir := filepath.Join(root, fmt.Sprintf("dir_%d", d))
		if err := os.MkdirAll(dir, 0755); err != nil {
			b.Fatal(err)
		}
		for f := 0; f < filesPerDir; f++ {
			ext := ".txt"
			if f%3 == 0 {
				ext = ".go"
			} else if f%3 == 1 {
				ext = ".json"
			}
			path := filepath.Join(dir, fmt.Sprintf("file_%d%s", f, ext))
			if err := os.WriteFile(path, []byte(fmt.Sprintf("content %d", f)), 0644); err != nil {
				b.Fatal(err)
			}
		}
	}
	return root
}

func createDeepDirTree(b *testing.B, depth, filesPerLevel int) string {
	b.Helper()
	root := b.TempDir()
	current := root
	for d := 0; d < depth; d++ {
		current = filepath.Join(current, fmt.Sprintf("level_%d", d))
		if err := os.MkdirAll(current, 0755); err != nil {
			b.Fatal(err)
		}
		for f := 0; f < filesPerLevel; f++ {
			path := filepath.Join(current, fmt.Sprintf("file_%d.txt", f))
			if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
				b.Fatal(err)
			}
		}
	}
	return root
}

func BenchmarkRunFind_NamePattern(b *testing.B) {
	root := createBenchDirTree(b, 20, 50)
	var buf bytes.Buffer
	opts := FindOptions{Name: "*.go"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_TypeFile(b *testing.B) {
	root := createBenchDirTree(b, 20, 50)
	var buf bytes.Buffer
	opts := FindOptions{Type: "f"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_TypeDirectory(b *testing.B) {
	root := createBenchDirTree(b, 20, 50)
	var buf bytes.Buffer
	opts := FindOptions{Type: "d"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_DeepTree(b *testing.B) {
	root := createDeepDirTree(b, 20, 10)
	var buf bytes.Buffer
	opts := FindOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_MaxDepth(b *testing.B) {
	root := createDeepDirTree(b, 20, 10)
	var buf bytes.Buffer
	opts := FindOptions{MaxDepth: 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_Empty(b *testing.B) {
	root := createBenchDirTree(b, 20, 50)
	// Create some empty files and dirs
	for i := 0; i < 10; i++ {
		path := filepath.Join(root, fmt.Sprintf("empty_%d.txt", i))
		_ = os.WriteFile(path, []byte{}, 0644)
		_ = os.MkdirAll(filepath.Join(root, fmt.Sprintf("emptydir_%d", i)), 0755)
	}
	var buf bytes.Buffer
	opts := FindOptions{Empty: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_LargeTree(b *testing.B) {
	root := createBenchDirTree(b, 50, 100)
	var buf bytes.Buffer
	opts := FindOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}
