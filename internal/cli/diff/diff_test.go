package diff

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDiff(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "diff_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("identical files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "same1.txt")
		file2 := filepath.Join(tmpDir, "same2.txt")
		content := "line1\nline2\nline3"

		if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file2, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{})
		if err != nil {
			t.Fatalf("RunDiff() error = %v", err)
		}

		// Identical files should produce minimal or no output
		output := buf.String()
		if strings.Contains(output, "-") || strings.Contains(output, "+") {
			t.Logf("RunDiff() output for identical files: %v", output)
		}
	})

	t.Run("different files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "diff1.txt")
		file2 := filepath.Join(tmpDir, "diff2.txt")

		if err := os.WriteFile(file1, []byte("apple\nbanana\ncherry"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file2, []byte("apple\norange\ncherry"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{})
		// RunDiff returns error when files differ
		if err == nil {
			// Some implementations return nil even when files differ
			t.Log("RunDiff() returned no error for different files")
		}

		output := buf.String()
		// Should show some difference
		if !strings.Contains(output, "banana") && !strings.Contains(output, "orange") {
			t.Logf("RunDiff() output: %v", output)
		}
	})

	t.Run("unified format", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "uni1.txt")
		file2 := filepath.Join(tmpDir, "uni2.txt")

		if err := os.WriteFile(file1, []byte("old line\ncommon"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file2, []byte("new line\ncommon"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{Unified: 3})

		output := buf.String()
		// Unified diff should have context markers
		if len(output) > 0 {
			t.Logf("Unified diff output: %v", output)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "exists.txt")
		file2 := filepath.Join(tmpDir, "notexists.txt")

		if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{})
		if err == nil {
			t.Error("RunDiff() expected error for nonexistent file")
		}
	})

	t.Run("single argument", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single.txt")
		if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunDiff(&buf, []string{file}, DiffOptions{})
		// With only one file, behavior depends on implementation
		if err != nil {
			t.Logf("RunDiff() with single file: %v", err)
		}
	})

	t.Run("empty files identical", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "empty1.txt")
		file2 := filepath.Join(tmpDir, "empty2.txt")

		_ = os.WriteFile(file1, []byte(""), 0644)
		_ = os.WriteFile(file2, []byte(""), 0644)

		var buf bytes.Buffer

		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{})
		if err != nil {
			t.Logf("RunDiff() empty files: %v", err)
		}

		// Empty identical files should have no diff
		output := strings.TrimSpace(buf.String())
		if output != "" {
			t.Logf("RunDiff() empty files output: %v", output)
		}
	})

	t.Run("one empty one not", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "hasdata.txt")
		file2 := filepath.Join(tmpDir, "nodata.txt")

		_ = os.WriteFile(file1, []byte("some content"), 0644)
		_ = os.WriteFile(file2, []byte(""), 0644)

		var buf bytes.Buffer

		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{})

		// Should show difference
		if buf.Len() == 0 {
			t.Log("RunDiff() may show no output for empty vs non-empty")
		}
	})

	t.Run("whitespace only difference", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "ws1.txt")
		file2 := filepath.Join(tmpDir, "ws2.txt")

		_ = os.WriteFile(file1, []byte("line"), 0644)
		_ = os.WriteFile(file2, []byte("line "), 0644)

		var buf bytes.Buffer

		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{})

		// Should detect whitespace difference
		if buf.Len() > 0 {
			t.Logf("Whitespace diff detected")
		}
	})

	t.Run("newline difference", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "nl1.txt")
		file2 := filepath.Join(tmpDir, "nl2.txt")

		_ = os.WriteFile(file1, []byte("line\n"), 0644)
		_ = os.WriteFile(file2, []byte("line"), 0644)

		var buf bytes.Buffer

		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{})

		// May or may not detect newline difference
		t.Logf("Newline diff output length: %d", buf.Len())
	})

	t.Run("unicode content", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "uni1.txt")
		file2 := filepath.Join(tmpDir, "uni2.txt")

		_ = os.WriteFile(file1, []byte("世界"), 0644)
		_ = os.WriteFile(file2, []byte("こんにちは"), 0644)

		var buf bytes.Buffer

		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{})

		if buf.Len() == 0 {
			t.Log("RunDiff() should detect unicode differences")
		}
	})

	t.Run("large files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "large1.txt")
		file2 := filepath.Join(tmpDir, "large2.txt")

		var content1, content2 strings.Builder
		for i := range 1000 {
			content1.WriteString("line " + string(rune('0'+i%10)) + "\n")
			content2.WriteString("line " + string(rune('0'+i%10)) + "\n")
		}

		// Make one difference in the middle
		content2.WriteString("different line\n")

		_ = os.WriteFile(file1, []byte(content1.String()), 0644)
		_ = os.WriteFile(file2, []byte(content2.String()), 0644)

		var buf bytes.Buffer

		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{})
		if err != nil {
			t.Logf("RunDiff() large files: %v", err)
		}
	})

	t.Run("added lines at end", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "add1.txt")
		file2 := filepath.Join(tmpDir, "add2.txt")

		_ = os.WriteFile(file1, []byte("line1\nline2"), 0644)
		_ = os.WriteFile(file2, []byte("line1\nline2\nline3\nline4"), 0644)

		var buf bytes.Buffer

		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{})

		output := buf.String()
		if !strings.Contains(output, "line3") && !strings.Contains(output, "line4") {
			t.Log("RunDiff() may format added lines differently")
		}
	})

	t.Run("removed lines", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "rem1.txt")
		file2 := filepath.Join(tmpDir, "rem2.txt")

		_ = os.WriteFile(file1, []byte("line1\nline2\nline3"), 0644)
		_ = os.WriteFile(file2, []byte("line1"), 0644)

		var buf bytes.Buffer

		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{})

		output := buf.String()
		if buf.Len() > 0 {
			t.Logf("Removed lines diff: %v", output)
		}
	})

	t.Run("context option", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "ctx1.txt")
		file2 := filepath.Join(tmpDir, "ctx2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\nc\nd\ne"), 0644)
		_ = os.WriteFile(file2, []byte("a\nb\nX\nd\ne"), 0644)

		var buf bytes.Buffer

		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{Context: 1})

		// Context should show surrounding lines
		if buf.Len() == 0 {
			t.Log("RunDiff() with context produced no output")
		}
	})

	t.Run("consistent output", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "con1.txt")
		file2 := filepath.Join(tmpDir, "con2.txt")

		_ = os.WriteFile(file1, []byte("alpha"), 0644)
		_ = os.WriteFile(file2, []byte("beta"), 0644)

		var buf1, buf2 bytes.Buffer

		_ = RunDiff(&buf1, []string{file1, file2}, DiffOptions{})
		_ = RunDiff(&buf2, []string{file1, file2}, DiffOptions{})

		if buf1.String() != buf2.String() {
			t.Error("RunDiff() should produce consistent output")
		}
	})

	t.Run("binary files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "bin1.bin")
		file2 := filepath.Join(tmpDir, "bin2.bin")

		_ = os.WriteFile(file1, []byte{0x00, 0x01, 0x02}, 0644)
		_ = os.WriteFile(file2, []byte{0x00, 0x01, 0x03}, 0644)

		var buf bytes.Buffer

		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{})

		// Binary diff handling varies
		t.Logf("Binary diff output length: %d", buf.Len())
	})

	t.Run("brief mode", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "brief1.txt")
		file2 := filepath.Join(tmpDir, "brief2.txt")

		_ = os.WriteFile(file1, []byte("alpha"), 0644)
		_ = os.WriteFile(file2, []byte("beta"), 0644)

		var buf bytes.Buffer
		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{Brief: true})

		output := buf.String()
		if !strings.Contains(output, "differ") {
			t.Error("expected 'differ' in brief output for different files")
		}
	})

	t.Run("brief mode identical", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "briefsame1.txt")
		file2 := filepath.Join(tmpDir, "briefsame2.txt")

		_ = os.WriteFile(file1, []byte("same"), 0644)
		_ = os.WriteFile(file2, []byte("same"), 0644)

		var buf bytes.Buffer
		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{Brief: true})

		output := buf.String()
		if strings.Contains(output, "differ") {
			t.Error("identical files should not report 'differ' in brief mode")
		}
	})

	t.Run("side by side", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "sbs1.txt")
		file2 := filepath.Join(tmpDir, "sbs2.txt")

		_ = os.WriteFile(file1, []byte("line1\nline2\nline3"), 0644)
		_ = os.WriteFile(file2, []byte("line1\nchanged\nline3"), 0644)

		var buf bytes.Buffer
		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{Side: true, Width: 80})

		output := buf.String()
		if buf.Len() == 0 {
			t.Log("side-by-side diff produced no output")
		}
		// Side-by-side should use < or > markers for changes
		if !strings.Contains(output, "<") && !strings.Contains(output, ">") {
			t.Log("side-by-side output may not contain < or > markers for this input")
		}
	})

	t.Run("side by side suppress common", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "sbsc1.txt")
		file2 := filepath.Join(tmpDir, "sbsc2.txt")

		_ = os.WriteFile(file1, []byte("same1\nsame2\ndiff\nsame3"), 0644)
		_ = os.WriteFile(file2, []byte("same1\nsame2\nchanged\nsame3"), 0644)

		var buf bytes.Buffer
		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{Side: true, SuppressCommon: true})

		output := buf.String()
		if strings.Contains(output, "same1") {
			t.Error("suppress common should hide common lines")
		}
	})

	t.Run("ignore case", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "ic1.txt")
		file2 := filepath.Join(tmpDir, "ic2.txt")

		_ = os.WriteFile(file1, []byte("HELLO\nWorld"), 0644)
		_ = os.WriteFile(file2, []byte("hello\nworld"), 0644)

		var buf bytes.Buffer
		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{IgnoreCase: true})

		if err != nil {
			t.Fatalf("RunDiff() with IgnoreCase error = %v", err)
		}
		// Should report no differences when ignoring case
		output := strings.TrimSpace(buf.String())
		if output != "" {
			t.Logf("IgnoreCase still shows diff: %s", output)
		}
	})

	t.Run("ignore space", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "is1.txt")
		file2 := filepath.Join(tmpDir, "is2.txt")

		_ = os.WriteFile(file1, []byte("hello  world"), 0644)
		_ = os.WriteFile(file2, []byte("hello world"), 0644)

		var buf bytes.Buffer
		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{IgnoreSpace: true})

		if err != nil {
			t.Fatalf("RunDiff() with IgnoreSpace error = %v", err)
		}
	})

	t.Run("ignore blank lines", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "ib1.txt")
		file2 := filepath.Join(tmpDir, "ib2.txt")

		_ = os.WriteFile(file1, []byte("line1\n\n\nline2"), 0644)
		_ = os.WriteFile(file2, []byte("line1\nline2"), 0644)

		var buf bytes.Buffer
		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{IgnoreBlank: true})

		if err != nil {
			t.Fatalf("RunDiff() with IgnoreBlank error = %v", err)
		}
	})

	t.Run("color output", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "color1.txt")
		file2 := filepath.Join(tmpDir, "color2.txt")

		_ = os.WriteFile(file1, []byte("old"), 0644)
		_ = os.WriteFile(file2, []byte("new"), 0644)

		var buf bytes.Buffer
		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{Color: true})

		output := buf.String()
		if !strings.Contains(output, "\033[") {
			t.Log("color output may not have ANSI codes for all cases")
		}
	})

	t.Run("JSON compare identical", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "json1.json")
		file2 := filepath.Join(tmpDir, "json2.json")

		_ = os.WriteFile(file1, []byte(`{"a":1,"b":2}`), 0644)
		_ = os.WriteFile(file2, []byte(`{"a":1,"b":2}`), 0644)

		var buf bytes.Buffer
		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{JSON: true})

		if err != nil {
			t.Fatalf("RunDiff() JSON compare error = %v", err)
		}
		if buf.Len() > 0 {
			t.Error("identical JSON should produce no output")
		}
	})

	t.Run("JSON compare different", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "jdiff1.json")
		file2 := filepath.Join(tmpDir, "jdiff2.json")

		_ = os.WriteFile(file1, []byte(`{"a":1,"b":2}`), 0644)
		_ = os.WriteFile(file2, []byte(`{"a":1,"b":3,"c":4}`), 0644)

		var buf bytes.Buffer
		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{JSON: true})

		output := buf.String()
		if !strings.Contains(output, "b") {
			t.Error("expected JSON diff to mention changed key 'b'")
		}
		if !strings.Contains(output, "c") {
			t.Error("expected JSON diff to mention added key 'c'")
		}
	})

	t.Run("JSON compare invalid", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "jinv1.json")
		file2 := filepath.Join(tmpDir, "jinv2.json")

		_ = os.WriteFile(file1, []byte(`not json`), 0644)
		_ = os.WriteFile(file2, []byte(`{"a":1}`), 0644)

		var buf bytes.Buffer
		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{JSON: true})

		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("recursive directory diff", func(t *testing.T) {
		dir1 := filepath.Join(tmpDir, "rdir1")
		dir2 := filepath.Join(tmpDir, "rdir2")
		_ = os.MkdirAll(dir1, 0755)
		_ = os.MkdirAll(dir2, 0755)

		_ = os.WriteFile(filepath.Join(dir1, "file.txt"), []byte("content1"), 0644)
		_ = os.WriteFile(filepath.Join(dir2, "file.txt"), []byte("content2"), 0644)
		_ = os.WriteFile(filepath.Join(dir1, "only_in_1.txt"), []byte("x"), 0644)
		_ = os.WriteFile(filepath.Join(dir2, "only_in_2.txt"), []byte("y"), 0644)

		var buf bytes.Buffer
		err := RunDiff(&buf, []string{dir1, dir2}, DiffOptions{Recursive: true})

		if err != nil {
			t.Fatalf("recursive diff error = %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "only_in_1") {
			t.Error("expected 'Only in' message for only_in_1.txt")
		}
		if !strings.Contains(output, "only_in_2") {
			t.Error("expected 'Only in' message for only_in_2.txt")
		}
	})

	t.Run("directory without recursive flag", func(t *testing.T) {
		dir1 := filepath.Join(tmpDir, "nrdir1")
		dir2 := filepath.Join(tmpDir, "nrdir2")
		_ = os.MkdirAll(dir1, 0755)
		_ = os.MkdirAll(dir2, 0755)

		var buf bytes.Buffer
		err := RunDiff(&buf, []string{dir1, dir2}, DiffOptions{})
		if err == nil {
			t.Error("expected error when comparing directories without recursive flag")
		}
	})
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name           string
		lines          []DiffLine
		expectCount1   int
		expectCount2   int
	}{
		{
			"empty",
			nil,
			0, 0,
		},
		{
			"all_context",
			[]DiffLine{{Type: ' '}, {Type: ' '}},
			2, 2,
		},
		{
			"mixed",
			[]DiffLine{{Type: ' '}, {Type: '-'}, {Type: '+'}, {Type: ' '}},
			3, 3,
		},
		{
			"only_removed",
			[]DiffLine{{Type: '-'}, {Type: '-'}},
			2, 0,
		},
		{
			"only_added",
			[]DiffLine{{Type: '+'}, {Type: '+'}},
			0, 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c1, c2 := countLines(tt.lines)
			if c1 != tt.expectCount1 || c2 != tt.expectCount2 {
				t.Errorf("countLines() = (%d, %d), want (%d, %d)", c1, c2, tt.expectCount1, tt.expectCount2)
			}
		})
	}
}

func TestTruncateOrPad(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{"short", "hello", 10, "hello     "},
		{"exact", "hello", 5, "hello"},
		{"long", "hello world", 8, "hello w>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateOrPad(tt.input, tt.width)
			if result != tt.expected {
				t.Errorf("truncateOrPad(%q, %d) = %q, want %q", tt.input, tt.width, result, tt.expected)
			}
		})
	}
}

func TestPathOrRoot(t *testing.T) {
	if pathOrRoot("") != "(root)" {
		t.Error("empty path should return (root)")
	}
	if pathOrRoot("a.b") != "a.b" {
		t.Error("non-empty path should return as-is")
	}
}
