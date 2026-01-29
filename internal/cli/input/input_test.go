package input

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenWithStdin(t *testing.T) {
	reader := strings.NewReader("test content")

	sources, err := Open(nil, reader)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}

	if sources[0].Name != "standard input" {
		t.Errorf("expected name 'standard input', got %q", sources[0].Name)
	}

	if sources[0].Reader != reader {
		t.Error("expected reader to be the provided reader")
	}
}

func TestOpenWithDash(t *testing.T) {
	reader := strings.NewReader("test content")

	sources, err := Open([]string{"-"}, reader)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}

	if sources[0].Name != "standard input" {
		t.Errorf("expected name 'standard input', got %q", sources[0].Name)
	}
}

func TestOpenWithFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test file content")

	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	sources, err := Open([]string{tmpFile}, nil)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer CloseAll(sources)

	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}

	if sources[0].Name != tmpFile {
		t.Errorf("expected name %q, got %q", tmpFile, sources[0].Name)
	}

	// Read the content
	data, err := io.ReadAll(sources[0].Reader)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("expected %q, got %q", content, data)
	}
}

func TestOpenWithMixedArgs(t *testing.T) {
	// Create temporary files
	tmpDir := t.TempDir()
	tmpFile1 := filepath.Join(tmpDir, "test1.txt")
	tmpFile2 := filepath.Join(tmpDir, "test2.txt")

	if err := os.WriteFile(tmpFile1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if err := os.WriteFile(tmpFile2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	reader := strings.NewReader("stdin content")

	sources, err := Open([]string{tmpFile1, "-", tmpFile2}, reader)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer CloseAll(sources)

	if len(sources) != 3 {
		t.Fatalf("expected 3 sources, got %d", len(sources))
	}

	// First source is file
	if sources[0].Name != tmpFile1 {
		t.Errorf("expected name %q, got %q", tmpFile1, sources[0].Name)
	}

	// Second source is stdin
	if sources[1].Name != "standard input" {
		t.Errorf("expected name 'standard input', got %q", sources[1].Name)
	}

	// Third source is file
	if sources[2].Name != tmpFile2 {
		t.Errorf("expected name %q, got %q", tmpFile2, sources[2].Name)
	}
}

func TestOpenNonexistentFile(t *testing.T) {
	sources, err := Open([]string{"/nonexistent/file/path"}, nil)
	if err == nil {
		CloseAll(sources)
		t.Error("expected error for nonexistent file")
	}
}

func TestOpenOne(t *testing.T) {
	reader := strings.NewReader("test")

	// Test with empty args
	src, err := OpenOne(nil, reader)
	if err != nil {
		t.Fatalf("OpenOne() error = %v", err)
	}

	if src.Name != "standard input" {
		t.Errorf("expected name 'standard input', got %q", src.Name)
	}

	// Test with dash
	src, err = OpenOne([]string{"-"}, reader)
	if err != nil {
		t.Fatalf("OpenOne() error = %v", err)
	}

	if src.Name != "standard input" {
		t.Errorf("expected name 'standard input', got %q", src.Name)
	}
}

func TestSourceClose(t *testing.T) {
	// Test closing a source with no closer
	src := Source{Reader: strings.NewReader("test"), Name: "test"}
	if err := src.Close(); err != nil {
		t.Errorf("Close() on source without closer should not error: %v", err)
	}

	// Test closing a source with closer
	closed := false

	src = Source{
		Reader: strings.NewReader("test"),
		Name:   "test",
		Closer: func() error {
			closed = true
			return nil
		},
	}
	if err := src.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if !closed {
		t.Error("closer should have been called")
	}
}
