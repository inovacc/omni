// Package cmderr provides a unified error model for CLI commands.
// Commands return errors instead of calling os.Exit directly.
// The root command maps errors to exit codes.
package cmderr

import (
	"errors"
	"fmt"
)

// Sentinel errors for common failure categories.
var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrPermission   = errors.New("permission denied")
	ErrIO           = errors.New("I/O error")
	ErrConflict     = errors.New("conflict")
	ErrTimeout      = errors.New("timeout")
	ErrUnsupported  = errors.New("unsupported")
)

// SilentError is an error that should set an exit code without printing a message.
// Cobra's SilenceErrors must be checked, or the root command must handle this type.
type SilentError struct {
	Code int
}

func (e *SilentError) Error() string {
	return ""
}

// SilentExit returns a SilentError with the given exit code.
// Use for commands like grep that exit non-zero on "no match" without printing errors.
func SilentExit(code int) error {
	return &SilentError{Code: code}
}

// ExitError wraps an error with a specific exit code.
type ExitError struct {
	Err  error
	Code int
}

func (e *ExitError) Error() string {
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

// WithExitCode wraps an error with a specific exit code.
func WithExitCode(err error, code int) error {
	if err == nil {
		return nil
	}

	return &ExitError{Err: err, Code: code}
}

// Wrap wraps a sentinel error with context.
func Wrap(sentinel error, msg string) error {
	return fmt.Errorf("%s: %w", msg, sentinel)
}

// IsNotFound reports whether err is, or wraps, ErrNotFound.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

// IsInvalidInput reports whether err is, or wraps, ErrInvalidInput.
func IsInvalidInput(err error) bool { return errors.Is(err, ErrInvalidInput) }

// IsPermission reports whether err is, or wraps, ErrPermission.
func IsPermission(err error) bool { return errors.Is(err, ErrPermission) }

// IsIO reports whether err is, or wraps, ErrIO.
func IsIO(err error) bool { return errors.Is(err, ErrIO) }

// IsConflict reports whether err is, or wraps, ErrConflict.
func IsConflict(err error) bool { return errors.Is(err, ErrConflict) }

// IsTimeout reports whether err is, or wraps, ErrTimeout.
func IsTimeout(err error) bool { return errors.Is(err, ErrTimeout) }

// IsUnsupported reports whether err is, or wraps, ErrUnsupported.
func IsUnsupported(err error) bool { return errors.Is(err, ErrUnsupported) }

// ExitCodeFor maps an error to an exit code.
// Returns 0 for nil errors.
func ExitCodeFor(err error) int {
	if err == nil {
		return 0
	}

	// Check for silent exit first.
	var silentErr *SilentError
	if errors.As(err, &silentErr) {
		return silentErr.Code
	}

	// Check for explicit exit code.
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code
	}

	// Map sentinel errors to exit codes.
	switch {
	case errors.Is(err, ErrNotFound):
		return 1
	case errors.Is(err, ErrConflict):
		return 1
	case errors.Is(err, ErrInvalidInput):
		return 2
	case errors.Is(err, ErrPermission):
		return 3
	case errors.Is(err, ErrIO):
		return 4
	case errors.Is(err, ErrTimeout):
		return 5
	case errors.Is(err, ErrUnsupported):
		return 6
	default:
		return 1
	}
}
