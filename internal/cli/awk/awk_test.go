package awk

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAwk(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "awk_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("print all", func(t *testing.T) {
		file := filepath.Join(tmpDir, "all.txt")
		_ = os.WriteFile(file, []byte("line1\nline2\nline3\n"), 0644)

		var buf bytes.Buffer

		err := RunAwk(&buf, []string{"{print}", file}, AwkOptions{})
		if err != nil {
			t.Fatalf("RunAwk() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "line1") || !strings.Contains(output, "line2") {
			t.Errorf("RunAwk() print all = %q", output)
		}
	})

	t.Run("print fields", func(t *testing.T) {
		file := filepath.Join(tmpDir, "fields.txt")
		_ = os.WriteFile(file, []byte("one two three\nfour five six\n"), 0644)

		var buf bytes.Buffer

		err := RunAwk(&buf, []string{"{print $1}", file}, AwkOptions{})
		if err != nil {
			t.Fatalf("RunAwk() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "one") || !strings.Contains(output, "four") {
			t.Errorf("RunAwk() print $1 = %q", output)
		}

		if strings.Contains(output, "two") || strings.Contains(output, "five") {
			t.Errorf("RunAwk() print $1 should not contain $2: %q", output)
		}
	})

	t.Run("print multiple fields", func(t *testing.T) {
		file := filepath.Join(tmpDir, "multi.txt")
		_ = os.WriteFile(file, []byte("a b c\n"), 0644)

		var buf bytes.Buffer

		err := RunAwk(&buf, []string{"{print $1, $3}", file}, AwkOptions{})
		if err != nil {
			t.Fatalf("RunAwk() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "a") || !strings.Contains(output, "c") {
			t.Errorf("RunAwk() print $1, $3 = %q", output)
		}
	})

	t.Run("field separator", func(t *testing.T) {
		file := filepath.Join(tmpDir, "sep.txt")
		_ = os.WriteFile(file, []byte("one:two:three\n"), 0644)

		var buf bytes.Buffer

		err := RunAwk(&buf, []string{"{print $2}", file}, AwkOptions{FieldSeparator: ":"})
		if err != nil {
			t.Fatalf("RunAwk() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "two") {
			t.Errorf("RunAwk() -F: print $2 = %q", output)
		}
	})

	t.Run("pattern match", func(t *testing.T) {
		file := filepath.Join(tmpDir, "pattern.txt")
		_ = os.WriteFile(file, []byte("apple\nbanana\napricot\n"), 0644)

		var buf bytes.Buffer

		err := RunAwk(&buf, []string{"/^a/{print}", file}, AwkOptions{})
		if err != nil {
			t.Fatalf("RunAwk() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "apple") || !strings.Contains(output, "apricot") {
			t.Errorf("RunAwk() pattern should match lines starting with 'a': %q", output)
		}

		if strings.Contains(output, "banana") {
			t.Errorf("RunAwk() pattern should not match 'banana': %q", output)
		}
	})

	t.Run("BEGIN block", func(t *testing.T) {
		file := filepath.Join(tmpDir, "begin.txt")
		_ = os.WriteFile(file, []byte("data\n"), 0644)

		var buf bytes.Buffer

		err := RunAwk(&buf, []string{`BEGIN{print "header"}{print}`, file}, AwkOptions{})
		if err != nil {
			t.Fatalf("RunAwk() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "header") {
			t.Errorf("RunAwk() BEGIN should print header: %q", output)
		}
	})

	t.Run("END block", func(t *testing.T) {
		file := filepath.Join(tmpDir, "end.txt")
		_ = os.WriteFile(file, []byte("data\n"), 0644)

		var buf bytes.Buffer

		err := RunAwk(&buf, []string{`{print}END{print "footer"}`, file}, AwkOptions{})
		if err != nil {
			t.Fatalf("RunAwk() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "footer") {
			t.Errorf("RunAwk() END should print footer: %q", output)
		}
	})

	t.Run("print $0", func(t *testing.T) {
		file := filepath.Join(tmpDir, "zero.txt")
		_ = os.WriteFile(file, []byte("entire line\n"), 0644)

		var buf bytes.Buffer

		err := RunAwk(&buf, []string{"{print $0}", file}, AwkOptions{})
		if err != nil {
			t.Fatalf("RunAwk() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "entire line") {
			t.Errorf("RunAwk() print $0 = %q", output)
		}
	})

	t.Run("string literal", func(t *testing.T) {
		file := filepath.Join(tmpDir, "literal.txt")
		_ = os.WriteFile(file, []byte("data\n"), 0644)

		var buf bytes.Buffer

		err := RunAwk(&buf, []string{`{print "prefix", $1}`, file}, AwkOptions{})
		if err != nil {
			t.Fatalf("RunAwk() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "prefix") {
			t.Errorf("RunAwk() string literal = %q", output)
		}
	})

	t.Run("NF variable", func(t *testing.T) {
		file := filepath.Join(tmpDir, "nf.txt")
		_ = os.WriteFile(file, []byte("a b c d\n"), 0644)

		var buf bytes.Buffer

		err := RunAwk(&buf, []string{"{print NF}", file}, AwkOptions{})
		if err != nil {
			t.Fatalf("RunAwk() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "4" {
			t.Errorf("RunAwk() NF = %q, want '4'", output)
		}
	})

	t.Run("no program", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunAwk(&buf, []string{}, AwkOptions{})
		if err == nil {
			t.Error("RunAwk() expected error for no program")
		}
	})
}

func TestParseAwkProgram(t *testing.T) {
	t.Run("simple action", func(t *testing.T) {
		prog, err := parseAwkProgram("{print}")
		if err != nil {
			t.Fatalf("parseAwkProgram() error = %v", err)
		}

		if len(prog.rules) == 0 {
			t.Error("parseAwkProgram() should have rules")
		}
	})

	t.Run("BEGIN and END", func(t *testing.T) {
		prog, err := parseAwkProgram(`BEGIN{print "start"}END{print "end"}`)
		if err != nil {
			t.Fatalf("parseAwkProgram() error = %v", err)
		}

		if prog.begin == nil {
			t.Error("parseAwkProgram() should have BEGIN")
		}

		if prog.end == nil {
			t.Error("parseAwkProgram() should have END")
		}
	})

	t.Run("pattern with action", func(t *testing.T) {
		prog, err := parseAwkProgram("/test/{print}")
		if err != nil {
			t.Fatalf("parseAwkProgram() error = %v", err)
		}

		if len(prog.rules) == 0 {
			t.Error("parseAwkProgram() should have rules")
		}

		if prog.rules[0].pattern == nil {
			t.Error("parseAwkProgram() rule should have pattern")
		}
	})

	t.Run("unterminated regex", func(t *testing.T) {
		_, err := parseAwkProgram("/unterminated{print}")
		if err == nil {
			t.Error("parseAwkProgram() expected error for unterminated regex")
		}
	})
}

func TestExpandAwkFields(t *testing.T) {
	fields := []string{"whole line", "one", "two", "three"}

	tests := []struct {
		expr     string
		expected string
	}{
		{"$0", "whole line"},
		{"$1", "one"},
		{"$2", "two"},
		{"$1, $2", "one two"},
		{`"prefix"`, "prefix"},
		{"NF", "3"},
		{"$99", ""}, // out of range
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := expandAwkFields(tt.expr, fields)
			if result != tt.expected {
				t.Errorf("expandAwkFields(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}

func TestExpandSingleField(t *testing.T) {
	fields := []string{"line", "a", "b", "c"}

	tests := []struct {
		expr     string
		expected string
	}{
		{"$0", "line"},
		{"$1", "a"},
		{`"literal"`, "literal"},
		{"NF", "3"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := expandSingleField(tt.expr, fields)
			if result != tt.expected {
				t.Errorf("expandSingleField(%q) = %q, want %q", tt.expr, result, tt.expected)
			}
		})
	}
}
