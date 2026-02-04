package rg

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestNewFormatter(t *testing.T) {
	var buf bytes.Buffer

	opts := FormatterOptions{
		UseColor:       true,
		Scheme:         DefaultScheme(),
		Format:         FormatDefault,
		ShowLineNumber: true,
		ShowColumn:     true,
		Pattern:        "test",
	}

	f := NewFormatter(&buf, opts)

	if f == nil {
		t.Fatal("NewFormatter() returned nil")
	}

	if !f.useColor {
		t.Error("NewFormatter() useColor should be true")
	}

	if !f.showLineNumber {
		t.Error("NewFormatter() showLineNumber should be true")
	}

	if f.pattern != "test" {
		t.Errorf("NewFormatter() pattern = %q, want 'test'", f.pattern)
	}
}

func TestPrintFileHeader(t *testing.T) {
	scheme := DefaultScheme()

	t.Run("default format", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{
			UseColor: false,
			Format:   FormatDefault,
		})

		f.PrintFileHeader("test.go")

		if !strings.Contains(buf.String(), "test.go") {
			t.Errorf("PrintFileHeader() = %q, want to contain 'test.go'", buf.String())
		}
	})

	t.Run("no heading format", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{
			Format: FormatNoHeading,
		})

		f.PrintFileHeader("test.go")

		if buf.String() != "" {
			t.Errorf("PrintFileHeader() with NoHeading should output nothing, got %q", buf.String())
		}
	})

	t.Run("with color", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{
			UseColor: true,
			Scheme:   scheme,
			Format:   FormatDefault,
		})

		f.PrintFileHeader("test.go")

		if !strings.Contains(buf.String(), scheme.Path) {
			t.Errorf("PrintFileHeader() with color should contain path color code")
		}
	})
}

func TestPrintMatch(t *testing.T) {
	scheme := DefaultScheme()
	re := regexp.MustCompile("world")

	t.Run("basic match no color", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{
			ShowLineNumber: true,
			Regex:          re,
		})

		f.PrintMatch("test.go", 10, 5, "hello world", false)

		output := buf.String()
		if !strings.Contains(output, "10") {
			t.Errorf("PrintMatch() should contain line number, got %q", output)
		}

		if !strings.Contains(output, "hello world") {
			t.Errorf("PrintMatch() should contain line content, got %q", output)
		}
	})

	t.Run("with trim", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{
			Trim: true,
		})

		f.PrintMatch("test.go", 1, 1, "  spaced line  ", false)

		output := buf.String()
		if strings.Contains(output, "  spaced") {
			t.Errorf("PrintMatch() with trim should remove leading spaces, got %q", output)
		}
	})

	t.Run("with replacement", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{
			Regex:   re,
			Replace: "universe",
		})

		f.PrintMatch("test.go", 1, 1, "hello world", false)

		output := buf.String()
		if !strings.Contains(output, "universe") {
			t.Errorf("PrintMatch() with replace should contain replacement, got %q", output)
		}
	})

	t.Run("no heading format", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{
			Format:         FormatNoHeading,
			ShowLineNumber: true,
			ShowColumn:     true,
		})

		f.PrintMatch("test.go", 10, 5, "hello world", false)

		output := buf.String()
		if !strings.Contains(output, "test.go") {
			t.Errorf("PrintMatch() NoHeading should include filename, got %q", output)
		}
	})

	t.Run("context line", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{
			UseColor:       true,
			Scheme:         scheme,
			ShowLineNumber: true,
		})

		f.PrintMatch("test.go", 10, 1, "context line", true)

		output := buf.String()
		if strings.Contains(output, scheme.Match) {
			t.Errorf("PrintMatch() context line should not have match highlighting")
		}
	})

	t.Run("with color highlighting", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{
			UseColor: true,
			Scheme:   scheme,
			Regex:    re,
		})

		f.PrintMatch("test.go", 1, 1, "hello world", false)

		output := buf.String()
		if !strings.Contains(output, scheme.Match) {
			t.Errorf("PrintMatch() should highlight match, got %q", output)
		}
	})
}

func TestPrintMatchOnlyMatching(t *testing.T) {
	re := regexp.MustCompile("foo")

	t.Run("only matching mode", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{
			OnlyMatching:   true,
			Regex:          re,
			ShowLineNumber: true,
		})

		f.PrintMatch("test.go", 5, 1, "foo bar foo", false)

		output := buf.String()

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 2 {
			t.Errorf("OnlyMatching should output 2 lines for 2 matches, got %d: %q", len(lines), output)
		}
	})

	t.Run("only matching with literal", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{
			OnlyMatching: true,
			UseLiteral:   true,
			Pattern:      "bar",
		})

		f.PrintMatch("test.go", 5, 1, "foo bar baz bar", false)

		output := buf.String()

		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 2 {
			t.Errorf("OnlyMatching literal should output 2 lines, got %d", len(lines))
		}
	})
}

