package cobra

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/inovacc/omni/internal/cli/scaffolding"
)

func TestRunCobraInit(t *testing.T) {
	t.Run("basic initialization", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/myapp", CobraInitOptions{
			Module:      "github.com/test/myapp",
			AppName:     "myapp",
			Description: "Test application",
		}, scaffolding.Options{})
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
			if _, err := fs.Stat("/myapp/" + f); err != nil {
				t.Errorf("Expected file %s not created", f)
			}
		}
	})

	t.Run("with viper", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/viperapp", CobraInitOptions{
			Module:   "github.com/test/viperapp",
			UseViper: true,
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		// Check config file was created
		if _, err := fs.Stat("/viperapp/internal/config/config.go"); err != nil {
			t.Error("Expected config.go to be created with viper option")
		}
	})

	t.Run("with MIT license", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/mitapp", CobraInitOptions{
			Module:  "github.com/test/mitapp",
			License: "MIT",
			Author:  "Test Author",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		content, err := afero.ReadFile(fs, "/mitapp/LICENSE")
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
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/apacheapp", CobraInitOptions{
			Module:  "github.com/test/apacheapp",
			License: "Apache-2.0",
			Author:  "Apache Author",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		content, err := afero.ReadFile(fs, "/apacheapp/LICENSE")
		if err != nil {
			t.Fatalf("Failed to read LICENSE: %v", err)
		}

		if !strings.Contains(string(content), "Apache License") {
			t.Error("LICENSE should contain Apache License")
		}
	})

	t.Run("json output", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/jsonapp", CobraInitOptions{
			Module: "github.com/test/jsonapp",
		}, scaffolding.Options{JSON: true})
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
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/tmp/test", CobraInitOptions{}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error for missing module")
		}
	})

	t.Run("extracts app name from module", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/extracted", CobraInitOptions{
			Module: "github.com/test/extractedapp",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		mainContent, _ := afero.ReadFile(fs, "/extracted/main.go")
		if !strings.Contains(string(mainContent), "github.com/test/extractedapp") {
			t.Error("main.go should contain the module path")
		}
	})

	t.Run("main.go content", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		_ = RunCobraInit(&buf, fs, "/maintest", CobraInitOptions{
			Module: "github.com/test/maintest",
		}, scaffolding.Options{})

		content, _ := afero.ReadFile(fs, "/maintest/main.go")
		mainStr := string(content)

		if !strings.Contains(mainStr, "package main") {
			t.Error("main.go should have package main")
		}

		if !strings.Contains(mainStr, "cmd.Execute()") {
			t.Error("main.go should call cmd.Execute()")
		}
	})

	t.Run("root.go content", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		_ = RunCobraInit(&buf, fs, "/roottest", CobraInitOptions{
			Module:      "github.com/test/roottest",
			AppName:     "roottest",
			Description: "Test description",
		}, scaffolding.Options{})

		content, _ := afero.ReadFile(fs, "/roottest/cmd/root.go")
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
	setupProject := func(t *testing.T) (afero.Fs, string) {
		fs := afero.NewMemMapFs()
		appDir := "/addtest"

		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, appDir, CobraInitOptions{
			Module: "github.com/test/addtest",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatal(err)
		}

		return fs, appDir
	}

	t.Run("add command to root", func(t *testing.T) {
		fs, appDir := setupProject(t)

		var buf bytes.Buffer

		err := RunCobraAdd(&buf, fs, appDir, CobraAddOptions{
			Name:        "serve",
			Parent:      "root",
			Description: "Start the server",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraAdd() error = %v", err)
		}

		// Check file was created
		if _, err := fs.Stat("/addtest/cmd/serve.go"); err != nil {
			t.Error("serve.go should be created")
		}

		content, _ := afero.ReadFile(fs, "/addtest/cmd/serve.go")
		serveStr := string(content)

		if !strings.Contains(serveStr, "serveCmd") {
			t.Error("serve.go should define serveCmd")
		}

		if !strings.Contains(serveStr, "rootCmd.AddCommand(serveCmd)") {
			t.Error("serve.go should add to rootCmd")
		}
	})

	t.Run("add command with default parent", func(t *testing.T) {
		fs, appDir := setupProject(t)

		var buf bytes.Buffer

		err := RunCobraAdd(&buf, fs, appDir, CobraAddOptions{
			Name: "config",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraAdd() error = %v", err)
		}

		content, _ := afero.ReadFile(fs, "/addtest/cmd/config.go")
		if !strings.Contains(string(content), "rootCmd.AddCommand") {
			t.Error("Should default to root parent")
		}
	})

	t.Run("add subcommand", func(t *testing.T) {
		fs, appDir := setupProject(t)

		// First add a parent command
		var buf bytes.Buffer

		_ = RunCobraAdd(&buf, fs, appDir, CobraAddOptions{
			Name:   "user",
			Parent: "root",
		}, scaffolding.Options{})

		// Add subcommand
		buf.Reset()

		err := RunCobraAdd(&buf, fs, appDir, CobraAddOptions{
			Name:   "list",
			Parent: "user",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraAdd() error = %v", err)
		}

		content, _ := afero.ReadFile(fs, "/addtest/cmd/list.go")
		if !strings.Contains(string(content), "userCmd.AddCommand(listCmd)") {
			t.Error("list.go should add to userCmd")
		}
	})

	t.Run("json output", func(t *testing.T) {
		fs, appDir := setupProject(t)

		var buf bytes.Buffer

		err := RunCobraAdd(&buf, fs, appDir, CobraAddOptions{
			Name: "status",
		}, scaffolding.Options{JSON: true})
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
		fs, appDir := setupProject(t)

		var buf bytes.Buffer

		err := RunCobraAdd(&buf, fs, appDir, CobraAddOptions{}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error for missing command name")
		}
	})

	t.Run("command already exists", func(t *testing.T) {
		fs, appDir := setupProject(t)

		var buf bytes.Buffer

		_ = RunCobraAdd(&buf, fs, appDir, CobraAddOptions{Name: "duplicate"}, scaffolding.Options{})

		buf.Reset()

		err := RunCobraAdd(&buf, fs, appDir, CobraAddOptions{Name: "duplicate"}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error for duplicate command")
		}
	})

	t.Run("not a cobra project", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		var buf bytes.Buffer

		err := RunCobraAdd(&buf, fs, "/empty", CobraAddOptions{Name: "test"}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error for non-Cobra project")
		}
	})
}

func TestLicenses(t *testing.T) {
	t.Run("unknown license", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/unknownlic", CobraInitOptions{
			Module:  "github.com/test/unknownlic",
			License: "UNKNOWN",
		}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error for unknown license type")
		}
	})

	t.Run("BSD license", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/bsdapp", CobraInitOptions{
			Module:  "github.com/test/bsdapp",
			License: "BSD-3",
			Author:  "BSD Author",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		content, _ := afero.ReadFile(fs, "/bsdapp/LICENSE")
		if !strings.Contains(string(content), "BSD 3-Clause") {
			t.Error("LICENSE should contain BSD 3-Clause")
		}
	})
}

