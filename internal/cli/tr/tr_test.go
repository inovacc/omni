package tr

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunTr(t *testing.T) {
	t.Run("simple translation", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("hello")

		err := RunTr(&buf, input, "el", "ip", TrOptions{})
		if err != nil {
			t.Fatalf("RunTr() error = %v", err)
		}

		if buf.String() != "hippo" {
			t.Errorf("RunTr() = %q, want 'hippo'", buf.String())
		}
	})

	t.Run("lowercase to uppercase", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("hello world")

		err := RunTr(&buf, input, "a-z", "A-Z", TrOptions{})
		if err != nil {
			t.Fatalf("RunTr() error = %v", err)
		}

		if buf.String() != "HELLO WORLD" {
			t.Errorf("RunTr() = %q, want 'HELLO WORLD'", buf.String())
		}
	})

	t.Run("delete characters", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("hello world")

		err := RunTr(&buf, input, "aeiou", "", TrOptions{Delete: true})
		if err != nil {
			t.Fatalf("RunTr() error = %v", err)
		}

		if buf.String() != "hll wrld" {
			t.Errorf("RunTr() = %q, want 'hll wrld'", buf.String())
		}
	})

	t.Run("squeeze repeated characters", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("heeelllo")

		err := RunTr(&buf, input, "el", "el", TrOptions{Squeeze: true})
		if err != nil {
			t.Fatalf("RunTr() error = %v", err)
		}

		if buf.String() != "helo" {
			t.Errorf("RunTr() = %q, want 'helo'", buf.String())
		}
	})

	t.Run("complement set", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("hello123")

		err := RunTr(&buf, input, "0-9", "", TrOptions{Delete: true, Complement: true})
		if err != nil {
			t.Fatalf("RunTr() error = %v", err)
		}

		if buf.String() != "123" {
			t.Errorf("RunTr() = %q, want '123'", buf.String())
		}
	})

	t.Run("escape sequences", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("hello\tworld")

		err := RunTr(&buf, input, "\\t", " ", TrOptions{})
		if err != nil {
			t.Fatalf("RunTr() error = %v", err)
		}

		if buf.String() != "hello world" {
			t.Errorf("RunTr() = %q, want 'hello world'", buf.String())
		}
	})

	t.Run("truncate mode", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("abcde")

		err := RunTr(&buf, input, "abcde", "xy", TrOptions{Truncate: true})
		if err != nil {
			t.Fatalf("RunTr() error = %v", err)
		}

		// With truncate, only a->x, b->y, c/d/e remain unchanged
		if buf.String() != "xycde" {
			t.Errorf("RunTr() = %q, want 'xycde'", buf.String())
		}
	})

	t.Run("extend set2 with last char", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("abcde")

		err := RunTr(&buf, input, "abcde", "xy", TrOptions{Truncate: false})
		if err != nil {
			t.Fatalf("RunTr() error = %v", err)
		}

		// Without truncate, set2 is extended: a->x, b->y, c->y, d->y, e->y
		if buf.String() != "xyyyy" {
			t.Errorf("RunTr() = %q, want 'xyyyy'", buf.String())
		}
	})

	t.Run("empty input", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("")

		err := RunTr(&buf, input, "a", "b", TrOptions{})
		if err != nil {
			t.Fatalf("RunTr() error = %v", err)
		}

		if buf.String() != "" {
			t.Errorf("RunTr() = %q, want ''", buf.String())
		}
	})

	t.Run("unicode characters", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("héllo")

		err := RunTr(&buf, input, "é", "e", TrOptions{})
		if err != nil {
			t.Fatalf("RunTr() error = %v", err)
		}

		if buf.String() != "hello" {
			t.Errorf("RunTr() = %q, want 'hello'", buf.String())
		}
	})
}

func TestExpandCharSet(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple chars", "abc", "abc"},
		{"range", "a-e", "abcde"},
		{"digit range", "0-5", "012345"},
		{"escape newline", "\\n", "\n"},
		{"escape tab", "\\t", "\t"},
		{"escape backslash", "\\\\", "\\"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandCharSet(tt.input)
			if result != tt.expected {
				t.Errorf("expandCharSet(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandClass(t *testing.T) {
	tests := []struct {
		class    string
		contains string
	}{
		{"lower", "abcdefghijklmnopqrstuvwxyz"},
		{"upper", "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"digit", "0123456789"},
		{"space", " \t\n"},
		{"blank", " \t"},
		{"xdigit", "0123456789ABCDEFabcdef"},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			result := expandClass(tt.class)
			for _, c := range tt.contains {
				if !strings.ContainsRune(result, c) {
					t.Errorf("expandClass(%q) missing %q", tt.class, string(c))
				}
			}
		})
	}
}

func TestBuildTransMap(t *testing.T) {
	t.Run("equal length sets", func(t *testing.T) {
		m := buildTransMap("abc", "xyz", false)
		if m['a'] != 'x' || m['b'] != 'y' || m['c'] != 'z' {
			t.Errorf("buildTransMap() incorrect mapping")
		}
	})

	t.Run("empty set2", func(t *testing.T) {
		m := buildTransMap("abc", "", false)
		if len(m) != 0 {
			t.Errorf("buildTransMap() should return empty map for empty set2")
		}
	})

	t.Run("truncate mode", func(t *testing.T) {
		m := buildTransMap("abcde", "xy", true)
		if m['a'] != 'x' || m['b'] != 'y' {
			t.Errorf("buildTransMap() truncate: a/b should map to x/y")
		}

		if _, ok := m['c']; ok {
			t.Errorf("buildTransMap() truncate: c should not be mapped")
		}
	})
}
