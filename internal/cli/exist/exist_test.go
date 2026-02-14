package exist

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestRunFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	spacesFile := filepath.Join(tmpDir, "file with spaces.txt")

	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(emptyFile, nil, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(spacesFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		target       string
		wantExists   bool
		wantNotFound bool
	}{
		{"existing file", tmpFile, true, false},
		{"empty file", emptyFile, true, false},
		{"file with spaces in path", spacesFile, true, false},
		{"directory is not a file", tmpDir, false, true},
		{"nonexistent", filepath.Join(tmpDir, "nope"), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunFile(&buf, tt.target, Options{})

			if tt.wantNotFound && !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}

			if !tt.wantNotFound && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantExists && !strings.Contains(buf.String(), "exists (file)") {
				t.Errorf("expected 'exists (file)' in output, got %q", buf.String())
			}

			if tt.wantNotFound && !strings.Contains(buf.String(), "not found") {
				t.Errorf("expected 'not found' in output, got %q", buf.String())
			}
		})
	}
}

func TestRunDir(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	nestedDir := filepath.Join(tmpDir, "sub", "nested")

	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		target       string
		wantExists   bool
		wantNotFound bool
	}{
		{"existing dir", tmpDir, true, false},
		{"nested subdir", nestedDir, true, false},
		{"file is not a dir", tmpFile, false, true},
		{"nonexistent", filepath.Join(tmpDir, "nope"), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunDir(&buf, tt.target, Options{})

			if tt.wantNotFound && !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}

			if !tt.wantNotFound && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantExists && !strings.Contains(buf.String(), "exists (dir)") {
				t.Errorf("expected 'exists (dir)' in output, got %q", buf.String())
			}
		})
	}
}

func TestRunPath(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")

	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		target       string
		wantExists   bool
		wantNotFound bool
		wantType     string
	}{
		{"existing file", tmpFile, true, false, "file"},
		{"existing dir", tmpDir, true, false, "dir"},
		{"nonexistent", filepath.Join(tmpDir, "nope"), false, true, ""},
	}

	// Add symlink test on non-Windows platforms
	if runtime.GOOS != "windows" {
		symlink := filepath.Join(tmpDir, "link")
		if err := os.Symlink(tmpFile, symlink); err == nil {
			tests = append(tests, struct {
				name         string
				target       string
				wantExists   bool
				wantNotFound bool
				wantType     string
			}{"symlink", symlink, true, false, "symlink"})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunPath(&buf, tt.target, Options{})

			if tt.wantNotFound && !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}

			if !tt.wantNotFound && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantType != "" && !strings.Contains(buf.String(), "exists ("+tt.wantType+")") {
				t.Errorf("expected 'exists (%s)' in output, got %q", tt.wantType, buf.String())
			}
		})
	}
}

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		wantExists   bool
		wantNotFound bool
	}{
		{"go command exists", "go", true, false},
		{"nonexistent command", "nonexistent_cmd_xyz_12345", false, true},
		{"empty command name", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunCommand(&buf, tt.target, Options{})

			if tt.wantNotFound && !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}

			if !tt.wantNotFound && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantExists && !strings.Contains(buf.String(), "exists (command)") {
				t.Errorf("expected 'exists (command)' in output, got %q", buf.String())
			}
		})
	}
}

func TestRunEnv(t *testing.T) {
	t.Setenv("TEST_EXIST_VAR", "test_value")
	t.Setenv("TEST_EXIST_EMPTY", "")

	tests := []struct {
		name         string
		target       string
		wantExists   bool
		wantNotFound bool
	}{
		{"existing var", "TEST_EXIST_VAR", true, false},
		{"empty value var", "TEST_EXIST_EMPTY", true, false},
		{"PATH exists", "PATH", true, false},
		{"nonexistent var", "NONEXISTENT_VAR_XYZ_12345", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunEnv(&buf, tt.target, Options{})

			if tt.wantNotFound && !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}

			if !tt.wantNotFound && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantExists && !strings.Contains(buf.String(), "exists (env)") {
				t.Errorf("expected 'exists (env)' in output, got %q", buf.String())
			}
		})
	}
}

func TestRunProcess(t *testing.T) {
	currentPID := strconv.Itoa(os.Getpid())

	tests := []struct {
		name         string
		target       string
		wantExists   bool
		wantNotFound bool
	}{
		{"current process by PID", currentPID, true, false},
		{"nonexistent PID", "999999999", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunProcess(&buf, tt.target, Options{})

			if tt.wantNotFound && !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}

			if !tt.wantNotFound && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantExists && !strings.Contains(buf.String(), "exists (process)") {
				t.Errorf("expected 'exists (process)' in output, got %q", buf.String())
			}
		})
	}
}

