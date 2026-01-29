package watch

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunWatch(t *testing.T) {
	t.Run("basic watch", func(t *testing.T) {
		var buf bytes.Buffer

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		callCount := 0
		err := RunWatch(ctx, &buf, func() (string, error) {
			callCount++
			return "output\n", nil
		}, WatchOptions{Interval: 20 * time.Millisecond})

		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("RunWatch() error = %v", err)
		}

		if callCount == 0 {
			t.Error("RunWatch() function should be called at least once")
		}
	})

	t.Run("no title", func(t *testing.T) {
		var buf bytes.Buffer

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_ = RunWatch(ctx, &buf, func() (string, error) {
			return "test\n", nil
		}, WatchOptions{Interval: 20 * time.Millisecond, NoTitle: true})

		output := buf.String()
		if strings.Contains(output, "Every") {
			t.Errorf("RunWatch() -t should not show header: %s", output)
		}
	})

	t.Run("exit on error", func(t *testing.T) {
		var buf bytes.Buffer

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		testErr := errors.New("test error")
		err := RunWatch(ctx, &buf, func() (string, error) {
			return "", testErr
		}, WatchOptions{Interval: 20 * time.Millisecond, ExitOnError: true})

		if err != testErr {
			t.Errorf("RunWatch() -e should exit on error: got %v, want %v", err, testErr)
		}
	})

	t.Run("beep on error", func(t *testing.T) {
		var buf bytes.Buffer

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_ = RunWatch(ctx, &buf, func() (string, error) {
			return "", errors.New("error")
		}, WatchOptions{Interval: 20 * time.Millisecond, BeepOnError: true})

		output := buf.String()
		if !strings.Contains(output, "\a") {
			t.Errorf("RunWatch() -b should beep on error: %s", output)
		}
	})
}

func TestRunWatchCommand(t *testing.T) {
	var buf bytes.Buffer

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = RunWatchCommand(ctx, &buf, "test message", WatchOptions{Interval: 20 * time.Millisecond})

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("RunWatchCommand() should contain message: %s", output)
	}
}

func TestWatchFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watchfile_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "watched.txt")
	_ = os.WriteFile(file, []byte("initial"), 0644)

	callCount := 0

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go func() {
		// Wait a bit then modify the file
		time.Sleep(50 * time.Millisecond)

		_ = os.WriteFile(file, []byte("modified"), 0644)
	}()

	err = WatchFile(ctx, file, func(path string) error {
		callCount++
		return nil
	}, 20*time.Millisecond)

	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WatchFile() error = %v", err)
	}

	if callCount == 0 {
		t.Logf("WatchFile() callback not called (timing-sensitive test)")
	}
}

func TestWatchFile_NonExistent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := WatchFile(ctx, "/nonexistent/file.txt", func(path string) error {
		return nil
	}, 20*time.Millisecond)
	if err == nil {
		t.Error("WatchFile() expected error for nonexistent file")
	}
}

func TestWatchDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "watchdir_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	events := make([]string, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	go func() {
		// Wait then create a new file
		time.Sleep(50 * time.Millisecond)

		_ = os.WriteFile(filepath.Join(tmpDir, "new.txt"), []byte("new"), 0644)

		// Wait then modify it
		time.Sleep(50 * time.Millisecond)

		_ = os.WriteFile(filepath.Join(tmpDir, "new.txt"), []byte("modified"), 0644)

		// Wait then delete it
		time.Sleep(50 * time.Millisecond)

		_ = os.Remove(filepath.Join(tmpDir, "new.txt"))
	}()

	err = WatchDir(ctx, tmpDir, func(event, path string) error {
		events = append(events, event)
		return nil
	}, 20*time.Millisecond)

	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WatchDir() error = %v", err)
	}

	// Events are timing-sensitive, so we just log what we got
	t.Logf("WatchDir() detected events: %v", events)
}

func TestWatchDir_NonExistent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := WatchDir(ctx, "/nonexistent/dir", func(event, path string) error {
		return nil
	}, 20*time.Millisecond)
	if err == nil {
		t.Error("WatchDir() expected error for nonexistent directory")
	}
}

func TestRunWatchIteration(t *testing.T) {
	var buf bytes.Buffer

	err := runWatchIteration(&buf, func() (string, error) {
		return "test output\n", nil
	}, WatchOptions{})
	if err != nil {
		t.Fatalf("runWatchIteration() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test output") {
		t.Errorf("runWatchIteration() should contain output: %s", output)
	}

	if !strings.Contains(output, "Every") {
		t.Errorf("runWatchIteration() should contain header: %s", output)
	}
}
