// Package tests provides comprehensive integration tests for all omni commands.
//
// Run with: go test -v ./tests/...
package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var omniBinary string

func TestMain(m *testing.M) {
	// Build omni binary for testing
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Go up to project root
	projectRoot := filepath.Dir(dir)

	if runtime.GOOS == "windows" {
		omniBinary = filepath.Join(projectRoot, "omni_test.exe")
	} else {
		omniBinary = filepath.Join(projectRoot, "omni_test")
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", omniBinary, ".")

	cmd.Dir = projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("failed to build omni: " + string(out))
	}

	code := m.Run()

	// Cleanup
	_ = os.Remove(omniBinary)

	os.Exit(code)
}

// Helper to run omni command
func runOmni(t *testing.T, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(omniBinary, args...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

// Helper to run omni with stdin
func runOmniWithStdin(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(omniBinary, args...)
	cmd.Stdin = strings.NewReader(stdin)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

// Helper to create temp file
func createTempFile(t *testing.T, content string) string {
	t.Helper()

	f, err := os.CreateTemp("", "omni_test_*")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}

	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Remove(f.Name()) })

	return f.Name()
}

// Helper to create temp dir
func createTempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "omni_test_*")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	return dir
}

// ============================================================================
// CORE COMMANDS
// ============================================================================

func TestEcho(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"basic", []string{"echo", "Hello World"}, "Hello World\n"},
		{"multiple args", []string{"echo", "Hello", "World"}, "Hello World\n"},
		{"no newline", []string{"echo", "-n", "test"}, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if out != tt.expected {
				t.Errorf("got %q, want %q", out, tt.expected)
			}
		})
	}
}

func TestPwd(t *testing.T) {
	out, err := runOmni(t, "pwd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestLs(t *testing.T) {
	dir := createTempDir(t)
	_ = os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644)

	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{"basic", []string{"ls", dir}, "test.txt"},
		{"long", []string{"ls", "-l", dir}, "test.txt"},
		{"all", []string{"ls", "-a", dir}, "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out, tt.contains) {
				t.Errorf("output %q does not contain %q", out, tt.contains)
			}
		})
	}
}

func TestCat(t *testing.T) {
	content := "Hello World\nLine 2"
	file := createTempFile(t, content)

	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{"basic", []string{"cat", file}, "Hello World"},
		{"number lines", []string{"cat", "-n", file}, "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out, tt.contains) {
				t.Errorf("output %q does not contain %q", out, tt.contains)
			}
		})
	}
}

func TestTree(t *testing.T) {
	dir := createTempDir(t)
	subdir := filepath.Join(dir, "subdir")
	_ = os.Mkdir(subdir, 0755)
	_ = os.WriteFile(filepath.Join(subdir, "file.txt"), []byte("test"), 0644)

	out, err := runOmni(t, "tree", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "subdir") {
		t.Errorf("output does not contain subdir")
	}
}

func TestDate(t *testing.T) {
	out, err := runOmni(t, "date")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestDirname(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/path/to/file.txt", filepath.Dir("/path/to/file.txt")},
		{"file.txt", "."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			out, err := runOmni(t, "dirname", tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(out) != tt.expected {
				t.Errorf("got %q, want %q", strings.TrimSpace(out), tt.expected)
			}
		})
	}
}

func TestBasename(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"basic", []string{"basename", "/path/to/file.txt"}, "file.txt"},
		{"with suffix", []string{"basename", "/path/to/file.txt", ".txt"}, "file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(out) != tt.expected {
				t.Errorf("got %q, want %q", strings.TrimSpace(out), tt.expected)
			}
		})
	}
}

func TestSeq(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{"basic", []string{"seq", "5"}, "1"},
		{"range", []string{"seq", "2", "5"}, "2"},
		{"separator", []string{"seq", "-s", ",", "3"}, "1,2,3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out, tt.contains) {
				t.Errorf("output %q does not contain %q", out, tt.contains)
			}
		})
	}
}

func TestArch(t *testing.T) {
	out, err := runOmni(t, "arch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestUname(t *testing.T) {
	out, err := runOmni(t, "uname")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		t.Error("expected non-empty output")
	}
}

// ============================================================================
// FILE OPERATIONS
// ============================================================================

func TestTouch(t *testing.T) {
	dir := createTempDir(t)
	file := filepath.Join(dir, "newfile.txt")

	_, err := runOmni(t, "touch", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Error("file was not created")
	}
}

