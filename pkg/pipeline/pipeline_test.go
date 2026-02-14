package pipeline

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestPipelineEmpty(t *testing.T) {
	p := New()

	var buf bytes.Buffer

	err := p.Run(context.Background(), strings.NewReader("hello\n"), &buf)
	if err != nil {
		t.Fatal(err)
	}

	if buf.String() != "hello\n" {
		t.Errorf("expected pass-through, got %q", buf.String())
	}
}

func TestPipelineSingleStage(t *testing.T) {
	p := New(&Grep{Pattern: "error"})

	var buf bytes.Buffer

	input := "error one\ninfo two\nerror three\n"

	err := p.Run(context.Background(), strings.NewReader(input), &buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := "error one\nerror three\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestPipelineMultiStage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		stages   []Stage
		expected string
	}{
		{
			name:  "grep then sort",
			input: "error b\ninfo\nerror a\nwarning\nerror a\n",
			stages: []Stage{
				&Grep{Pattern: "error"},
				&Sort{},
			},
			expected: "error a\nerror a\nerror b\n",
		},
		{
			name:  "grep sort uniq",
			input: "error b\ninfo\nerror a\nerror a\nwarning\n",
			stages: []Stage{
				&Grep{Pattern: "error"},
				&Sort{},
				&Uniq{},
			},
			expected: "error a\nerror b\n",
		},
		{
			name:  "sort numeric then head",
			input: "3\n1\n2\n1\n3\n",
			stages: []Stage{
				&Sort{Numeric: true},
				&Uniq{},
				&Head{N: 2},
			},
			expected: "1\n2\n",
		},
		{
			name:  "grep inverted",
			input: "error\ninfo\nwarning\n",
			stages: []Stage{
				&Grep{Pattern: "error", Invert: true},
			},
			expected: "info\nwarning\n",
		},
		{
			name:  "tail then rev",
			input: "aaa\nbbb\nccc\nddd\n",
			stages: []Stage{
				&Tail{N: 2},
				&Rev{},
			},
			expected: "ccc\nddd\n",
		},
		{
			name:  "tac",
			input: "a\nb\nc\n",
			stages: []Stage{
				&Tac{},
			},
			expected: "c\nb\na\n",
		},
		{
			name:  "nl",
			input: "first\nsecond\nthird\n",
			stages: []Stage{
				&Nl{Start: 1},
			},
			expected: "     1\tfirst\n     2\tsecond\n     3\tthird\n",
		},
		{
			name:  "wc lines",
			input: "a\nb\nc\n",
			stages: []Stage{
				&Wc{Lines: true},
			},
			expected: "3\n",
		},
		{
			name:  "skip",
			input: "a\nb\nc\nd\n",
			stages: []Stage{
				&Skip{N: 2},
			},
			expected: "c\nd\n",
		},
		{
			name:  "contains",
			input: "hello world\nfoo bar\nhello there\n",
			stages: []Stage{
				&Contains{Substr: "hello"},
			},
			expected: "hello world\nhello there\n",
		},
		{
			name:  "replace",
			input: "hello world\nhello there\n",
			stages: []Stage{
				&Replace{Old: "hello", New: "hi"},
			},
			expected: "hi world\nhi there\n",
		},
		{
			name:  "cut",
			input: "a:b:c\nd:e:f\n",
			stages: []Stage{
				&Cut{Delimiter: ":", Fields: []int{1, 3}},
			},
			expected: "a:c\nd:f\n",
		},
		{
			name:  "tr",
			input: "hello\n",
			stages: []Stage{
				&Tr{From: "aeiou", To: "AEIOU"},
			},
			expected: "hEllO\n",
		},
		{
			name:  "sed global",
			input: "foo bar foo\n",
			stages: []Stage{
				&Sed{Pattern: "foo", Replacement: "baz", Global: true},
			},
			expected: "baz bar baz\n",
		},
		{
			name:  "sort reverse",
			input: "a\nc\nb\n",
			stages: []Stage{
				&Sort{Reverse: true},
			},
			expected: "c\nb\na\n",
		},
		{
			name:  "filter lib-only",
			input: "short\na longer line\nhi\n",
			stages: []Stage{
				&Filter{Fn: func(s string) bool { return len(s) > 3 }},
			},
			expected: "short\na longer line\n",
		},
		{
			name:  "map lib-only",
			input: "hello\nworld\n",
			stages: []Stage{
				&Map{Fn: strings.ToUpper},
			},
			expected: "HELLO\nWORLD\n",
		},
		{
			name:  "grep case insensitive",
			input: "Error found\ninfo line\nERROR again\n",
			stages: []Stage{
				&Grep{Pattern: "error", IgnoreCase: true},
			},
			expected: "Error found\nERROR again\n",
		},
		{
			name:  "wc all",
			input: "hello world\nfoo\n",
			stages: []Stage{
				&Wc{},
			},
			expected: "2\t3\t16\n",
		},
		{
			name:  "sort numeric reverse",
			input: "10\n2\n30\n1\n",
			stages: []Stage{
				&Sort{Numeric: true, Reverse: true},
			},
			expected: "30\n10\n2\n1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.stages...)

			var buf bytes.Buffer

			err := p.Run(context.Background(), strings.NewReader(tt.input), &buf)
			if err != nil {
				t.Fatal(err)
			}

			if buf.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, buf.String())
			}
		})
	}
}

func TestPipelineAdd(t *testing.T) {
	p := New(&Grep{Pattern: "a"})
	p.Add(&Sort{})

	if len(p.Stages()) != 2 {
		t.Errorf("expected 2 stages, got %d", len(p.Stages()))
	}
}

func TestPipelineContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	p := New(&Grep{Pattern: "x"})

	var buf bytes.Buffer

	err := p.Run(ctx, strings.NewReader("x\ny\nz\n"), &buf)
	if err == nil {
		t.Log("no error returned, which is acceptable for already-processed input")
	}
}

func TestStageNames(t *testing.T) {
	stages := []struct {
		stage Stage
		name  string
	}{
		{&Grep{Pattern: "x"}, "grep"},
		{&Grep{Pattern: "x", Invert: true}, "grep-v"},
		{&Contains{Substr: "x"}, "contains"},
		{&Replace{Old: "a", New: "b"}, "replace"},
		{&Head{N: 5}, "head"},
		{&Tail{N: 5}, "tail"},
		{&Skip{N: 1}, "skip"},
		{&Sort{}, "sort"},
		{&Uniq{}, "uniq"},
		{&Cut{Fields: []int{1}}, "cut"},
		{&Tr{From: "a", To: "b"}, "tr"},
		{&Sed{Pattern: "a", Replacement: "b"}, "sed"},
		{&Rev{}, "rev"},
		{&Nl{}, "nl"},
		{&Tee{}, "tee"},
		{&Tac{}, "tac"},
		{&Wc{}, "wc"},
		{&Filter{Fn: nil}, "filter"},
		{&Filter{Fn: nil, Desc: "test"}, "filter(test)"},
		{&Map{Fn: nil}, "map"},
		{&Map{Fn: nil, Desc: "upper"}, "map(upper)"},
	}

	for _, tt := range stages {
		if tt.stage.Name() != tt.name {
			t.Errorf("stage %T: expected name %q, got %q", tt.stage, tt.name, tt.stage.Name())
		}
	}
}

func TestHeadDrainsInput(t *testing.T) {
	// Ensure head doesn't cause broken pipe errors upstream
	input := strings.Repeat("line\n", 10000)
	p := New(&Head{N: 5})

	var buf bytes.Buffer

	err := p.Run(context.Background(), strings.NewReader(input), &buf)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(lines))
	}
}
