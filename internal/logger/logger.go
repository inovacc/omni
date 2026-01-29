package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/inovacc/omni/internal/flags"
	"github.com/segmentio/ksuid"
)

const (
	// EnvLogEnabled is the environment variable that enables logging.
	// Set to "true" or "1" to enable command logging.
	EnvLogEnabled = "OMNI_LOGGER_ENABLED"
)

var (
	instance *Logger
	once     sync.Once
)

// Logger handles command logging for omni.
type Logger struct {
	slog    *slog.Logger
	file    *os.File
	active  bool
	command string
}

// Init initializes the global logger instance with the command name.
// It checks for OMNI_LOG_ENABLED and OMNI_LOG_PATH environment variables.
// Log files are created as: OMNI_LOG_PATH/ksuid-command.log
// Safe to call multiple times; initialization happens only once.
func Init(command string) *Logger {
	once.Do(func() {
		instance = initLogger(command)
	})

	return instance
}

// initLogger creates a new logger based on environment configuration.
func initLogger(command string) *Logger {
	l := &Logger{
		command: command,
	}

	enabled := flags.IsFeatureEnabled("logger")
	if !enabled {
		return l
	}

	logDir := flags.GetFeatureData("logger")
	if logDir == "" {
		_, _ = os.Stderr.WriteString("omni: OMNI_LOGGER_ENABLED set but empty logger path\n")
		return l
	}

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		_, _ = os.Stderr.WriteString("omni: failed to create log directory: " + err.Error() + "\n")
		return l
	}

	// Generate unique log file path: dir/ksuid-command.log
	logPath, err := generateLogPath(logDir, command)
	if err != nil {
		_, _ = os.Stderr.WriteString("omni: failed to generate log path: " + err.Error() + "\n")
		return l
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		_, _ = os.Stderr.WriteString("omni: failed to open log file: " + err.Error() + "\n")
		return l
	}

	l.file = file
	l.slog = slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	l.active = true

	return l
}

// generateLogPath creates a unique log file path using ksuid and command name.
// Format: dir/ksuid-command.log
func generateLogPath(logDir, command string) (string, error) {
	id, err := ksuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate ksuid: %w", err)
	}

	if command == "" {
		command = "omni"
	}

	filename := fmt.Sprintf("%s-%s.log", id.String(), command)

	return filepath.Join(logDir, filename), nil
}

// New creates a new logger that writes to a unique file in the specified directory.
// Uses ksuid to generate unique file names to avoid race conditions between instances.
// Format: logDir/ksuid-command.log
// This is useful for testing or when you need a logger independent of the global instance.
func New(logDir, command string) (*Logger, error) {
	l := &Logger{
		command: command,
	}

	if logDir == "" {
		return l, nil
	}

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath, err := generateLogPath(logDir, command)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	l.file = file
	l.slog = slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	l.active = true

	return l, nil
}

// NewWithExactPath creates a logger that writes to the exact path specified.
// Unlike New, this does not generate a unique path - use when you need a specific file.
func NewWithExactPath(logPath string) (*Logger, error) {
	l := &Logger{}

	if logPath == "" {
		return l, nil
	}

	// Ensure log directory exists
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	l.file = file
	l.slog = slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	l.active = true

	return l, nil
}

// Get returns the global logger instance.
// Returns nil if Init has not been called.
func Get() *Logger {
	return instance
}

// IsActive returns true if logging is enabled.
func (l *Logger) IsActive() bool {
	if l == nil {
		return false
	}

	return l.active
}

// LogCommand logs a command execution with its arguments.
func (l *Logger) LogCommand(args []string) {
	if l == nil || !l.active {
		return
	}

	l.slog.Info("command", "cmd", l.command, "args", args, "timestamp", time.Now().Format(time.RFC3339), "pid", os.Getpid())
}

