package xargs

import (
	"bytes"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
)

func TestRunXargs(t *testing.T) {
	t.Run("basic execution", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("a b c")

		var received []string

		worker := func(args []string) error {
			received = args
			return nil
		}

		err := RunXargs(&buf, input, nil, XargsOptions{}, worker)
		if err != nil {
			t.Fatalf("RunXargs() error = %v", err)
		}

		if len(received) != 3 {
			t.Errorf("RunXargs() received %d args, want 3", len(received))
		}
	})

	t.Run("with initial args", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("file1 file2")

		var received []string

		worker := func(args []string) error {
			received = args
			return nil
		}

		err := RunXargs(&buf, input, []string{"echo"}, XargsOptions{}, worker)
		if err != nil {
			t.Fatalf("RunXargs() error = %v", err)
		}

		if len(received) != 3 || received[0] != "echo" {
			t.Errorf("RunXargs() received = %v, want [echo file1 file2]", received)
		}
	})

	t.Run("max args batching", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("a b c d e")

		callCount := 0
		worker := func(args []string) error {
			callCount++

			if len(args) > 2 {
				t.Errorf("RunXargs() batch has %d args, want <= 2", len(args))
			}

			return nil
		}

		err := RunXargs(&buf, input, nil, XargsOptions{MaxArgs: 2}, worker)
		if err != nil {
			t.Fatalf("RunXargs() error = %v", err)
		}

		if callCount != 3 {
			t.Errorf("RunXargs() called %d times, want 3", callCount)
		}
	})

	t.Run("null input delimiter", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("file1\x00file2\x00file3")

		var received []string

		worker := func(args []string) error {
			received = args
			return nil
		}

		err := RunXargs(&buf, input, nil, XargsOptions{NullInput: true}, worker)
		if err != nil {
			t.Fatalf("RunXargs() error = %v", err)
		}

		if len(received) != 3 {
			t.Errorf("RunXargs() null input got %d args, want 3", len(received))
		}
	})

	t.Run("custom delimiter", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("a,b,c")

		var received []string

		worker := func(args []string) error {
			received = args
			return nil
		}

		err := RunXargs(&buf, input, nil, XargsOptions{Delimiter: ","}, worker)
		if err != nil {
			t.Fatalf("RunXargs() error = %v", err)
		}

		if len(received) != 3 {
			t.Errorf("RunXargs() custom delim got %d args, want 3", len(received))
		}
	})

	t.Run("replace string", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("file1 file2")

		var receivedBatches [][]string

		worker := func(args []string) error {
			receivedBatches = append(receivedBatches, args)
			return nil
		}

		err := RunXargs(&buf, input, []string{"echo", "{}"}, XargsOptions{ReplaceStr: "{}", MaxArgs: 1}, worker)
		if err != nil {
			t.Fatalf("RunXargs() error = %v", err)
		}

		// Should replace {} with each input
		if len(receivedBatches) < 2 {
			t.Errorf("RunXargs() replace got %d batches", len(receivedBatches))
		}
	})

	t.Run("no run empty", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("")

		called := false
		worker := func(args []string) error {
			called = true
			return nil
		}

		err := RunXargs(&buf, input, nil, XargsOptions{NoRunEmpty: true}, worker)
		if err != nil {
			t.Fatalf("RunXargs() error = %v", err)
		}

		if called {
			t.Error("RunXargs() should not call worker with empty input when NoRunEmpty is set")
		}
	})

	t.Run("worker error", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("a b c")

		expectedErr := errors.New("worker error")
		worker := func(args []string) error {
			return expectedErr
		}

		err := RunXargs(&buf, input, nil, XargsOptions{}, worker)
		if err != expectedErr {
			t.Errorf("RunXargs() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("nil worker prints args", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("a b c")

		err := RunXargs(&buf, input, nil, XargsOptions{}, nil)
		if err != nil {
			t.Fatalf("RunXargs() error = %v", err)
		}

		if !strings.Contains(buf.String(), "a b c") {
			t.Errorf("RunXargs() output = %q, want to contain 'a b c'", buf.String())
		}
	})

	t.Run("parallel execution", func(t *testing.T) {
		var buf bytes.Buffer

		input := strings.NewReader("a b c d")

		var callCount atomic.Int32
		worker := func(args []string) error {
			callCount.Add(1)
			return nil
		}

		err := RunXargs(&buf, input, nil, XargsOptions{MaxArgs: 1, MaxProcs: 2}, worker)
		if err != nil {
			t.Fatalf("RunXargs() error = %v", err)
		}

		if callCount.Load() != 4 {
			t.Errorf("RunXargs() parallel called %d times, want 4", callCount.Load())
		}
	})
}

func TestParseXargsInput(t *testing.T) {
	t.Run("whitespace separated", func(t *testing.T) {
		input := strings.NewReader("  a   b   c  ")

		args, err := parseXargsInput(input, XargsOptions{})
		if err != nil {
			t.Fatalf("parseXargsInput() error = %v", err)
		}

		if len(args) != 3 {
			t.Errorf("parseXargsInput() = %v, want 3 args", args)
		}
	})

	t.Run("newlines as whitespace", func(t *testing.T) {
		input := strings.NewReader("a\nb\nc")

		args, err := parseXargsInput(input, XargsOptions{})
		if err != nil {
			t.Fatalf("parseXargsInput() error = %v", err)
		}

		if len(args) != 3 {
			t.Errorf("parseXargsInput() = %v, want 3 args", args)
		}
	})

	t.Run("null terminated", func(t *testing.T) {
		input := strings.NewReader("path with spaces\x00another path\x00")

		args, err := parseXargsInput(input, XargsOptions{NullInput: true})
		if err != nil {
			t.Fatalf("parseXargsInput() error = %v", err)
		}

		if len(args) != 2 {
			t.Errorf("parseXargsInput() = %v, want 2 args", args)
		}

		if args[0] != "path with spaces" {
			t.Errorf("parseXargsInput() first arg = %q", args[0])
		}
	})

	t.Run("custom delimiter", func(t *testing.T) {
		input := strings.NewReader("a:b:c")

		args, err := parseXargsInput(input, XargsOptions{Delimiter: ":"})
		if err != nil {
			t.Fatalf("parseXargsInput() error = %v", err)
		}

		if len(args) != 3 {
			t.Errorf("parseXargsInput() = %v, want 3 args", args)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		input := strings.NewReader("")

		args, err := parseXargsInput(input, XargsOptions{})
		if err != nil {
			t.Fatalf("parseXargsInput() error = %v", err)
		}

		if len(args) != 0 {
			t.Errorf("parseXargsInput() = %v, want empty", args)
		}
	})
}

func TestRunXargsWithPrint(t *testing.T) {
	var buf bytes.Buffer

	input := strings.NewReader("hello world")

	err := RunXargsWithPrint(&buf, input, XargsOptions{})
	if err != nil {
		t.Fatalf("RunXargsWithPrint() error = %v", err)
	}

	if !strings.Contains(buf.String(), "hello world") {
		t.Errorf("RunXargsWithPrint() = %q", buf.String())
	}
}
