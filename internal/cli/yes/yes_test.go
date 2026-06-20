package yes

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// errWriter fails on the first write, used to exercise the broken-pipe branch.
type errWriter struct{ err error }

func (e errWriter) Write([]byte) (int, error) { return 0, e.err }

// TestRunYesContextCancelled asserts RunYes honours a cancelled context and
// returns context.Canceled instead of looping forever.
func TestRunYesContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already done before the loop starts

	var buf bytes.Buffer
	err := RunYes(ctx, &buf, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunYes(cancelled) error = %v, want context.Canceled", err)
	}
}

// TestRunYesWriteError asserts a write failure is classified as cmderr.ErrIO.
func TestRunYesWriteError(t *testing.T) {
	w := errWriter{err: errors.New("broken pipe")}
	err := RunYes(context.Background(), w, []string{"n"})
	if !cmderr.IsIO(err) {
		t.Fatalf("RunYes(write error) error = %v, want cmderr.ErrIO", err)
	}
	if !strings.Contains(err.Error(), "yes: write") {
		t.Errorf("error %q does not mention 'yes: write'", err.Error())
	}
}

// TestRunYesCustomOutput asserts that args are joined and written before
// cancellation, and that the default "y" is used with no args.
func TestRunYesCustomOutput(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"default", nil, "y"},
		{"single arg", []string{"yep"}, "yep"},
		{"joined args", []string{"a", "b", "c"}, "a b c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// limitWriter cancels the context after the first successful write so
			// the otherwise-infinite loop terminates deterministically.
			ctx, cancel := context.WithCancel(context.Background())
			var buf bytes.Buffer
			lw := &cancelAfterWriteWriter{buf: &buf, cancel: cancel}
			err := RunYes(ctx, lw, tt.args)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("RunYes error = %v, want context.Canceled", err)
			}
			line := strings.SplitN(buf.String(), "\n", 2)[0]
			if line != tt.want {
				t.Errorf("first line = %q, want %q", line, tt.want)
			}
		})
	}
}

// cancelAfterWriteWriter records output then cancels the context so RunYes exits.
type cancelAfterWriteWriter struct {
	buf    *bytes.Buffer
	cancel context.CancelFunc
}

func (c *cancelAfterWriteWriter) Write(p []byte) (int, error) {
	n, err := c.buf.Write(p)
	c.cancel()
	return n, err
}

func TestYes(t *testing.T) {
	t.Run("default output", func(t *testing.T) {
		result := Yes("", 5)

		if len(result) != 5 {
			t.Errorf("Yes() got %d items, want 5", len(result))
		}

		for i, v := range result {
			if v != "y" {
				t.Errorf("Yes()[%d] = %v, want 'y'", i, v)
			}
		}
	})

	t.Run("custom output", func(t *testing.T) {
		result := Yes("hello", 3)

		if len(result) != 3 {
			t.Errorf("Yes() got %d items, want 3", len(result))
		}

		for i, v := range result {
			if v != "hello" {
				t.Errorf("Yes()[%d] = %v, want 'hello'", i, v)
			}
		}
	})

	t.Run("zero count", func(t *testing.T) {
		result := Yes("y", 0)

		if len(result) != 0 {
			t.Errorf("Yes() with 0 count got %d items, want 0", len(result))
		}
	})

	t.Run("single item", func(t *testing.T) {
		result := Yes("test", 1)

		if len(result) != 1 || result[0] != "test" {
			t.Errorf("Yes() single = %v", result)
		}
	})

	t.Run("large count", func(t *testing.T) {
		result := Yes("x", 1000)

		if len(result) != 1000 {
			t.Errorf("Yes() got %d items, want 1000", len(result))
		}
	})

	t.Run("unicode output", func(t *testing.T) {
		result := Yes("世界", 3)

		if len(result) != 3 {
			t.Errorf("Yes() got %d items, want 3", len(result))
		}

		for i, v := range result {
			if v != "世界" {
				t.Errorf("Yes()[%d] = %v, want '世界'", i, v)
			}
		}
	})

	t.Run("whitespace output", func(t *testing.T) {
		result := Yes("  ", 2)

		if len(result) != 2 {
			t.Errorf("Yes() got %d items, want 2", len(result))
		}

		for i, v := range result {
			if v != "  " {
				t.Errorf("Yes()[%d] = %v, want '  '", i, v)
			}
		}
	})

	t.Run("special characters", func(t *testing.T) {
		result := Yes("!@#$%", 2)

		if len(result) != 2 || result[0] != "!@#$%" {
			t.Errorf("Yes() special = %v", result)
		}
	})
}
