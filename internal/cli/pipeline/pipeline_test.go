package pipeline

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		args     []string
		opts     Options
		expected string
		wantErr  bool
	}{
		{
			name:     "grep and sort",
			input:    "error b\ninfo\nerror a\n",
			args:     []string{"grep error", "sort"},
			expected: "error a\nerror b\n",
		},
		{
			name:     "sort uniq head",
			input:    "3\n1\n2\n1\n3\n",
			args:     []string{"sort -n", "uniq", "head 2"},
			expected: "1\n2\n",
		},
		{
			name:     "grep case insensitive sort uniq",
			input:    "ERROR a\nerror b\nInfo\nerror a\nWarning\n",
			args:     []string{"grep -i error", "sort", "uniq"},
			expected: "ERROR a\nerror a\nerror b\n",
		},
		{
			name:     "sed and rev",
			input:    "hello world\n",
			args:     []string{"sed s/hello/hi/g", "rev"},
			expected: "dlrow ih\n",
		},
		{
			name:     "wc lines",
			input:    "a\nb\nc\n",
			args:     []string{"wc -l"},
			expected: "3\n",
		},
		{
			name:    "no args",
			input:   "hello\n",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "invalid stage",
			input:   "hello\n",
			args:    []string{"notacommand"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := Run(&buf, strings.NewReader(tt.input), tt.args, tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if buf.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, buf.String())
			}
		})
	}
}

func TestRunVerbose(t *testing.T) {
	var buf bytes.Buffer

	err := Run(&buf, strings.NewReader("a\nb\n"), []string{"sort", "uniq"}, Options{Verbose: true})
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "--- stage 1: sort ---") {
		t.Error("expected verbose output for stage 1")
	}

	if !strings.Contains(output, "--- stage 2: uniq ---") {
		t.Error("expected verbose output for stage 2")
	}
}

func TestRunFileNotFound(t *testing.T) {
	var buf bytes.Buffer

	err := Run(&buf, nil, []string{"sort"}, Options{File: "/nonexistent/file.txt"})
	if err == nil {
		t.Error("expected error for missing file")
	}
}
