package pipe

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseCommands(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		separator string
		want      []string
	}{
		{
			name:      "brace syntax simple",
			args:      []string{"{ls -la}", "{grep .go}", "{wc -l}"},
			separator: "|",
			want:      []string{"ls -la", "grep .go", "wc -l"},
		},
		{
			name:      "brace syntax with commas",
			args:      []string{"{cat file.txt}, {sort}, {uniq}"},
			separator: "|",
			want:      []string{"cat file.txt", "sort", "uniq"},
		},
		{
			name:      "brace syntax single string",
			args:      []string{"{cat file.txt}, {grep pattern}, {sort}"},
			separator: "|",
			want:      []string{"cat file.txt", "grep pattern", "sort"},
		},
		{
			name:      "brace syntax with jq",
			args:      []string{"{cat data.json}", "{jq '.users[]'}"},
			separator: "|",
			want:      []string{"cat data.json", "jq '.users[]'"},
		},
		{
			name:      "single string with pipes",
			args:      []string{"cat file.txt | grep pattern | sort"},
			separator: "|",
			want:      []string{"cat file.txt", "grep pattern", "sort"},
		},
		{
			name:      "args with pipe separator",
			args:      []string{"cat", "file.txt", "|", "grep", "pattern"},
			separator: "|",
			want:      []string{"cat file.txt", "grep pattern"},
		},
		{
			name:      "custom separator",
			args:      []string{"cat file.txt -> grep pattern -> sort"},
			separator: "->",
			want:      []string{"cat file.txt", "grep pattern", "sort"},
		},
		{
			name:      "separate quoted args",
			args:      []string{"cat file.txt", "grep pattern", "sort"},
			separator: "|",
			want:      []string{"cat file.txt", "grep pattern", "sort"},
		},
		{
			name:      "empty separator parts ignored",
			args:      []string{"cat file.txt | | grep pattern"},
			separator: "|",
			want:      []string{"cat file.txt", "grep pattern"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommands(tt.args, tt.separator)

			if len(got) != len(tt.want) {
				t.Errorf("parseCommands() got %d commands, want %d", len(got), len(tt.want))
				t.Errorf("got: %v", got)
				t.Errorf("want: %v", tt.want)

				return
			}

			for i, cmd := range got {
				if cmd != tt.want[i] {
					t.Errorf("parseCommands()[%d] = %q, want %q", i, cmd, tt.want[i])
				}
			}
		})
	}
}

func TestParseCommandLine(t *testing.T) {
	tests := []struct {
		name    string
		cmdLine string
		want    []string
	}{
		{
			name:    "simple command",
			cmdLine: "cat file.txt",
			want:    []string{"cat", "file.txt"},
		},
		{
			name:    "with flags",
			cmdLine: "grep -i pattern file.txt",
			want:    []string{"grep", "-i", "pattern", "file.txt"},
		},
		{
			name:    "quoted string",
			cmdLine: `grep "hello world" file.txt`,
			want:    []string{"grep", "hello world", "file.txt"},
		},
		{
			name:    "single quoted",
			cmdLine: `grep 'hello world' file.txt`,
			want:    []string{"grep", "hello world", "file.txt"},
		},
		{
			name:    "multiple spaces",
			cmdLine: "cat    file.txt",
			want:    []string{"cat", "file.txt"},
		},
		{
			name:    "tabs",
			cmdLine: "cat\tfile.txt",
			want:    []string{"cat", "file.txt"},
		},
		{
			name:    "escaped space",
			cmdLine: `cat file\ name.txt`,
			want:    []string{"cat", "file name.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommandLine(tt.cmdLine)

			if len(got) != len(tt.want) {
				t.Errorf("parseCommandLine() got %d parts, want %d", len(got), len(tt.want))
				t.Errorf("got: %v", got)

				return
			}

			for i, part := range got {
				if part != tt.want[i] {
					t.Errorf("parseCommandLine()[%d] = %q, want %q", i, part, tt.want[i])
				}
			}
		})
	}
}

