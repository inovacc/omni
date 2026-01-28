package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoggerNotActive(t *testing.T) {
	log, err := New("")
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if log.IsActive() {
		t.Error("expected logger to be inactive when path is empty")
	}

	// These should not panic
	log.LogCommand("test", []string{"arg1", "arg2"})
	log.LogCommandWithResult("test", []string{"arg1"}, nil)
	log.LogRaw("test message")
}

func TestLoggerActive(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "omni.log")

	log, err := New(logFile)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if !log.IsActive() {
		t.Fatal("expected logger to be active when path is set")
	}

	// Log a command
	log.LogCommand("cat", []string{"file.txt", "-n"})

	// Close the logger
	if err := log.Close(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	// Read and verify log file
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var logEntry map[string]any
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}

	if logEntry["cmd"] != "cat" {
		t.Errorf("expected cmd=cat, got %v", logEntry["cmd"])
	}

	args, ok := logEntry["args"].([]any)
	if !ok {
		t.Fatalf("expected args to be array, got %T", logEntry["args"])
	}

	if len(args) != 2 || args[0] != "file.txt" || args[1] != "-n" {
		t.Errorf("unexpected args: %v", args)
	}
}

func TestLogCommandWithResult(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "omni.log")

	log, err := New(logFile)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Log success
	log.LogCommandWithResult("ls", []string{"-la"}, nil)

	// Log error
	log.LogCommandWithResult("rm", []string{"missing.txt"}, os.ErrNotExist)

	if err := log.Close(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log entries, got %d", len(lines))
	}

	// Check first entry (success)
	var entry1 map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry1); err != nil {
		t.Fatalf("failed to parse first entry: %v", err)
	}

	if entry1["status"] != "success" {
		t.Errorf("expected status=success, got %v", entry1["status"])
	}

	// Check second entry (error)
	var entry2 map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &entry2); err != nil {
		t.Fatalf("failed to parse second entry: %v", err)
	}

	if entry2["status"] != "error" {
		t.Errorf("expected status=error, got %v", entry2["status"])
	}

	if entry2["error"] == nil {
		t.Error("expected error field to be present")
	}
}

func TestFormatArgs(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{}, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b", "c"}, "a b c"},
		{[]string{"-n", "file.txt"}, "-n file.txt"},
	}

	for _, tt := range tests {
		got := FormatArgs(tt.args)
		if got != tt.want {
			t.Errorf("FormatArgs(%v) = %q, want %q", tt.args, got, tt.want)
		}
	}
}

func TestNilLogger(t *testing.T) {
	var log *Logger

	// All methods should be safe to call on nil
	if log.IsActive() {
		t.Error("nil logger should not be active")
	}

	// These should not panic
	log.LogCommand("test", nil)
	log.LogCommandWithResult("test", nil, nil)
	log.LogRaw("test")

	if err := log.Close(); err != nil {
		t.Errorf("Close() on nil logger should return nil, got %v", err)
	}

	w := log.Writer()
	if w == nil {
		t.Error("Writer() should never return nil")
	}
}

func TestNewInvalidPath(t *testing.T) {
	// Try to create logger with invalid path
	_, err := New("/nonexistent/path/to/dir/log.txt")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestLogRaw(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "omni.log")

	log, err := New(logFile)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	log.LogRaw("custom message", "key1", "value1", "key2", 42)

	if err := log.Close(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal(content, &entry); err != nil {
		t.Fatalf("failed to parse entry: %v", err)
	}

	if entry["msg"] != "custom message" {
		t.Errorf("expected msg='custom message', got %v", entry["msg"])
	}

	if entry["key1"] != "value1" {
		t.Errorf("expected key1='value1', got %v", entry["key1"])
	}

	if entry["key2"] != float64(42) {
		t.Errorf("expected key2=42, got %v", entry["key2"])
	}
}

func TestWriter(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "omni.log")

	log, err := New(logFile)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	defer func() { _ = log.Close() }()

	w := log.Writer()
	if w == nil {
		t.Error("Writer() should not return nil for active logger")
	}

	// Write directly to the writer
	_, err = w.Write([]byte("direct write\n"))
	if err != nil {
		t.Errorf("Write() failed: %v", err)
	}
}
