package pipeline

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// run drives a single stage with the given input and returns its output.
func run(t *testing.T, s Stage, in string) string {
	t.Helper()
	var out bytes.Buffer
	if err := s.Process(context.Background(), strings.NewReader(in), &out); err != nil {
		t.Fatalf("%s.Process: %v", s.Name(), err)
	}
	return out.String()
}

func TestStageProcess_Table(t *testing.T) {
	tests := []struct {
		name  string
		stage Stage
		in    string
		want  string
	}{
		{"grep match", &Grep{Pattern: "foo"}, "foo\nbar\nfoobar\n", "foo\nfoobar\n"},
		{"grep ignorecase", &Grep{Pattern: "FOO", IgnoreCase: true}, "foo\nbar\n", "foo\n"},
		{"grep invert", &Grep{Pattern: "foo", Invert: true}, "foo\nbar\n", "bar\n"},
		{"contains", &Contains{Substr: "ar"}, "foo\nbar\ncar\n", "bar\ncar\n"},
		{"contains ignorecase", &Contains{Substr: "AR", IgnoreCase: true}, "Bar\nfoo\n", "Bar\n"},
		{"replace", &Replace{Old: "a", New: "X"}, "banana\n", "bXnXnX\n"},
		{"head 2", &Head{N: 2}, "1\n2\n3\n4\n", "1\n2\n"},
		{"head default", &Head{}, "1\n2\n", "1\n2\n"},
		{"skip 1", &Skip{N: 1}, "1\n2\n3\n", "2\n3\n"},
		{"uniq", &Uniq{}, "a\na\nb\nb\na\n", "a\nb\na\n"},
		{"uniq ignorecase", &Uniq{IgnoreCase: true}, "A\na\nB\n", "A\nB\n"},
		{"cut tab", &Cut{Delimiter: "\t", Fields: []int{2}}, "a\tb\tc\n", "b\n"},
		{"cut multi", &Cut{Delimiter: ",", Fields: []int{1, 3}}, "a,b,c\n", "a,c\n"},
		{"cut out of range", &Cut{Delimiter: ",", Fields: []int{9}}, "a,b\n", "\n"},
		{"cut empty delim default", &Cut{Fields: []int{1}}, "a\tb\n", "a\n"},
		{"tr", &Tr{From: "abc", To: "xyz"}, "cab\n", "zxy\n"},
		{"tr shorter to", &Tr{From: "abc", To: "x"}, "abc\n", "xxx\n"},
		{"sed global", &Sed{Pattern: "a", Replacement: "X", Global: true}, "aaa\n", "XXX\n"},
		{"sed first only", &Sed{Pattern: "a", Replacement: "X"}, "aaa\n", "Xaa\n"},
		{"sed no match", &Sed{Pattern: "z", Replacement: "X"}, "aaa\n", "aaa\n"},
		{"rev", &Rev{}, "abc\n", "cba\n"},
		{"nl", &Nl{Start: 1}, "x\ny\n", "     1\tx\n     2\ty\n"},
		{"nl default start", &Nl{}, "x\n", "     1\tx\n"},
		{"sort", &Sort{}, "c\na\nb\n", "a\nb\nc\n"},
		{"sort reverse", &Sort{Reverse: true}, "a\nc\nb\n", "c\nb\na\n"},
		{"sort numeric", &Sort{Numeric: true}, "10\n2\n1\n", "1\n2\n10\n"},
		{"sort numeric reverse", &Sort{Numeric: true, Reverse: true}, "1\n10\n2\n", "10\n2\n1\n"},
		{"tail 2", &Tail{N: 2}, "1\n2\n3\n4\n", "3\n4\n"},
		{"tail default", &Tail{}, "1\n2\n", "1\n2\n"},
		{"tail fewer than n", &Tail{N: 5}, "1\n2\n", "1\n2\n"},
		{"tac", &Tac{}, "1\n2\n3\n", "3\n2\n1\n"},
		{"wc all", &Wc{}, "a b\nc\n", "2\t3\t6\n"},
		{"wc lines", &Wc{Lines: true}, "a\nb\n", "2\n"},
		{"wc words", &Wc{Words: true}, "a b c\n", "3\n"},
		{"wc chars", &Wc{Chars: true}, "ab\n", "3\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := run(t, tc.stage, tc.in); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFilterAndMapStages(t *testing.T) {
	f := &Filter{Fn: func(s string) bool { return strings.HasPrefix(s, "a") }, Desc: "starts-a"}
	if f.Name() != "filter(starts-a)" {
		t.Errorf("filter name = %q", f.Name())
	}
	if got := run(t, f, "apple\nbanana\navocado\n"); got != "apple\navocado\n" {
		t.Errorf("filter got %q", got)
	}
	if (&Filter{}).Name() != "filter" {
		t.Error("default filter name")
	}

	m := &Map{Fn: strings.ToUpper, Desc: "upper"}
	if m.Name() != "map(upper)" {
		t.Errorf("map name = %q", m.Name())
	}
	if got := run(t, m, "a\nb\n"); got != "A\nB\n" {
		t.Errorf("map got %q", got)
	}
	if (&Map{}).Name() != "map" {
		t.Error("default map name")
	}
}

func TestTeeWritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	tee := &Tee{Path: path}
	if got := run(t, tee, "one\ntwo\n"); got != "one\ntwo\n" {
		t.Errorf("tee stdout got %q", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read tee file: %v", err)
	}
	if string(data) != "one\ntwo\n" {
		t.Errorf("tee file got %q", string(data))
	}
}

func TestTeeNoPath(t *testing.T) {
	tee := &Tee{}
	if got := run(t, tee, "x\n"); got != "x\n" {
		t.Errorf("tee no path got %q", got)
	}
}

func TestTeeBadPathErrors(t *testing.T) {
	// A path inside a non-existent directory cannot be created.
	tee := &Tee{Path: filepath.Join(t.TempDir(), "nope", "deep", "f.txt")}
	var out bytes.Buffer
	err := tee.Process(context.Background(), strings.NewReader("x\n"), &out)
	if err == nil {
		t.Error("tee with uncreatable path should error")
	}
}

func TestGrepInvalidPattern(t *testing.T) {
	g := &Grep{Pattern: "("}
	var out bytes.Buffer
	if err := g.Process(context.Background(), strings.NewReader("x\n"), &out); err == nil {
		t.Error("grep with invalid regex should error")
	}
}

func TestSedInvalidPattern(t *testing.T) {
	s := &Sed{Pattern: "(", Replacement: "x", Global: true}
	var out bytes.Buffer
	if err := s.Process(context.Background(), strings.NewReader("x\n"), &out); err == nil {
		t.Error("sed with invalid regex should error")
	}
}

func TestStageContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stages := []Stage{
		&Grep{Pattern: "x"}, &Contains{Substr: "x"}, &Replace{Old: "a", New: "b"},
		&Head{N: 1}, &Skip{N: 0}, &Uniq{}, &Cut{Fields: []int{1}}, &Tr{From: "a", To: "b"},
		&Rev{}, &Nl{}, &Sort{}, &Tail{N: 1}, &Tac{},
		&Filter{Fn: func(string) bool { return true }}, &Map{Fn: func(s string) string { return s }},
	}
	for _, s := range stages {
		var out bytes.Buffer
		err := s.Process(ctx, strings.NewReader("a\nb\n"), &out)
		if err == nil {
			t.Errorf("%s: expected context error", s.Name())
		}
	}
}

func TestCreateFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cf.txt")
	f, err := createFile(path)
	if err != nil {
		t.Fatalf("createFile: %v", err)
	}
	_, _ = f.Write([]byte("hi"))
	_ = f.Close()
	data, _ := os.ReadFile(path)
	if string(data) != "hi" {
		t.Errorf("createFile content = %q", string(data))
	}
}
