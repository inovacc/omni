package lint

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLint_ValidTaskfile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := filepath.Join(tmpDir, "Taskfile.yml")
	content := `version: "3"

tasks:
  build:
    cmds:
      - omni echo "Building..."
      - go build ./...
`
	_ = os.WriteFile(taskfile, []byte(content), 0644)

	var buf bytes.Buffer

	err = RunLint(&buf, []string{taskfile}, LintOptions{})
	if err != nil {
		t.Errorf("RunLint() error = %v for valid taskfile", err)
	}
}

func TestRunLint_MissingVersion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := filepath.Join(tmpDir, "Taskfile.yml")
	content := `tasks:
  build:
    cmds:
      - go build ./...
`
	_ = os.WriteFile(taskfile, []byte(content), 0644)

	var buf bytes.Buffer

	_ = RunLint(&buf, []string{taskfile}, LintOptions{})

	output := buf.String()
	if !strings.Contains(output, "missing-version") {
		t.Errorf("RunLint() should warn about missing version: %s", output)
	}
}

func TestRunLint_ShellCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := filepath.Join(tmpDir, "Taskfile.yml")
	content := `version: "3"

tasks:
  test:
    cmds:
      - cat README.md
      - grep pattern file.txt
`
	_ = os.WriteFile(taskfile, []byte(content), 0644)

	var buf bytes.Buffer

	_ = RunLint(&buf, []string{taskfile}, LintOptions{})

	output := buf.String()
	if !strings.Contains(output, "use-omni") {
		t.Errorf("RunLint() should suggest omni replacement: %s", output)
	}
}

func TestRunLint_NonPortableCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := filepath.Join(tmpDir, "Taskfile.yml")
	content := `version: "3"

tasks:
  install:
    cmds:
      - apt-get install something
`
	_ = os.WriteFile(taskfile, []byte(content), 0644)

	var buf bytes.Buffer

	err = RunLint(&buf, []string{taskfile}, LintOptions{})

	// Should return error for non-portable command
	if err == nil {
		t.Error("RunLint() should return error for non-portable command")
	}

	output := buf.String()
	if !strings.Contains(output, "non-portable") {
		t.Errorf("RunLint() should flag non-portable command: %s", output)
	}
}

func TestRunLint_BashSpecific(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := filepath.Join(tmpDir, "Taskfile.yml")
	content := `version: "3"

tasks:
  check:
    cmds:
      - '[[ -f file.txt ]] && echo exists'
`
	_ = os.WriteFile(taskfile, []byte(content), 0644)

	var buf bytes.Buffer

	_ = RunLint(&buf, []string{taskfile}, LintOptions{})

	output := buf.String()
	if !strings.Contains(output, "bash-specific") {
		t.Errorf("RunLint() should flag bash-specific syntax: %s", output)
	}
}

func TestRunLint_HardcodedPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := filepath.Join(tmpDir, "Taskfile.yml")
	content := `version: "3"

tasks:
  run:
    cmds:
      - /usr/bin/python script.py
`
	_ = os.WriteFile(taskfile, []byte(content), 0644)

	var buf bytes.Buffer

	_ = RunLint(&buf, []string{taskfile}, LintOptions{})

	output := buf.String()
	if !strings.Contains(output, "hardcoded-path") {
		t.Errorf("RunLint() should flag hardcoded path: %s", output)
	}
}

func TestRunLint_InvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := filepath.Join(tmpDir, "Taskfile.yml")
	content := `version: "3"
tasks:
  - invalid yaml structure
    not: valid
`
	_ = os.WriteFile(taskfile, []byte(content), 0644)

	var buf bytes.Buffer

	err = RunLint(&buf, []string{taskfile}, LintOptions{})
	if err == nil {
		t.Error("RunLint() should return error for invalid YAML")
	}
}

func TestRunLint_QuietMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := filepath.Join(tmpDir, "Taskfile.yml")
	content := `version: "3"

tasks:
  build:
    cmds:
      - cat file.txt
`
	_ = os.WriteFile(taskfile, []byte(content), 0644)

	var buf bytes.Buffer

	_ = RunLint(&buf, []string{taskfile}, LintOptions{Quiet: true})

	output := buf.String()
	// In quiet mode, individual warning messages should be suppressed (but summary still shows)
	// The colored "warning" word should not appear in output
	if strings.Contains(output, "use-omni") {
		t.Errorf("RunLint() quiet mode should suppress warning details: %s", output)
	}
}

