package column

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunColumn(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "column_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("basic fill columns", func(t *testing.T) {
		file := filepath.Join(tmpDir, "list.txt")
		_ = os.WriteFile(file, []byte("apple\nbanana\ncherry\ndate\n"), 0644)

		var buf bytes.Buffer

		err := RunColumn(&buf, []string{file}, ColumnOptions{Columns: 40})
		if err != nil {
			t.Fatalf("RunColumn() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "apple") {
			t.Errorf("RunColumn() output missing content")
		}
	})

	t.Run("table mode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "table.txt")
		_ = os.WriteFile(file, []byte("name age\nalice 30\nbob 25\n"), 0644)

		var buf bytes.Buffer

		err := RunColumn(&buf, []string{file}, ColumnOptions{Table: true})
		if err != nil {
			t.Fatalf("RunColumn() error = %v", err)
		}

		output := buf.String()
		// Should be aligned
		if !strings.Contains(output, "name") && !strings.Contains(output, "alice") {
			t.Errorf("RunColumn() table output = %q", output)
		}
	})

	t.Run("table with custom separator", func(t *testing.T) {
		file := filepath.Join(tmpDir, "sep.txt")
		_ = os.WriteFile(file, []byte("a,b,c\n1,2,3\n"), 0644)

		var buf bytes.Buffer

		err := RunColumn(&buf, []string{file}, ColumnOptions{Table: true, Separator: ","})
		if err != nil {
			t.Fatalf("RunColumn() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "a") && !strings.Contains(output, "1") {
			t.Errorf("RunColumn() output = %q", output)
		}
	})

	t.Run("table with output separator", func(t *testing.T) {
		file := filepath.Join(tmpDir, "outsep.txt")
		_ = os.WriteFile(file, []byte("a b\n1 2\n"), 0644)

		var buf bytes.Buffer

		err := RunColumn(&buf, []string{file}, ColumnOptions{Table: true, OutputSep: " | "})
		if err != nil {
			t.Fatalf("RunColumn() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, " | ") {
			t.Errorf("RunColumn() should use custom output separator")
		}
	})

	t.Run("fill rows mode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "rows.txt")
		_ = os.WriteFile(file, []byte("1\n2\n3\n4\n"), 0644)

		var buf bytes.Buffer

		err := RunColumn(&buf, []string{file}, ColumnOptions{FillRows: true, Columns: 20})
		if err != nil {
			t.Fatalf("RunColumn() error = %v", err)
		}

		// Should produce output with rows filled first
		if buf.Len() == 0 {
			t.Errorf("RunColumn() produced no output")
		}
	})

	t.Run("right align columns", func(t *testing.T) {
		file := filepath.Join(tmpDir, "right.txt")
		_ = os.WriteFile(file, []byte("a b\n111 2222\n"), 0644)

		var buf bytes.Buffer

		err := RunColumn(&buf, []string{file}, ColumnOptions{Table: true, Right: true})
		if err != nil {
			t.Fatalf("RunColumn() error = %v", err)
		}

		// Right-aligned should have leading spaces
		output := buf.String()
		if !strings.Contains(output, "a") {
			t.Errorf("RunColumn() right aligned output = %q", output)
		}
	})

	t.Run("column headers", func(t *testing.T) {
		file := filepath.Join(tmpDir, "headers.txt")
		_ = os.WriteFile(file, []byte("alice 30\nbob 25\n"), 0644)

		var buf bytes.Buffer

		err := RunColumn(&buf, []string{file}, ColumnOptions{Table: true, ColumnHeaders: "Name,Age"})
		if err != nil {
			t.Fatalf("RunColumn() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Name") {
			t.Errorf("RunColumn() should include headers")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")
		_ = os.WriteFile(file, []byte(""), 0644)

		var buf bytes.Buffer

		err := RunColumn(&buf, []string{file}, ColumnOptions{})
		if err != nil {
			t.Fatalf("RunColumn() error = %v", err)
		}
	})

	t.Run("no merge delimiters", func(t *testing.T) {
		file := filepath.Join(tmpDir, "nomerge.txt")
		_ = os.WriteFile(file, []byte("a,,b\n1,,2\n"), 0644)

		var buf bytes.Buffer

		err := RunColumn(&buf, []string{file}, ColumnOptions{Table: true, Separator: ",", NoMerge: true})
		if err != nil {
			t.Fatalf("RunColumn() error = %v", err)
		}

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		// With no merge, empty field should be preserved
		if len(lines) < 2 {
			t.Errorf("RunColumn() no merge got %d lines", len(lines))
		}
	})
}

func TestColumnTable(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		var buf bytes.Buffer

		err := columnTable(&buf, []string{}, ColumnOptions{})
		if err != nil {
			t.Errorf("columnTable() error = %v", err)
		}
	})

	t.Run("single line", func(t *testing.T) {
		var buf bytes.Buffer

		err := columnTable(&buf, []string{"a b c"}, ColumnOptions{})
		if err != nil {
			t.Errorf("columnTable() error = %v", err)
		}

		if !strings.Contains(buf.String(), "a") {
			t.Errorf("columnTable() output = %q", buf.String())
		}
	})
}

func TestColumnFill(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		var buf bytes.Buffer

		err := columnFill(&buf, []string{}, ColumnOptions{Columns: 80})
		if err != nil {
			t.Errorf("columnFill() error = %v", err)
		}
	})

	t.Run("few items", func(t *testing.T) {
		var buf bytes.Buffer

		err := columnFill(&buf, []string{"a", "b", "c"}, ColumnOptions{Columns: 80})
		if err != nil {
			t.Errorf("columnFill() error = %v", err)
		}
	})
}
