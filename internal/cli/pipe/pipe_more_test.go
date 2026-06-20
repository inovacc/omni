package pipe

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/command"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
	"github.com/spf13/cobra"
)

// newTestUnifiedRegistry builds an in-process command.Registry with a handful of
// pure commands so the pipe package can be exercised entirely offline.
func newTestUnifiedRegistry() *command.Registry {
	reg := command.NewRegistry()

	// "upper" uppercases all of stdin.
	reg.Register("upper", command.CommandFunc(func(_ context.Context, w io.Writer, r io.Reader, _ []string) error {
		b, _ := io.ReadAll(r)
		_, _ = w.Write([]byte(strings.ToUpper(string(b))))
		return nil
	}))

	// "echoarg" writes its args joined by spaces (ignores stdin).
	reg.Register("echoarg", command.CommandFunc(func(_ context.Context, w io.Writer, _ io.Reader, args []string) error {
		_, _ = io.WriteString(w, strings.Join(args, " "))
		return nil
	}))

	// "passthru" copies stdin to stdout verbatim.
	reg.Register("passthru", command.CommandFunc(func(_ context.Context, w io.Writer, r io.Reader, _ []string) error {
		_, _ = io.Copy(w, r)
		return nil
	}))

	// "boom" always fails.
	reg.Register("boom", command.CommandFunc(func(_ context.Context, _ io.Writer, _ io.Reader, _ []string) error {
		return io.ErrUnexpectedEOF
	}))

	return reg
}

// newTestCobraRoot builds a minimal Cobra tree to exercise the Cobra fallback path.
func newTestCobraRoot() *cobra.Command {
	root := &cobra.Command{Use: "omni"}
	cat := &cobra.Command{
		Use: "ccat",
		RunE: func(cmd *cobra.Command, _ []string) error {
			b, _ := io.ReadAll(cmd.InOrStdin())
			_, _ = cmd.OutOrStdout().Write(b)
			return nil
		},
	}
	root.AddCommand(cat)
	return root
}

