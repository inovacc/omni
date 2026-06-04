package handler

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/scaffolding"
)

// TestRunHandlerInit_RejectsTraversalName guards the scaffolding name path-
// traversal finding (HARDENING 2026-06-04 second pass, CWE-22): the name is
// joined into the output path, so a value containing a path separator or ".."
// must be rejected before any file is written.
func TestRunHandlerInit_RejectsTraversalName(t *testing.T) {
	for _, name := range []string{"../evil", `..\evil`, "sub/evil", `abs\path`, "/etc", ".."} {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer
		err := RunHandlerInit(&buf, fs, name, HandlerOptions{}, scaffolding.Options{})
		if !errors.Is(err, cmderr.ErrInvalidInput) {
			t.Errorf("RunHandlerInit(name=%q) err = %v, want cmderr.ErrInvalidInput", name, err)
		}
	}
}
