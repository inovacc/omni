package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
)

func TestCommandFunc(t *testing.T) {
	tests := []struct {
		name    string
		fn      CommandFunc
		wantErr error
	}{
		{
			name: "returns nil",
			fn: func(_ context.Context, _ io.Writer, _ io.Reader, _ []string) error {
				return nil
			},
			wantErr: nil,
		},
		{
			name: "returns error",
			fn: func(_ context.Context, _ io.Writer, _ io.Reader, _ []string) error {
				return errors.New("boom")
			},
			wantErr: errors.New("boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn.Run(context.Background(), &bytes.Buffer{}, strings.NewReader(""), nil)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if tt.wantErr != nil && (err == nil || err.Error() != tt.wantErr.Error()) {
				t.Fatalf("expected error %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	if r.Len() != 0 {
		t.Fatalf("expected empty registry, got %d", r.Len())
	}

	// Register commands
	noop := CommandFunc(func(_ context.Context, _ io.Writer, _ io.Reader, _ []string) error { return nil })
	r.Register("zebra", noop)
	r.Register("alpha", noop)
	r.Register("middle", noop)

	// Len
	if r.Len() != 3 {
		t.Fatalf("expected 3 commands, got %d", r.Len())
	}

	// Get existing
	cmd, ok := r.Get("alpha")
	if !ok || cmd == nil {
		t.Fatal("expected to find 'alpha'")
	}

	// Get missing
	_, ok = r.Get("missing")
	if ok {
		t.Fatal("expected 'missing' to not be found")
	}

	// Names sorted
	names := r.Names()
	expected := []string{"alpha", "middle", "zebra"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(names))
	}
	for i, name := range names {
		if name != expected[i] {
			t.Fatalf("expected names[%d] = %q, got %q", i, expected[i], name)
		}
	}
}

func TestRegistryConcurrency(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	noop := CommandFunc(func(_ context.Context, _ io.Writer, _ io.Reader, _ []string) error { return nil })

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r.Register(fmt.Sprintf("cmd-%d", n), noop)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, _ = r.Get(fmt.Sprintf("cmd-%d", n))
			_ = r.Names()
			_ = r.Len()
		}(i)
	}

	wg.Wait()

	if r.Len() != 100 {
		t.Fatalf("expected 100 commands, got %d", r.Len())
	}
}

func TestAdaptWriterArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantOutput string
	}{
		{name: "no args", args: nil, wantOutput: "got 0 args\n"},
		{name: "with args", args: []string{"a", "b"}, wantOutput: "got 2 args\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := func(w io.Writer, args []string) error {
				_, _ = fmt.Fprintf(w, "got %d args\n", len(args))
				return nil
			}
			cmd := AdaptWriterArgs(fn)

			var buf bytes.Buffer
			err := cmd.Run(context.Background(), &buf, strings.NewReader(""), tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if buf.String() != tt.wantOutput {
				t.Fatalf("expected %q, got %q", tt.wantOutput, buf.String())
			}
		})
	}
}

func TestAdaptWriterReaderArgs(t *testing.T) {
	fn := func(w io.Writer, r io.Reader, args []string) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(w, "read=%s args=%v\n", string(data), args)
		return nil
	}
	cmd := AdaptWriterReaderArgs(fn)

	var buf bytes.Buffer
	err := cmd.Run(context.Background(), &buf, strings.NewReader("hello"), []string{"x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "read=hello args=[x]\n"
	if buf.String() != want {
		t.Fatalf("expected %q, got %q", want, buf.String())
	}
}

func TestAdaptFull(t *testing.T) {
	fn := func(ctx context.Context, w io.Writer, r io.Reader, args []string) error {
		if ctx == nil {
			return errors.New("nil context")
		}
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(w, "ctx=ok read=%s args=%v\n", string(data), args)
		return nil
	}
	cmd := AdaptFull(fn)

	var buf bytes.Buffer
	err := cmd.Run(context.Background(), &buf, strings.NewReader("world"), []string{"y", "z"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "ctx=ok read=world args=[y z]\n"
	if buf.String() != want {
		t.Fatalf("expected %q, got %q", want, buf.String())
	}
}