func TestMkdir(t *testing.T) {
	dir := createTempDir(t)
	newDir := filepath.Join(dir, "newdir")

	_, err := runOmni(t, "mkdir", newDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(newDir)
	if os.IsNotExist(err) {
		t.Error("directory was not created")
	}

	if !info.IsDir() {
		t.Error("created path is not a directory")
	}
}

func TestMkdirParents(t *testing.T) {
	dir := createTempDir(t)
	deepDir := filepath.Join(dir, "a", "b", "c")

	_, err := runOmni(t, "mkdir", "-p", deepDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(deepDir); os.IsNotExist(err) {
		t.Error("nested directories were not created")
	}
}

func TestCp(t *testing.T) {
	dir := createTempDir(t)
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	_ = os.WriteFile(src, []byte("content"), 0644)

	_, err := runOmni(t, "cp", src, dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read destination: %v", err)
	}

	if string(content) != "content" {
		t.Errorf("got %q, want %q", string(content), "content")
	}
}

func TestMv(t *testing.T) {
	dir := createTempDir(t)
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	_ = os.WriteFile(src, []byte("content"), 0644)

	_, err := runOmni(t, "mv", src, dst)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source file still exists")
	}

	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read destination: %v", err)
	}

	if string(content) != "content" {
		t.Errorf("got %q, want %q", string(content), "content")
	}
}

func TestRm(t *testing.T) {
	file := createTempFile(t, "content")

	_, err := runOmni(t, "rm", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Error("file still exists")
	}
}

func TestStat(t *testing.T) {
	file := createTempFile(t, "content")

	out, err := runOmni(t, "stat", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestFind(t *testing.T) {
	dir := createTempDir(t)
	_ = os.WriteFile(filepath.Join(dir, "test.txt"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "test.json"), []byte("{}"), 0644)

	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{"basic", []string{"find", dir}, "test.txt"},
		{"name pattern", []string{"find", dir, "--name", "*.txt"}, "test.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out, tt.contains) {
				t.Errorf("output does not contain %q", tt.contains)
			}
		})
	}
}

// ============================================================================
// TEXT PROCESSING
// ============================================================================

func TestHead(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5"
	file := createTempFile(t, content)

	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{"basic", []string{"head", file}, "line1"},
		{"n 2", []string{"head", "-n", "2", file}, "line1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out, tt.contains) {
				t.Errorf("output does not contain %q", tt.contains)
			}
		})
	}
}

func TestTail(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5"
	file := createTempFile(t, content)

	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{"basic", []string{"tail", file}, "line5"},
		{"n 2", []string{"tail", "-n", "2", file}, "line4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out, tt.contains) {
				t.Errorf("output does not contain %q", tt.contains)
			}
		})
	}
}

func TestWc(t *testing.T) {
	content := "hello world\nfoo bar"
	file := createTempFile(t, content)

	out, err := runOmni(t, "wc", "-l", file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "2") {
		t.Errorf("expected 2 lines, got: %s", out)
	}
}

func TestSort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		args     []string
		expected string
	}{
		{"basic", "c\na\nb", []string{"sort"}, "a\nb\nc"},
		{"reverse", "a\nb\nc", []string{"sort", "-r"}, "c\nb\na"},
		{"numeric", "10\n2\n1", []string{"sort", "-n"}, "1\n2\n10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmniWithStdin(t, tt.input, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(out) != tt.expected {
				t.Errorf("got %q, want %q", strings.TrimSpace(out), tt.expected)
			}
		})
	}
}

func TestUniq(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		args     []string
		expected string
	}{
		{"basic", "a\na\nb\n", []string{"uniq"}, "a\nb"},
		{"count", "a\nb", []string{"uniq", "-c"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmniWithStdin(t, tt.input, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expected != "" && strings.TrimSpace(out) != tt.expected {
				t.Errorf("got %q, want %q", strings.TrimSpace(out), tt.expected)
			}
		})
	}
}

func TestGrep(t *testing.T) {
	content := "hello world\nfoo bar\nhello again"
	file := createTempFile(t, content)

	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{"basic", []string{"grep", "hello", file}, "hello"},
		{"ignore case", []string{"grep", "-i", "HELLO", file}, "hello"},
		{"invert", []string{"grep", "-v", "hello", file}, "foo bar"},
		{"count", []string{"grep", "-c", "hello", file}, "2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out, tt.contains) {
				t.Errorf("output does not contain %q: got %q", tt.contains, out)
			}
		})
	}
}

