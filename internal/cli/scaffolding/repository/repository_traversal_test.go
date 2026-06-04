package repository

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/scaffolding"
)

// TestRunRepositoryInit_RejectsTraversalName guards the scaffolding name path-
// traversal finding (HARDENING 2026-06-04 second pass, CWE-22).
func TestRunRepositoryInit_RejectsTraversalName(t *testing.T) {
	for _, name := range []string{"../evil", `..\evil`, "sub/evil", `abs\path`, "/etc", ".."} {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer
		err := RunRepositoryInit(&buf, fs, name, RepositoryOptions{}, scaffolding.Options{})
		if !errors.Is(err, cmderr.ErrInvalidInput) {
			t.Errorf("RunRepositoryInit(name=%q) err = %v, want cmderr.ErrInvalidInput", name, err)
		}
	}
}
