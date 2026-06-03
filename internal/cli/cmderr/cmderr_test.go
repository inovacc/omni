package cmderr_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

func TestExitCodeFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"raw ErrNotFound", cmderr.ErrNotFound, 1},
		{"raw ErrConflict", cmderr.ErrConflict, 1},
		{"raw ErrInvalidInput", cmderr.ErrInvalidInput, 2},
		{"raw ErrPermission", cmderr.ErrPermission, 3},
		{"raw ErrIO", cmderr.ErrIO, 4},
		{"raw ErrTimeout", cmderr.ErrTimeout, 5},
		{"raw ErrUnsupported", cmderr.ErrUnsupported, 6},
		{"wrapped ErrNotFound", cmderr.Wrap(cmderr.ErrNotFound, "missing"), 1},
		{"wrapped ErrConflict", cmderr.Wrap(cmderr.ErrConflict, "conflict"), 1},
		{"wrapped ErrInvalidInput", cmderr.Wrap(cmderr.ErrInvalidInput, "bad flag"), 2},
		{"wrapped ErrPermission", cmderr.Wrap(cmderr.ErrPermission, "denied"), 3},
		{"wrapped ErrIO", cmderr.Wrap(cmderr.ErrIO, "io"), 4},
		{"wrapped ErrTimeout", cmderr.Wrap(cmderr.ErrTimeout, "timeout"), 5},
		{"wrapped ErrUnsupported", cmderr.Wrap(cmderr.ErrUnsupported, "nope"), 6},
		{"fmt-wrapped not found", fmt.Errorf("file: %w", cmderr.ErrNotFound), 1},
		{"unclassified", errors.New("something went wrong"), 1},
		{"explicit code via WithExitCode", cmderr.WithExitCode(errors.New("custom"), 42), 42},
		{"silent exit", cmderr.SilentExit(7), 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cmderr.ExitCodeFor(tt.err); got != tt.want {
				t.Errorf("ExitCodeFor(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	sentinels := []struct {
		name     string
		sentinel error
		suffix   string
	}{
		{"ErrNotFound", cmderr.ErrNotFound, "not found"},
		{"ErrInvalidInput", cmderr.ErrInvalidInput, "invalid input"},
		{"ErrPermission", cmderr.ErrPermission, "permission denied"},
		{"ErrIO", cmderr.ErrIO, "I/O error"},
		{"ErrConflict", cmderr.ErrConflict, "conflict"},
		{"ErrTimeout", cmderr.ErrTimeout, "timeout"},
		{"ErrUnsupported", cmderr.ErrUnsupported, "unsupported"},
	}

	for _, tt := range sentinels {
		t.Run(tt.name, func(t *testing.T) {
			const msg = "context message"
			err := cmderr.Wrap(tt.sentinel, msg)
			if err == nil {
				t.Fatal("Wrap returned nil")
			}
			if !errors.Is(err, tt.sentinel) {
				t.Errorf("errors.Is(err, %v) = false, want true", tt.sentinel)
			}
			if !strings.HasPrefix(err.Error(), msg) {
				t.Errorf("err.Error() = %q, want prefix %q", err.Error(), msg)
			}
			if !strings.HasSuffix(err.Error(), tt.suffix) {
				t.Errorf("err.Error() = %q, want suffix %q", err.Error(), tt.suffix)
			}
		})
	}
}

func TestSilentExit(t *testing.T) {
	err := cmderr.SilentExit(7)
	if err == nil {
		t.Fatal("SilentExit returned nil")
	}

	var silentErr *cmderr.SilentError
	if !errors.As(err, &silentErr) {
		t.Fatalf("errors.As(err, *SilentError) = false, want true")
	}
	if silentErr.Code != 7 {
		t.Errorf("SilentError.Code = %d, want 7", silentErr.Code)
	}
	if err.Error() != "" {
		t.Errorf("SilentError.Error() = %q, want empty", err.Error())
	}
	if got := cmderr.ExitCodeFor(err); got != 7 {
		t.Errorf("ExitCodeFor(SilentExit(7)) = %d, want 7", got)
	}
}

func TestSilentErrorDirect(t *testing.T) {
	var err error = &cmderr.SilentError{Code: 3}
	if err.Error() != "" {
		t.Errorf("(&SilentError{}).Error() = %q, want empty", err.Error())
	}
	if got := cmderr.ExitCodeFor(err); got != 3 {
		t.Errorf("ExitCodeFor(&SilentError{Code: 3}) = %d, want 3", got)
	}
}

func TestWithExitCode(t *testing.T) {
	t.Run("nil passes through", func(t *testing.T) {
		if got := cmderr.WithExitCode(nil, 42); got != nil {
			t.Errorf("WithExitCode(nil, 42) = %v, want nil", got)
		}
	})

	t.Run("preserves sentinel chain", func(t *testing.T) {
		inner := cmderr.Wrap(cmderr.ErrNotFound, "missing file")
		wrapped := cmderr.WithExitCode(inner, 42)

		if !errors.Is(wrapped, cmderr.ErrNotFound) {
			t.Error("errors.Is(wrapped, ErrNotFound) = false, want true")
		}

		var exitErr *cmderr.ExitError
		if !errors.As(wrapped, &exitErr) {
			t.Fatal("errors.As(wrapped, *ExitError) = false, want true")
		}
		if exitErr.Code != 42 {
			t.Errorf("ExitError.Code = %d, want 42", exitErr.Code)
		}

		// ExitCodeFor uses the explicit code, not the sentinel mapping.
		if got := cmderr.ExitCodeFor(wrapped); got != 42 {
			t.Errorf("ExitCodeFor = %d, want 42", got)
		}

		if unwrapped := errors.Unwrap(wrapped); unwrapped == nil {
			t.Error("errors.Unwrap(wrapped) = nil, want inner error")
		}
		if wrapped.Error() != inner.Error() {
			t.Errorf("wrapped.Error() = %q, want %q", wrapped.Error(), inner.Error())
		}
	})

	t.Run("wraps plain error", func(t *testing.T) {
		err := cmderr.WithExitCode(errors.New("boom"), 1)
		if err.Error() != "boom" {
			t.Errorf("Error() = %q, want %q", err.Error(), "boom")
		}
	})
}

// TestWrapDoubleWrap guards Pitfall 3: when the outer Wrap's message
// happens to contain another sentinel's textual rendering, the top-level
// sentinel must still be the outer one. Documents current behavior so a
// future refactor doesn't silently change classification semantics.
func TestWrapDoubleWrap(t *testing.T) {
	inner := cmderr.Wrap(cmderr.ErrNotFound, "x")
	outer := cmderr.Wrap(cmderr.ErrIO, inner.Error())

	if !errors.Is(outer, cmderr.ErrIO) {
		t.Error("errors.Is(outer, ErrIO) = false, want true")
	}
	if got := cmderr.ExitCodeFor(outer); got != 4 {
		t.Errorf("ExitCodeFor(double-wrapped) = %d, want 4 (ErrIO)", got)
	}
	// errors.Is should NOT find ErrNotFound — we wrapped the string, not the error.
	if errors.Is(outer, cmderr.ErrNotFound) {
		t.Error("errors.Is(outer, ErrNotFound) = true, want false (string-wrap must not leak sentinel)")
	}
}

func TestIsClassHelpers(t *testing.T) {
	cases := []struct {
		name     string
		fn       func(error) bool
		sentinel error
	}{
		{"IsNotFound", cmderr.IsNotFound, cmderr.ErrNotFound},
		{"IsInvalidInput", cmderr.IsInvalidInput, cmderr.ErrInvalidInput},
		{"IsPermission", cmderr.IsPermission, cmderr.ErrPermission},
		{"IsIO", cmderr.IsIO, cmderr.ErrIO},
		{"IsConflict", cmderr.IsConflict, cmderr.ErrConflict},
		{"IsTimeout", cmderr.IsTimeout, cmderr.ErrTimeout},
		{"IsUnsupported", cmderr.IsUnsupported, cmderr.ErrUnsupported},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.fn(tc.sentinel) {
				t.Errorf("%s(sentinel) = false, want true", tc.name)
			}
			if !tc.fn(cmderr.Wrap(tc.sentinel, "context")) {
				t.Errorf("%s(wrapped) = false, want true", tc.name)
			}
			if tc.fn(errors.New("unrelated")) {
				t.Errorf("%s(unrelated) = true, want false", tc.name)
			}
			if tc.fn(nil) {
				t.Errorf("%s(nil) = true, want false", tc.name)
			}
		})
	}
}

func TestExitErrorUnwrap(t *testing.T) {
	base := fmt.Errorf("base error")
	wrapped := &cmderr.ExitError{Err: base, Code: 99}
	if errors.Unwrap(wrapped) != base {
		t.Error("ExitError.Unwrap did not return inner error")
	}
	if wrapped.Error() != "base error" {
		t.Errorf("ExitError.Error() = %q, want %q", wrapped.Error(), "base error")
	}
}
