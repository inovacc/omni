package handler

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"

	"github.com/inovacc/omni/internal/cli/scaffolding"
)

func TestRunHandlerInit(t *testing.T) {
	t.Run("basic stdlib handler", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		var buf bytes.Buffer

		err := RunHandlerInit(&buf, fs, "user", HandlerOptions{
			Dir: "/tmp/handler",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunHandlerInit() error = %v", err)
		}

		// Check handler file exists
		if _, err := fs.Stat("/tmp/handler/user.go"); err != nil {
			t.Error("user.go should be created")
		}

		// Check test file exists
		if _, err := fs.Stat("/tmp/handler/user_test.go"); err != nil {
			t.Error("user_test.go should be created")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunHandlerInit(&buf, fs, "", HandlerOptions{}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error for missing name")
		}
	})
}
