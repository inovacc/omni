package runtimeps

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// TestRunTop_NoTTYIsUnsupported pins the environment-refusal classification:
// when stdout is not a TTY (as under `go test`), RunTop must refuse with
// cmderr.ErrUnsupported (exit 6) rather than a raw, unclassified error.
func TestRunTop_NoTTYIsUnsupported(t *testing.T) {
	if isTTY() {
		t.Skip("stdout is a TTY; RunTop would launch the interactive program")
	}

	err := RunTop(context.Background(), time.Second, false)
	if err == nil {
		t.Fatal("expected error when no TTY is present")
	}
	if !errors.Is(err, cmderr.ErrUnsupported) {
		t.Errorf("no-TTY refusal should be ErrUnsupported (exit 6): %v", err)
	}
}
