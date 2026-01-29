package forloop

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunRangeDryRun(t *testing.T) {
	tests := []struct {
		name    string
		start   int
		end     int
		step    int
		command string
		opts    Options
		want    []string
	}{
		{
			name:    "simple range",
			start:   1,
			end:     3,
			step:    0,
			command: "echo $i",
			opts:    Options{DryRun: true},
			want:    []string{"echo 1", "echo 2", "echo 3"},
		},
		{
			name:    "range with step",
			start:   0,
			end:     10,
			step:    2,
			command: "echo $i",
			opts:    Options{DryRun: true},
			want:    []string{"echo 0", "echo 2", "echo 4", "echo 6", "echo 8", "echo 10"},
		},
		{
			name:    "descending range",
			start:   5,
			end:     1,
			step:    -1,
			command: "echo $i",
			opts:    Options{DryRun: true},
			want:    []string{"echo 5", "echo 4", "echo 3", "echo 2", "echo 1"},
		},
		{
			name:    "custom variable",
			start:   1,
			end:     2,
			step:    0,
			command: "echo ${num}",
			opts:    Options{DryRun: true, Variable: "num"},
			want:    []string{"echo 1", "echo 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunRange(&buf, tt.start, tt.end, tt.step, tt.command, tt.opts)
			if err != nil {
				t.Errorf("RunRange() error = %v", err)
				return
			}

			lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
			if len(lines) != len(tt.want) {
				t.Errorf("RunRange() got %d lines, want %d", len(lines), len(tt.want))
				return
			}

			for i, want := range tt.want {
				if lines[i] != want {
					t.Errorf("RunRange() line %d = %q, want %q", i, lines[i], want)
				}
			}
		})
	}
}

func TestRunEachDryRun(t *testing.T) {
	tests := []struct {
		name    string
		items   []string
		command string
		opts    Options
		want    []string
	}{
		{
			name:    "simple list",
			items:   []string{"a", "b", "c"},
			command: "echo $item",
			opts:    Options{DryRun: true},
			want:    []string{"echo a", "echo b", "echo c"},
		},
		{
			name:    "with spaces",
			items:   []string{"hello world", "foo bar"},
			command: "echo ${item}",
			opts:    Options{DryRun: true},
			want:    []string{"echo hello world", "echo foo bar"},
		},
		{
			name:    "custom variable",
			items:   []string{"x", "y"},
			command: "process $val",
			opts:    Options{DryRun: true, Variable: "val"},
			want:    []string{"process x", "process y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunEach(&buf, tt.items, tt.command, tt.opts)
			if err != nil {
				t.Errorf("RunEach() error = %v", err)
				return
			}

			lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
			if len(lines) != len(tt.want) {
				t.Errorf("RunEach() got %d lines, want %d", len(lines), len(tt.want))
				return
			}

			for i, want := range tt.want {
				if lines[i] != want {
					t.Errorf("RunEach() line %d = %q, want %q", i, lines[i], want)
				}
			}
		})
	}
}

func TestRunLinesDryRun(t *testing.T) {
	input := "line1\nline2\nline3"
	reader := strings.NewReader(input)

	var buf bytes.Buffer

	opts := Options{DryRun: true}

	err := RunLines(&buf, reader, "echo $line", opts)
	if err != nil {
		t.Fatalf("RunLines() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	want := []string{"echo line1", "echo line2", "echo line3"}

	if len(lines) != len(want) {
		t.Errorf("RunLines() got %d lines, want %d", len(lines), len(want))
		return
	}

	for i, w := range want {
		if lines[i] != w {
			t.Errorf("RunLines() line %d = %q, want %q", i, lines[i], w)
		}
	}
}

func TestRunLinesWithLineNumber(t *testing.T) {
	input := "a\nb"
	reader := strings.NewReader(input)

	var buf bytes.Buffer

	opts := Options{DryRun: true}

	err := RunLines(&buf, reader, "echo $n: $line", opts)
	if err != nil {
		t.Fatalf("RunLines() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")

	if lines[0] != "echo 1: a" {
		t.Errorf("Line 1 = %q, want %q", lines[0], "echo 1: a")
	}

	if lines[1] != "echo 2: b" {
		t.Errorf("Line 2 = %q, want %q", lines[1], "echo 2: b")
	}
}

func TestRunSplitDryRun(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		delimiter string
		command   string
		want      []string
	}{
		{
			name:      "comma separated",
			input:     "a,b,c",
			delimiter: ",",
			command:   "echo $item",
			want:      []string{"echo a", "echo b", "echo c"},
		},
		{
			name:      "colon separated",
			input:     "/bin:/usr/bin:/usr/local/bin",
			delimiter: ":",
			command:   "ls $item",
			want:      []string{"ls /bin", "ls /usr/bin", "ls /usr/local/bin"},
		},
		{
			name:      "with index",
			input:     "x,y",
			delimiter: ",",
			command:   "echo $i: $item",
			want:      []string{"echo 0: x", "echo 1: y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			opts := Options{DryRun: true}

			err := RunSplit(&buf, tt.input, tt.delimiter, tt.command, opts)
			if err != nil {
				t.Errorf("RunSplit() error = %v", err)
				return
			}

			lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
			if len(lines) != len(tt.want) {
				t.Errorf("RunSplit() got %d lines, want %d", len(lines), len(tt.want))
				return
			}

			for i, want := range tt.want {
				if lines[i] != want {
					t.Errorf("RunSplit() line %d = %q, want %q", i, lines[i], want)
				}
			}
		})
	}
}

func TestReplaceVariable(t *testing.T) {
	tests := []struct {
		cmd     string
		varName string
		value   string
		want    string
	}{
		{"echo $item", "item", "test", "echo test"},
		{"echo ${item}", "item", "test", "echo test"},
		{"echo $item $item", "item", "x", "echo x x"},
		{"$item-suffix", "item", "pre", "pre-suffix"},
		{"echo $items", "item", "x", "echo $items"}, // Should not replace partial
		{"echo $i", "i", "1", "echo 1"},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := replaceVariable(tt.cmd, tt.varName, tt.value)
			if got != tt.want {
				t.Errorf("replaceVariable(%q, %q, %q) = %q, want %q",
					tt.cmd, tt.varName, tt.value, got, tt.want)
			}
		})
	}
}

func TestEmptyItems(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{DryRun: true}

	err := RunEach(&buf, []string{}, "echo $item", opts)
	if err != nil {
		t.Errorf("RunEach() with empty items error = %v", err)
	}

	if buf.String() != "" {
		t.Errorf("RunEach() with empty items should produce no output")
	}
}
