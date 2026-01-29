package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoggerNotActive(t *testing.T) {
	log, err := New("", "")
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if log.IsActive() {
		t.Error("expected logger to be inactive when path is empty")
	}

	// These should not panic
	log.LogCommand([]string{"arg1", "arg2"})
	log.LogCommandWithResult("test", []string{"arg1"}, nil)
	log.LogRaw("test message")
}

func TestLoggerActive(t *testing.T) {
	tmpDir := t.TempDir()

	log, err := New(tmpDir, "cat")
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if !log.IsActive() {
		t.Fatal("expected logger to be active when path is set")
	}

	// Log a command
	log.LogCommand([]string{"file.txt", "-n"})

	// Close the logger
	if err := log.Close(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	// Find the log file (ksuid-cat.log)
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	var logFile string

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), "-cat.log") {
			logFile = filepath.Join(tmpDir, entry.Name())
			break
		}
	}

	if logFile == "" {
		t.Fatal("log file not found")
	}

	// Read and verify the log file
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
	logFile := filepath.Join(tmpDir, "test.log")

	log, err := NewWithExactPath(logFile)
	if err != nil {
		t.Fatalf("NewWithExactPath() failed: %v", err)
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

	// Check the first entry (success)
	var entry1 map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry1); err != nil {
		t.Fatalf("failed to parse first entry: %v", err)
	}

	if entry1["status"] != "success" {
		t.Errorf("expected status=success, got %v", entry1["status"])
	}

	// Check the second entry (error)
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
	log.LogCommand(nil)
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
	// Try to create a logger with an invalid path (path that can't be created)
	// On Windows, NUL is reserved; on Unix, /dev/null/subdir is invalid
	_, err := New("/dev/null/invalid/path", "test")
	if err == nil {
		// Some systems might succeed, so this test is best-effort
		t.Log("expected error for invalid path, but none received")
	}
}

func TestLogRaw(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	log, err := NewWithExactPath(logFile)
	if err != nil {
		t.Fatalf("NewWithExactPath() failed: %v", err)
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
	logFile := filepath.Join(tmpDir, "test.log")

	log, err := NewWithExactPath(logFile)
	if err != nil {
		t.Fatalf("NewWithExactPath() failed: %v", err)
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

func TestGenerateLogPath(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		command string
		wantSfx string
	}{
		{"normal command", filepath.Join("var", "log"), "cat", "-cat.log"},
		{"empty command", filepath.Join("var", "log"), "", "-omni.log"},
		{"complex command", "tmp", "my-cmd", "-my-cmd.log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := generateLogPath(tt.dir, tt.command)
			if err != nil {
				t.Fatalf("generateLogPath() failed: %v", err)
			}

			// Use filepath.Clean to normalize paths for comparison
			expectedPrefix := filepath.Clean(tt.dir)
			if !strings.HasPrefix(filepath.Clean(path), expectedPrefix) {
				t.Errorf("path should start with %s, got %s", expectedPrefix, path)
			}

			if !strings.HasSuffix(path, tt.wantSfx) {
				t.Errorf("path should end with %s, got %s", tt.wantSfx, path)
			}
		})
	}
}

func TestLogQuery(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	log, err := NewWithExactPath(logFile)
	if err != nil {
		t.Fatalf("NewWithExactPath() failed: %v", err)
	}

	log.LogQuery("/path/to/db.sqlite", "SELECT * FROM users")

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

	if entry["msg"] != "query" {
		t.Errorf("expected msg='query', got %v", entry["msg"])
	}

	if entry["database"] != "/path/to/db.sqlite" {
		t.Errorf("expected database='/path/to/db.sqlite', got %v", entry["database"])
	}

	if entry["query"] != "SELECT * FROM users" {
		t.Errorf("expected query='SELECT * FROM users', got %v", entry["query"])
	}
}

func TestLogQueryResult(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	log, err := NewWithExactPath(logFile)
	if err != nil {
		t.Fatalf("NewWithExactPath() failed: %v", err)
	}

	// Log success
	log.LogQueryResult("/path/to/db.sqlite", "SELECT * FROM users", 10, 50*time.Millisecond, nil)

	// Log error
	log.LogQueryResult("/path/to/db.sqlite", "SELECT * FROM invalid", 0, 5*time.Millisecond, os.ErrNotExist)

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

	// Check success entry
	var entry1 map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry1); err != nil {
		t.Fatalf("failed to parse first entry: %v", err)
	}

	if entry1["msg"] != "query_result" {
		t.Errorf("expected msg='query_result', got %v", entry1["msg"])
	}

	if entry1["status"] != "success" {
		t.Errorf("expected status='success', got %v", entry1["status"])
	}

	if entry1["rows"] != float64(10) {
		t.Errorf("expected rows=10, got %v", entry1["rows"])
	}

	// Check error entry
	var entry2 map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &entry2); err != nil {
		t.Fatalf("failed to parse second entry: %v", err)
	}

	if entry2["status"] != "error" {
		t.Errorf("expected status='error', got %v", entry2["status"])
	}

	if entry2["error"] == nil {
		t.Error("expected error field to be present")
	}
}

