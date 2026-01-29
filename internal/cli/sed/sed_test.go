package sed

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRunSed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sed_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("simple substitution", func(t *testing.T) {
		file := filepath.Join(tmpDir, "sub.txt")
		_ = os.WriteFile(file, []byte("hello world\n"), 0644)

		var buf bytes.Buffer

		err := RunSed(&buf, []string{"s/world/universe/", file}, SedOptions{})
		if err != nil {
			t.Fatalf("RunSed() error = %v", err)
		}

		if !strings.Contains(buf.String(), "hello universe") {
			t.Errorf("RunSed() output = %q, want 'hello universe'", buf.String())
		}
	})

	t.Run("global substitution", func(t *testing.T) {
		file := filepath.Join(tmpDir, "global.txt")
		_ = os.WriteFile(file, []byte("aaa bbb aaa\n"), 0644)

		var buf bytes.Buffer

		err := RunSed(&buf, []string{"s/aaa/XXX/g", file}, SedOptions{})
		if err != nil {
			t.Fatalf("RunSed() error = %v", err)
		}

		if !strings.Contains(buf.String(), "XXX bbb XXX") {
			t.Errorf("RunSed() output = %q, want 'XXX bbb XXX'", buf.String())
		}
	})

	t.Run("delete line", func(t *testing.T) {
		file := filepath.Join(tmpDir, "delete.txt")
		_ = os.WriteFile(file, []byte("line1\ndelete this\nline3\n"), 0644)

		var buf bytes.Buffer

		err := RunSed(&buf, []string{"/delete/d", file}, SedOptions{})
		if err != nil {
			t.Fatalf("RunSed() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "delete this") {
			t.Errorf("RunSed() should delete matching line: %q", output)
		}

		if !strings.Contains(output, "line1") || !strings.Contains(output, "line3") {
			t.Errorf("RunSed() should keep other lines: %q", output)
		}
	})

	t.Run("delete by line number", func(t *testing.T) {
		file := filepath.Join(tmpDir, "deln.txt")
		_ = os.WriteFile(file, []byte("line1\nline2\nline3\n"), 0644)

		var buf bytes.Buffer

		err := RunSed(&buf, []string{"2d", file}, SedOptions{})
		if err != nil {
			t.Fatalf("RunSed() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "line2") {
			t.Errorf("RunSed() should delete line 2: %q", output)
		}
	})

	t.Run("delete range", func(t *testing.T) {
		file := filepath.Join(tmpDir, "range.txt")
		_ = os.WriteFile(file, []byte("line1\nline2\nline3\nline4\n"), 0644)

		var buf bytes.Buffer

		err := RunSed(&buf, []string{"2,3d", file}, SedOptions{})
		if err != nil {
			t.Fatalf("RunSed() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "line2") || strings.Contains(output, "line3") {
			t.Errorf("RunSed() should delete lines 2-3: %q", output)
		}

		if !strings.Contains(output, "line1") || !strings.Contains(output, "line4") {
			t.Errorf("RunSed() should keep other lines: %q", output)
		}
	})

	t.Run("quiet mode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "quiet.txt")
		_ = os.WriteFile(file, []byte("line1\nmatch this\nline3\n"), 0644)

		var buf bytes.Buffer

		err := RunSed(&buf, []string{"s/match/MATCH/", file}, SedOptions{Quiet: true})
		if err != nil {
			t.Fatalf("RunSed() error = %v", err)
		}

		// In quiet mode, output should be suppressed
		output := buf.String()
		if output != "" {
			t.Errorf("RunSed() -n should suppress output, got: %q", output)
		}
	})

	t.Run("expression flag", func(t *testing.T) {
		file := filepath.Join(tmpDir, "expr.txt")
		_ = os.WriteFile(file, []byte("hello world\n"), 0644)

		var buf bytes.Buffer

		err := RunSed(&buf, []string{file}, SedOptions{Expression: []string{"s/hello/hi/"}})
		if err != nil {
			t.Fatalf("RunSed() error = %v", err)
		}

		if !strings.Contains(buf.String(), "hi world") {
			t.Errorf("RunSed() -e output = %q, want 'hi world'", buf.String())
		}
	})

	t.Run("in-place edit", func(t *testing.T) {
		file := filepath.Join(tmpDir, "inplace.txt")
		_ = os.WriteFile(file, []byte("original content\n"), 0644)

		var buf bytes.Buffer

		err := RunSed(&buf, []string{"s/original/modified/", file}, SedOptions{InPlace: true})
		if err != nil {
			t.Fatalf("RunSed() error = %v", err)
		}

		data, _ := os.ReadFile(file)
		if !strings.Contains(string(data), "modified") {
			t.Errorf("RunSed() -i should modify file: %q", data)
		}
	})

	t.Run("in-place with backup", func(t *testing.T) {
		file := filepath.Join(tmpDir, "backup.txt")
		_ = os.WriteFile(file, []byte("original content\n"), 0644)

		var buf bytes.Buffer

		err := RunSed(&buf, []string{"s/original/modified/", file}, SedOptions{InPlace: true, InPlaceExt: ".bak"})
		if err != nil {
			t.Fatalf("RunSed() error = %v", err)
		}

		backup := file + ".bak"
		if _, err := os.Stat(backup); os.IsNotExist(err) {
			t.Error("RunSed() -i.bak should create backup file")
		}

		data, _ := os.ReadFile(backup)
		if !strings.Contains(string(data), "original") {
			t.Errorf("RunSed() backup should contain original: %q", data)
		}
	})

	t.Run("no expression", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSed(&buf, []string{}, SedOptions{})
		if err == nil {
			t.Error("RunSed() expected error for no expression")
		}
	})
}

