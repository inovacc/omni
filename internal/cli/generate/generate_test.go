package generate

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCobraInit(t *testing.T) {
	t.Run("basic initialization", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_init_test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "myapp")

		var buf bytes.Buffer
		err = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:      "github.com/test/myapp",
			AppName:     "myapp",
			Description: "Test application",
		}, Options{})

		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		// Check files were created
		expectedFiles := []string{
			"main.go",
			"go.mod",
			"cmd/root.go",
			"cmd/version.go",
			"README.md",
			"Taskfile.yml",
			".gitignore",
		}

		for _, f := range expectedFiles {
			path := filepath.Join(appDir, f)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Expected file %s not created", f)
			}
		}
	})

	t.Run("with viper", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_viper_test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "viperapp")

		var buf bytes.Buffer
		err = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:   "github.com/test/viperapp",
			UseViper: true,
		}, Options{})

		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		// Check config file was created
		configPath := filepath.Join(appDir, "internal", "config", "config.go")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Expected config.go to be created with viper option")
		}
	})

	t.Run("with MIT license", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_license_test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "mitapp")

		var buf bytes.Buffer
		err = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:  "github.com/test/mitapp",
			License: "MIT",
			Author:  "Test Author",
		}, Options{})

		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		licensePath := filepath.Join(appDir, "LICENSE")
		content, err := os.ReadFile(licensePath)
		if err != nil {
			t.Fatalf("Failed to read LICENSE: %v", err)
		}

		if !strings.Contains(string(content), "MIT License") {
			t.Error("LICENSE should contain MIT License")
		}

		if !strings.Contains(string(content), "Test Author") {
			t.Error("LICENSE should contain author name")
		}
	})

	t.Run("with Apache license", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_apache_test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "apacheapp")

		var buf bytes.Buffer
		err = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:  "github.com/test/apacheapp",
			License: "Apache-2.0",
			Author:  "Apache Author",
		}, Options{})

		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		licensePath := filepath.Join(appDir, "LICENSE")
		content, err := os.ReadFile(licensePath)
		if err != nil {
			t.Fatalf("Failed to read LICENSE: %v", err)
		}

		if !strings.Contains(string(content), "Apache License") {
			t.Error("LICENSE should contain Apache License")
		}
	})

	t.Run("json output", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_json_test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "jsonapp")

		var buf bytes.Buffer
		err = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module: "github.com/test/jsonapp",
		}, Options{JSON: true})

		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		var result InitResult
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Invalid JSON output: %v", err)
		}

		if result.Status != "created" {
			t.Errorf("Status = %v, want created", result.Status)
		}

		if result.Module != "github.com/test/jsonapp" {
			t.Errorf("Module = %v, want github.com/test/jsonapp", result.Module)
		}

		if len(result.FilesCreated) == 0 {
			t.Error("FilesCreated should not be empty")
		}
	})

	t.Run("missing module", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunCobraInit(&buf, "/tmp/test", CobraInitOptions{}, Options{})

		if err == nil {
			t.Error("Expected error for missing module")
		}
	})

	t.Run("extracts app name from module", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_name_test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "extracted")

		var buf bytes.Buffer
		err = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module: "github.com/test/extractedapp",
		}, Options{})

		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		// Check that the app name was extracted
		mainContent, _ := os.ReadFile(filepath.Join(appDir, "main.go"))
		if !strings.Contains(string(mainContent), "github.com/test/extractedapp") {
			t.Error("main.go should contain the module path")
		}
	})

	t.Run("main.go content", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_main_test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "maintest")

		var buf bytes.Buffer
		_ = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module: "github.com/test/maintest",
		}, Options{})

		content, _ := os.ReadFile(filepath.Join(appDir, "main.go"))
		mainStr := string(content)

		if !strings.Contains(mainStr, "package main") {
			t.Error("main.go should have package main")
		}

		if !strings.Contains(mainStr, "cmd.Execute()") {
			t.Error("main.go should call cmd.Execute()")
		}
	})

	t.Run("root.go content", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_root_test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "roottest")

		var buf bytes.Buffer
		_ = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:      "github.com/test/roottest",
			AppName:     "roottest",
			Description: "Test description",
		}, Options{})

		content, _ := os.ReadFile(filepath.Join(appDir, "cmd", "root.go"))
		rootStr := string(content)

		if !strings.Contains(rootStr, "package cmd") {
			t.Error("root.go should have package cmd")
		}

		if !strings.Contains(rootStr, "roottest") {
			t.Error("root.go should contain app name")
		}

		if !strings.Contains(rootStr, "Test description") {
			t.Error("root.go should contain description")
		}
	})
}