func TestTaskfileContent(t *testing.T) {
	fs := afero.NewMemMapFs()
	var buf bytes.Buffer

	_ = RunCobraInit(&buf, fs, "/tasktest", CobraInitOptions{
		Module:  "github.com/test/tasktest",
		AppName: "tasktest",
	}, scaffolding.Options{})

	content, _ := afero.ReadFile(fs, "/tasktest/Taskfile.yml")
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
	fs := afero.NewMemMapFs()
	var buf bytes.Buffer

	_ = RunCobraInit(&buf, fs, "/ignoretest", CobraInitOptions{
		Module:  "github.com/test/ignoretest",
		AppName: "ignoretest",
	}, scaffolding.Options{})

	content, _ := afero.ReadFile(fs, "/ignoretest/.gitignore")
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
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/fullapp", CobraInitOptions{
			Module: "github.com/test/fullapp",
			Full:   true,
			Author: "Test Author",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

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
			if _, err := fs.Stat("/fullapp/" + f); err != nil {
				t.Errorf("Expected file %s not created in full mode", f)
			}
		}
	})

	t.Run("full version.go content", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		_ = RunCobraInit(&buf, fs, "/fullverapp", CobraInitOptions{
			Module: "github.com/test/fullverapp",
			Full:   true,
		}, scaffolding.Options{})

		content, _ := afero.ReadFile(fs, "/fullverapp/cmd/version.go")
		versionStr := string(content)

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
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		_ = RunCobraInit(&buf, fs, "/gorelapp", CobraInitOptions{
			Module: "github.com/test/gorelapp",
			Full:   true,
		}, scaffolding.Options{})

		content, _ := afero.ReadFile(fs, "/gorelapp/.goreleaser.yaml")
		gorelStr := string(content)

		if !strings.Contains(gorelStr, "version: 2") {
			t.Error(".goreleaser.yaml should have version 2")
		}

		if !strings.Contains(gorelStr, "CGO_ENABLED=0") {
			t.Error(".goreleaser.yaml should set CGO_ENABLED=0")
		}
	})

	t.Run("golangci-lint content", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		_ = RunCobraInit(&buf, fs, "/lintapp", CobraInitOptions{
			Module: "github.com/test/lintapp",
			Full:   true,
		}, scaffolding.Options{})

		content, _ := afero.ReadFile(fs, "/lintapp/.golangci.yml")
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
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/serviceapp", CobraInitOptions{
			Module:     "github.com/test/serviceapp",
			UseService: true,
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		expectedFiles := []string{
			"main.go",
			"go.mod",
			"cmd/root.go",
			"cmd/version.go",
			"internal/parameters/config.go",
			"internal/service/service.go",
		}

		for _, f := range expectedFiles {
			if _, err := fs.Stat("/serviceapp/" + f); err != nil {
				t.Errorf("Expected file %s not created in service mode", f)
			}
		}
	})

	t.Run("service root.go content", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		_ = RunCobraInit(&buf, fs, "/servicerootapp", CobraInitOptions{
			Module:     "github.com/test/servicerootapp",
			UseService: true,
		}, scaffolding.Options{})

		content, _ := afero.ReadFile(fs, "/servicerootapp/cmd/root.go")
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
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		_ = RunCobraInit(&buf, fs, "/paramsapp", CobraInitOptions{
			Module:     "github.com/test/paramsapp",
			UseService: true,
		}, scaffolding.Options{})

		content, _ := afero.ReadFile(fs, "/paramsapp/internal/parameters/config.go")
		paramsStr := string(content)

		if !strings.Contains(paramsStr, "package parameters") {
			t.Error("parameters/config.go should have package parameters")
		}

		if !strings.Contains(paramsStr, "type Service struct") {
			t.Error("parameters/config.go should define Service struct")
		}
	})

	t.Run("service content", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		_ = RunCobraInit(&buf, fs, "/svcapp", CobraInitOptions{
			Module:     "github.com/test/svcapp",
			UseService: true,
		}, scaffolding.Options{})

		content, _ := afero.ReadFile(fs, "/svcapp/internal/service/service.go")
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
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		_ = RunCobraInit(&buf, fs, "/gomodapp", CobraInitOptions{
			Module:     "github.com/test/gomodapp",
			UseService: true,
		}, scaffolding.Options{})

		content, _ := afero.ReadFile(fs, "/gomodapp/go.mod")
		gomodStr := string(content)

		if !strings.Contains(gomodStr, "github.com/inovacc/config") {
			t.Error("go.mod should include inovacc/config dependency in service mode")
		}
	})
}

