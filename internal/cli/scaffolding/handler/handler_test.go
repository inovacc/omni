package handler

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/internal/cli/scaffolding"
)

func TestRunHandlerInit(t *testing.T) {
	t.Run("basic stdlib handler", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "handler_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		var buf bytes.Buffer

		err = RunHandlerInit(&buf, "user", HandlerOptions{
			Dir: tmpDir,
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunHandlerInit() error = %v", err)
		}

		// Check handler file exists
		handlerPath := filepath.Join(tmpDir, "user.go")
		if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
			t.Error("user.go should be created")
		}

		// Check test file exists
		testPath := filepath.Join(tmpDir, "user_test.go")
		if _, err := os.Stat(testPath); os.IsNotExist(err) {
			t.Error("user_test.go should be created")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunHandlerInit(&buf, "", HandlerOptions{}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error for missing name")
		}
	})
}
