package repository

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"

	"github.com/inovacc/omni/internal/cli/scaffolding"
)

func TestRunRepositoryInit(t *testing.T) {
	t.Run("basic postgres repository", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		var buf bytes.Buffer

		err := RunRepositoryInit(&buf, fs, "user", RepositoryOptions{
			Dir:       "/tmp/repo",
			Interface: true,
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunRepositoryInit() error = %v", err)
		}

		// Check files exist
		if _, err := fs.Stat("/tmp/repo/user.go"); err != nil {
			t.Error("user.go should be created")
		}

		if _, err := fs.Stat("/tmp/repo/interface.go"); err != nil {
			t.Error("interface.go should be created")
		}

		if _, err := fs.Stat("/tmp/repo/user_test.go"); err != nil {
			t.Error("user_test.go should be created")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunRepositoryInit(&buf, fs, "", RepositoryOptions{}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error for missing name")
		}
	})
}
