package testgen

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/inovacc/omni/internal/cli/scaffolding"
)

func TestRunTestInit(t *testing.T) {
	t.Run("generates test for source file", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a sample Go source file (on real fs since go/parser needs it)
		srcPath := filepath.Join(tmpDir, "sample.go")
		srcContent := `package sample

func Add(a, b int) int {
	return a + b
}

func Multiply(x, y int) int {
	return x * y
}
`
		if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
			t.Fatal(err)
		}

		fs := afero.NewOsFs()
		var buf bytes.Buffer

		err := RunTestInit(&buf, fs, srcPath, TestOptions{Table: true}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunTestInit() error = %v", err)
		}

		// Check test file was created
		testPath := filepath.Join(tmpDir, "sample_test.go")
		if _, err := os.Stat(testPath); os.IsNotExist(err) {
			t.Error("sample_test.go should be created")
		}
	})

	t.Run("missing source path", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunTestInit(&buf, fs, "", TestOptions{}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error for missing source path")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunTestInit(&buf, fs, "/tmp/nonexistent_file.go", TestOptions{}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})
}
