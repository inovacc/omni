package repository

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/internal/cli/scaffolding"
)

func TestRunRepositoryInit(t *testing.T) {
	t.Run("basic postgres repository", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "repo_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		var buf bytes.Buffer

		err = RunRepositoryInit(&buf, "user", RepositoryOptions{
			Dir:       tmpDir,
			Interface: true,
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunRepositoryInit() error = %v", err)
		}

		// Check files exist
		repoPath := filepath.Join(tmpDir, "user.go")
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			t.Error("user.go should be created")
		}

		interfacePath := filepath.Join(tmpDir, "interface.go")
		if _, err := os.Stat(interfacePath); os.IsNotExist(err) {
			t.Error("interface.go should be created")
		}

		testPath := filepath.Join(tmpDir, "user_test.go")
		if _, err := os.Stat(testPath); os.IsNotExist(err) {
			t.Error("user_test.go should be created")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRepositoryInit(&buf, "", RepositoryOptions{}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error for missing name")
		}
	})
}
