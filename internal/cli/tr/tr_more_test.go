package tr

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestExpandClassMore covers each POSIX character class branch.
func TestExpandClassMore(t *testing.T) {
	tests := []struct {
		class   string
		check   func(string) bool
		desc    string
	}{
		{"digit", func(s string) bool { return s == "0123456789" }, "digits"},
		{"lower", func(s string) bool { return s == "abcdefghijklmnopqrstuvwxyz" }, "lower"},
		{"upper", func(s string) bool { return s == "ABCDEFGHIJKLMNOPQRSTUVWXYZ" }, "upper"},
		{"alpha", func(s string) bool { return len(s) == 52 }, "alpha 52"},
		{"alnum", func(s string) bool { return len(s) == 62 }, "alnum 62"},
		{"space", func(s string) bool { return strings.Contains(s, " ") && strings.Contains(s, "\t") }, "space"},
		{"blank", func(s string) bool { return s == " \t" }, "blank"},
		{"xdigit", func(s string) bool { return s == "0123456789ABCDEFabcdef" }, "xdigit"},
		{"punct", func(s string) bool { return strings.Contains(s, "!") && strings.Contains(s, "~") }, "punct"},
		{"graph", func(s string) bool { return strings.HasPrefix(s, "!") }, "graph"},
		{"print", func(s string) bool { return strings.HasPrefix(s, " ") }, "print"},
		{"unknown", func(s string) bool { return s == "" }, "empty for unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			got := expandClass(tt.class)
			if !tt.check(got) {
				t.Errorf("expandClass(%q) failed %s: got %q", tt.class, tt.desc, got)
			}
		})
	}
}

// TestExpandCharSetClasses exercises class expansion embedded in a set.
func TestExpandCharSetClasses(t *testing.T) {
	if got := expandCharSet("[:digit:]"); got != "0123456789" {
		t.Errorf("digit class via set = %q", got)
	}
	if got := expandCharSet("a-e"); got != "abcde" {
		t.Errorf("range = %q", got)
	}
	if got := expandCharSet(`\t\n`); got != "\t\n" {
		t.Errorf("escapes = %q", got)
	}
}

// TestRunTrFromStdinViaRunTr exercises the translate/delete/squeeze paths.
// RunTrFromStdin itself reads os.Stdin which we cannot inject offline, so we
// drive the same logic through RunTr with an injected reader and only smoke-test
// RunTrFromStdin's wiring with empty stdin via a redirect-free call is skipped.
func TestRunTrTranslate(t *testing.T) {
	tests := []struct {
		name string
		in   string
		set1 string
		set2 string
		opts TrOptions
		want string
	}{
		{"upcase via class", "hello", "[:lower:]", "[:upper:]", TrOptions{}, "HELLO"},
		{"rot range", "abc", "a-c", "x-z", TrOptions{}, "xyz"},
		{"delete digits", "a1b2c3", "[:digit:]", "", TrOptions{Delete: true}, "abc"},
		{"squeeze", "aaabbbccc", "a-c", "a-c", TrOptions{Squeeze: true}, "abc"},
		{"complement delete", "a1b2", "[:digit:]", "", TrOptions{Complement: true, Delete: true}, "12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RunTr(&buf, strings.NewReader(tt.in), tt.set1, tt.set2, tt.opts); err != nil {
				t.Fatal(err)
			}
			if buf.String() != tt.want {
				t.Errorf("RunTr(%q) = %q, want %q", tt.in, buf.String(), tt.want)
			}
		})
	}
}

// TestRunTrFromStdin drives RunTrFromStdin by temporarily redirecting the
// process stdin to a real temp file (no network, no exec, fully offline).
func TestRunTrFromStdin(t *testing.T) {
	dir := t.TempDir()
	tmp, err := os.CreateTemp(dir, "stdin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.WriteString("hello"); err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tmp.Close() }()

	old := os.Stdin
	os.Stdin = tmp
	defer func() { os.Stdin = old }()

	var buf bytes.Buffer
	if err := RunTrFromStdin(&buf, "a-z", "A-Z", TrOptions{}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "HELLO" {
		t.Errorf("RunTrFromStdin = %q, want %q", buf.String(), "HELLO")
	}
}