func TestPrintContextSeparator(t *testing.T) {
	t.Run("no color", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{})

		f.PrintContextSeparator()

		if !strings.Contains(buf.String(), "--") {
			t.Errorf("PrintContextSeparator() = %q, want '--'", buf.String())
		}
	})

	t.Run("with color", func(t *testing.T) {
		var buf bytes.Buffer

		scheme := DefaultScheme()
		f := NewFormatter(&buf, FormatterOptions{
			UseColor: true,
			Scheme:   scheme,
		})

		f.PrintContextSeparator()

		if !strings.Contains(buf.String(), scheme.Separator) {
			t.Errorf("PrintContextSeparator() should contain separator color")
		}
	})
}

func TestPrintFilesWithMatch(t *testing.T) {
	var buf bytes.Buffer

	f := NewFormatter(&buf, FormatterOptions{})

	f.PrintFilesWithMatch("src/main.go")

	if !strings.Contains(buf.String(), "src/main.go") {
		t.Errorf("PrintFilesWithMatch() = %q, want 'src/main.go'", buf.String())
	}
}

func TestPrintCount(t *testing.T) {
	t.Run("basic count", func(t *testing.T) {
		var buf bytes.Buffer

		f := NewFormatter(&buf, FormatterOptions{})

		f.PrintCount("test.go", 42)

		output := buf.String()
		if !strings.Contains(output, "test.go") || !strings.Contains(output, "42") {
			t.Errorf("PrintCount() = %q, want path and count", output)
		}
	})

	t.Run("with color", func(t *testing.T) {
		var buf bytes.Buffer

		scheme := DefaultScheme()
		f := NewFormatter(&buf, FormatterOptions{
			UseColor: true,
			Scheme:   scheme,
		})

		f.PrintCount("test.go", 10)

		if !strings.Contains(buf.String(), scheme.Path) {
			t.Errorf("PrintCount() with color should have path coloring")
		}
	})
}

func TestFindMatches(t *testing.T) {
	re := regexp.MustCompile("foo")

	t.Run("regex matches", func(t *testing.T) {
		f := &Formatter{re: re}
		matches := f.findMatches("foo bar foo baz")

		if len(matches) != 2 {
			t.Errorf("findMatches() found %d matches, want 2", len(matches))
		}
	})

	t.Run("literal matches", func(t *testing.T) {
		f := &Formatter{
			useLiteral: true,
			pattern:    "bar",
		}
		matches := f.findMatches("foo bar baz bar")

		if len(matches) != 2 {
			t.Errorf("findMatches() literal found %d matches, want 2", len(matches))
		}
	})

	t.Run("no regex", func(t *testing.T) {
		f := &Formatter{}
		matches := f.findMatches("foo bar")

		if matches != nil {
			t.Errorf("findMatches() with no regex should return nil")
		}
	})
}

func TestFindLiteralMatches(t *testing.T) {
	t.Run("case sensitive", func(t *testing.T) {
		f := &Formatter{pattern: "foo"}
		matches := f.findLiteralMatches("foo Foo FOO foo")

		if len(matches) != 2 {
			t.Errorf("findLiteralMatches() case sensitive found %d, want 2", len(matches))
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		f := &Formatter{
			pattern:         "foo",
			caseInsensitive: true,
		}
		matches := f.findLiteralMatches("foo Foo FOO foo")

		if len(matches) != 4 {
			t.Errorf("findLiteralMatches() case insensitive found %d, want 4", len(matches))
		}
	})
}

func TestByteOffset(t *testing.T) {
	bo := NewByteOffset()

	if bo == nil {
		t.Fatal("NewByteOffset() returned nil")
	}

	t.Run("initial offset", func(t *testing.T) {
		offset := bo.GetMatchOffset(1)
		if offset != 0 {
			t.Errorf("Initial offset should be 0, got %d", offset)
		}
	})

	t.Run("after adding line", func(t *testing.T) {
		bo.AddLine("hello") // 5 chars + 1 newline = 6 bytes

		offset := bo.GetMatchOffset(1)
		if offset != 6 {
			t.Errorf("Offset after 'hello' should be 6, got %d", offset)
		}
	})

	t.Run("column offset", func(t *testing.T) {
		bo2 := NewByteOffset()
		bo2.AddLine("hello") // 6 bytes total

		offset := bo2.GetMatchOffset(3) // column 3 = byte 2 (0-indexed)
		if offset != 8 {                // 6 + (3-1) = 8
			t.Errorf("Offset at column 3 should be 8, got %d", offset)
		}
	})
}

func TestStats(t *testing.T) {
	var buf bytes.Buffer

	stats := &Stats{
		FilesSearched: 100,
		FilesMatched:  25,
		TotalMatches:  150,
	}

	stats.PrintStats(&buf)

	output := buf.String()
	if !strings.Contains(output, "150 matches") {
		t.Errorf("PrintStats() should contain match count, got %q", output)
	}

	if !strings.Contains(output, "25 files contained") {
		t.Errorf("PrintStats() should contain files matched, got %q", output)
	}

	if !strings.Contains(output, "100 files searched") {
		t.Errorf("PrintStats() should contain files searched, got %q", output)
	}
}

func TestOutputFormatConstants(t *testing.T) {
	// Verify format constants are distinct
	formats := []OutputFormat{FormatDefault, FormatNoHeading, FormatJSON, FormatJSONStream}
	seen := make(map[OutputFormat]bool)

	for _, f := range formats {
		if seen[f] {
			t.Errorf("Duplicate format constant: %v", f)
		}

		seen[f] = true
	}
}