func TestLogQueryWithData(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	log, err := NewWithExactPath(logFile)
	if err != nil {
		t.Fatalf("NewWithExactPath() failed: %v", err)
	}

	columns := []string{"id", "name", "email"}
	rows := []map[string]any{
		{"id": 1, "name": "Alice", "email": "alice@example.com"},
		{"id": 2, "name": "Bob", "email": "bob@example.com"},
	}

	log.LogQueryWithData("/path/to/db.sqlite", "SELECT * FROM users", columns, rows, 25*time.Millisecond, nil)

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

	if entry["msg"] != "query_result" {
		t.Errorf("expected msg='query_result', got %v", entry["msg"])
	}

	if entry["row_count"] != float64(2) {
		t.Errorf("expected row_count=2, got %v", entry["row_count"])
	}

	cols, ok := entry["columns"].([]any)
	if !ok {
		t.Fatalf("expected columns to be array, got %T", entry["columns"])
	}

	if len(cols) != 3 {
		t.Errorf("expected 3 columns, got %d", len(cols))
	}

	data, ok := entry["data"].([]any)
	if !ok {
		t.Fatalf("expected data to be array, got %T", entry["data"])
	}

	if len(data) != 2 {
		t.Errorf("expected 2 data rows, got %d", len(data))
	}
}

func TestQueryLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	log, err := NewWithExactPath(logFile)
	if err != nil {
		t.Fatalf("NewWithExactPath() failed: %v", err)
	}

	ql := NewQueryLogger(log, "/path/to/db.sqlite")

	ql.Log("SELECT 1", 1, time.Millisecond, nil)

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

	if entry["database"] != "/path/to/db.sqlite" {
		t.Errorf("expected database='/path/to/db.sqlite', got %v", entry["database"])
	}
}

func TestNilQueryLogger(t *testing.T) {
	// Test nil QueryLogger
	var ql *QueryLogger

	// Should not panic
	ql.Log("SELECT 1", 1, time.Millisecond, nil)
	ql.LogWithData("SELECT 1", nil, nil, time.Millisecond, nil)

	// Test QueryLogger with nil Logger
	ql = NewQueryLogger(nil, "test.db")
	ql.Log("SELECT 1", 1, time.Millisecond, nil)
	ql.LogWithData("SELECT 1", nil, nil, time.Millisecond, nil)
}

func TestNilLoggerQueryMethods(t *testing.T) {
	var log *Logger

	// All query methods should be safe to call on nil
	log.LogQuery("test.db", "SELECT 1")
	log.LogQueryResult("test.db", "SELECT 1", 0, time.Millisecond, nil)
	log.LogQueryWithData("test.db", "SELECT 1", nil, nil, time.Millisecond, nil)
}
