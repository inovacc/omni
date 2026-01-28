package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// EnvLogEnabled is the environment variable that enables logging.
	// Set to "true" or "1" to enable command logging.
	EnvLogEnabled = "OMNI_LOG_ENABLED"

	// EnvLogPath is the environment variable that specifies the log file path.
	// When OMNI_LOG_ENABLED is set, commands will be logged to this file.
	EnvLogPath = "OMNI_LOG_PATH"
)

var (
	instance *Logger
	once     sync.Once
)

// Logger handles command logging for omni.
type Logger struct {
	slog   *slog.Logger
	file   *os.File
	active bool
}

// Init initializes the global logger instance.
// It checks for OMNI_LOG and OMNI_LOG_PATH environment variables.
// Safe to call multiple times; initialization happens only once.
func Init() *Logger {
	once.Do(func() {
		instance = initLogger()
	})

	return instance
}

// initLogger creates a new logger based on environment configuration.
func initLogger() *Logger {
	l := &Logger{}

	enabled := os.Getenv(EnvLogEnabled)
	if !isEnabled(enabled) {
		return l
	}

	logPath := os.Getenv(EnvLogPath)
	if logPath == "" {
		_, _ = os.Stderr.WriteString("omni: OMNI_LOG_ENABLED set but OMNI_LOG_PATH not set\n")
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

// isEnabled checks if the value represents an enabled state.
func isEnabled(val string) bool {
	v := strings.ToLower(strings.TrimSpace(val))
	return v == "true" || v == "1" || v == "yes"
}

// New creates a new logger that writes to the specified file path.
// This is useful for testing or when you need a logger independent of the global instance.
func New(logPath string) (*Logger, error) {
	l := &Logger{}

	if logPath == "" {
		return l, nil
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
func (l *Logger) LogCommand(command string, args []string) {
	if l == nil || !l.active {
		return
	}

	l.slog.Info("command",
		"cmd", command,
		"args", args,
		"timestamp", time.Now().Format(time.RFC3339),
		"pid", os.Getpid(),
	)
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
