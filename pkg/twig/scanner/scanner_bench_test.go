package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func createBenchScanDir(b *testing.B, dirs, filesPerDir int) string {
	b.Helper()

	root := b.TempDir()
	for d := range dirs {
		dir := filepath.Join(root, fmt.Sprintf("dir_%d", d))
		if err := os.MkdirAll(dir, 0755); err != nil {
			b.Fatal(err)
		}

		for f := range filesPerDir {
			path := filepath.Join(dir, fmt.Sprintf("file_%d.txt", f))

			data := fmt.Sprintf("content for file %d in dir %d", f, d)
			if err := os.WriteFile(path, []byte(data), 0644); err != nil {
				b.Fatal(err)
			}
		}
	}

	return root
}

func createDeepScanDir(b *testing.B, depth, filesPerLevel int) string {
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

func BenchmarkScan_Shallow(b *testing.B) {
	root := createBenchScanDir(b, 5, 20)
	config := DefaultConfig()
	config.ShowHidden = true
	s := NewScanner(config)
	ctx := context.Background()

	for b.Loop() {
		_, _ = s.Scan(ctx, root)
	}
}

func BenchmarkScan_Deep(b *testing.B) {
	root := createDeepScanDir(b, 15, 10)
	config := DefaultConfig()
	config.ShowHidden = true
	s := NewScanner(config)
	ctx := context.Background()

	for b.Loop() {
		_, _ = s.Scan(ctx, root)
	}
}

func BenchmarkScan_Large(b *testing.B) {
	root := createBenchScanDir(b, 50, 100)
	config := DefaultConfig()
	config.ShowHidden = true
	s := NewScanner(config)
	ctx := context.Background()

	for b.Loop() {
		_, _ = s.Scan(ctx, root)
	}
}

func BenchmarkScan_WithHash(b *testing.B) {
	root := createBenchScanDir(b, 20, 50)
	config := DefaultConfig()
	config.ShowHidden = true
	config.ShowHash = true
	s := NewScanner(config)
	ctx := context.Background()

	for b.Loop() {
		_, _ = s.Scan(ctx, root)
	}
}

func BenchmarkScan_MaxFiles(b *testing.B) {
	root := createBenchScanDir(b, 50, 100)
	config := DefaultConfig()
	config.ShowHidden = true
	config.MaxFiles = 500
	s := NewScanner(config)
	ctx := context.Background()

	for b.Loop() {
		_, _ = s.Scan(ctx, root)
	}
}

func BenchmarkScan_Parallel_2(b *testing.B) {
	root := createBenchScanDir(b, 50, 100)
	config := DefaultConfig()
	config.ShowHidden = true
	config.Parallel = 2
	s := NewScanner(config)
	ctx := context.Background()

	for b.Loop() {
		_, _ = s.Scan(ctx, root)
	}
}

func BenchmarkScan_Parallel_4(b *testing.B) {
	root := createBenchScanDir(b, 50, 100)
	config := DefaultConfig()
	config.ShowHidden = true
	config.Parallel = 4
	s := NewScanner(config)
	ctx := context.Background()

	for b.Loop() {
		_, _ = s.Scan(ctx, root)
	}
}

func BenchmarkScan_Parallel_8(b *testing.B) {
	root := createBenchScanDir(b, 50, 100)
	config := DefaultConfig()
	config.ShowHidden = true
	config.Parallel = 8
	s := NewScanner(config)
	ctx := context.Background()

	for b.Loop() {
		_, _ = s.Scan(ctx, root)
	}
}

func BenchmarkScan_Sequential(b *testing.B) {
	root := createBenchScanDir(b, 50, 100)
	config := DefaultConfig()
	config.ShowHidden = true
	config.Parallel = 1
	s := NewScanner(config)
	ctx := context.Background()

	for b.Loop() {
		_, _ = s.Scan(ctx, root)
	}
}

func BenchmarkScan_DirsOnly(b *testing.B) {
	root := createBenchScanDir(b, 50, 100)
	config := DefaultConfig()
	config.ShowHidden = true
	config.DirsOnly = true
	s := NewScanner(config)
	ctx := context.Background()

	for b.Loop() {
		_, _ = s.Scan(ctx, root)
	}
}