func TestCut(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		args     []string
		expected string
	}{
		{"field", "a,b,c", []string{"cut", "-d", ",", "-f", "2"}, "b"},
		{"chars", "hello", []string{"cut", "-c", "1-3"}, "hel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmniWithStdin(t, tt.input, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(out) != tt.expected {
				t.Errorf("got %q, want %q", strings.TrimSpace(out), tt.expected)
			}
		})
	}
}

func TestTr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		args     []string
		expected string
	}{
		{"translate", "hello", []string{"tr", "a-z", "A-Z"}, "HELLO"},
		{"delete", "hello123", []string{"tr", "-d", "0-9"}, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmniWithStdin(t, tt.input, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(out) != tt.expected {
				t.Errorf("got %q, want %q", strings.TrimSpace(out), tt.expected)
			}
		})
	}
}

func TestSed(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		args     []string
		expected string
	}{
		{"substitute", "hello world", []string{"sed", "s/world/universe/"}, "hello universe"},
		{"global", "aaa", []string{"sed", "s/a/b/g"}, "bbb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmniWithStdin(t, tt.input, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(out) != tt.expected {
				t.Errorf("got %q, want %q", strings.TrimSpace(out), tt.expected)
			}
		})
	}
}

func TestAwk(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		args     []string
		expected string
	}{
		{"print field", "a b c", []string{"awk", "{print $2}"}, "b"},
		{"separator", "a,b,c", []string{"awk", "-F", ",", "{print $2}"}, "b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmniWithStdin(t, tt.input, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(out) != tt.expected {
				t.Errorf("got %q, want %q", strings.TrimSpace(out), tt.expected)
			}
		})
	}
}

func TestTac(t *testing.T) {
	out, err := runOmniWithStdin(t, "a\nb\nc", "tac")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(out) != "c\nb\na" {
		t.Errorf("got %q, want %q", strings.TrimSpace(out), "c\nb\na")
	}
}

func TestRev(t *testing.T) {
	out, err := runOmniWithStdin(t, "hello", "rev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(out) != "olleh" {
		t.Errorf("got %q, want %q", strings.TrimSpace(out), "olleh")
	}
}

// ============================================================================
// HASH & ENCODING
// ============================================================================

func TestHash(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		input    string
		contains string
	}{
		{"sha256", []string{"hash", "-a", "sha256"}, "test", "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"},
		{"md5", []string{"hash", "-a", "md5"}, "test", "098f6bcd4621d373cade4e832627b4f6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmniWithStdin(t, tt.input, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out, tt.contains) {
				t.Errorf("output does not contain expected hash")
			}
		})
	}
}

func TestBase64(t *testing.T) {
	// Encode
	out, err := runOmniWithStdin(t, "hello", "base64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(out) != "aGVsbG8=" {
		t.Errorf("encode: got %q, want %q", strings.TrimSpace(out), "aGVsbG8=")
	}

	// Decode
	out, err = runOmniWithStdin(t, "aGVsbG8=", "base64", "-d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(out) != "hello" {
		t.Errorf("decode: got %q, want %q", strings.TrimSpace(out), "hello")
	}
}

func TestBase32(t *testing.T) {
	out, err := runOmniWithStdin(t, "hello", "base32")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "NBSWY3DP") {
		t.Errorf("encode: got %q, expected to contain NBSWY3DP", out)
	}
}

func TestHex(t *testing.T) {
	// Encode
	out, err := runOmniWithStdin(t, "hello", "hex", "encode")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(out) != "68656c6c6f" {
		t.Errorf("encode: got %q, want %q", strings.TrimSpace(out), "68656c6c6f")
	}

	// Decode
	out, err = runOmniWithStdin(t, "68656c6c6f", "hex", "decode")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(out) != "hello" {
		t.Errorf("decode: got %q, want %q", strings.TrimSpace(out), "hello")
	}
}

func TestUrl(t *testing.T) {
	// Encode
	out, err := runOmni(t, "url", "encode", "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(out) != "hello%20world" {
		t.Errorf("encode: got %q, want %q", strings.TrimSpace(out), "hello%20world")
	}

	// Decode
	out, err = runOmni(t, "url", "decode", "hello%20world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(out) != "hello world" {
		t.Errorf("decode: got %q, want %q", strings.TrimSpace(out), "hello world")
	}
}

