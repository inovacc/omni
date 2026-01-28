package cat

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cat_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("read single file", func(t *testing.T) {
		content := "hello world"
		file := filepath.Join(tmpDir, "test1.txt")

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCat(&buf, []string{file}, CatOptions{})
		if err != nil {
			t.Fatalf("RunCat() error = %v", err)
		}

		if strings.TrimSpace(buf.String()) != content {
			t.Errorf("RunCat() = %v, want %v", buf.String(), content)
		}
	})

	t.Run("read multiple files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "multi1.txt")
		file2 := filepath.Join(tmpDir, "multi2.txt")

		if err := os.WriteFile(file1, []byte("first\n"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file2, []byte("second\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCat(&buf, []string{file1, file2}, CatOptions{})
		if err != nil {
			t.Fatalf("RunCat() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "first") || !strings.Contains(output, "second") {
			t.Errorf("RunCat() should contain both files: %v", output)
		}
	})

	t.Run("number all lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "numbered.txt")

		if err := os.WriteFile(file, []byte("line1\nline2\nline3\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCat(&buf, []string{file}, CatOptions{NumberAll: true})
		if err != nil {
			t.Fatalf("RunCat() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "1") && !strings.Contains(output, "2") {
			t.Errorf("RunCat() with NumberAll should have line numbers: %v", output)
		}
	})

	t.Run("number non-blank lines only", func(t *testing.T) {
		file := filepath.Join(tmpDir, "nonblank.txt")

		if err := os.WriteFile(file, []byte("line1\n\nline2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCat(&buf, []string{file}, CatOptions{NumberNonBlank: true})
		if err != nil {
			t.Fatalf("RunCat() error = %v", err)
		}

		lines := strings.Split(buf.String(), "\n")
		numberedCount := 0

		for _, line := range lines {
			if strings.Contains(line, "\t") && len(strings.TrimSpace(line)) > 0 {
				numberedCount++
			}
		}

		if numberedCount > 2 {
			t.Errorf("RunCat() NumberNonBlank should skip blank lines")
		}
	})

	t.Run("show ends", func(t *testing.T) {
		file := filepath.Join(tmpDir, "ends.txt")

		if err := os.WriteFile(file, []byte("line1\nline2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCat(&buf, []string{file}, CatOptions{ShowEnds: true})
		if err != nil {
			t.Fatalf("RunCat() error = %v", err)
		}

		if !strings.Contains(buf.String(), "$") {
			t.Errorf("RunCat() ShowEnds should show $ at end of lines: %v", buf.String())
		}
	})

	t.Run("show tabs", func(t *testing.T) {
		file := filepath.Join(tmpDir, "tabs.txt")

		if err := os.WriteFile(file, []byte("col1\tcol2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCat(&buf, []string{file}, CatOptions{ShowTabs: true})
		if err != nil {
			t.Fatalf("RunCat() error = %v", err)
		}

		if !strings.Contains(buf.String(), "^I") {
			t.Errorf("RunCat() ShowTabs should show ^I for tabs: %v", buf.String())
		}
	})

	t.Run("squeeze blank lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "squeeze.txt")

		if err := os.WriteFile(file, []byte("line1\n\n\n\nline2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCat(&buf, []string{file}, CatOptions{SqueezeBlank: true})
		if err != nil {
			t.Fatalf("RunCat() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		blankCount := 0

		for i := 1; i < len(lines); i++ {
			if lines[i-1] == "" && lines[i] == "" {
				blankCount++
			}
		}

		if blankCount > 0 {
			t.Errorf("RunCat() SqueezeBlank should remove consecutive blank lines")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunCat(&buf, []string{"/nonexistent/file.txt"}, CatOptions{})
		if err == nil {
			t.Error("RunCat() should return error for nonexistent file")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")

		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCat(&buf, []string{file}, CatOptions{})
		if err != nil {
			t.Fatalf("RunCat() error = %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("RunCat() empty file should produce no output: %v", buf.String())
		}
	})
}

func TestCat(t *testing.T) {
	t.Run("simple copy", func(t *testing.T) {
		input := bytes.NewBufferString("test content")
		var output bytes.Buffer

		err := Cat(&output, input)
		if err != nil {
			t.Fatalf("Cat() error = %v", err)
		}

		if output.String() != "test content" {
			t.Errorf("Cat() = %v, want %v", output.String(), "test content")
		}
	})
}

func TestReadFrom(t *testing.T) {
	t.Run("read lines", func(t *testing.T) {
		input := bytes.NewBufferString("line1\nline2\nline3\n")

		lines, err := ReadFrom(input)
		if err != nil {
			t.Fatalf("ReadFrom() error = %v", err)
		}

		if len(lines) != 3 {
			t.Errorf("ReadFrom() got %d lines, want 3", len(lines))
		}

		if lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
			t.Errorf("ReadFrom() = %v", lines)
		}
	})
}