func TestRunWithInput(t *testing.T) {
	// Test without a real registry - just test parsing and structure
	tests := []struct {
		name    string
		input   string
		args    []string
		opts    Options
		wantErr bool
	}{
		{
			name:    "empty args",
			input:   "hello",
			args:    []string{},
			opts:    Options{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunWithInput(&buf, tt.input, tt.args, tt.opts, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunWithInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResultJSON(t *testing.T) {
	result := Result{
		Commands: []CommandResult{
			{Command: "cat file.txt", Output: "hello\nworld\n"},
			{Command: "grep hello", Output: "hello\n"},
		},
		Output:  "hello\n",
		Success: true,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var decoded Result
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if len(decoded.Commands) != 2 {
		t.Errorf("Commands length = %d, want 2", len(decoded.Commands))
	}

	if !decoded.Success {
		t.Errorf("Success = false, want true")
	}

	if decoded.Output != "hello\n" {
		t.Errorf("Output = %q, want %q", decoded.Output, "hello\n")
	}
}

func TestCommandResultWithError(t *testing.T) {
	result := Result{
		Commands: []CommandResult{
			{Command: "cat file.txt", Output: "content"},
			{Command: "invalid-cmd", Error: "unknown command"},
		},
		Output:  "",
		Success: false,
		Error:   "command 2 failed: unknown command",
	}

	if result.Success {
		t.Errorf("Success should be false")
	}

	if !strings.Contains(result.Error, "command 2 failed") {
		t.Errorf("Error should mention which command failed")
	}
}

func TestSubstituteVariables(t *testing.T) {
	tests := []struct {
		name        string
		cmdStr      string
		output      string
		varName     string
		wantCmds    []string
		wantIsIter  bool
	}{
		{
			name:       "no substitution",
			cmdStr:     "mkdir test",
			output:     "some-uuid\n",
			varName:    "OUT",
			wantCmds:   []string{"mkdir test"},
			wantIsIter: false,
		},
		{
			name:       "single var substitution with $OUT",
			cmdStr:     "mkdir $OUT",
			output:     "my-uuid-123\n",
			varName:    "OUT",
			wantCmds:   []string{"mkdir my-uuid-123"},
			wantIsIter: false,
		},
		{
			name:       "single var substitution with ${OUT}",
			cmdStr:     "mkdir ${OUT}",
			output:     "my-uuid-456\n",
			varName:    "OUT",
			wantCmds:   []string{"mkdir my-uuid-456"},
			wantIsIter: false,
		},
		{
			name:       "custom variable name",
			cmdStr:     "mkdir $UUID",
			output:     "custom-id\n",
			varName:    "UUID",
			wantCmds:   []string{"mkdir custom-id"},
			wantIsIter: false,
		},
		{
			name:       "iteration with [$OUT...]",
			cmdStr:     "mkdir [$OUT...]",
			output:     "uuid1\nuuid2\nuuid3\n",
			varName:    "OUT",
			wantCmds:   []string{"mkdir uuid1", "mkdir uuid2", "mkdir uuid3"},
			wantIsIter: true,
		},
		{
			name:       "iteration with ${OUT} braces",
			cmdStr:     "mkdir [${OUT}...]",
			output:     "id1\nid2\n",
			varName:    "OUT",
			wantCmds:   []string{"mkdir id1", "mkdir id2"},
			wantIsIter: true,
		},
		{
			name:       "custom var iteration",
			cmdStr:     "touch [$UUID...]",
			output:     "file1\nfile2\nfile3\n",
			varName:    "UUID",
			wantCmds:   []string{"touch file1", "touch file2", "touch file3"},
			wantIsIter: true,
		},
		{
			name:       "empty var name defaults to OUT",
			cmdStr:     "mkdir $OUT",
			output:     "default-test\n",
			varName:    "",
			wantCmds:   []string{"mkdir default-test"},
			wantIsIter: false,
		},
		{
			name:       "multiple lines uses last non-empty",
			cmdStr:     "mkdir $OUT",
			output:     "first\nsecond\nthird\n",
			varName:    "OUT",
			wantCmds:   []string{"mkdir third"},
			wantIsIter: false,
		},
		{
			name:       "skip empty lines in iteration",
			cmdStr:     "mkdir [$OUT...]",
			output:     "one\n\ntwo\n\n",
			varName:    "OUT",
			wantCmds:   []string{"mkdir one", "mkdir two"},
			wantIsIter: true,
		},
		{
			name:       "var in middle of command",
			cmdStr:     "echo prefix-$OUT-suffix",
			output:     "value\n",
			varName:    "OUT",
			wantCmds:   []string{"echo prefix-value-suffix"},
			wantIsIter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmds, gotIsIter := substituteVariables(tt.cmdStr, tt.output, tt.varName)

			if gotIsIter != tt.wantIsIter {
				t.Errorf("substituteVariables() isIteration = %v, want %v", gotIsIter, tt.wantIsIter)
			}

			if len(gotCmds) != len(tt.wantCmds) {
				t.Errorf("substituteVariables() got %d commands, want %d", len(gotCmds), len(tt.wantCmds))
				t.Errorf("got: %v", gotCmds)
				t.Errorf("want: %v", tt.wantCmds)

				return
			}

			for i, cmd := range gotCmds {
				if cmd != tt.wantCmds[i] {
					t.Errorf("substituteVariables()[%d] = %q, want %q", i, cmd, tt.wantCmds[i])
				}
			}
		})
	}
}
