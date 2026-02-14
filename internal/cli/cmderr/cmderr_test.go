package cmderr

import (
	"errors"
	"fmt"
	"testing"
)

func TestExitCodeFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"not found", ErrNotFound, 1},
		{"conflict", ErrConflict, 1},
		{"invalid input", ErrInvalidInput, 2},
		{"permission", ErrPermission, 3},
		{"io", ErrIO, 4},
		{"timeout", ErrTimeout, 5},
		{"unsupported", ErrUnsupported, 6},
		{"unknown", errors.New("unknown"), 1},
		{"wrapped not found", fmt.Errorf("file: %w", ErrNotFound), 1},
		{"explicit code", WithExitCode(errors.New("custom"), 42), 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExitCodeFor(tt.err); got != tt.want {
				t.Errorf("ExitCodeFor() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestExitError_Unwrap(t *testing.T) {
	inner := ErrNotFound
	err := WithExitCode(inner, 99)

	if !errors.Is(err, ErrNotFound) {
		t.Error("WithExitCode should preserve error chain")
	}
}

func TestExitError_Error(t *testing.T) {
	err := WithExitCode(errors.New("boom"), 1)
	if err.Error() != "boom" {
		t.Errorf("Error() = %q, want %q", err.Error(), "boom")
	}
}

func TestWithExitCode_Nil(t *testing.T) {
	if WithExitCode(nil, 1) != nil {
		t.Error("WithExitCode(nil) should return nil")
	}
}

func TestWrap(t *testing.T) {
	err := Wrap(ErrNotFound, "file go.mod")
	if !errors.Is(err, ErrNotFound) {
		t.Error("Wrap should preserve sentinel")
	}

	if err.Error() != "file go.mod: not found" {
		t.Errorf("Error() = %q", err.Error())
	}
}