func TestRunCobraAdd(t *testing.T) {
	// Create a test project first
	setupProject := func(t *testing.T) (string, func()) {
		tmpDir, err := os.MkdirTemp("", "cobra_add_test")
		if err != nil {
			t.Fatal(err)
		}

		appDir := filepath.Join(tmpDir, "addtest")

		var buf bytes.Buffer
		err = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module: "github.com/test/addtest",
		}, Options{})

		if err != nil {
			_ = os.RemoveAll(tmpDir)
			t.Fatal(err)
		}

		return appDir, func() { _ = os.RemoveAll(tmpDir) }
	}

	t.Run("add command to root", func(t *testing.T) {
		appDir, cleanup := setupProject(t)
		defer cleanup()

		var buf bytes.Buffer
		err := RunCobraAdd(&buf, appDir, CobraAddOptions{
			Name:        "serve",
			Parent:      "root",
			Description: "Start the server",
		}, Options{})

		if err != nil {
			t.Fatalf("RunCobraAdd() error = %v", err)
		}

		// Check file was created
		servePath := filepath.Join(appDir, "cmd", "serve.go")
		if _, err := os.Stat(servePath); os.IsNotExist(err) {
			t.Error("serve.go should be created")
		}

		content, _ := os.ReadFile(servePath)
		serveStr := string(content)

		if !strings.Contains(serveStr, "serveCmd") {
			t.Error("serve.go should define serveCmd")
		}

		if !strings.Contains(serveStr, "rootCmd.AddCommand(serveCmd)") {
			t.Error("serve.go should add to rootCmd")
		}
	})

	t.Run("add command with default parent", func(t *testing.T) {
		appDir, cleanup := setupProject(t)
		defer cleanup()

		var buf bytes.Buffer
		err := RunCobraAdd(&buf, appDir, CobraAddOptions{
			Name: "config",
		}, Options{})

		if err != nil {
			t.Fatalf("RunCobraAdd() error = %v", err)
		}

		content, _ := os.ReadFile(filepath.Join(appDir, "cmd", "config.go"))
		if !strings.Contains(string(content), "rootCmd.AddCommand") {
			t.Error("Should default to root parent")
		}
	})

	t.Run("add subcommand", func(t *testing.T) {
		appDir, cleanup := setupProject(t)
		defer cleanup()

		// First add a parent command
		var buf bytes.Buffer
		_ = RunCobraAdd(&buf, appDir, CobraAddOptions{
			Name:   "user",
			Parent: "root",
		}, Options{})

		// Add subcommand
		buf.Reset()
		err := RunCobraAdd(&buf, appDir, CobraAddOptions{
			Name:   "list",
			Parent: "user",
		}, Options{})

		if err != nil {
			t.Fatalf("RunCobraAdd() error = %v", err)
		}

		content, _ := os.ReadFile(filepath.Join(appDir, "cmd", "list.go"))
		if !strings.Contains(string(content), "userCmd.AddCommand(listCmd)") {
			t.Error("list.go should add to userCmd")
		}
	})

	t.Run("json output", func(t *testing.T) {
		appDir, cleanup := setupProject(t)
		defer cleanup()

		var buf bytes.Buffer
		err := RunCobraAdd(&buf, appDir, CobraAddOptions{
			Name: "status",
		}, Options{JSON: true})

		if err != nil {
			t.Fatalf("RunCobraAdd() error = %v", err)
		}

		var result AddResult
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Invalid JSON output: %v", err)
		}

		if result.Status != "created" {
			t.Errorf("Status = %v, want created", result.Status)
		}

		if result.Command != "status" {
			t.Errorf("Command = %v, want status", result.Command)
		}
	})

	t.Run("missing command name", func(t *testing.T) {
		appDir, cleanup := setupProject(t)
		defer cleanup()

		var buf bytes.Buffer
		err := RunCobraAdd(&buf, appDir, CobraAddOptions{}, Options{})

		if err == nil {
			t.Error("Expected error for missing command name")
		}
	})

	t.Run("command already exists", func(t *testing.T) {
		appDir, cleanup := setupProject(t)
		defer cleanup()

		var buf bytes.Buffer
		_ = RunCobraAdd(&buf, appDir, CobraAddOptions{Name: "duplicate"}, Options{})

		buf.Reset()
		err := RunCobraAdd(&buf, appDir, CobraAddOptions{Name: "duplicate"}, Options{})

		if err == nil {
			t.Error("Expected error for duplicate command")
		}
	})

	t.Run("not a cobra project", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "not_cobra_test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		var buf bytes.Buffer
		err = RunCobraAdd(&buf, tmpDir, CobraAddOptions{Name: "test"}, Options{})

		if err == nil {
			t.Error("Expected error for non-Cobra project")
		}
	})
}

