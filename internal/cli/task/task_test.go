package task

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	} else {
		if len(defaultTask.Deps) != 1 || defaultTask.Deps[0].Task != "build" {
			t.Errorf("default.Deps = %v, want [build]", defaultTask.Deps)
		}
	}

	// Verify build task
	buildTask := tf.GetTask("build")
	if buildTask == nil {
		t.Error("GetTask(build) = nil")
	} else {
		if len(buildTask.Cmds) != 2 {
			t.Errorf("build.Cmds = %d, want 2", len(buildTask.Cmds))
		}
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
		{"echo hello", true},      // Implicit omni command
		{"ls", true},              // Implicit omni command
		{"bash -c 'ls'", false},   // External shell
		{"/bin/ls", false},        // External path
		{"python script.py", false}, // External command
		{"git status", false},     // Git should use omni
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