func TestRun_UnifiedPipeline(t *testing.T) {
	reg := NewRegistryWithUnified(nil, newTestUnifiedRegistry())

	tests := []struct {
		name    string
		args    []string
		opts    Options
		want    string
		wantErr bool
	}{
		{
			name: "single upper",
			args: []string{"echoarg hello | upper"},
			want: "HELLO",
		},
		{
			name: "passthru then upper",
			args: []string{"echoarg abc | passthru | upper"},
			want: "ABC",
		},
		{
			name:    "no commands",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "failing command",
			args:    []string{"echoarg x | boom"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := Run(&buf, tt.args, tt.opts, reg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Run() err=%v wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && buf.String() != tt.want {
				t.Errorf("Run() output=%q want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestRun_VerboseAndJSON(t *testing.T) {
	reg := NewRegistryWithUnified(nil, newTestUnifiedRegistry())

	t.Run("verbose", func(t *testing.T) {
		var buf strings.Builder
		if err := Run(&buf, []string{"echoarg hi | upper"}, Options{Verbose: true}, reg); err != nil {
			t.Fatalf("Run verbose err=%v", err)
		}
		if !strings.Contains(buf.String(), "=== Step") {
			t.Errorf("verbose output missing step markers: %q", buf.String())
		}
	})

	t.Run("json", func(t *testing.T) {
		var buf strings.Builder
		if err := Run(&buf, []string{"echoarg hi | upper"}, Options{OutputFormat: output.FormatJSON}, reg); err != nil {
			t.Fatalf("Run json err=%v", err)
		}
		var res Result
		if err := json.Unmarshal([]byte(buf.String()), &res); err != nil {
			t.Fatalf("json output not parseable: %v\n%s", err, buf.String())
		}
		if !res.Success || res.Output != "HI" {
			t.Errorf("unexpected result: %+v", res)
		}
	})
}

func TestRunWithInput_Pipeline(t *testing.T) {
	reg := NewRegistryWithUnified(nil, newTestUnifiedRegistry())

	tests := []struct {
		name    string
		input   string
		args    []string
		opts    Options
		want    string
		wantErr bool
	}{
		{
			name:  "input through upper",
			input: "hello world",
			args:  []string{"upper"},
			want:  "HELLO WORLD",
		},
		{
			name:  "input passthru then upper",
			input: "abc",
			args:  []string{"passthru | upper"},
			want:  "ABC",
		},
		{
			name:    "input failing",
			input:   "x",
			args:    []string{"boom"},
			wantErr: true,
		},
		{
			name:  "json output",
			input: "z",
			args:  []string{"upper"},
			opts:  Options{OutputFormat: output.FormatJSON},
			want:  "", // checked via JSON below
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := RunWithInput(&buf, tt.input, tt.args, tt.opts, reg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RunWithInput() err=%v wantErr=%v", err, tt.wantErr)
			}
			if tt.opts.OutputFormat == output.FormatJSON && !tt.wantErr {
				var res Result
				if err := json.Unmarshal([]byte(buf.String()), &res); err != nil {
					t.Fatalf("json parse: %v", err)
				}
				if res.Output != "Z" {
					t.Errorf("json Output=%q want Z", res.Output)
				}
				return
			}
			if !tt.wantErr && tt.want != "" && buf.String() != tt.want {
				t.Errorf("RunWithInput() output=%q want %q", buf.String(), tt.want)
			}
		})
	}
}

func TestRunWithInput_Verbose(t *testing.T) {
	reg := NewRegistryWithUnified(nil, newTestUnifiedRegistry())
	var buf strings.Builder
	if err := RunWithInput(&buf, "hello", []string{"upper"}, Options{Verbose: true}, reg); err != nil {
		t.Fatalf("err=%v", err)
	}
	if !strings.Contains(buf.String(), "=== Step 1") {
		t.Errorf("missing step marker: %q", buf.String())
	}
}

func TestExecuteCommand_CobraFallback(t *testing.T) {
	// No unified registry → forces the Cobra dispatch path.
	reg := NewRegistry(newTestCobraRoot())

	var out strings.Builder
	err := executeCommand(reg, []string{"ccat"}, strings.NewReader("piped"), &out)
	if err != nil {
		t.Fatalf("executeCommand cobra err=%v", err)
	}
	if out.String() != "piped" {
		t.Errorf("cobra fallback output=%q want piped", out.String())
	}
}

func TestExecuteCommand_Errors(t *testing.T) {
	t.Run("nil registry", func(t *testing.T) {
		var out strings.Builder
		if err := executeCommand(nil, []string{"x"}, nil, &out); err == nil {
			t.Error("expected error for nil registry")
		}
	})

	t.Run("unknown command cobra", func(t *testing.T) {
		reg := NewRegistry(newTestCobraRoot())
		var out strings.Builder
		if err := executeCommand(reg, []string{"nope"}, nil, &out); err == nil {
			t.Error("expected unknown command error")
		}
	})

	t.Run("nil root cobra", func(t *testing.T) {
		reg := NewRegistry(nil)
		var out strings.Builder
		if err := executeCommand(reg, []string{"x"}, nil, &out); err == nil {
			t.Error("expected error for nil root")
		}
	})
}

func TestRun_UnifiedFallsBackToCobra(t *testing.T) {
	// Unified registry present but does not contain "ccat" → must fall through to Cobra.
	reg := NewRegistryWithUnified(newTestCobraRoot(), newTestUnifiedRegistry())
	var buf strings.Builder
	if err := RunWithInput(&buf, "data", []string{"ccat"}, Options{}, reg); err != nil {
		t.Fatalf("err=%v", err)
	}
	if buf.String() != "data" {
		t.Errorf("output=%q want data", buf.String())
	}
}