func TestLicenses(t *testing.T) {
	t.Run("unknown license", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "unknown_license_test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "unknownlic")

		var buf bytes.Buffer
		err = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:  "github.com/test/unknownlic",
			License: "UNKNOWN",
		}, Options{})

		if err == nil {
			t.Error("Expected error for unknown license type")
		}
	})

	t.Run("BSD license", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "bsd_license_test")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "bsdapp")

		var buf bytes.Buffer
		err = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:  "github.com/test/bsdapp",
			License: "BSD-3",
			Author:  "BSD Author",
		}, Options{})

		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		content, _ := os.ReadFile(filepath.Join(appDir, "LICENSE"))
		if !strings.Contains(string(content), "BSD 3-Clause") {
			t.Error("LICENSE should contain BSD 3-Clause")
		}
	})
}

func TestTaskfileContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "taskfile_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	appDir := filepath.Join(tmpDir, "tasktest")

	var buf bytes.Buffer
	_ = RunCobraInit(&buf, appDir, CobraInitOptions{
		Module:  "github.com/test/tasktest",
		AppName: "tasktest",
	}, Options{})

	content, _ := os.ReadFile(filepath.Join(appDir, "Taskfile.yml"))
	taskStr := string(content)

	if !strings.Contains(taskStr, "version: '3'") {
		t.Error("Taskfile should have version 3")
	}

	if !strings.Contains(taskStr, "build:") {
		t.Error("Taskfile should have build task")
	}

	if !strings.Contains(taskStr, "test:") {
		t.Error("Taskfile should have test task")
	}

	if !strings.Contains(taskStr, "lint:") {
		t.Error("Taskfile should have lint task")
	}
}

func TestGitignoreContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitignore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	appDir := filepath.Join(tmpDir, "ignoretest")

	var buf bytes.Buffer
	_ = RunCobraInit(&buf, appDir, CobraInitOptions{
		Module:  "github.com/test/ignoretest",
		AppName: "ignoretest",
	}, Options{})

	content, _ := os.ReadFile(filepath.Join(appDir, ".gitignore"))
	gitignoreStr := string(content)

	if !strings.Contains(gitignoreStr, "ignoretest") {
		t.Error(".gitignore should include app binary name")
	}

	if !strings.Contains(gitignoreStr, ".exe") {
		t.Error(".gitignore should include .exe")
	}

	if !strings.Contains(gitignoreStr, ".idea/") {
		t.Error(".gitignore should include IDE directories")
	}
}
