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

func TestFullMode(t *testing.T) {
	t.Run("full initialization", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_full_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "fullapp")

		var buf bytes.Buffer

		err = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module: "github.com/test/fullapp",
			Full:   true,
			Author: "Test Author",
		}, Options{})
		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		// Check full mode files were created
		expectedFiles := []string{
			"main.go",
			"go.mod",
			"cmd/root.go",
			"cmd/version.go",
			"README.md",
			"Taskfile.yml",
			".gitignore",
			".editorconfig",
			".goreleaser.yaml",
			".golangci.yml",
			"tools.go",
			".github/workflows/build.yml",
			".github/workflows/test.yml",
			".github/workflows/release.yaml",
		}

		for _, f := range expectedFiles {
			path := filepath.Join(appDir, f)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Expected file %s not created in full mode", f)
			}
		}
	})

	t.Run("full version.go content", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_fullversion_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "fullverapp")

		var buf bytes.Buffer

		_ = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module: "github.com/test/fullverapp",
			Full:   true,
		}, Options{})

		content, _ := os.ReadFile(filepath.Join(appDir, "cmd", "version.go"))
		versionStr := string(content)

		// Full mode should have extended version info
		if !strings.Contains(versionStr, "BuildHash") {
			t.Error("Full mode version.go should have BuildHash")
		}

		if !strings.Contains(versionStr, "GoVersion") {
			t.Error("Full mode version.go should have GoVersion")
		}

		if !strings.Contains(versionStr, "GOOS") {
			t.Error("Full mode version.go should have GOOS")
		}

		if !strings.Contains(versionStr, "GOARCH") {
			t.Error("Full mode version.go should have GOARCH")
		}
	})

	t.Run("goreleaser content", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_goreleaser_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "gorelapp")

		var buf bytes.Buffer

		_ = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module: "github.com/test/gorelapp",
			Full:   true,
		}, Options{})

		content, _ := os.ReadFile(filepath.Join(appDir, ".goreleaser.yaml"))
		gorelStr := string(content)

		if !strings.Contains(gorelStr, "version: 2") {
			t.Error(".goreleaser.yaml should have version 2")
		}

		if !strings.Contains(gorelStr, "CGO_ENABLED=0") {
			t.Error(".goreleaser.yaml should set CGO_ENABLED=0")
		}
	})

	t.Run("golangci-lint content", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_golangci_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "lintapp")

		var buf bytes.Buffer

		_ = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module: "github.com/test/lintapp",
			Full:   true,
		}, Options{})

		content, _ := os.ReadFile(filepath.Join(appDir, ".golangci.yml"))
		lintStr := string(content)

		if !strings.Contains(lintStr, `version: "2"`) {
			t.Error(".golangci.yml should have version 2")
		}

		if !strings.Contains(lintStr, "govet") {
			t.Error(".golangci.yml should enable govet")
		}
	})
}

func TestServiceMode(t *testing.T) {
	t.Run("service pattern initialization", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_service_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "serviceapp")

		var buf bytes.Buffer

		err = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:     "github.com/test/serviceapp",
			UseService: true,
		}, Options{})
		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		// Check service pattern files were created
		expectedFiles := []string{
			"main.go",
			"go.mod",
			"cmd/root.go",
			"cmd/version.go",
			"internal/parameters/config.go",
			"internal/service/service.go",
		}

		for _, f := range expectedFiles {
			path := filepath.Join(appDir, f)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Expected file %s not created in service mode", f)
			}
		}
	})

	t.Run("service root.go content", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_serviceroot_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "servicerootapp")

		var buf bytes.Buffer

		_ = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:     "github.com/test/servicerootapp",
			UseService: true,
		}, Options{})

		content, _ := os.ReadFile(filepath.Join(appDir, "cmd", "root.go"))
		rootStr := string(content)

		if !strings.Contains(rootStr, "inovacc/config") {
			t.Error("Service mode root.go should import inovacc/config")
		}

		if !strings.Contains(rootStr, "service.Handler") {
			t.Error("Service mode root.go should use service.Handler")
		}

		if !strings.Contains(rootStr, "internal/parameters") {
			t.Error("Service mode root.go should import parameters")
		}

		if !strings.Contains(rootStr, "internal/service") {
			t.Error("Service mode root.go should import service")
		}
	})

	t.Run("parameters content", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_params_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "paramsapp")

		var buf bytes.Buffer

		_ = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:     "github.com/test/paramsapp",
			UseService: true,
		}, Options{})

		content, _ := os.ReadFile(filepath.Join(appDir, "internal", "parameters", "config.go"))
		paramsStr := string(content)

		if !strings.Contains(paramsStr, "package parameters") {
			t.Error("parameters/config.go should have package parameters")
		}

		if !strings.Contains(paramsStr, "type Service struct") {
			t.Error("parameters/config.go should define Service struct")
		}
	})

	t.Run("service content", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_svc_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "svcapp")

		var buf bytes.Buffer

		_ = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:     "github.com/test/svcapp",
			UseService: true,
		}, Options{})

		content, _ := os.ReadFile(filepath.Join(appDir, "internal", "service", "service.go"))
		svcStr := string(content)

		if !strings.Contains(svcStr, "package service") {
			t.Error("service/service.go should have package service")
		}

		if !strings.Contains(svcStr, "func Handler") {
			t.Error("service/service.go should define Handler function")
		}

		if !strings.Contains(svcStr, "config.GetServiceConfig") {
			t.Error("service/service.go should use config.GetServiceConfig")
		}
	})

	t.Run("go.mod includes config dependency", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_gomod_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		appDir := filepath.Join(tmpDir, "gomodapp")

		var buf bytes.Buffer

		_ = RunCobraInit(&buf, appDir, CobraInitOptions{
			Module:     "github.com/test/gomodapp",
			UseService: true,
		}, Options{})

		content, _ := os.ReadFile(filepath.Join(appDir, "go.mod"))
		gomodStr := string(content)

		if !strings.Contains(gomodStr, "github.com/inovacc/config") {
			t.Error("go.mod should include inovacc/config dependency in service mode")
		}
	})
}

func TestEditorConfigContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "editorconfig_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	appDir := filepath.Join(tmpDir, "editortest")

	var buf bytes.Buffer

	_ = RunCobraInit(&buf, appDir, CobraInitOptions{
		Module:  "github.com/test/editortest",
		AppName: "editortest",
	}, Options{})

	content, _ := os.ReadFile(filepath.Join(appDir, ".editorconfig"))
	editorStr := string(content)

	if !strings.Contains(editorStr, "root = true") {
		t.Error(".editorconfig should have root = true")
	}

	if !strings.Contains(editorStr, "[*.{go,go2}]") {
		t.Error(".editorconfig should have Go file section")
	}

	if !strings.Contains(editorStr, "indent_style = tab") {
		t.Error(".editorconfig should use tabs for Go files")
	}
}

func TestCobraConfig(t *testing.T) {
	t.Run("write and read config", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "cobra_config_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		configPath := filepath.Join(tmpDir, ".cobra.yaml")

		cfg := &CobraConfig{
			Author:     "Test Author <test@example.com>",
			License:    "MIT",
			UseViper:   true,
			UseService: false,
			Full:       true,
		}

		err = WriteDefaultConfig(configPath, cfg)
		if err != nil {
			t.Fatalf("WriteDefaultConfig() error = %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatal("Config file was not created")
		}

		// Read it back
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "author: Test Author <test@example.com>") {
			t.Error("Config should contain author")
		}

		if !strings.Contains(contentStr, "license: MIT") {
			t.Error("Config should contain license")
		}

		if !strings.Contains(contentStr, "useViper: true") {
			t.Error("Config should contain useViper")
		}

		if !strings.Contains(contentStr, "full: true") {
			t.Error("Config should contain full")
		}
	})

	t.Run("merge config with flags", func(t *testing.T) {
		cfg := &CobraConfig{
			Author:     "Config Author",
			License:    "Apache-2.0",
			UseViper:   true,
			UseService: true,
			Full:       true,
		}

		opts := &CobraInitOptions{
			Module:  "github.com/test/app",
			Author:  "Flag Author", // explicitly set via flag
			License: "",            // not set via flag
		}

		// Simulate flags that were explicitly set
		flagsSet := map[string]bool{
			"author": true, // author was set via flag
			// license was NOT set via flag
		}

		cfg.MergeWithFlags(opts, flagsSet)

		// Author should NOT be overwritten (flag was set)
		if opts.Author != "Flag Author" {
			t.Errorf("Author should be 'Flag Author', got '%s'", opts.Author)
		}

		// License should be overwritten from config (flag was not set)
		if opts.License != "Apache-2.0" {
			t.Errorf("License should be 'Apache-2.0', got '%s'", opts.License)
		}

		// UseViper should be set from config
		if !opts.UseViper {
			t.Error("UseViper should be true from config")
		}

		// UseService should be set from config
		if !opts.UseService {
			t.Error("UseService should be true from config")
		}

		// Full should be set from config
		if !opts.Full {
			t.Error("Full should be true from config")
		}
	})

	t.Run("empty config when no file exists", func(t *testing.T) {
		// This test relies on LoadCobraConfig searching for files that don't exist
		// Since we can't easily override the home directory, we just verify the function
		// returns an empty config when files don't exist (it should not error)
		cfg, path, err := LoadCobraConfig()
		if err != nil {
			// If there's an actual config file in the user's home, that's fine
			// We just want to make sure the function works
			t.Logf("LoadCobraConfig returned error (may have existing config): %v", err)
		}

		// If no config file found, path should be empty and cfg should be non-nil
		if path == "" && cfg == nil {
			t.Error("LoadCobraConfig should return non-nil config even when no file exists")
		}
	})

	t.Run("default config path", func(t *testing.T) {
		path := DefaultConfigPath()
		if path == "" {
			t.Skip("Could not determine home directory")
		}

		if !strings.HasSuffix(path, ".cobra.yaml") {
			t.Errorf("Default config path should end with .cobra.yaml, got %s", path)
		}
	})
}
