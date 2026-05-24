package cobra

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"go/token"
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
			"cmd/myapp/myapp.go",
			"go.mod",
			"cmd/myapp/cmd_version.go",
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

		// Verify the entry point file was created with the extracted app name
		if _, err := fs.Stat("/extracted/cmd/extractedapp/extractedapp.go"); err != nil {
			t.Error("entry point should be created at cmd/extractedapp/extractedapp.go")
		}

		mainContent, _ := afero.ReadFile(fs, "/extracted/cmd/extractedapp/extractedapp.go")
		if !strings.Contains(string(mainContent), "extractedapp") {
			t.Error("entry point should contain the extracted app name")
		}
	})

	t.Run("entry point content", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		_ = RunCobraInit(&buf, fs, "/maintest", CobraInitOptions{
			Module: "github.com/test/maintest",
		}, scaffolding.Options{})

		content, _ := afero.ReadFile(fs, "/maintest/cmd/maintest/maintest.go")
		mainStr := string(content)

		if !strings.Contains(mainStr, "package main") {
			t.Error("entry point should have package main")
		}

		if !strings.Contains(mainStr, "func main()") {
			t.Error("entry point should have func main()")
		}
	})

	t.Run("entry point root content", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		_ = RunCobraInit(&buf, fs, "/roottest", CobraInitOptions{
			Module:      "github.com/test/roottest",
			AppName:     "roottest",
			Description: "Test description",
		}, scaffolding.Options{})

		content, _ := afero.ReadFile(fs, "/roottest/cmd/roottest/roottest.go")
		rootStr := string(content)

		if !strings.Contains(rootStr, "package main") {
			t.Error("entry point should have package main")
		}

		if !strings.Contains(rootStr, "roottest") {
			t.Error("entry point should contain app name")
		}

		if !strings.Contains(rootStr, "Test description") {
			t.Error("entry point should contain description")
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
		if _, err := fs.Stat("/addtest/cmd/addtest/cmd_serve.go"); err != nil {
			t.Error("cmd_serve.go should be created")
		}

		content, _ := afero.ReadFile(fs, "/addtest/cmd/addtest/cmd_serve.go")
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

		content, _ := afero.ReadFile(fs, "/addtest/cmd/addtest/cmd_config.go")
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

		content, _ := afero.ReadFile(fs, "/addtest/cmd/addtest/cmd_list.go")
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

	t.Run("platform-split emits shared + three platform files", func(t *testing.T) {
		fs, appDir := setupProject(t)

		var buf bytes.Buffer

		err := RunCobraAdd(&buf, fs, appDir, CobraAddOptions{
			Name:          "daemon",
			Parent:        "root",
			Description:   "Run as daemon",
			PlatformSplit: true,
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraAdd() error = %v", err)
		}

		want := map[string]string{
			"/addtest/cmd/addtest/cmd_daemon.go":         "daemonCmd",
			"/addtest/cmd/addtest/cmd_daemon_windows.go": "//go:build windows",
			"/addtest/cmd/addtest/cmd_daemon_darwin.go":  "//go:build darwin",
			"/addtest/cmd/addtest/cmd_daemon_unix.go":    "//go:build unix && !darwin",
		}
		for path, marker := range want {
			content, err := afero.ReadFile(fs, path)
			if err != nil {
				t.Fatalf("expected file %s: %v", path, err)
			}
			if !strings.Contains(string(content), marker) {
				t.Errorf("%s missing marker %q", path, marker)
			}
		}

		// Shared file must delegate to runDaemon, not inline Println.
		shared, _ := afero.ReadFile(fs, "/addtest/cmd/addtest/cmd_daemon.go")
		if !strings.Contains(string(shared), "runDaemon(cmd, args)") {
			t.Error("shared file should delegate to runDaemon")
		}

		// Each platform file must define runDaemon so exactly one is compiled per OS.
		for _, p := range []string{"cmd_daemon_windows.go", "cmd_daemon_darwin.go", "cmd_daemon_unix.go"} {
			content, _ := afero.ReadFile(fs, "/addtest/cmd/addtest/"+p)
			if !strings.Contains(string(content), "func runDaemon(") {
				t.Errorf("%s should define runDaemon", p)
			}
		}
	})

	t.Run("platform-split conflict if any target file exists", func(t *testing.T) {
		fs, appDir := setupProject(t)

		// Pre-create one of the platform files so the pre-flight should reject.
		_ = afero.WriteFile(fs, "/addtest/cmd/addtest/cmd_daemon_darwin.go", []byte("// pre-existing\n"), 0644)

		var buf bytes.Buffer
		err := RunCobraAdd(&buf, fs, appDir, CobraAddOptions{
			Name:          "daemon",
			PlatformSplit: true,
		}, scaffolding.Options{})
		if err == nil {
			t.Error("expected conflict error when a platform file already exists")
		}
	})

	t.Run("not a cobra project", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		// Write go.mod so it passes that check, but no cmd/{appName} dir
		_ = afero.WriteFile(fs, "/empty/go.mod", []byte("module github.com/test/empty\n\ngo 1.21\n"), 0644)

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
			"cmd/fullapp/fullapp.go",
			"go.mod",
			"cmd/fullapp/cmd_version.go",
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

		content, _ := afero.ReadFile(fs, "/fullverapp/cmd/fullverapp/cmd_version.go")
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
			"cmd/serviceapp/serviceapp.go",
			"go.mod",
			"cmd/serviceapp/cmd_version.go",
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

		content, _ := afero.ReadFile(fs, "/servicerootapp/cmd/servicerootapp/servicerootapp.go")
		rootStr := string(content)

		if !strings.Contains(rootStr, "inovacc/config") {
			t.Error("Service mode entry point should import inovacc/config")
		}

		// Behavior change: the entry point is now a command group; the
		// long-running handler is owned by the `service run` subcommand
		// (cmd_service.go), so the root file must NOT bind service.Handler
		// nor import internal/service.
		if strings.Contains(rootStr, "RunE: service.Handler") {
			t.Error("Service mode entry point must not bind rootCmd.RunE; service run owns it")
		}

		if !strings.Contains(rootStr, "internal/parameters") {
			t.Error("Service mode entry point should import parameters")
		}

		if strings.Contains(rootStr, `"github.com/test/servicerootapp/internal/service"`) {
			t.Error("Service mode entry point should no longer import internal/service")
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

	// cmd_cmdtree.go should always be created
	if _, err := fs.Stat("/cmdtreeapp/cmd/cmdtreeapp/cmd_cmdtree.go"); err != nil {
		t.Error("cmd/cmdtreeapp/cmd_cmdtree.go should always be created")
	}

	content, _ := afero.ReadFile(fs, "/cmdtreeapp/cmd/cmdtreeapp/cmd_cmdtree.go")
	if !strings.Contains(string(content), "cmdtreeCmd") {
		t.Error("cmdtree.go should define cmdtreeCmd")
	}

	if !strings.Contains(string(content), "collectPersistentFlags") {
		t.Error("cmdtree.go should define collectPersistentFlags")
	}

	// cmd_aicontext.go should NOT be created by default
	if _, err := fs.Stat("/cmdtreeapp/cmd/cmdtreeapp/cmd_aicontext.go"); err == nil {
		t.Error("cmd/cmdtreeapp/cmd_aicontext.go should NOT be created by default")
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
	if _, err := fs.Stat("/aiapp/cmd/aiapp/cmd_cmdtree.go"); err != nil {
		t.Error("cmd/aiapp/cmd_cmdtree.go should be created")
	}

	if _, err := fs.Stat("/aiapp/cmd/aiapp/cmd_aicontext.go"); err != nil {
		t.Error("cmd/aiapp/cmd_aicontext.go should be created when AIContext=true")
	}

	content, _ := afero.ReadFile(fs, "/aiapp/cmd/aiapp/cmd_aicontext.go")
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
		if err := fs.MkdirAll(appDir+"/cmd/toolsapp", 0755); err != nil {
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

		if _, err := fs.Stat(appDir + "/cmd/toolsapp/cmd_cmdtree.go"); err != nil {
			t.Error("cmd/toolsapp/cmd_cmdtree.go should be created")
		}

		if _, err := fs.Stat(appDir + "/cmd/toolsapp/cmd_aicontext.go"); err == nil {
			t.Error("cmd/toolsapp/cmd_aicontext.go should NOT be created without AIContext flag")
		}
	})

	t.Run("creates both with AIContext", func(t *testing.T) {
		fs, appDir := setupMinimalProject(t)

		var buf bytes.Buffer

		err := RunCobraAddTools(&buf, fs, appDir, AddToolsOptions{AIContext: true}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraAddTools() error = %v", err)
		}

		if _, err := fs.Stat(appDir + "/cmd/toolsapp/cmd_cmdtree.go"); err != nil {
			t.Error("cmd/toolsapp/cmd_cmdtree.go should be created")
		}

		if _, err := fs.Stat(appDir + "/cmd/toolsapp/cmd_aicontext.go"); err != nil {
			t.Error("cmd/toolsapp/cmd_aicontext.go should be created with AIContext flag")
		}
	})

	t.Run("error when no cmd directory", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		// Write go.mod so it passes that check, but no cmd/{appName} dir
		_ = afero.WriteFile(fs, "/empty/go.mod", []byte("module github.com/test/empty\n\ngo 1.21\n"), 0644)

		var buf bytes.Buffer

		err := RunCobraAddTools(&buf, fs, "/empty", AddToolsOptions{}, scaffolding.Options{})
		if err == nil {
			t.Error("Expected error when cmd/ directory doesn't exist")
		}
	})

	t.Run("error when cmdtree already exists", func(t *testing.T) {
		fs, appDir := setupMinimalProject(t)

		// Create existing cmd_cmdtree.go
		_ = afero.WriteFile(fs, appDir+"/cmd/toolsapp/cmd_cmdtree.go", []byte("package main"), 0644)

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

func TestServiceCommandMode(t *testing.T) {
	fs := afero.NewMemMapFs()
	var buf bytes.Buffer

	err := RunCobraInit(&buf, fs, "/svcapp", CobraInitOptions{
		Module:     "example.com/svcapp",
		UseService: true,
	}, scaffolding.Options{})
	if err != nil {
		t.Fatalf("RunCobraInit: %v", err)
	}

	svcFile := "/svcapp/cmd/svcapp/cmd_service.go"
	b, err := afero.ReadFile(fs, svcFile)
	if err != nil {
		t.Fatalf("expected %s: %v", svcFile, err)
	}
	src := string(b)
	for _, want := range []string{
		"serviceCmd", "serviceProgram",
		"func (p *serviceProgram) Start", "func (p *serviceProgram) Stop",
		"kardianos/service",
		`Use:   "install"`, `Use:   "uninstall"`, `Use:   "start"`,
		`Use:   "stop"`, `Use:   "restart"`, `Use:   "status"`, `Use:   "run"`,
		"rootCmd.AddCommand(serviceCmd)",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("cmd_service.go missing %q", want)
		}
	}

	gomod, _ := afero.ReadFile(fs, "/svcapp/go.mod")
	if !strings.Contains(string(gomod), "github.com/kardianos/service") {
		t.Errorf("go.mod missing kardianos/service require")
	}

	mainB, _ := afero.ReadFile(fs, "/svcapp/cmd/svcapp/svcapp.go")
	if strings.Contains(string(mainB), "RunE: service.Handler") {
		t.Errorf("service mode must not hard-bind rootCmd.RunE; run owns it")
	}
}

func TestRunCobraInitDaemon(t *testing.T) {
	t.Run("emits all daemon files and they parse as Go", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/daemonapp", CobraInitOptions{
			Module:    "github.com/test/daemonapp",
			AppName:   "daemonapp",
			UseDaemon: true,
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraInit(--daemon) error = %v", err)
		}

		want := []string{
			"/daemonapp/internal/serverinfo/serverinfo.go",
			"/daemonapp/cmd/daemonapp/cmd_server.go",
			"/daemonapp/cmd/daemonapp/server.go",
			"/daemonapp/cmd/daemonapp/server_unix.go",
			"/daemonapp/cmd/daemonapp/server_systemd.go",
			"/daemonapp/cmd/daemonapp/server_darwin.go",
			"/daemonapp/cmd/daemonapp/server_windows.go",
		}
		fset := token.NewFileSet()
		for _, p := range want {
			content, err := afero.ReadFile(fs, p)
			if err != nil {
				t.Fatalf("expected %s: %v", p, err)
			}
			// Each generated file must parse as valid Go.
			if _, err := parser.ParseFile(fset, p, content, parser.SkipObjectResolution); err != nil {
				t.Errorf("%s did not parse: %v", p, err)
			}
		}

		// Spot-check semantics that templates must preserve.
		serverGo, _ := afero.ReadFile(fs, "/daemonapp/cmd/daemonapp/server.go")
		if !strings.Contains(string(serverGo), `daemonEnvVar = "DAEMONAPP_DAEMON_CHILD"`) {
			t.Error("server.go should use uppercased AppName for daemon env var")
		}
		if !strings.Contains(string(serverGo), `"github.com/test/daemonapp/internal/serverinfo"`) {
			t.Error("server.go should import the generated serverinfo package")
		}

		// Build tags must be exact — wrong tags cause silent platform breakage.
		cases := map[string]string{
			"/daemonapp/cmd/daemonapp/server_unix.go":    "//go:build !windows",
			"/daemonapp/cmd/daemonapp/server_systemd.go": "//go:build !windows && !darwin",
			"/daemonapp/cmd/daemonapp/server_darwin.go":  "//go:build darwin",
			"/daemonapp/cmd/daemonapp/server_windows.go": "//go:build windows",
		}
		for path, tag := range cases {
			b, _ := afero.ReadFile(fs, path)
			if !strings.Contains(string(b), tag) {
				t.Errorf("%s missing %q build tag", path, tag)
			}
		}

		// go.mod must list the daemon-only deps so `go mod tidy` resolves cleanly.
		goMod, _ := afero.ReadFile(fs, "/daemonapp/go.mod")
		for _, dep := range []string{"github.com/shirou/gopsutil/v3", "golang.org/x/sys"} {
			if !strings.Contains(string(goMod), dep) {
				t.Errorf("go.mod missing daemon dep %s", dep)
			}
		}
	})

	t.Run("--service and --daemon are mutually exclusive", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/clash", CobraInitOptions{
			Module:     "github.com/test/clash",
			UseService: true,
			UseDaemon:  true,
		}, scaffolding.Options{})
		if err == nil {
			t.Error("expected error when both --service and --daemon are set")
		}
	})
}
