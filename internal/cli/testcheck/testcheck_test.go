package testcheck

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/output"
)

func TestCheck(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "testcheck_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create package with tests
	pkgWithTests := filepath.Join(tmpDir, "withtest")
	_ = os.MkdirAll(pkgWithTests, 0755)
	_ = os.WriteFile(filepath.Join(pkgWithTests, "main.go"), []byte("package withtest\n"), 0644)
	_ = os.WriteFile(filepath.Join(pkgWithTests, "main_test.go"), []byte("package withtest\n"), 0644)

	// Create package without tests
	pkgNoTests := filepath.Join(tmpDir, "notest")
	_ = os.MkdirAll(pkgNoTests, 0755)
	_ = os.WriteFile(filepath.Join(pkgNoTests, "main.go"), []byte("package notest\n"), 0644)

	// Create package with only test files (should be excluded)
	pkgOnlyTests := filepath.Join(tmpDir, "onlytest")
	_ = os.MkdirAll(pkgOnlyTests, 0755)
	_ = os.WriteFile(filepath.Join(pkgOnlyTests, "main_test.go"), []byte("package onlytest\n"), 0644)

	result, err := Check(tmpDir)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.Total != 2 {
		t.Errorf("Check() total = %d, want 2", result.Total)
	}

	if result.WithTests != 1 {
		t.Errorf("Check() withTests = %d, want 1", result.WithTests)
	}

	if result.NoTests != 1 {
		t.Errorf("Check() noTests = %d, want 1", result.NoTests)
	}

	if result.Coverage != 50.0 {
		t.Errorf("Check() coverage = %.1f, want 50.0", result.Coverage)
	}
}

func TestCheck_SkipsHiddenDirs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testcheck_hidden")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create hidden directory
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	_ = os.MkdirAll(hiddenDir, 0755)
	_ = os.WriteFile(filepath.Join(hiddenDir, "main.go"), []byte("package hidden\n"), 0644)

	// Create vendor directory
	vendorDir := filepath.Join(tmpDir, "vendor")
	_ = os.MkdirAll(vendorDir, 0755)
	_ = os.WriteFile(filepath.Join(vendorDir, "main.go"), []byte("package vendor\n"), 0644)

	// Create normal package
	normalDir := filepath.Join(tmpDir, "normal")
	_ = os.MkdirAll(normalDir, 0755)
	_ = os.WriteFile(filepath.Join(normalDir, "main.go"), []byte("package normal\n"), 0644)

	result, err := Check(tmpDir)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	// Should only find the normal package
	if result.Total != 1 {
		t.Errorf("Check() should skip hidden/vendor dirs, total = %d, want 1", result.Total)
	}
}

func TestRun_TextOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testcheck_output")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a package without tests
	pkgDir := filepath.Join(tmpDir, "mypackage")
	_ = os.MkdirAll(pkgDir, 0755)
	_ = os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package mypackage\n"), 0644)

	var buf bytes.Buffer

	err = Run(&buf, tmpDir, Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Run() output should not be empty")
	}

	// Should contain "NO TEST" since we created a package without tests
	if !strings.Contains(output, "NO TEST") {
		t.Errorf("Run() output should contain 'NO TEST': %s", output)
	}
}

func TestRun_JSONOutput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testcheck_json")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	pkgDir := filepath.Join(tmpDir, "jsonpkg")
	_ = os.MkdirAll(pkgDir, 0755)
	_ = os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package jsonpkg\n"), 0644)
	_ = os.WriteFile(filepath.Join(pkgDir, "main_test.go"), []byte("package jsonpkg\n"), 0644)

	var buf bytes.Buffer

	err = Run(&buf, tmpDir, Options{OutputFormat: output.FormatJSON})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Run() JSON output invalid: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Run() JSON total = %d, want 1", result.Total)
	}

	if result.WithTests != 1 {
		t.Errorf("Run() JSON withTests = %d, want 1", result.WithTests)
	}
}

func TestRun_Summary(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testcheck_summary")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	pkgDir := filepath.Join(tmpDir, "pkg1")
	_ = os.MkdirAll(pkgDir, 0755)
	_ = os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package pkg1\n"), 0644)

	var buf bytes.Buffer

	err = Run(&buf, tmpDir, Options{Summary: true})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()
	// Summary should be a single line
	if !strings.Contains(output, "Total:") {
		t.Errorf("Run() summary should contain 'Total:': %s", output)
	}
}

func TestRun_ShowAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testcheck_all")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Package with tests
	pkg1 := filepath.Join(tmpDir, "pkg1")
	_ = os.MkdirAll(pkg1, 0755)
	_ = os.WriteFile(filepath.Join(pkg1, "main.go"), []byte("package pkg1\n"), 0644)
	_ = os.WriteFile(filepath.Join(pkg1, "main_test.go"), []byte("package pkg1\n"), 0644)

	// Package without tests
	pkg2 := filepath.Join(tmpDir, "pkg2")
	_ = os.MkdirAll(pkg2, 0755)
	_ = os.WriteFile(filepath.Join(pkg2, "main.go"), []byte("package pkg2\n"), 0644)

	var buf bytes.Buffer

	err = Run(&buf, tmpDir, Options{ShowAll: true})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()
	// Should show both HAS TEST and NO TEST
	if !strings.Contains(output, "HAS TEST") {
		t.Errorf("Run() --all should show 'HAS TEST': %s", output)
	}

	if !strings.Contains(output, "NO TEST") {
		t.Errorf("Run() --all should show 'NO TEST': %s", output)
	}
}

func TestCheck_EmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testcheck_empty")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	result, err := Check(tmpDir)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.Total != 0 {
		t.Errorf("Check() empty dir total = %d, want 0", result.Total)
	}

	if result.Coverage != 0 {
		t.Errorf("Check() empty dir coverage = %.1f, want 0", result.Coverage)
	}
}

func TestCheck_MultipleTestFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testcheck_multi")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	pkgDir := filepath.Join(tmpDir, "multipkg")
	_ = os.MkdirAll(pkgDir, 0755)
	_ = os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte("package multipkg\n"), 0644)
	_ = os.WriteFile(filepath.Join(pkgDir, "helper.go"), []byte("package multipkg\n"), 0644)
	_ = os.WriteFile(filepath.Join(pkgDir, "main_test.go"), []byte("package multipkg\n"), 0644)
	_ = os.WriteFile(filepath.Join(pkgDir, "helper_test.go"), []byte("package multipkg\n"), 0644)

	result, err := Check(tmpDir)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Check() total = %d, want 1", result.Total)
	}

	if len(result.Packages) != 1 {
		t.Fatalf("Check() packages = %d, want 1", len(result.Packages))
	}

	pkg := result.Packages[0]
	if pkg.GoFiles != 2 {
		t.Errorf("Check() goFiles = %d, want 2", pkg.GoFiles)
	}

	if len(pkg.TestFiles) != 2 {
		t.Errorf("Check() testFiles = %d, want 2", len(pkg.TestFiles))
	}
}