func TestRunCobraInitWithCmdtree(t *testing.T) {
	fs := afero.NewMemMapFs()
	var buf bytes.Buffer

	err := RunCobraInit(&buf, fs, "/cmdtreeapp", CobraInitOptions{
		Module: "github.com/test/cmdtreeapp",
	}, scaffolding.Options{})
	if err != nil {
		t.Fatalf("RunCobraInit() error = %v", err)
	}

	// cmdtree.go should always be created
	if _, err := fs.Stat("/cmdtreeapp/cmd/cmdtree.go"); err != nil {
		t.Error("cmd/cmdtree.go should always be created")
	}

	content, _ := afero.ReadFile(fs, "/cmdtreeapp/cmd/cmdtree.go")
	if !strings.Contains(string(content), "cmdtreeCmd") {
		t.Error("cmdtree.go should define cmdtreeCmd")
	}

	if !strings.Contains(string(content), "collectPersistentFlags") {
		t.Error("cmdtree.go should define collectPersistentFlags")
	}

	// aicontext.go should NOT be created by default
	if _, err := fs.Stat("/cmdtreeapp/cmd/aicontext.go"); err == nil {
		t.Error("cmd/aicontext.go should NOT be created by default")
	}
}

func TestRunCobraInitWithAIContext(t *testing.T) {
	fs := afero.NewMemMapFs()
	var buf bytes.Buffer

	err := RunCobraInit(&buf, fs, "/aiapp", CobraInitOptions{
		Module:      "github.com/test/aiapp",
		AppName:     "aiapp",
		Description: "AI test app",
		AIContext:   true,
	}, scaffolding.Options{})
	if err != nil {
		t.Fatalf("RunCobraInit() error = %v", err)
	}

	// Both should be created
	if _, err := fs.Stat("/aiapp/cmd/cmdtree.go"); err != nil {
		t.Error("cmd/cmdtree.go should be created")
	}

	if _, err := fs.Stat("/aiapp/cmd/aicontext.go"); err != nil {
		t.Error("cmd/aicontext.go should be created when AIContext=true")
	}

	content, _ := afero.ReadFile(fs, "/aiapp/cmd/aicontext.go")
	aiStr := string(content)

	if !strings.Contains(aiStr, "aiapp") {
		t.Error("aicontext.go should contain app name")
	}

	if !strings.Contains(aiStr, "AI test app") {
		t.Error("aicontext.go should contain description")
	}

	if !strings.Contains(aiStr, "aicontextCmd") {
		t.Error("aicontext.go should define aicontextCmd")
	}
}

