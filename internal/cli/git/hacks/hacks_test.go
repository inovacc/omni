package hacks

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing.
func setupTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = dir
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	_ = cmd.Run()

	return dir
}

func TestQuickCommit(t *testing.T) {
	dir := setupTestRepo(t)
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Create a test file
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test quick commit
	err := QuickCommit("test commit", true)
	if err != nil {
		t.Fatalf("QuickCommit failed: %v", err)
	}

	// Verify commit was created
	cmd := exec.Command("git", "log", "--oneline", "-1")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}

	if !strings.Contains(string(out), "test commit") {
		t.Errorf("commit message not found in log: %s", out)
	}
}

func TestBranchClean_NoBranches(t *testing.T) {
	dir := setupTestRepo(t)
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Create initial commit
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = dir
	_ = cmd.Run()

	// Test branch clean with no merged branches
	deleted, err := BranchClean(true)
	if err != nil {
		t.Fatalf("BranchClean failed: %v", err)
	}

	if len(deleted) != 0 {
		t.Errorf("expected no branches to delete, got %d", len(deleted))
	}
}

func TestUndo(t *testing.T) {
	dir := setupTestRepo(t)
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Create initial commit
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "first commit")
	cmd.Dir = dir
	_ = cmd.Run()

	// Create second commit
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command("git", "add", "-A")
	cmd.Dir = dir
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "second commit")
	cmd.Dir = dir
	_ = cmd.Run()

	// Test undo
	err := Undo()
	if err != nil {
		t.Fatalf("Undo failed: %v", err)
	}

	// Verify we're back to first commit
	cmd = exec.Command("git", "log", "--oneline", "-1")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}

	if !strings.Contains(string(out), "first commit") {
		t.Errorf("expected first commit, got: %s", out)
	}
}

func TestLogGraph(t *testing.T) {
	dir := setupTestRepo(t)
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Create initial commit
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "test log")
	cmd.Dir = dir
	_ = cmd.Run()

	// Test log graph
	out, err := LogGraph(5)
	if err != nil {
		t.Fatalf("LogGraph failed: %v", err)
	}

	if !strings.Contains(out, "test log") {
		t.Errorf("expected commit message in log, got: %s", out)
	}
}

func TestStatus(t *testing.T) {
	dir := setupTestRepo(t)
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Test status on empty repo
	out, err := Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	// Should contain branch info
	if out == "" {
		t.Error("expected non-empty status output")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	dir := setupTestRepo(t)
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Create initial commit (needed for branch to exist)
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = dir
	_ = cmd.Run()

	// Test getCurrentBranch
	branch, err := getCurrentBranch()
	if err != nil {
		t.Fatalf("getCurrentBranch failed: %v", err)
	}

	// Default branch should be master or main
	if branch != "master" && branch != "main" {
		t.Errorf("unexpected branch name: %s", branch)
	}
}