// LogCommandWithResult logs a command execution including its result.
func (l *Logger) LogCommandWithResult(command string, args []string, err error) {
	if l == nil || !l.active {
		return
	}

	status := "success"

	var errMsg string

	if err != nil {
		status = "error"
		errMsg = err.Error()
	}

	attrs := []any{
		"cmd", command,
		"args", args,
		"status", status,
		"timestamp", time.Now().Format(time.RFC3339),
		"pid", os.Getpid(),
	}

	if errMsg != "" {
		attrs = append(attrs, "error", errMsg)
	}

	l.slog.Info("command", attrs...)
}

// LogRaw logs a raw message with optional key-value pairs.
func (l *Logger) LogRaw(msg string, args ...any) {
	if l == nil || !l.active {
		return
	}

	l.slog.Info(msg, args...)
}

// LogQuery logs a database query before execution.
func (l *Logger) LogQuery(database, query string) {
	if l == nil || !l.active {
		return
	}

	l.slog.Info("query",
		"database", database,
		"query", query,
		"timestamp", time.Now().Format(time.RFC3339),
		"pid", os.Getpid(),
	)
}

// LogQueryResult logs a database query with its result.
func (l *Logger) LogQueryResult(database, query string, rowCount int, duration time.Duration, err error) {
	if l == nil || !l.active {
		return
	}

	status := "success"

	var errMsg string

	if err != nil {
		status = "error"
		errMsg = err.Error()
	}

	attrs := []any{
		"database", database,
		"query", query,
		"status", status,
		"rows", rowCount,
		"duration_ms", duration.Milliseconds(),
		"timestamp", time.Now().Format(time.RFC3339),
		"pid", os.Getpid(),
	}

	if errMsg != "" {
		attrs = append(attrs, "error", errMsg)
	}

	l.slog.Info("query_result", attrs...)
}

// LogQueryWithData logs a query along with its result data (for small result sets).
// Use this sparingly as it can produce large log entries.
func (l *Logger) LogQueryWithData(database, query string, columns []string, rows []map[string]any, duration time.Duration, err error) {
	if l == nil || !l.active {
		return
	}

	status := "success"

	var errMsg string

	if err != nil {
		status = "error"
		errMsg = err.Error()
	}

	attrs := []any{
		"database", database,
		"query", query,
		"status", status,
		"row_count", len(rows),
		"columns", columns,
		"duration_ms", duration.Milliseconds(),
		"timestamp", time.Now().Format(time.RFC3339),
		"pid", os.Getpid(),
	}

	if errMsg != "" {
		attrs = append(attrs, "error", errMsg)
	}

	// Only include data if there are rows and logging data is desired
	if len(rows) > 0 && len(rows) <= 100 { // Limit to avoid huge logs
		attrs = append(attrs, "data", rows)
	}

	l.slog.Info("query_result", attrs...)
}

// QueryLogger provides a convenient wrapper for logging database queries.
type QueryLogger struct {
	logger   *Logger
	database string
}

// NewQueryLogger creates a query logger for a specific database.
func NewQueryLogger(l *Logger, database string) *QueryLogger {
	return &QueryLogger{
		logger:   l,
		database: database,
	}
}

// Log logs a query with its result.
func (ql *QueryLogger) Log(query string, rowCount int, duration time.Duration, err error) {
	if ql == nil || ql.logger == nil {
		return
	}

	ql.logger.LogQueryResult(ql.database, query, rowCount, duration, err)
}

// LogWithData logs a query with its result data.
func (ql *QueryLogger) LogWithData(query string, columns []string, rows []map[string]any, duration time.Duration, err error) {
	if ql == nil || ql.logger == nil {
		return
	}

	ql.logger.LogQueryWithData(ql.database, query, columns, rows, duration, err)
}

// Close closes the log file if open.
func (l *Logger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}

	return l.file.Close()
}

// Writer returns the underlying io.Writer for the logger.
// Returns io.Discard if logging is not active.
func (l *Logger) Writer() io.Writer {
	if l == nil || l.file == nil {
		return io.Discard
	}

	return l.file
}

// FormatArgs formats command arguments as a single string.
func FormatArgs(args []string) string {
	return strings.Join(args, " ")
}