func TestRunProcessByName(t *testing.T) {
	// "go" test runner should be running as a process
	var buf bytes.Buffer

	// Search for a nonexistent process name
	err := RunProcess(&buf, "nonexistent_process_xyz_12345", Options{})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for nonexistent process name, got %v", err)
	}
}

func TestRunPort(t *testing.T) {
	// Start a listener on a random port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = ln.Close() }()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())

	tests := []struct {
		name         string
		target       string
		wantExists   bool
		wantNotFound bool
		wantParseErr bool
	}{
		{"listening port", portStr, true, false, false},
		{"unused port", "19", false, true, false},
		{"invalid port string", "abc", false, false, true},
		{"out of range high", "99999", false, false, true},
		{"zero port", "0", false, false, true},
		{"negative port", "-1", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunPort(&buf, tt.target, Options{})

			if tt.wantParseErr {
				if err == nil {
					t.Error("expected parse error")
				}

				if errors.Is(err, ErrNotFound) {
					t.Error("expected parse error, not ErrNotFound")
				}

				return
			}

			if tt.wantNotFound && !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}

			if !tt.wantNotFound && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantExists && !strings.Contains(buf.String(), "exists (port)") {
				t.Errorf("expected 'exists (port)' in output, got %q", buf.String())
			}
		})
	}
}

func TestJSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")

	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TEST_JSON_VAR", "json_value")

	t.Run("file exists", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunFile(&buf, tmpFile, Options{JSON: true})
		if err != nil {
			t.Fatal(err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if !result.Exists {
			t.Error("expected exists=true")
		}

		if result.Type != "file" {
			t.Errorf("expected type=file, got %s", result.Type)
		}

		if result.Target != tmpFile {
			t.Errorf("expected target=%s, got %s", tmpFile, result.Target)
		}

		// Verify details are present
		if result.Details == nil {
			t.Error("expected details to be populated")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunFile(&buf, filepath.Join(tmpDir, "nope"), Options{JSON: true})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if result.Exists {
			t.Error("expected exists=false")
		}

		if result.Details != nil {
			t.Error("expected details=nil for not found")
		}
	})

	t.Run("dir exists", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDir(&buf, tmpDir, Options{JSON: true})
		if err != nil {
			t.Fatal(err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if !result.Exists || result.Type != "dir" {
			t.Errorf("expected exists=true type=dir, got exists=%v type=%s", result.Exists, result.Type)
		}
	})

	t.Run("command exists", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunCommand(&buf, "go", Options{JSON: true})
		if err != nil {
			t.Fatal(err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if !result.Exists || result.Type != "command" {
			t.Errorf("expected exists=true type=command, got exists=%v type=%s", result.Exists, result.Type)
		}
	})

	t.Run("env exists", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEnv(&buf, "TEST_JSON_VAR", Options{JSON: true})
		if err != nil {
			t.Fatal(err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if !result.Exists || result.Type != "env" {
			t.Errorf("expected exists=true type=env, got exists=%v type=%s", result.Exists, result.Type)
		}
	})

	t.Run("process exists", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunProcess(&buf, strconv.Itoa(os.Getpid()), Options{JSON: true})
		if err != nil {
			t.Fatal(err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if !result.Exists || result.Type != "process" {
			t.Errorf("expected exists=true type=process, got exists=%v type=%s", result.Exists, result.Type)
		}
	})

	t.Run("port listening", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = ln.Close() }()

		_, portStr, _ := net.SplitHostPort(ln.Addr().String())

		var buf bytes.Buffer

		err = RunPort(&buf, portStr, Options{JSON: true})
		if err != nil {
			t.Fatal(err)
		}

		var result Result
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if !result.Exists || result.Type != "port" {
			t.Errorf("expected exists=true type=port, got exists=%v type=%s", result.Exists, result.Type)
		}
	})
}

func TestQuietMode(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")

	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TEST_QUIET_VAR", "value")

	tests := []struct {
		name    string
		run     func(w *bytes.Buffer) error
		wantErr bool
	}{
		{
			"file exists",
			func(w *bytes.Buffer) error { return RunFile(w, tmpFile, Options{Quiet: true}) },
			false,
		},
		{
			"file not found",
			func(w *bytes.Buffer) error { return RunFile(w, filepath.Join(tmpDir, "nope"), Options{Quiet: true}) },
			true,
		},
		{
			"dir exists",
			func(w *bytes.Buffer) error { return RunDir(w, tmpDir, Options{Quiet: true}) },
			false,
		},
		{
			"dir not found",
			func(w *bytes.Buffer) error { return RunDir(w, filepath.Join(tmpDir, "nope"), Options{Quiet: true}) },
			true,
		},
		{
			"path exists",
			func(w *bytes.Buffer) error { return RunPath(w, tmpFile, Options{Quiet: true}) },
			false,
		},
		{
			"command exists",
			func(w *bytes.Buffer) error { return RunCommand(w, "go", Options{Quiet: true}) },
			false,
		},
		{
			"command not found",
			func(w *bytes.Buffer) error { return RunCommand(w, "nonexistent_cmd_xyz", Options{Quiet: true}) },
			true,
		},
		{
			"env exists",
			func(w *bytes.Buffer) error { return RunEnv(w, "TEST_QUIET_VAR", Options{Quiet: true}) },
			false,
		},
		{
			"env not found",
			func(w *bytes.Buffer) error { return RunEnv(w, "NONEXISTENT_VAR_XYZ", Options{Quiet: true}) },
			true,
		},
		{
			"process exists",
			func(w *bytes.Buffer) error { return RunProcess(w, strconv.Itoa(os.Getpid()), Options{Quiet: true}) },
			false,
		},
		{
			"process not found",
			func(w *bytes.Buffer) error { return RunProcess(w, "999999999", Options{Quiet: true}) },
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := tt.run(&buf)

			if buf.Len() != 0 {
				t.Errorf("expected empty output in quiet mode, got %q", buf.String())
			}

			if tt.wantErr && !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestOutputFormat(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")

	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("exists format", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunFile(&buf, tmpFile, Options{})

		want := tmpFile + ": exists (file)\n"
		if buf.String() != want {
			t.Errorf("got %q, want %q", buf.String(), want)
		}
	})

	t.Run("not found format", func(t *testing.T) {
		var buf bytes.Buffer

		target := filepath.Join(tmpDir, "nope")
		_ = RunFile(&buf, target, Options{})

		want := target + ": not found\n"
		if buf.String() != want {
			t.Errorf("got %q, want %q", buf.String(), want)
		}
	})

	t.Run("dir type in output", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunDir(&buf, tmpDir, Options{})

		want := tmpDir + ": exists (dir)\n"
		if buf.String() != want {
			t.Errorf("got %q, want %q", buf.String(), want)
		}
	})
}

func TestJSONDetailsContent(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")

	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TEST_DETAIL_VAR", "detail_value")

	t.Run("file details has size", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunFile(&buf, tmpFile, Options{JSON: true}); err != nil {
			t.Fatal(err)
		}

		var raw map[string]any
		if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
			t.Fatal(err)
		}

		details, ok := raw["details"].(map[string]any)
		if !ok {
			t.Fatal("details is not a map")
		}

		size, ok := details["size"].(float64)
		if !ok || size != 5 {
			t.Errorf("expected size=5, got %v", details["size"])
		}

		if _, ok := details["mode"]; !ok {
			t.Error("expected mode in details")
		}

		if _, ok := details["mod_time"]; !ok {
			t.Error("expected mod_time in details")
		}
	})

	t.Run("command details has path", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunCommand(&buf, "go", Options{JSON: true}); err != nil {
			t.Fatal(err)
		}

		var raw map[string]any
		if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
			t.Fatal(err)
		}

		details, ok := raw["details"].(map[string]any)
		if !ok {
			t.Fatal("details is not a map")
		}

		path, ok := details["path"].(string)
		if !ok || path == "" {
			t.Errorf("expected non-empty path in details, got %v", details["path"])
		}
	})

	t.Run("env details has value", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunEnv(&buf, "TEST_DETAIL_VAR", Options{JSON: true}); err != nil {
			t.Fatal(err)
		}

		var raw map[string]any
		if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
			t.Fatal(err)
		}

		details, ok := raw["details"].(map[string]any)
		if !ok {
			t.Fatal("details is not a map")
		}

		val, ok := details["value"].(string)
		if !ok || val != "detail_value" {
			t.Errorf("expected value=detail_value, got %v", details["value"])
		}
	})

	t.Run("process details has pid", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunProcess(&buf, strconv.Itoa(os.Getpid()), Options{JSON: true}); err != nil {
			t.Fatal(err)
		}

		var raw map[string]any
		if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
			t.Fatal(err)
		}

		details, ok := raw["details"].(map[string]any)
		if !ok {
			t.Fatal("details is not a map")
		}

		pid, ok := details["pid"].(float64)
		if !ok || int(pid) != os.Getpid() {
			t.Errorf("expected pid=%d, got %v", os.Getpid(), details["pid"])
		}

		if _, ok := details["name"]; !ok {
			t.Error("expected name in details")
		}
	})
}
