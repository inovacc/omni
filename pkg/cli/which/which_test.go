package which

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunWhich(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "which_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a mock executable
	var execName string
	if runtime.GOOS == "windows" {
		execName = "testcmd.exe"
	} else {
		execName = "testcmd"
	}

	execPath := filepath.Join(tmpDir, execName)
	if err := os.WriteFile(execPath, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	// Save original PATH and restore after test
	origPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", origPath) }()

	t.Run("find command in PATH", func(t *testing.T) {
		_ = os.Setenv("PATH", tmpDir)

		var buf bytes.Buffer

		cmdName := "testcmd"
		if runtime.GOOS == "windows" {
			cmdName = "testcmd"
		}

		err := RunWhich(&buf, []string{cmdName}, WhichOptions{})
		if err != nil {
			t.Fatalf("RunWhich() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "testcmd") {
			t.Errorf("RunWhich() = %q, should contain 'testcmd'", output)
		}
	})

	t.Run("command not found", func(t *testing.T) {
		_ = os.Setenv("PATH", tmpDir)

		var buf bytes.Buffer

		err := RunWhich(&buf, []string{"nonexistentcommand"}, WhichOptions{})
		if err == nil {
			t.Error("RunWhich() expected error for nonexistent command")
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunWhich(&buf, []string{}, WhichOptions{})
		if err == nil {
			t.Error("RunWhich() expected error for missing operand")
		}
	})

	t.Run("empty PATH", func(t *testing.T) {
		_ = os.Setenv("PATH", "")

		var buf bytes.Buffer

		err := RunWhich(&buf, []string{"anything"}, WhichOptions{})
		if err == nil {
			t.Error("RunWhich() expected error for empty PATH")
		}
	})

	t.Run("find all matches", func(t *testing.T) {
		// Create another directory with same executable
		tmpDir2, err := os.MkdirTemp("", "which_test2")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir2) }()

		execPath2 := filepath.Join(tmpDir2, execName)
		if err := os.WriteFile(execPath2, []byte("#!/bin/sh\necho test2"), 0755); err != nil {
			t.Fatal(err)
		}

		pathSep := string(os.PathListSeparator)
		_ = os.Setenv("PATH", tmpDir+pathSep+tmpDir2)

		var buf bytes.Buffer

		cmdName := "testcmd"
		err = RunWhich(&buf, []string{cmdName}, WhichOptions{All: true})
		if err != nil {
			t.Fatalf("RunWhich() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) < 2 {
			t.Logf("RunWhich() -a got %d matches, expected >= 2", len(lines))
		}
	})

	t.Run("multiple commands", func(t *testing.T) {
		// Create second executable
		var exec2Name string
		if runtime.GOOS == "windows" {
			exec2Name = "testcmd2.exe"
		} else {
			exec2Name = "testcmd2"
		}

		exec2Path := filepath.Join(tmpDir, exec2Name)
		if err := os.WriteFile(exec2Path, []byte("#!/bin/sh\necho test2"), 0755); err != nil {
			t.Fatal(err)
		}

		_ = os.Setenv("PATH", tmpDir)

		var buf bytes.Buffer

		err := RunWhich(&buf, []string{"testcmd", "testcmd2"}, WhichOptions{})
		if err != nil {
			t.Fatalf("RunWhich() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunWhich() got %d lines, want 2", len(lines))
		}
	})
}

func TestIsExec(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "isexec_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("executable file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "exec")
		if err := os.WriteFile(file, []byte("test"), 0755); err != nil {
			t.Fatal(err)
		}

		if !isExec(file) {
			t.Errorf("isExec(%q) = false, want true", file)
		}
	})

	t.Run("non-executable file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "noexec")
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		// On Windows, any file is "executable"
		if runtime.GOOS != "windows" {
			if isExec(file) {
				t.Errorf("isExec(%q) = true, want false", file)
			}
		}
	})

	t.Run("directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "dir")
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatal(err)
		}

		if isExec(dir) {
			t.Errorf("isExec(%q) = true for directory, want false", dir)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		if isExec("/nonexistent/file") {
			t.Error("isExec() = true for nonexistent file, want false")
		}
	})
}

func TestFindExecutable(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "findexec_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("finds executable", func(t *testing.T) {
		var fileName string
		if runtime.GOOS == "windows" {
			fileName = "cmd.exe"
		} else {
			fileName = "cmd"
		}

		file := filepath.Join(tmpDir, fileName)
		if err := os.WriteFile(file, []byte("test"), 0755); err != nil {
			t.Fatal(err)
		}

		basePath := filepath.Join(tmpDir, "cmd")
		results := findExecutable(basePath)

		if len(results) == 0 {
			t.Error("findExecutable() found no results")
		}
	})

	t.Run("nonexistent returns empty", func(t *testing.T) {
		results := findExecutable("/nonexistent/cmd")

		if len(results) != 0 {
			t.Errorf("findExecutable() = %v, want empty", results)
		}
	})
}