func TestRunCobraAddTools(t *testing.T) {
	setupMinimalProject := func(t *testing.T) (afero.Fs, string) {
		fs := afero.NewMemMapFs()
		appDir := "/toolsapp"

		// Create minimal project structure
		if err := fs.MkdirAll(appDir+"/cmd", 0755); err != nil {
			t.Fatal(err)
		}

		// Write go.mod
		goMod := []byte("module github.com/test/toolsapp\n\ngo 1.21\n")
		if err := afero.WriteFile(fs, appDir+"/go.mod", goMod, 0644); err != nil {
			t.Fatal(err)
		}

		return fs, appDir
	}

	t.Run("creates cmdtree only", func(t *testing.T) {
		fs, appDir := setupMinimalProject(t)

		var buf bytes.Buffer

		err := RunCobraAddTools(&buf, fs, appDir, AddToolsOptions{}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraAddTools() error = %v", err)
		}

		if _, err := fs.Stat(appDir + "/cmd/cmdtree.go"); err != nil {
			t.Error("cmd/cmdtree.go should be created")
		}

		if _, err := fs.Stat(appDir + "/cmd/aicontext.go"); err == nil {
			t.Error("cmd/aicontext.go should NOT be created without AIContext flag")
		}
	})

	t.Run("creates both with AIContext", func(t *testing.T) {
		fs, appDir := setupMinimalProject(t)

		var buf bytes.Buffer

		err := RunCobraAddTools(&buf, fs, appDir, AddToolsOptions{AIContext: true}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraAddTools() error = %v", err)
		}

		if _, err := fs.Stat(appDir + "/cmd/cmdtree.go"); err != nil {
			t.Error("cmd/cmdtree.go should be created")
		}

		if _, err := fs.Stat(appDir + "/cmd/aicontext.go"); err != nil {
			t.Error("cmd/aicontext.go should be created with AIContext flag")
		}
	})

	t.Run("error when no cmd directory", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		var buf bytes.Buffer

		err := RunCobraAddTools(&buf, fs, "/empty", AddToolsOptions{}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error when cmd/ directory doesn't exist")
		}
	})

	t.Run("error when cmdtree already exists", func(t *testing.T) {
		fs, appDir := setupMinimalProject(t)

		// Create existing cmdtree.go
		_ = afero.WriteFile(fs, appDir+"/cmd/cmdtree.go", []byte("package cmd"), 0644)

		var buf bytes.Buffer

		err := RunCobraAddTools(&buf, fs, appDir, AddToolsOptions{}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error when cmdtree.go already exists")
		}
	})
}

func TestParseModuleName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "module github.com/test/app\n\ngo 1.21\n", "github.com/test/app"},
		{"with spaces", "module  github.com/test/app \n", "github.com/test/app"},
		{"empty", "", ""},
		{"no module", "go 1.21\n", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseModuleName([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("parseModuleName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEditorConfigContent(t *testing.T) {
	fs := afero.NewMemMapFs()
	var buf bytes.Buffer

	_ = RunCobraInit(&buf, fs, "/editortest", CobraInitOptions{
		Module:  "github.com/test/editortest",
		AppName: "editortest",
	}, scaffolding.Options{})

	content, _ := afero.ReadFile(fs, "/editortest/.editorconfig")
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
		fs := afero.NewMemMapFs()

		configPath := "/tmp/.cobra.yaml"

		cfg := &CobraConfig{
			Author:     "Test Author <test@example.com>",
			License:    "MIT",
			UseViper:   true,
			UseService: false,
			Full:       true,
		}

		err := WriteDefaultConfig(fs, configPath, cfg)
		if err != nil {
			t.Fatalf("WriteDefaultConfig() error = %v", err)
		}

		// Verify file was created
		if _, err := fs.Stat(configPath); err != nil {
			t.Fatal("Config file was not created")
		}

		// Read it back
		content, err := afero.ReadFile(fs, configPath)
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
		cfg, path, err := LoadCobraConfig()
		if err != nil {
			t.Logf("LoadCobraConfig returned error (may have existing config): %v", err)
		}

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