func TestParseSubstitute(t *testing.T) {
	tests := []struct {
		expr   string
		global bool
		hasErr bool
	}{
		{"s/foo/bar/", false, false},
		{"s/foo/bar/g", true, false},
		{"s|foo|bar|", false, false},
		{"s#foo#bar#g", true, false},
		{"s/foo", false, true},
		{"invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			sub, err := parseSubstitute(tt.expr)

			if tt.hasErr {
				if err == nil {
					t.Errorf("parseSubstitute(%q) expected error", tt.expr)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseSubstitute(%q) error = %v", tt.expr, err)
			}

			if sub.global != tt.global {
				t.Errorf("parseSubstitute(%q).global = %v, want %v", tt.expr, sub.global, tt.global)
			}
		})
	}
}

func TestSedSubstitute_execute(t *testing.T) {
	t.Run("first occurrence", func(t *testing.T) {
		sub := &sedSubstitute{
			pattern:     regexp.MustCompile("a"),
			replacement: "X",
		}

		result, _ := sub.execute("aaa", 1)
		if result != "Xaa" {
			t.Errorf("substitute first = %q, want 'Xaa'", result)
		}
	})

	t.Run("global", func(t *testing.T) {
		sub := &sedSubstitute{
			pattern:     regexp.MustCompile("a"),
			replacement: "X",
			global:      true,
		}

		result, _ := sub.execute("aaa", 1)
		if result != "XXX" {
			t.Errorf("substitute global = %q, want 'XXX'", result)
		}
	})

	t.Run("nth occurrence", func(t *testing.T) {
		sub := &sedSubstitute{
			pattern:     regexp.MustCompile("a"),
			replacement: "X",
			nthMatch:    2,
		}

		result, _ := sub.execute("aaa", 1)
		if result != "aXa" {
			t.Errorf("substitute nth = %q, want 'aXa'", result)
		}
	})
}

func TestSedDelete_execute(t *testing.T) {
	t.Run("pattern match", func(t *testing.T) {
		del := &sedDelete{pattern: regexp.MustCompile("delete")}

		_, shouldPrint := del.execute("delete this", 1)
		if shouldPrint {
			t.Error("delete should suppress matching line")
		}

		_, shouldPrint = del.execute("keep this", 1)
		if !shouldPrint {
			t.Error("delete should keep non-matching line")
		}
	})

	t.Run("line number", func(t *testing.T) {
		del := &sedDelete{addressStart: 2}

		_, shouldPrint := del.execute("line", 1)
		if !shouldPrint {
			t.Error("delete should keep line 1")
		}

		_, shouldPrint = del.execute("line", 2)
		if shouldPrint {
			t.Error("delete should suppress line 2")
		}
	})
}
