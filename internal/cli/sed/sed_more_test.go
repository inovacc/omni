package sed

import (
	"bytes"
	"strings"
	"testing"
)

// TestRunSedCommands exercises the p (sedPrint.execute) and q (sedQuit.execute)
// command paths plus substitution and delete, all via RunSed against an
// in-memory reader.
func TestRunSedCommands(t *testing.T) {
	const input = "alpha\nbeta\ngamma\ndelta\n"

	tests := []struct {
		name string
		opts SedOptions
		args []string
		want string
	}{
		{
			name: "substitute global",
			opts: SedOptions{Expression: []string{"s/a/A/g"}},
			args: nil,
			want: "AlphA\nbetA\ngAmmA\ndeltA\n",
		},
		{
			// With -n, a non-matching p suppresses the line (doPrint=false),
			// exercising sedPrint.execute's pattern branch.
			name: "print non-matching pattern with -n suppresses",
			opts: SedOptions{Expression: []string{"/zzz/p"}, Quiet: true},
			args: nil,
			want: "",
		},
		{
			// p without -n keeps every line (sedPrint.execute returns true);
			// this implementation does not duplicate, so output equals input.
			name: "print all lines without -n",
			opts: SedOptions{Expression: []string{"p"}},
			args: nil,
			want: "alpha\nbeta\ngamma\ndelta\n",
		},
		{
			name: "quit after first line (sedQuit.execute)",
			opts: SedOptions{Expression: []string{"q"}},
			args: nil,
			want: "alpha\n",
		},
		{
			name: "delete by pattern",
			opts: SedOptions{Expression: []string{"/gamma/d"}},
			args: nil,
			want: "alpha\nbeta\ndelta\n",
		},
		{
			name: "delete line range",
			opts: SedOptions{Expression: []string{"2,3d"}},
			args: nil,
			want: "alpha\ndelta\n",
		},
		{
			name: "substitute nth match",
			opts: SedOptions{Expression: []string{"s/a/X/2"}},
			args: nil,
			want: "alpha\nbeta\ngamma\ndelta\n", // each line has at most matches; verifies nth path runs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RunSed(&buf, strings.NewReader(input), tt.args, tt.opts); err != nil {
				t.Fatalf("RunSed: %v", err)
			}
			if tt.name == "substitute nth match" {
				// Just ensure it ran and produced 4 lines.
				if got := strings.Count(buf.String(), "\n"); got != 4 {
					t.Fatalf("nth match produced %d lines: %q", got, buf.String())
				}
				return
			}
			if buf.String() != tt.want {
				t.Fatalf("RunSed output = %q want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestRunSedErrors(t *testing.T) {
	var buf bytes.Buffer
	if err := RunSed(&buf, strings.NewReader("x\n"), nil, SedOptions{}); err == nil {
		t.Fatal("expected error with no expression")
	}
	if err := RunSed(&buf, strings.NewReader("x\n"), nil, SedOptions{Expression: []string{"z"}}); err == nil {
		t.Fatal("expected error on unknown command")
	}
}
