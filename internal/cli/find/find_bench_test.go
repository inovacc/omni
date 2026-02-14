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
	for d := range dirs {
		dir := filepath.Join(root, fmt.Sprintf("dir_%d", d))
		if err := os.MkdirAll(dir, 0755); err != nil {
			b.Fatal(err)
		}

		for f := range filesPerDir {
			ext := ".txt"

			switch f % 3 {
			case 0:
				ext = ".go"
			case 1:
				ext = ".json"
			}

			path := filepath.Join(dir, fmt.Sprintf("file_%d%s", f, ext))
			if err := os.WriteFile(path, fmt.Appendf(nil, "content %d", f), 0644); err != nil {
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
	for d := range depth {
		current = filepath.Join(current, fmt.Sprintf("level_%d", d))
		if err := os.MkdirAll(current, 0755); err != nil {
			b.Fatal(err)
		}

		for f := range filesPerLevel {
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

	for b.Loop() {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_TypeFile(b *testing.B) {
	root := createBenchDirTree(b, 20, 50)

	var buf bytes.Buffer

	opts := FindOptions{Type: "f"}

	for b.Loop() {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_TypeDirectory(b *testing.B) {
	root := createBenchDirTree(b, 20, 50)

	var buf bytes.Buffer

	opts := FindOptions{Type: "d"}

	for b.Loop() {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_DeepTree(b *testing.B) {
	root := createDeepDirTree(b, 20, 10)

	var buf bytes.Buffer

	opts := FindOptions{}

	for b.Loop() {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_MaxDepth(b *testing.B) {
	root := createDeepDirTree(b, 20, 10)

	var buf bytes.Buffer

	opts := FindOptions{MaxDepth: 5}

	for b.Loop() {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_Empty(b *testing.B) {
	root := createBenchDirTree(b, 20, 50)
	// Create some empty files and dirs
	for i := range 10 {
		path := filepath.Join(root, fmt.Sprintf("empty_%d.txt", i))
		_ = os.WriteFile(path, []byte{}, 0644)
		_ = os.MkdirAll(filepath.Join(root, fmt.Sprintf("emptydir_%d", i)), 0755)
	}

	var buf bytes.Buffer

	opts := FindOptions{Empty: true}

	for b.Loop() {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}

func BenchmarkRunFind_LargeTree(b *testing.B) {
	root := createBenchDirTree(b, 50, 100)

	var buf bytes.Buffer

	opts := FindOptions{}

	for b.Loop() {
		buf.Reset()
		_ = RunFind(&buf, []string{root}, opts)
	}
}
