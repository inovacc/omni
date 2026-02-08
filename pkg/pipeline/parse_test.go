package pipeline

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"grep simple", "grep error", "grep", false},
		{"grep case insensitive", "grep -i error", "grep", false},
		{"grep-v", "grep-v pattern", "grep-v", false},
		{"grep with -v flag", "grep -v pattern", "grep-v", false},
		{"contains", "contains hello", "contains", false},
		{"contains -i", "contains -i hello", "contains", false},
		{"replace", "replace old new", "replace", false},
		{"head default", "head", "head", false},
		{"head with n", "head 5", "head", false},
		{"head with -n", "head -n 5", "head", false},
		{"take alias", "take 3", "head", false},
		{"tail", "tail 5", "tail", false},
		{"tail -n", "tail -n 5", "tail", false},
		{"skip", "skip 3", "skip", false},
		{"sort", "sort", "sort", false},
		{"sort -r", "sort -r", "sort", false},
		{"sort -n", "sort -n", "sort", false},
		{"sort -rn", "sort -rn", "sort", false},
		{"uniq", "uniq", "uniq", false},
		{"uniq -i", "uniq -i", "uniq", false},
		{"cut", "cut -d: -f1,3", "cut", false},
		{"cut spaces", "cut -d \" \" -f 1", "cut", false},
		{"tr", "tr aeiou AEIOU", "tr", false},
		{"sed", "sed s/foo/bar/g", "sed", false},
		{"rev", "rev", "rev", false},
		{"nl", "nl", "nl", false},
		{"tee", "tee /tmp/out.txt", "tee", false},
		{"tac", "tac", "tac", false},
		{"wc", "wc", "wc", false},
		{"wc -l", "wc -l", "wc", false},

		// Errors
		{"empty", "", "", true},
		{"unknown", "foobar", "", true},
		{"grep no pattern", "grep", "", true},
		{"contains no arg", "contains", "", true},
		{"replace one arg", "replace old", "", true},
		{"tr one arg", "tr abc", "", true},
		{"sed no arg", "sed", "", true},
		{"cut no field", "cut -d:", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if stage.Name() != tt.want {
				t.Errorf("expected stage name %q, got %q", tt.want, stage.Name())
			}
		})
	}
}

func TestParseAll(t *testing.T) {
	stages, err := ParseAll([]string{"grep error", "sort", "uniq", "head 10"})
	if err != nil {
		t.Fatal(err)
	}

	if len(stages) != 4 {
		t.Errorf("expected 4 stages, got %d", len(stages))
	}

	names := []string{"grep", "sort", "uniq", "head"}
	for i, s := range stages {
		if s.Name() != names[i] {
			t.Errorf("stage %d: expected %q, got %q", i, names[i], s.Name())
		}
	}
}

func TestParseAllError(t *testing.T) {
	_, err := ParseAll([]string{"grep error", "unknown_cmd"})
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestParseCommandLine(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{`grep -i "hello world"`, []string{"grep", "-i", "hello world"}},
		{`sed s/foo/bar/g`, []string{"sed", "s/foo/bar/g"}},
		{`cut -d' ' -f1`, []string{"cut", "-d ", "-f1"}},
		{`head 10`, []string{"head", "10"}},
		{`tr 'abc' 'ABC'`, []string{"tr", "abc", "ABC"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseCommandLine(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d parts, got %d: %v", len(tt.want), len(got), got)
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("part %d: expected %q, got %q", i, tt.want[i], got[i])
				}
			}
		})
	}
}

func TestParseSed(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pattern string
		repl    string
		global  bool
	}{
		{"basic", "sed s/foo/bar/", "foo", "bar", false},
		{"global", "sed s/foo/bar/g", "foo", "bar", true},
		{"pipe delim", "sed s|foo|bar|g", "foo", "bar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage, err := Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}

			sed := stage.(*Sed)
			if sed.Pattern != tt.pattern {
				t.Errorf("pattern: expected %q, got %q", tt.pattern, sed.Pattern)
			}

			if sed.Replacement != tt.repl {
				t.Errorf("replacement: expected %q, got %q", tt.repl, sed.Replacement)
			}

			if sed.Global != tt.global {
				t.Errorf("global: expected %v, got %v", tt.global, sed.Global)
			}
		})
	}
}

func TestParseGrepOptions(t *testing.T) {
	stage, err := Parse("grep -i -v pattern")
	if err != nil {
		t.Fatal(err)
	}

	g := stage.(*Grep)
	if !g.IgnoreCase {
		t.Error("expected IgnoreCase=true")
	}

	if !g.Invert {
		t.Error("expected Invert=true")
	}
}

func TestParseSortOptions(t *testing.T) {
	tests := []struct {
		input   string
		numeric bool
		reverse bool
	}{
		{"sort", false, false},
		{"sort -n", true, false},
		{"sort -r", false, true},
		{"sort -rn", true, true},
		{"sort -nr", true, true},
		{"sort --reverse", false, true},
		{"sort --numeric", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			stage, err := Parse(tt.input)
			if err != nil {
				t.Fatal(err)
			}

			s := stage.(*Sort)
			if s.Numeric != tt.numeric {
				t.Errorf("numeric: expected %v, got %v", tt.numeric, s.Numeric)
			}

			if s.Reverse != tt.reverse {
				t.Errorf("reverse: expected %v, got %v", tt.reverse, s.Reverse)
			}
		})
	}
}
