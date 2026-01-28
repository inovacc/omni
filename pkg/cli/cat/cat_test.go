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

	t.Run("files concatenated in order", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "order1.txt")
		file2 := filepath.Join(tmpDir, "order2.txt")
		file3 := filepath.Join(tmpDir, "order3.txt")

		_ = os.WriteFile(file1, []byte("AAA\n"), 0644)
		_ = os.WriteFile(file2, []byte("BBB\n"), 0644)
		_ = os.WriteFile(file3, []byte("CCC\n"), 0644)

		var buf bytes.Buffer

		_ = RunCat(&buf, []string{file1, file2, file3}, CatOptions{})

		output := buf.String()
		aPos := strings.Index(output, "AAA")
		bPos := strings.Index(output, "BBB")
		cPos := strings.Index(output, "CCC")

		if aPos >= bPos || bPos >= cPos {
			t.Errorf("RunCat() files not concatenated in order: %v", output)
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

	t.Run("number all lines includes blanks", func(t *testing.T) {
		file := filepath.Join(tmpDir, "numbered_blanks.txt")

		if err := os.WriteFile(file, []byte("line1\n\nline3\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunCat(&buf, []string{file}, CatOptions{NumberAll: true})

		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		if len(lines) != 3 {
			t.Errorf("RunCat() NumberAll should number blank lines too, got %d lines", len(lines))
		}

		// Check that line 3 has number 3
		if !strings.Contains(lines[2], "3") {
			t.Errorf("RunCat() NumberAll line 3 should be numbered 3: %v", lines[2])
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

	t.Run("number non-blank continues numbering", func(t *testing.T) {
		file := filepath.Join(tmpDir, "nonblank_continue.txt")

		if err := os.WriteFile(file, []byte("a\n\nb\n\nc\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunCat(&buf, []string{file}, CatOptions{NumberNonBlank: true})

		// Should have numbers 1, 2, 3 for non-blank lines
		output := buf.String()
		if !strings.Contains(output, "1") || !strings.Contains(output, "2") || !strings.Contains(output, "3") {
			t.Errorf("RunCat() NumberNonBlank should continue numbering: %v", output)
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

	t.Run("show ends on empty lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "ends_empty.txt")

		if err := os.WriteFile(file, []byte("line1\n\nline3\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunCat(&buf, []string{file}, CatOptions{ShowEnds: true})

		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		// Empty line should just be "$"
		if lines[1] != "$" {
			t.Errorf("RunCat() ShowEnds empty line should be just $: %v", lines[1])
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

	t.Run("show multiple tabs", func(t *testing.T) {
		file := filepath.Join(tmpDir, "multitabs.txt")

		if err := os.WriteFile(file, []byte("a\tb\tc\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunCat(&buf, []string{file}, CatOptions{ShowTabs: true})

		count := strings.Count(buf.String(), "^I")
		if count != 2 {
			t.Errorf("RunCat() ShowTabs should show 2 ^I, got %d", count)
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

	t.Run("squeeze preserves single blank", func(t *testing.T) {
		file := filepath.Join(tmpDir, "squeeze_single.txt")

		if err := os.WriteFile(file, []byte("line1\n\n\n\nline2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunCat(&buf, []string{file}, CatOptions{SqueezeBlank: true})

		// Should have exactly one blank line between line1 and line2
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		if len(lines) != 3 {
			t.Errorf("RunCat() SqueezeBlank should have 3 lines (line1, blank, line2), got %d", len(lines))
		}
	})

	t.Run("show non-printable characters", func(t *testing.T) {
		file := filepath.Join(tmpDir, "nonprint.txt")

		// Write a file with control character (bell = 0x07)
		if err := os.WriteFile(file, []byte("hello\x07world\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunCat(&buf, []string{file}, CatOptions{ShowNonPrint: true})

		// Control char 0x07 should be displayed as ^G
		if !strings.Contains(buf.String(), "^G") {
			t.Errorf("RunCat() ShowNonPrint should show ^G for bell: %v", buf.String())
		}
	})

	t.Run("combined options", func(t *testing.T) {
		file := filepath.Join(tmpDir, "combined.txt")

		if err := os.WriteFile(file, []byte("line1\t\n\n\nline2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunCat(&buf, []string{file}, CatOptions{
			NumberAll:    true,
			ShowEnds:     true,
			ShowTabs:     true,
			SqueezeBlank: true,
		})

		output := buf.String()
		// Should have line numbers, ^I for tabs, and $ for ends
		if !strings.Contains(output, "^I") {
			t.Errorf("RunCat() combined should show tabs")
		}

		if !strings.Contains(output, "$") {
			t.Errorf("RunCat() combined should show ends")
		}

		if !strings.Contains(output, "1") {
			t.Errorf("RunCat() combined should show line numbers")
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

	t.Run("file with only newlines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "newlines.txt")

		if err := os.WriteFile(file, []byte("\n\n\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunCat(&buf, []string{file}, CatOptions{})

		lines := strings.Split(buf.String(), "\n")
		if len(lines) < 3 {
			t.Errorf("RunCat() should preserve newlines")
		}
	})

	t.Run("binary file content", func(t *testing.T) {
		file := filepath.Join(tmpDir, "binary.txt")
		content := []byte{0x00, 0x01, 0x02, 'h', 'e', 'l', 'l', 'o', '\n'}

		if err := os.WriteFile(file, content, 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCat(&buf, []string{file}, CatOptions{})
		// Should not error on binary content
		if err != nil {
			t.Fatalf("RunCat() error = %v", err)
		}
	})

	t.Run("very long line", func(t *testing.T) {
		file := filepath.Join(tmpDir, "longline.txt")
		longLine := strings.Repeat("x", 10000) + "\n"

		if err := os.WriteFile(file, []byte(longLine), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCat(&buf, []string{file}, CatOptions{})
		if err != nil {
			t.Fatalf("RunCat() error = %v", err)
		}

		if len(strings.TrimSpace(buf.String())) != 10000 {
			t.Errorf("RunCat() should preserve long lines")
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unicode.txt")
		content := "Hello 世界 🌍\n日本語テスト\n"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunCat(&buf, []string{file}, CatOptions{})
		if err != nil {
			t.Fatalf("RunCat() error = %v", err)
		}

		if !strings.Contains(buf.String(), "世界") || !strings.Contains(buf.String(), "🌍") {
			t.Errorf("RunCat() should preserve unicode: %v", buf.String())
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

	t.Run("empty input", func(t *testing.T) {
		input := bytes.NewBufferString("")
		var output bytes.Buffer

		err := Cat(&output, input)
		if err != nil {
			t.Fatalf("Cat() error = %v", err)
		}

		if output.Len() != 0 {
			t.Errorf("Cat() empty input should produce empty output")
		}
	})

	t.Run("large input", func(t *testing.T) {
		largeContent := strings.Repeat("data", 10000)
		input := bytes.NewBufferString(largeContent)
		var output bytes.Buffer

		err := Cat(&output, input)
		if err != nil {
			t.Fatalf("Cat() error = %v", err)
		}

		if output.String() != largeContent {
			t.Errorf("Cat() should copy large content exactly")
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

	t.Run("empty input", func(t *testing.T) {
		input := bytes.NewBufferString("")

		lines, err := ReadFrom(input)
		if err != nil {
			t.Fatalf("ReadFrom() error = %v", err)
		}

		if len(lines) != 0 {
			t.Errorf("ReadFrom() empty input should return empty slice")
		}
	})

	t.Run("no trailing newline", func(t *testing.T) {
		input := bytes.NewBufferString("line1\nline2")

		lines, err := ReadFrom(input)
		if err != nil {
			t.Fatalf("ReadFrom() error = %v", err)
		}

		if len(lines) != 2 {
			t.Errorf("ReadFrom() got %d lines, want 2", len(lines))
		}
	})

	t.Run("single line", func(t *testing.T) {
		input := bytes.NewBufferString("single")

		lines, err := ReadFrom(input)
		if err != nil {
			t.Fatalf("ReadFrom() error = %v", err)
		}

		if len(lines) != 1 || lines[0] != "single" {
			t.Errorf("ReadFrom() = %v, want [single]", lines)
		}
	})
}

func TestWriteTo(t *testing.T) {
	t.Run("write lines", func(t *testing.T) {
		var buf bytes.Buffer
		lines := []string{"line1", "line2", "line3"}

		err := WriteTo(&buf, lines)
		if err != nil {
			t.Fatalf("WriteTo() error = %v", err)
		}

		expected := "line1\nline2\nline3\n"
		if buf.String() != expected {
			t.Errorf("WriteTo() = %v, want %v", buf.String(), expected)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		var buf bytes.Buffer

		err := WriteTo(&buf, []string{})
		if err != nil {
			t.Fatalf("WriteTo() error = %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("WriteTo() empty slice should produce no output")
		}
	})

	t.Run("single line", func(t *testing.T) {
		var buf bytes.Buffer

		err := WriteTo(&buf, []string{"single"})
		if err != nil {
			t.Fatalf("WriteTo() error = %v", err)
		}

		if buf.String() != "single\n" {
			t.Errorf("WriteTo() = %v, want 'single\\n'", buf.String())
		}
	})
}

func TestCatFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "catfiles_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("concatenate files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "f1.txt")
		file2 := filepath.Join(tmpDir, "f2.txt")

		_ = os.WriteFile(file1, []byte("first\n"), 0644)
		_ = os.WriteFile(file2, []byte("second\n"), 0644)

		var buf bytes.Buffer

		err := CatFiles(&buf, []string{file1, file2})
		if err != nil {
			t.Fatalf("CatFiles() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "first") || !strings.Contains(output, "second") {
			t.Errorf("CatFiles() = %v", output)
		}
	})
}