func TestXxd(t *testing.T) {
	// Basic dump
	out, err := runOmniWithStdin(t, "hello", "xxd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "00000000:") {
		t.Error("output should contain address")
	}

	if !strings.Contains(out, "hello") {
		t.Error("output should contain ASCII")
	}
}

func TestXxdPlain(t *testing.T) {
	out, err := runOmniWithStdin(t, "hello", "xxd", "-p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(out) != "68656c6c6f" {
		t.Errorf("got %q, want %q", strings.TrimSpace(out), "68656c6c6f")
	}
}

func TestXxdReverse(t *testing.T) {
	out, err := runOmniWithStdin(t, "68656c6c6f", "xxd", "-r", "-p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(out) != "hello" {
		t.Errorf("got %q, want %q", strings.TrimSpace(out), "hello")
	}
}

func TestXxdInclude(t *testing.T) {
	out, err := runOmniWithStdin(t, "Hi", "xxd", "-i")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "unsigned char") {
		t.Error("output should contain C array definition")
	}

	if !strings.Contains(out, "0x48") {
		t.Error("output should contain hex values")
	}
}

func TestXxdBits(t *testing.T) {
	out, err := runOmniWithStdin(t, "A", "xxd", "-b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 'A' = 0x41 = 01000001
	if !strings.Contains(out, "01000001") {
		t.Errorf("output should contain binary for 'A'\nGot: %s", out)
	}
}

func TestXxdRoundtrip(t *testing.T) {
	original := "Hello, World!"

	// Dump to plain hex
	dumpOut, err := runOmniWithStdin(t, original, "xxd", "-p")
	if err != nil {
		t.Fatalf("dump error: %v", err)
	}

	// Reverse back
	reverseOut, err := runOmniWithStdin(t, strings.TrimSpace(dumpOut), "xxd", "-r", "-p")
	if err != nil {
		t.Fatalf("reverse error: %v", err)
	}

	if reverseOut != original {
		t.Errorf("roundtrip failed: got %q, want %q", reverseOut, original)
	}
}

// ============================================================================
// DATA PROCESSING
// ============================================================================

func TestJq(t *testing.T) {
	jsonContent := `{"name":"test","version":"1.0.0"}`
	file := createTempFile(t, jsonContent)

	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{"query", []string{"jq", ".name", file}, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(out, tt.contains) {
				t.Errorf("output does not contain %q", tt.contains)
			}
		})
	}
}

func TestJsonFmt(t *testing.T) {
	out, err := runOmniWithStdin(t, `{"a":1}`, "json", "fmt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "\"a\"") {
		t.Errorf("output does not look formatted")
	}
}

func TestJsonValidate(t *testing.T) {
	// Valid JSON
	_, err := runOmniWithStdin(t, `{"a":1}`, "json", "validate")
	if err != nil {
		t.Errorf("valid JSON should pass: %v", err)
	}

	// Invalid JSON
	_, err = runOmniWithStdin(t, `not json`, "json", "validate")
	if err == nil {
		t.Error("invalid JSON should fail")
	}
}

func TestYamlValidate(t *testing.T) {
	// Valid YAML
	_, err := runOmniWithStdin(t, "name: test\nversion: 1.0.0", "yaml", "validate")
	if err != nil {
		t.Errorf("valid YAML should pass: %v", err)
	}
}

// ============================================================================
// ID GENERATORS
// ============================================================================

func TestUuid(t *testing.T) {
	out, err := runOmni(t, "uuid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(strings.TrimSpace(out)) != 36 {
		t.Errorf("unexpected UUID length: %d", len(strings.TrimSpace(out)))
	}
}

func TestRandom(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"string", []string{"random", "string", "16"}},
		{"hex", []string{"random", "hex", "16"}},
		{"int", []string{"random", "int", "1", "100"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(strings.TrimSpace(out)) == 0 {
				t.Error("expected non-empty output")
			}
		})
	}
}

