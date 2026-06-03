package forloop

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
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

// TestForLoop_InjectionSafe is a permanent regression test pinning the
// shell-command-injection hardening of forloop's execution path.
//
// Threat model: the per-iteration loop VALUE (here, an item passed to RunEach)
// is attacker-controlled data. Before the fix it was string-concatenated into
// the executed `sh -c` command, so a value such as `x; touch MARKER` or
// `$(touch MARKER)` / backtick form would be reparsed by the shell as
// additional command syntax (injection). The fix passes each value via the
// child process environment (cmd.Env) and leaves the trusted command TEMPLATE's
// $var/${var} references intact, so on POSIX sh the shell expands the value as a
// single inert word and its metacharacters never become commands.
//
// The command template is benign (`echo $item`) and references the loop
// variable; the attack lives entirely in the data VALUE. The security assertion
// is that the marker file is never created.
//
// Windows cmd (supply-01): the env-binding strategy alone is NOT sufficient for
// cmd.exe, which performs `%var%` substitution at PARSE TIME — the raw value
// (including `&`/`&&`/`|` separators) would be spliced into the command line
// before tokenization, so `echo %item%` with item="x & type nul > MARKER" would
// run the injected command and create the marker. The fix invokes the shell as
// `cmd /V:ON /C ...` and rewrites references to the DELAYED-expansion form
// `!item!`, which expands AFTER tokenization: the value lands as a single inert
// token and its metacharacters are not reparsed as command syntax. This test
// exercises both hosts (sh and cmd) against their respective separators and
// asserts the marker is never created on either.
func TestForLoop_InjectionSafe(t *testing.T) {
	tmp := t.TempDir()
	// Marker path must NOT exist after execution. Keep it free of shell
	// metacharacters and spaces so a successful injection would actually create it.
	marker := filepath.Join(tmp, "INJECTED_MARKER")

	// Benign template that references the loop variable. The template never
	// contains the malicious payload; only the data VALUE carries the attack.
	const command = "echo $item"

	// Injection vectors differ per shell. Each tries to create <marker>, which
	// would happen only if the value were reparsed as command syntax.
	var payloads []string
	if runtime.GOOS == "windows" {
		// cmd.exe separators: `&`, `&&`, `|`, plus a caret/quote variant. Use
		// `type nul > <marker>` / `echo > <marker>` which create the file.
		payloads = []string{
			"x & type nul > " + marker,
			"x && echo pwned > " + marker,
			"x | echo > " + marker,
			`x &^ echo "pwned" > ` + marker,
		}
	} else {
		// POSIX sh vectors: `;`, `&&`, `$(...)`, and backticks. Each tries to
		// run `touch <marker>`, which would create the file if reparsed as code.
		payloads = []string{
			"x; touch " + marker,
			"x && touch " + marker,
			"x$(touch " + marker + ")",
			"x`touch " + marker + "`",
		}
	}

	for _, payload := range payloads {
		var buf bytes.Buffer
		// Real execution path (not dry-run): RunEach -> executeCommand ->
		// sh -c (POSIX) or cmd /V:ON /C (Windows).
		if err := RunEach(&buf, []string{payload}, command, Options{}); err != nil {
			// A non-nil error is acceptable (the shell may complain about the
			// inert value); the security property is the marker's absence.
			t.Logf("RunEach(payload=%q) returned err=%v (tolerated)", payload, err)
		}

		_, statErr := os.Stat(marker)
		switch {
		case statErr == nil:
			t.Fatalf("command injection: marker file %q was created from loop value %q; "+
				"the malicious value was executed as a command instead of treated as inert data",
				marker, payload)
		case !errors.Is(statErr, fs.ErrNotExist):
			t.Fatalf("unexpected error stat-ing marker %q: %v", marker, statErr)
		}
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