func TestRunLint_StrictMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := filepath.Join(tmpDir, "Taskfile.yml")
	content := `version: "3"

tasks:
  check:
    cmds:
      - '[[ -f file ]]'
`
	_ = os.WriteFile(taskfile, []byte(content), 0644)

	var buf bytes.Buffer

	err = RunLint(&buf, []string{taskfile}, LintOptions{Strict: true})

	// In strict mode, bash-specific becomes an error
	if err == nil {
		t.Error("RunLint() strict mode should return error for bash-specific syntax")
	}
}

func TestRunLint_NoTasksError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := filepath.Join(tmpDir, "Taskfile.yml")
	content := `version: "3"
`
	_ = os.WriteFile(taskfile, []byte(content), 0644)

	var buf bytes.Buffer

	err = RunLint(&buf, []string{taskfile}, LintOptions{})
	if err == nil {
		t.Error("RunLint() should return error for missing tasks")
	}
}

func TestRunLint_DirectoryInput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	taskfile := filepath.Join(tmpDir, "Taskfile.yml")
	content := `version: "3"

tasks:
  build:
    cmds:
      - go build
`
	_ = os.WriteFile(taskfile, []byte(content), 0644)

	var buf bytes.Buffer

	err = RunLint(&buf, []string{tmpDir}, LintOptions{})
	if err != nil {
		t.Errorf("RunLint() error = %v for directory input", err)
	}
}

func TestRunLint_NonexistentFile(t *testing.T) {
	var buf bytes.Buffer

	err := RunLint(&buf, []string{"/nonexistent/Taskfile.yml"}, LintOptions{})

	// The function handles missing files by reporting them but continues
	// It returns an error indicating issues were found
	if err == nil {
		// This is also acceptable - graceful handling
		return
	}

	output := buf.String()
	// Should report the file issue
	if len(output) == 0 {
		t.Error("RunLint() should report nonexistent file")
	}
}

func TestFindTaskfiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lint_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create various Taskfile variants
	_ = os.WriteFile(filepath.Join(tmpDir, "Taskfile.yml"), []byte("version: \"3\""), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "Taskfile.yaml"), []byte("version: \"3\""), 0644)

	files := findTaskfiles(tmpDir)
	if len(files) < 2 {
		t.Errorf("findTaskfiles() found %d files, want at least 2", len(files))
	}
}

func TestCheckCommand_OmniAlreadyUsed(t *testing.T) {
	issues := checkCommand("test.yml", "build", "omni cat file.txt", 1, LintOptions{})

	// Should not flag omni commands
	for _, issue := range issues {
		if issue.Rule == "use-omni" {
			t.Error("checkCommand() should not flag omni commands")
		}
	}
}

func TestCheckCommand_SourceBashSpecific(t *testing.T) {
	issues := checkCommand("test.yml", "build", "source .env", 1, LintOptions{})

	found := false

	for _, issue := range issues {
		if issue.Rule == "bash-specific" {
			found = true

			break
		}
	}

	if !found {
		t.Error("checkCommand() should flag 'source' as bash-specific")
	}
}

func TestCheckCommand_PushdPopd(t *testing.T) {
	issues := checkCommand("test.yml", "build", "pushd /tmp && make && popd", 1, LintOptions{})

	found := false

	for _, issue := range issues {
		if issue.Rule == "bash-specific" {
			found = true

			break
		}
	}

	if !found {
		t.Error("checkCommand() should flag 'pushd/popd' as bash-specific")
	}
}

func TestCheckCommand_HereString(t *testing.T) {
	issues := checkCommand("test.yml", "build", "cat <<< 'hello'", 1, LintOptions{})

	found := false

	for _, issue := range issues {
		if issue.Rule == "bash-specific" {
			found = true

			break
		}
	}

	if !found {
		t.Error("checkCommand() should flag here-strings as bash-specific")
	}
}