func TestKsuid(t *testing.T) {
	out, err := runOmni(t, "ksuid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestUlid(t *testing.T) {
	out, err := runOmni(t, "ulid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ULID is 26 characters
	if len(strings.TrimSpace(out)) != 26 {
		t.Errorf("unexpected ULID length: %d", len(strings.TrimSpace(out)))
	}
}

func TestNanoid(t *testing.T) {
	out, err := runOmni(t, "nanoid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		t.Error("expected non-empty output")
	}
}

// ============================================================================
// SYSTEM COMMANDS
// ============================================================================

func TestEnv(t *testing.T) {
	out, err := runOmni(t, "env")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should contain at least PATH or similar
	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestWhoami(t *testing.T) {
	out, err := runOmni(t, "whoami")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(strings.TrimSpace(out)) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestId(t *testing.T) {
	out, err := runOmni(t, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestUptime(t *testing.T) {
	out, err := runOmni(t, "uptime")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestDf(t *testing.T) {
	out, err := runOmni(t, "df")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestDu(t *testing.T) {
	dir := createTempDir(t)
	_ = os.WriteFile(filepath.Join(dir, "test.txt"), []byte("content"), 0644)

	out, err := runOmni(t, "du", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestPs(t *testing.T) {
	out, err := runOmni(t, "ps")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestKillList(t *testing.T) {
	out, err := runOmni(t, "kill", "-l")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

// ============================================================================
// ARCHIVE & COMPRESSION
// ============================================================================

func TestTar(t *testing.T) {
	dir := createTempDir(t)
	file := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(file, []byte("content"), 0644)

	tarFile := filepath.Join(dir, "test.tar")

	// Create
	_, err := runOmni(t, "tar", "-cf", tarFile, "-C", dir, "test.txt")
	if err != nil {
		t.Fatalf("create: unexpected error: %v", err)
	}

	// List
	out, err := runOmni(t, "tar", "-tf", tarFile)
	if err != nil {
		t.Fatalf("list: unexpected error: %v", err)
	}

	if !strings.Contains(out, "test.txt") {
		t.Error("list should contain test.txt")
	}
}

func TestZip(t *testing.T) {
	dir := createTempDir(t)
	file := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(file, []byte("content"), 0644)

	zipFile := filepath.Join(dir, "test.zip")

	// Create
	_, err := runOmni(t, "zip", zipFile, file)
	if err != nil {
		t.Fatalf("create: unexpected error: %v", err)
	}

	// List
	out, err := runOmni(t, "unzip", "-l", zipFile)
	if err != nil {
		t.Fatalf("list: unexpected error: %v", err)
	}

	if !strings.Contains(out, "test.txt") {
		t.Error("list should contain test.txt")
	}
}

func TestGzip(t *testing.T) {
	dir := createTempDir(t)
	file := filepath.Join(dir, "test.txt")
	_ = os.WriteFile(file, []byte("content"), 0644)

	// Compress
	_, err := runOmni(t, "gzip", file)
	if err != nil {
		t.Fatalf("compress: unexpected error: %v", err)
	}

	gzFile := file + ".gz"
	if _, err := os.Stat(gzFile); os.IsNotExist(err) {
		t.Error("gz file was not created")
	}

	// Decompress
	_, err = runOmni(t, "gunzip", gzFile)
	if err != nil {
		t.Fatalf("decompress: unexpected error: %v", err)
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Error("decompressed file was not created")
	}
}

// ============================================================================
// UTILITIES
// ============================================================================

func TestCase(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"upper", []string{"case", "upper", "hello"}, "HELLO"},
		{"lower", []string{"case", "lower", "HELLO"}, "hello"},
		{"title", []string{"case", "title", "hello world"}, "Hello World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runOmni(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if strings.TrimSpace(out) != tt.expected {
				t.Errorf("got %q, want %q", strings.TrimSpace(out), tt.expected)
			}
		})
	}
}

func TestCmdtree(t *testing.T) {
	out, err := runOmni(t, "cmdtree")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestLoc(t *testing.T) {
	dir := createTempDir(t)
	goFile := filepath.Join(dir, "test.go")
	_ = os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0644)

	out, err := runOmni(t, "loc", dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

// ============================================================================
// ERROR HANDLING
// ============================================================================

func TestErrorNonexistentFile(t *testing.T) {
	_, err := runOmni(t, "cat", "/nonexistent/file/path")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestErrorInvalidJson(t *testing.T) {
	_, err := runOmniWithStdin(t, "not json", "json", "validate")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestErrorGrepNoMatch(t *testing.T) {
	file := createTempFile(t, "hello world")

	_, err := runOmni(t, "grep", "xyz123", file)
	if err == nil {
		t.Error("expected error when grep finds no matches")
	}
}
