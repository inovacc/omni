package task

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseTaskfile(t *testing.T) {
	// Create a temporary taskfile
	tmpDir, err := os.MkdirTemp("", "task_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfileContent := `
version: '3'

vars:
  BUILD_DIR: ./build
  VERSION: "1.0.0"

env:
  GO_ENV: production

tasks:
  default:
    deps: [build]
    desc: Default task

  build:
    desc: Build the project
    cmds:
      - omni mkdir -p {{.BUILD_DIR}}
      - omni echo "Building version {{.VERSION}}"

  test:
    desc: Run tests
    deps: [build]
    cmds:
      - omni echo "Running tests"

  clean:
    desc: Clean build artifacts
    internal: true
    cmds:
      - omni rm -rf {{.BUILD_DIR}}
`

	taskfilePath := filepath.Join(tmpDir, "Taskfile.yml")
	if err := os.WriteFile(taskfilePath, []byte(taskfileContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Parse the taskfile
	tf, err := ParseTaskfile(taskfilePath)
	if err != nil {
		t.Fatalf("ParseTaskfile() error = %v", err)
	}

	// Verify version
	if tf.Version != "3" {
		t.Errorf("Version = %q, want %q", tf.Version, "3")
	}

	// Verify vars
	if tf.Vars["BUILD_DIR"] != "./build" {
		t.Errorf("Vars[BUILD_DIR] = %v, want %v", tf.Vars["BUILD_DIR"], "./build")
	}

	// Verify env
	if tf.Env["GO_ENV"] != "production" {
		t.Errorf("Env[GO_ENV] = %v, want %v", tf.Env["GO_ENV"], "production")
	}

	// Verify tasks
	if len(tf.Tasks) != 4 {
		t.Errorf("len(Tasks) = %d, want 4", len(tf.Tasks))
	}

	// Verify default task
	defaultTask := tf.GetTask("default")
	if defaultTask == nil {
		t.Error("GetTask(default) = nil")
	} else if len(defaultTask.Deps) != 1 || defaultTask.Deps[0].Task != "build" {
		t.Errorf("default.Deps = %v, want [build]", defaultTask.Deps)
	}

	// Verify build task
	buildTask := tf.GetTask("build")
	if buildTask == nil {
		t.Error("GetTask(build) = nil")
	} else if len(buildTask.Cmds) != 2 {
		t.Errorf("build.Cmds = %d, want 2", len(buildTask.Cmds))
	}

	// Verify internal task is hidden from list
	names := tf.ListTaskNames()
	for _, name := range names {
		if name == "clean" {
			t.Error("ListTaskNames() should not include internal task 'clean'")
		}
	}
}

func TestVarResolver(t *testing.T) {
	globalVars := map[string]any{
		"NAME":    "test",
		"VERSION": "1.0",
		"COUNT":   42,
	}
	taskVars := map[string]any{
		"VERSION": "2.0", // Override
		"LOCAL":   "local",
	}
	envVars := map[string]string{
		"ENV_VAR": "env_value",
	}

	resolver := NewVarResolver(globalVars, taskVars, envVars)

	tests := []struct {
		input    string
		expected string
	}{
		{"{{.NAME}}", "test"},
		{"{{.VERSION}}", "2.0"}, // Task var overrides global
		{"{{.LOCAL}}", "local"},
		{"{{.COUNT}}", "42"},
		{"$ENV_VAR", "env_value"},
		{"Building {{.NAME}} v{{.VERSION}}", "Building test v2.0"},
		{"{{.UNDEFINED}}", ""}, // Undefined vars expand to empty
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := resolver.Expand(tt.input)
			if result != tt.expected {
				t.Errorf("Expand(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDependencyResolver(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"a": {Deps: []Dependency{{Task: "b"}, {Task: "c"}}},
			"b": {Deps: []Dependency{{Task: "d"}}},
			"c": {Deps: []Dependency{{Task: "d"}}},
			"d": {},
		},
	}

	resolver := NewDependencyResolver(tf)

	// Test dependency resolution
	order, err := resolver.ResolveDeps("a")
	if err != nil {
		t.Fatalf("ResolveDeps(a) error = %v", err)
	}

	// d should come before b and c, which should come before a
	dIdx, bIdx, cIdx, aIdx := -1, -1, -1, -1

	for i, name := range order {
		switch name {
		case "a":
			aIdx = i
		case "b":
			bIdx = i
		case "c":
			cIdx = i
		case "d":
			dIdx = i
		}
	}

	if dIdx >= bIdx || dIdx >= cIdx {
		t.Errorf("d should come before b and c: order = %v", order)
	}

	if bIdx >= aIdx || cIdx >= aIdx {
		t.Errorf("b and c should come before a: order = %v", order)
	}
}

func TestDependencyResolverCycle(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"a": {Deps: []Dependency{{Task: "b"}}},
			"b": {Deps: []Dependency{{Task: "c"}}},
			"c": {Deps: []Dependency{{Task: "a"}}}, // Cycle!
		},
	}

	resolver := NewDependencyResolver(tf)

	_, err := resolver.ResolveDeps("a")
	if err == nil {
		t.Error("ResolveDeps() should fail on cyclic dependency")
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"omni echo hello", []string{"omni", "echo", "hello"}},
		{"omni mkdir -p ./build", []string{"omni", "mkdir", "-p", "./build"}},
		{`omni echo "hello world"`, []string{"omni", "echo", "hello world"}},
		{`omni echo 'single quoted'`, []string{"omni", "echo", "single quoted"}},
		{"  omni  ls  ", []string{"omni", "ls"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseCommand(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseCommand(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("parseCommand(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestIsOmniCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"omni echo hello", true},
		{"omni mkdir -p ./build", true},
		{"echo hello", true},        // Implicit omni command
		{"ls", true},                // Implicit omni command
		{"bash -c 'ls'", false},     // External shell
		{"/bin/ls", false},          // External path
		{"python script.py", false}, // External command
		{"git status", false},       // Git should use omni
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isOmniCommand(tt.input)
			if result != tt.expected {
				t.Errorf("isOmniCommand(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExecutorListTasks(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"build": {Desc: "Build the project"},
			"test":  {Desc: "Run tests"},
			"clean": {Desc: "Clean up", Internal: true},
		},
	}

	var buf bytes.Buffer

	exec := NewExecutor(&buf, tf, Options{})

	if err := exec.ListTasks(); err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "build") {
		t.Error("ListTasks() should include 'build'")
	}

	if !strings.Contains(output, "test") {
		t.Error("ListTasks() should include 'test'")
	}

	if strings.Contains(output, "clean") {
		t.Error("ListTasks() should not include internal task 'clean'")
	}
}

func TestExecutorDryRun(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"build": {
				Cmds: []Command{{Cmd: "omni echo hello"}},
			},
		},
	}

	var buf bytes.Buffer

	exec := NewExecutor(&buf, tf, Options{DryRun: true})

	// Use mock runner to avoid actual execution
	mock := NewMockCommandRunner()
	exec.SetCommandRunner(mock)

	ctx := context.Background()
	if err := exec.RunTask(ctx, "build"); err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}

	// In dry-run mode, no commands should be executed
	if len(mock.Commands) != 0 {
		t.Errorf("Dry-run should not execute commands, got %v", mock.Commands)
	}

	output := buf.String()
	if !strings.Contains(output, "[dry-run]") {
		t.Error("Dry-run output should contain '[dry-run]'")
	}
}

func TestExecutorWithMock(t *testing.T) {
	tf := &Taskfile{
		Vars: map[string]any{"NAME": "test"},
		Tasks: map[string]*Task{
			"greet": {
				Cmds: []Command{{Cmd: "omni echo Hello {{.NAME}}"}},
			},
		},
	}

	var buf bytes.Buffer

	exec := NewExecutor(&buf, tf, Options{})

	mock := NewMockCommandRunner()
	mock.SetOutput("echo", "Hello test\n")
	exec.SetCommandRunner(mock)

	ctx := context.Background()
	if err := exec.RunTask(ctx, "greet"); err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}

	// Verify command was executed with expanded variable
	if len(mock.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(mock.Commands))
	}

	cmd := mock.Commands[0]
	// Variable gets expanded, command gets parsed as two args: "Hello" and "test"
	if len(cmd) < 1 || cmd[0] != "echo" {
		t.Errorf("Unexpected command: %v", cmd)
	}
}

func TestRun(t *testing.T) {
	// Create a temporary taskfile
	tmpDir, err := os.MkdirTemp("", "task_run_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfileContent := `
version: '3'
tasks:
  default:
    cmds:
      - omni echo default task
  hello:
    desc: Say hello
    cmds:
      - omni echo hello
`

	taskfilePath := filepath.Join(tmpDir, "Taskfile.yml")
	if err := os.WriteFile(taskfilePath, []byte(taskfileContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original factory and restore after
	origFactory := CommandRunnerFactory

	defer func() { CommandRunnerFactory = origFactory }()

	mock := NewMockCommandRunner()
	CommandRunnerFactory = func() CommandRunner { return mock }

	ctx := context.Background()

	var buf bytes.Buffer

	t.Run("run default task", func(t *testing.T) {
		mock.Commands = nil

		buf.Reset()

		err := Run(ctx, &buf, nil, Options{Taskfile: taskfilePath})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if len(mock.Commands) != 1 {
			t.Errorf("Expected 1 command, got %d", len(mock.Commands))
		}
	})

	t.Run("run named task", func(t *testing.T) {
		mock.Commands = nil

		buf.Reset()

		err := Run(ctx, &buf, []string{"hello"}, Options{Taskfile: taskfilePath})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if len(mock.Commands) != 1 {
			t.Errorf("Expected 1 command, got %d", len(mock.Commands))
		}
	})

	t.Run("list tasks", func(t *testing.T) {
		buf.Reset()

		err := Run(ctx, &buf, nil, Options{Taskfile: taskfilePath, List: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "hello") {
			t.Errorf("List should include 'hello', got: %s", output)
		}
	})
}

func TestFindTaskfile(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "task_find_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("explicit path", func(t *testing.T) {
		taskfilePath := filepath.Join(tmpDir, "custom.yml")
		if err := os.WriteFile(taskfilePath, []byte("version: '3'\ntasks: {}"), 0644); err != nil {
			t.Fatal(err)
		}

		found, err := findTaskfile(taskfilePath, "")
		if err != nil {
			t.Fatalf("findTaskfile() error = %v", err)
		}

		if found != taskfilePath {
			t.Errorf("findTaskfile() = %q, want %q", found, taskfilePath)
		}
	})

	t.Run("explicit path not found", func(t *testing.T) {
		_, err := findTaskfile(filepath.Join(tmpDir, "nonexistent.yml"), "")
		if err == nil {
			t.Error("findTaskfile() should error for missing file")
		}
	})

	t.Run("search in directory", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "subdir")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		taskfilePath := filepath.Join(subDir, "Taskfile.yml")
		if err := os.WriteFile(taskfilePath, []byte("version: '3'\ntasks: {}"), 0644); err != nil {
			t.Fatal(err)
		}

		found, err := findTaskfile("", subDir)
		if err != nil {
			t.Fatalf("findTaskfile() error = %v", err)
		}

		if found != taskfilePath {
			t.Errorf("findTaskfile() = %q, want %q", found, taskfilePath)
		}
	})

	t.Run("no taskfile found", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		if err := os.MkdirAll(emptyDir, 0755); err != nil {
			t.Fatal(err)
		}

		_, err := findTaskfile("", emptyDir)
		if err == nil {
			t.Error("findTaskfile() should error when no taskfile found")
		}
	})
}

func TestShowSummary(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"build": {
				Desc:    "Build the project",
				Summary: "This task builds the project\nusing the Go compiler.",
				Deps:    []Dependency{{Task: "clean"}},
				Cmds: []Command{
					{Cmd: "omni echo Building..."},
					{Task: "test"},
				},
			},
			"clean": {Desc: "Clean build artifacts"},
			"test":  {Desc: "Run tests"},
		},
	}

	var buf bytes.Buffer

	exec := NewExecutor(&buf, tf, Options{})

	t.Run("show summary", func(t *testing.T) {
		buf.Reset()

		err := exec.ShowSummary([]string{"build"})
		if err != nil {
			t.Fatalf("ShowSummary() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Task: build") {
			t.Errorf("Should contain task name, got: %s", output)
		}

		if !strings.Contains(output, "Build the project") {
			t.Errorf("Should contain description, got: %s", output)
		}

		if !strings.Contains(output, "Go compiler") {
			t.Errorf("Should contain summary, got: %s", output)
		}

		if !strings.Contains(output, "Dependencies: clean") {
			t.Errorf("Should contain dependencies, got: %s", output)
		}

		if !strings.Contains(output, "task: test") {
			t.Errorf("Should contain task reference in commands, got: %s", output)
		}
	})

	t.Run("task not found", func(t *testing.T) {
		buf.Reset()

		err := exec.ShowSummary([]string{"nonexistent"})
		if err == nil {
			t.Error("ShowSummary() should error for unknown task")
		}
	})
}

func TestCheckStatus(t *testing.T) {
	tf := &Taskfile{
		Vars: map[string]any{},
		Tasks: map[string]*Task{
			"up-to-date": {
				Status: []string{"omni echo check"},
			},
		},
	}

	var buf bytes.Buffer

	exec := NewExecutor(&buf, tf, Options{})

	// Mock that succeeds (task is up-to-date)
	successMock := NewMockCommandRunner()
	exec.SetCommandRunner(successMock)

	ctx := context.Background()
	task := tf.Tasks["up-to-date"]

	t.Run("status check passes", func(t *testing.T) {
		upToDate, err := exec.checkStatus(ctx, task)
		if err != nil {
			t.Fatalf("checkStatus() error = %v", err)
		}

		if !upToDate {
			t.Error("checkStatus() should return true when status checks pass")
		}
	})

	t.Run("status check fails", func(t *testing.T) {
		failMock := NewMockCommandRunner()
		failMock.SetError("echo", errors.New("check failed"))
		exec.SetCommandRunner(failMock)

		upToDate, err := exec.checkStatus(ctx, task)
		if err != nil {
			t.Fatalf("checkStatus() error = %v", err)
		}

		if upToDate {
			t.Error("checkStatus() should return false when status check fails")
		}
	})

	t.Run("non-omni status command", func(t *testing.T) {
		badTask := &Task{
			Status: []string{"python check.py"},
		}

		_, err := exec.checkStatus(ctx, badTask)
		if err == nil {
			t.Error("checkStatus() should error for non-omni command")
		}
	})
}

func TestProcessIncludes(t *testing.T) {
	// Create temporary directory structure
	tmpDir, err := os.MkdirTemp("", "task_include_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create included taskfile
	includeDir := filepath.Join(tmpDir, "included")
	if err := os.MkdirAll(includeDir, 0755); err != nil {
		t.Fatal(err)
	}

	includedContent := `
version: '3'
vars:
  INCLUDED_VAR: included_value
  SHARED_VAR: from_included
tasks:
  build:
    desc: Build from included
    cmds:
      - omni echo included build
  test:
    desc: Test from included
`
	if err := os.WriteFile(filepath.Join(includeDir, "Taskfile.yml"), []byte(includedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create main taskfile with includes
	mainContent := `
version: '3'
vars:
  MAIN_VAR: main_value
  SHARED_VAR: from_main
includes:
  sub: ./included
tasks:
  default:
    desc: Main default task
`

	mainPath := filepath.Join(tmpDir, "Taskfile.yml")
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}

	tf, err := ParseTaskfile(mainPath)
	if err != nil {
		t.Fatalf("ParseTaskfile() error = %v", err)
	}

	t.Run("included tasks have namespace", func(t *testing.T) {
		if tf.GetTask("sub:build") == nil {
			t.Error("Should have 'sub:build' task")
		}

		if tf.GetTask("sub:test") == nil {
			t.Error("Should have 'sub:test' task")
		}
	})

	t.Run("main tasks remain", func(t *testing.T) {
		if tf.GetTask("default") == nil {
			t.Error("Should have 'default' task")
		}
	})

	t.Run("vars merged correctly", func(t *testing.T) {
		if tf.Vars["MAIN_VAR"] != "main_value" {
			t.Errorf("MAIN_VAR = %v, want main_value", tf.Vars["MAIN_VAR"])
		}

		if tf.Vars["INCLUDED_VAR"] != "included_value" {
			t.Errorf("INCLUDED_VAR = %v, want included_value", tf.Vars["INCLUDED_VAR"])
		}
		// Main vars take precedence
		if tf.Vars["SHARED_VAR"] != "from_main" {
			t.Errorf("SHARED_VAR = %v, want from_main (main should override)", tf.Vars["SHARED_VAR"])
		}
	})
}

func TestProcessIncludesErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "task_include_err_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("include not found", func(t *testing.T) {
		content := `
version: '3'
includes:
  missing: ./nonexistent
tasks: {}
`

		path := filepath.Join(tmpDir, "missing.yml")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := ParseTaskfile(path)
		if err == nil {
			t.Error("ParseTaskfile() should error for missing include")
		}
	})

	t.Run("include dir without taskfile", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty_include")
		if err := os.MkdirAll(emptyDir, 0755); err != nil {
			t.Fatal(err)
		}

		content := `
version: '3'
includes:
  empty: ./empty_include
tasks: {}
`

		path := filepath.Join(tmpDir, "empty_dir.yml")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := ParseTaskfile(path)
		if err == nil {
			t.Error("ParseTaskfile() should error for include dir without taskfile")
		}
	})
}

func TestExecutorDeferredCommands(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"with-defer": {
				Cmds: []Command{
					{Cmd: "omni echo step1"},
					{Cmd: "omni echo cleanup", Defer: true},
					{Cmd: "omni echo step2"},
				},
			},
		},
	}

	var buf bytes.Buffer

	exec := NewExecutor(&buf, tf, Options{})

	mock := NewMockCommandRunner()
	exec.SetCommandRunner(mock)

	ctx := context.Background()
	if err := exec.RunTask(ctx, "with-defer"); err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}

	// Should execute: step1, step2, cleanup (deferred runs at end)
	if len(mock.Commands) != 3 {
		t.Fatalf("Expected 3 commands, got %d: %v", len(mock.Commands), mock.Commands)
	}

	// Verify order: step1, step2, then cleanup
	if mock.Commands[0][1] != "step1" {
		t.Errorf("First command should be step1, got %v", mock.Commands[0])
	}

	if mock.Commands[1][1] != "step2" {
		t.Errorf("Second command should be step2, got %v", mock.Commands[1])
	}

	if mock.Commands[2][1] != "cleanup" {
		t.Errorf("Third command should be cleanup (deferred), got %v", mock.Commands[2])
	}
}

func TestExecutorTaskReference(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"main": {
				Cmds: []Command{
					{Cmd: "omni echo main-before"},
					{Task: "helper"},
					{Cmd: "omni echo main-after"},
				},
			},
			"helper": {
				Cmds: []Command{
					{Cmd: "omni echo helper"},
				},
			},
		},
	}

	var buf bytes.Buffer

	exec := NewExecutor(&buf, tf, Options{})

	mock := NewMockCommandRunner()
	exec.SetCommandRunner(mock)

	ctx := context.Background()
	if err := exec.RunTask(ctx, "main"); err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}

	if len(mock.Commands) != 3 {
		t.Fatalf("Expected 3 commands, got %d", len(mock.Commands))
	}
}

func TestExecutorIgnoreError(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"ignore-error": {
				Cmds: []Command{
					{Cmd: "omni cat /nonexistent", IgnoreError: true},
					{Cmd: "omni echo continue"},
				},
			},
		},
	}

	var buf bytes.Buffer

	exec := NewExecutor(&buf, tf, Options{})

	mock := NewMockCommandRunner()
	mock.SetError("cat", errors.New("file not found"))
	exec.SetCommandRunner(mock)

	ctx := context.Background()
	// Should not return error because ignore_error is set on the failing command
	err := exec.RunTask(ctx, "ignore-error")
	if err != nil {
		t.Fatalf("RunTask() should ignore error, got: %v", err)
	}

	// Both commands should have been attempted
	if len(mock.Commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(mock.Commands))
	}
}

func TestExecutorVerbose(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"verbose-test": {
				Cmds: []Command{{Cmd: "omni echo hello"}},
			},
		},
	}

	var buf bytes.Buffer

	exec := NewExecutor(&buf, tf, Options{Verbose: true})

	mock := NewMockCommandRunner()
	exec.SetCommandRunner(mock)

	ctx := context.Background()
	if err := exec.RunTask(ctx, "verbose-test"); err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "$ omni echo hello") {
		t.Errorf("Verbose output should show command, got: %s", output)
	}
}

func TestExecutorSilent(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"silent-test": {
				Silent: true,
				Cmds:   []Command{{Cmd: "omni echo hello"}},
			},
		},
	}

	var buf bytes.Buffer

	exec := NewExecutor(&buf, tf, Options{})

	mock := NewMockCommandRunner()
	exec.SetCommandRunner(mock)

	ctx := context.Background()
	if err := exec.RunTask(ctx, "silent-test"); err != nil {
		t.Fatalf("RunTask() error = %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "task: silent-test") {
		t.Errorf("Silent task should not print task name, got: %s", output)
	}
}

func TestTaskAliases(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"build": {
				Desc:    "Build the project",
				Aliases: []string{"b", "compile"},
			},
		},
	}

	t.Run("get by name", func(t *testing.T) {
		task := tf.GetTask("build")
		if task == nil {
			t.Error("GetTask(build) should return task")
		}
	})

	t.Run("get by alias", func(t *testing.T) {
		task := tf.GetTask("b")
		if task == nil {
			t.Error("GetTask(b) should return task via alias")
		}
	})

	t.Run("get by second alias", func(t *testing.T) {
		task := tf.GetTask("compile")
		if task == nil {
			t.Error("GetTask(compile) should return task via alias")
		}
	})

	t.Run("unknown task", func(t *testing.T) {
		task := tf.GetTask("unknown")
		if task != nil {
			t.Error("GetTask(unknown) should return nil")
		}
	})
}

func TestCommandUnmarshalYAML(t *testing.T) {
	// Test string shorthand
	yamlStr := `
- omni echo hello
- cmd: omni echo world
  silent: true
- task: build
`

	var cmds []Command
	if err := yaml.Unmarshal([]byte(yamlStr), &cmds); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(cmds) != 3 {
		t.Fatalf("Expected 3 commands, got %d", len(cmds))
	}

	if cmds[0].Cmd != "omni echo hello" {
		t.Errorf("cmds[0].Cmd = %q, want 'omni echo hello'", cmds[0].Cmd)
	}

	if cmds[1].Cmd != "omni echo world" || !cmds[1].Silent {
		t.Errorf("cmds[1] = %+v, want cmd='omni echo world', silent=true", cmds[1])
	}

	if cmds[2].Task != "build" {
		t.Errorf("cmds[2].Task = %q, want 'build'", cmds[2].Task)
	}
}

func TestDependencyUnmarshalYAML(t *testing.T) {
	yamlStr := `
- build
- task: test
  vars:
    FAST: true
`

	var deps []Dependency
	if err := yaml.Unmarshal([]byte(yamlStr), &deps); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if len(deps) != 2 {
		t.Fatalf("Expected 2 dependencies, got %d", len(deps))
	}

	if deps[0].Task != "build" {
		t.Errorf("deps[0].Task = %q, want 'build'", deps[0].Task)
	}

	if deps[1].Task != "test" {
		t.Errorf("deps[1].Task = %q, want 'test'", deps[1].Task)
	}

	if deps[1].Vars["FAST"] != true {
		t.Errorf("deps[1].Vars[FAST] = %v, want true", deps[1].Vars["FAST"])
	}
}

func TestRunNoDefaultTask(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "task_nodefault_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	content := `
version: '3'
tasks:
  build:
    cmds:
      - omni echo build
`

	path := filepath.Join(tmpDir, "Taskfile.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	var buf bytes.Buffer

	err = Run(ctx, &buf, nil, Options{Taskfile: path})
	if err == nil {
		t.Error("Run() should error when no task specified and no default task")
	}

	if !strings.Contains(err.Error(), "no task specified") {
		t.Errorf("Error should mention 'no task specified', got: %v", err)
	}
}

func TestRunSummaryMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "task_summary_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	content := `
version: '3'
tasks:
  build:
    desc: Build the project
    summary: |
      This is a longer summary
      with multiple lines.
    cmds:
      - omni echo build
`

	path := filepath.Join(tmpDir, "Taskfile.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	var buf bytes.Buffer

	err = Run(ctx, &buf, []string{"build"}, Options{Taskfile: path, Summary: true})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Task: build") {
		t.Errorf("Summary should contain task name, got: %s", output)
	}

	if !strings.Contains(output, "longer summary") {
		t.Errorf("Summary should contain summary text, got: %s", output)
	}
}

func TestExecutorForceRun(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"check": {
				Status: []string{"omni echo status-check"},
				Cmds:   []Command{{Cmd: "omni echo running"}},
			},
		},
	}

	ctx := context.Background()

	t.Run("with status check (up-to-date)", func(t *testing.T) {
		var buf bytes.Buffer

		exec := NewExecutor(&buf, tf, Options{Verbose: true})
		mock := NewMockCommandRunner()
		exec.SetCommandRunner(mock)

		if err := exec.RunTask(ctx, "check"); err != nil {
			t.Fatalf("RunTask() error = %v", err)
		}

		// Status check passes, so main command shouldn't run
		output := buf.String()
		if !strings.Contains(output, "up to date") {
			t.Errorf("Should show 'up to date' message, got: %s", output)
		}
	})

	t.Run("with force flag", func(t *testing.T) {
		var buf bytes.Buffer

		exec := NewExecutor(&buf, tf, Options{Force: true})
		mock := NewMockCommandRunner()
		exec.SetCommandRunner(mock)

		if err := exec.RunTask(ctx, "check"); err != nil {
			t.Fatalf("RunTask() error = %v", err)
		}

		// With force, should run the task even if up-to-date
		found := false

		for _, cmd := range mock.Commands {
			if len(cmd) > 1 && cmd[1] == "running" {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Force should run task, commands: %v", mock.Commands)
		}
	})
}

func TestVarResolverEnvExpansion(t *testing.T) {
	// Set environment variable for test
	os.Setenv("TEST_VAR_123", "env_value_123")

	defer os.Unsetenv("TEST_VAR_123")

	resolver := NewVarResolver(nil, nil, nil)

	result := resolver.Expand("Value is $TEST_VAR_123")
	if result != "Value is env_value_123" {
		t.Errorf("Expand() = %q, want 'Value is env_value_123'", result)
	}
}

func TestGetDirectDeps(t *testing.T) {
	tf := &Taskfile{
		Tasks: map[string]*Task{
			"a": {Deps: []Dependency{{Task: "b"}, {Task: "c"}}},
			"b": {},
			"c": {},
		},
	}

	resolver := NewDependencyResolver(tf)

	t.Run("existing task", func(t *testing.T) {
		deps, err := resolver.GetDirectDeps("a")
		if err != nil {
			t.Fatalf("GetDirectDeps() error = %v", err)
		}

		if len(deps) != 2 {
			t.Fatalf("GetDirectDeps(a) = %d deps, want 2", len(deps))
		}

		depMap := make(map[string]bool)
		for _, d := range deps {
			depMap[d.Task] = true
		}

		if !depMap["b"] || !depMap["c"] {
			t.Errorf("GetDirectDeps(a) = %v, want [b, c]", deps)
		}
	})

	t.Run("unknown task", func(t *testing.T) {
		_, err := resolver.GetDirectDeps("unknown")
		if err == nil {
			t.Error("GetDirectDeps() should error for unknown task")
		}
	})
}
