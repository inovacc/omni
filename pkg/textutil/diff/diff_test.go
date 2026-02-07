package diff

import (
	"bytes"
	"strings"
	"testing"
)

func TestComputeDiff(t *testing.T) {
	t.Run("identical", func(t *testing.T) {
		lines := []string{"a", "b", "c"}
		hunks := ComputeDiff(lines, lines)
		if len(hunks) != 0 {
			t.Errorf("expected no hunks for identical input, got %d", len(hunks))
		}
	})

	t.Run("single change", func(t *testing.T) {
		lines1 := []string{"a", "b", "c"}
		lines2 := []string{"a", "x", "c"}
		hunks := ComputeDiff(lines1, lines2)
		if len(hunks) == 0 {
			t.Error("expected hunks for different input")
		}
	})

	t.Run("added lines", func(t *testing.T) {
		lines1 := []string{"a", "b"}
		lines2 := []string{"a", "b", "c", "d"}
		hunks := ComputeDiff(lines1, lines2)
		if len(hunks) == 0 {
			t.Error("expected hunks for added lines")
		}
	})

	t.Run("removed lines", func(t *testing.T) {
		lines1 := []string{"a", "b", "c"}
		lines2 := []string{"a"}
		hunks := ComputeDiff(lines1, lines2)
		if len(hunks) == 0 {
			t.Error("expected hunks for removed lines")
		}
	})

	t.Run("empty inputs", func(t *testing.T) {
		hunks := ComputeDiff(nil, nil)
		if len(hunks) != 0 {
			t.Error("expected no hunks for empty inputs")
		}
	})

	t.Run("one empty", func(t *testing.T) {
		hunks := ComputeDiff([]string{"a"}, nil)
		if len(hunks) == 0 {
			t.Error("expected hunks when one input is empty")
		}
	})

	t.Run("with context option", func(t *testing.T) {
		lines1 := []string{"a", "b", "c", "d", "e"}
		lines2 := []string{"a", "b", "X", "d", "e"}
		hunks := ComputeDiff(lines1, lines2, WithContext(1))
		if len(hunks) == 0 {
			t.Error("expected hunks")
		}
	})
}

func TestFormatUnified(t *testing.T) {
	lines1 := []string{"old line", "common"}
	lines2 := []string{"new line", "common"}
	hunks := ComputeDiff(lines1, lines2)

	var buf bytes.Buffer
	FormatUnified(&buf, "a.txt", "b.txt", hunks)

	output := buf.String()
	if !strings.Contains(output, "--- a.txt") {
		t.Error("expected --- header")
	}
	if !strings.Contains(output, "+++ b.txt") {
		t.Error("expected +++ header")
	}
	if !strings.Contains(output, "@@") {
		t.Error("expected @@ hunk header")
	}
}

func TestFormatUnifiedEmpty(t *testing.T) {
	var buf bytes.Buffer
	FormatUnified(&buf, "a.txt", "b.txt", nil)
	if buf.Len() != 0 {
		t.Error("expected empty output for nil hunks")
	}
}

func TestCompareJSON(t *testing.T) {
	t.Run("identical", func(t *testing.T) {
		v := map[string]any{"a": float64(1)}
		diffs := CompareJSON(v, v)
		if len(diffs) != 0 {
			t.Errorf("expected no diffs, got %v", diffs)
		}
	})

	t.Run("value change", func(t *testing.T) {
		v1 := map[string]any{"a": float64(1)}
		v2 := map[string]any{"a": float64(2)}
		diffs := CompareJSON(v1, v2)
		if len(diffs) == 0 {
			t.Error("expected diffs for value change")
		}
		if !strings.Contains(diffs[0], "~") {
			t.Errorf("expected ~ prefix, got %v", diffs[0])
		}
	})

	t.Run("added key", func(t *testing.T) {
		v1 := map[string]any{"a": float64(1)}
		v2 := map[string]any{"a": float64(1), "b": float64(2)}
		diffs := CompareJSON(v1, v2)
		found := false
		for _, d := range diffs {
			if strings.Contains(d, "+ b") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected + b diff, got %v", diffs)
		}
	})

	t.Run("removed key", func(t *testing.T) {
		v1 := map[string]any{"a": float64(1), "b": float64(2)}
		v2 := map[string]any{"a": float64(1)}
		diffs := CompareJSON(v1, v2)
		found := false
		for _, d := range diffs {
			if strings.Contains(d, "- b") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected - b diff, got %v", diffs)
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		v1 := map[string]any{"a": float64(1)}
		v2 := "string"
		diffs := CompareJSON(v1, v2)
		if len(diffs) == 0 {
			t.Error("expected type mismatch diff")
		}
	})

	t.Run("array diff", func(t *testing.T) {
		v1 := []any{float64(1), float64(2)}
		v2 := []any{float64(1), float64(3)}
		diffs := CompareJSON(v1, v2)
		if len(diffs) == 0 {
			t.Error("expected array diff")
		}
	})
}

func TestCompareJSONBytes(t *testing.T) {
	t.Run("identical", func(t *testing.T) {
		diffs, err := CompareJSONBytes([]byte(`{"a":1}`), []byte(`{"a":1}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(diffs) != 0 {
			t.Errorf("expected no diffs, got %v", diffs)
		}
	})

	t.Run("different", func(t *testing.T) {
		diffs, err := CompareJSONBytes([]byte(`{"a":1}`), []byte(`{"a":2}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(diffs) == 0 {
			t.Error("expected diffs")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := CompareJSONBytes([]byte(`not json`), []byte(`{"a":1}`))
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name         string
		lines        []Line
		expectCount1 int
		expectCount2 int
	}{
		{"empty", nil, 0, 0},
		{"all context", []Line{{Type: ' '}, {Type: ' '}}, 2, 2},
		{"mixed", []Line{{Type: ' '}, {Type: '-'}, {Type: '+'}, {Type: ' '}}, 3, 3},
		{"only removed", []Line{{Type: '-'}, {Type: '-'}}, 2, 0},
		{"only added", []Line{{Type: '+'}, {Type: '+'}}, 0, 2},
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
		input    string
		width    int
		expected string
	}{
		{"hello", 10, "hello     "},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello w>"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := TruncateOrPad(tt.input, tt.width); got != tt.expected {
				t.Errorf("TruncateOrPad(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.expected)
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
