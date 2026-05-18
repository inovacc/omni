# Testing Patterns

**Analysis Date:** 2026-04-11

## Test Framework

**Runner:**
- `testing` package (Go standard library)
- Config: None required (uses Go defaults)

**Assertion Library:**
- None (direct error/output comparison, no assertion package imported)

**Run Commands:**
```bash
go test -race -cover ./...              # Run all tests with race detection
go test -race ./internal/cli/pipe/...   # Run specific package tests
go test -race ./... -run TestSubstituteVariables -v  # Run specific test
go test -race -coverprofile=coverage.out ./...       # Generate coverage report
go tool cover -html=coverage.out                     # View coverage HTML
```

**CI:**
- Tool: GitHub Actions with reusable workflow
- Config: `.github/workflows/test.yml` → `inovacc/workflows/.github/workflows/reusable-go-check.yml`
- Checks: golangci-lint, gofmt, govulncheck, go test -race
- Timeout: 15m for tests, 30m overall

## Test File Organization

**Location:**
- Colocated with source code
- Pattern: `<name>_test.go` in same package as `<name>.go`
- Example: `internal/cli/hash/hash_test.go` tests `internal/cli/hash/hash.go`

**Naming:**
- File: `<name>_test.go` (always suffix, never prefix)
- Function: `Test<FunctionName>(t *testing.T)` (PascalCase with Test prefix)
- Subtests: `t.Run("descriptive name", func(t *testing.T) { ... })`

**Structure:**
```
internal/cli/<command>/
├── <command>.go          # Source
├── <command>_test.go     # Tests for main Run function
├── <command>_unix.go     # Platform-specific source
└── <command>_unix_test.go # Platform-specific tests (if needed)
```

## Test Structure

**Suite Organization:**

```go
func TestRunHead(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "head_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Helper function
	createTestFile := func(name string, numLines int) string {
		file := filepath.Join(tmpDir, name)
		// ... create file
		return file
	}

	// Table-driven tests
	tests := []struct {
		name      string
		args      []string
		opts      HeadOptions
		contains  string
		wantError bool
	}{
		{
			name:     "default 10 lines",
			args:     []string{testFile},
			opts:     HeadOptions{},
			contains: "line",
		},
		// more cases...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := RunHead(&buf, nil, tt.args, tt.opts)
			
			if (err != nil) != tt.wantError {
				t.Errorf("RunHead() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if tt.contains != "" && !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("RunHead() output = %v, want contains %v", buf.String(), tt.contains)
			}
		})
	}
}
```

**Patterns:**

**Setup and Teardown:**
- Setup: Inline before tests or in helper functions
- Teardown: Deferred cleanup with anonymous func:
```go
defer func() { _ = os.RemoveAll(tmpDir) }()  // mute close/remove errors
```

**Assertion Pattern:**
```go
if (err != nil) != tt.wantError {
	t.Errorf("RunHash() error = %v, wantError %v", err, tt.wantError)
	return
}

if strings.Contains(buf.String(), tt.contains) {
	t.Errorf("output mismatch: got %q, want %q", buf.String(), expected)
}
```

**Output Verification:**
```go
var buf bytes.Buffer
err := Run(&buf, args, opts)
output := buf.String()
// Compare output directly
```

## Mocking

**Framework:** None (use composition and interfaces)

**Patterns:**
- Mock `io.Reader` and `io.Writer` with `bytes.Buffer` for I/O
- Mock file operations with temporary directories (`t.TempDir()`)
- Mock functions by passing them as parameters (functional approach)

**What to Mock:**
- I/O: Always mock with `bytes.Buffer` or `strings.NewReader`
- File system: Use `t.TempDir()` for isolated tests
- External services: Use test fixtures (JSON files, YAML files)

**What NOT to Mock:**
- Standard library errors (`os.ErrNotExist`)
- Time-based operations (use real time, control sleep duration if testing)
- Internal command logic (test via public interface)

**Example - Mocking stdin/stdout:**
```go
var buf bytes.Buffer
err := RunHead(&buf, strings.NewReader("line1\nline2\n"), []string{"-"}, opts)
output := buf.String()
```

**Example - Mocking file system:**
```go
tmpDir := t.TempDir()  // auto-cleanup
testFile := filepath.Join(tmpDir, "test.txt")
os.WriteFile(testFile, []byte("content"), 0644)
// file auto-cleaned after test
```

## Fixtures and Factories

**Test Data:**

Factory pattern for building test objects:
```go
// Helper to create test files
createTestFile := func(name string, content string) string {
	file := filepath.Join(tmpDir, name)
	_ = os.WriteFile(file, []byte(content), 0644)
	return file
}

// Usage
file := createTestFile("input.txt", "line1\nline2\n")
```

**Location:**
- Colocated in `_test.go` files
- Embedded as helper functions within test functions
- No separate fixture directory (use temp directories)

**Data Files:**
- JSON fixtures: Embedded in test code as literals or loaded from `.json` files in test dir
- Large test data: Use `testdata/` subdirectory (sibling to `_test.go`)
- Snapshots: Use golden master system (see below)

## Coverage

**Requirements:** ~25.8% overall (skewed by vendored buf packages); ~78% average for owned `pkg/` packages

**View Coverage:**
```bash
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
go tool cover -func=coverage.out | grep total  # Show summary
```

**Coverage Config:**
- No enforcement in CI (informational only)
- Aim for high coverage on public APIs
- Platform-specific code may be excluded via build tags

## Test Types

**Unit Tests:**
- Scope: Single function or struct method
- Approach: Table-driven with `t.Run()` subtests
- Location: `<name>_test.go` in same package
- Example: `TestRunHash()`, `TestParseColorSpec()`

**Integration Tests:**
- Scope: Multiple packages or commands working together
- Approach: End-to-end command execution
- Location: `testing/scripts/` or `testing/golden/`
- Example: Golden master tests, black-box CLI tests

**E2E Tests:**
- Framework: Python scripts in `testing/scripts/` (comparison with reference tools)
- Tools: `test_video.py` (omni video vs yt-dlp comparison)
- CI: Docker container for deterministic environment
- Location: `testing/scripts/`

## Golden Master Tests

Characterization tests that capture exact command outputs as snapshots and detect regressions.

**Run Tests:**
```bash
# Verify outputs match snapshots
python testing/scripts/test_golden.py

# After intentional output changes, regenerate snapshots
python testing/scripts/test_golden.py --update

# Update specific category only
python testing/scripts/test_golden.py --update encoding

# CI pre-flight: verify all snapshot files exist
python testing/scripts/test_golden.py --check

# Run with verbose diffs on failure
python testing/scripts/test_golden.py --verbose

# List all registered tests
python testing/scripts/test_golden.py --list

# Filter tests by substring
python testing/scripts/test_golden.py --filter base64
```

**Registry:**
- Location: `testing/golden/golden_tests.yaml` (lightweight system)
- Also: `tools/golden/golden_tests.yaml` (full system)
- Format: YAML with categories and test cases
- Each test: name, args, optional stdin, normalizations

**Test Definition (YAML):**
```yaml
categories:
  - name: encoding
    tests:
      - name: base64_encode
        args: ["base64"]
        stdin: "hello world"
        
      - name: base64_decode
        args: ["base64", "-d"]
        stdin: "aGVsbG8gd29ybGQ="
```

**Snapshots:**
- Location: `testing/golden/snapshots/`
- Files: `<category>_<name>.stdout` (command output)
- Metadata: `<category>_<name>.json` (test metadata)
- Auto-generated by `--update` flag
- Committed to git for regression detection

**Adding Golden Tests:**
1. Add test case to `testing/golden/golden_tests.yaml` AND `tools/golden/golden_tests.yaml`
2. Run `python testing/scripts/test_golden.py --update` to generate snapshots
3. Review generated `.stdout` files
4. Commit both YAML and snapshot files

**Full-Featured System (tools/golden/):**
```bash
task golden:compare             # Compare with SHA-256 fast path
task golden:record              # Record baselines with manifest
task golden:list                # List test cases
task golden:map:table           # Test map as table
task golden:docker:build        # Build Docker image for testing
```

**Lightweight System (testing/golden/):**
```bash
task docker:test:golden          # Verify in Linux container
task docker:test:golden:update   # Regenerate in container
```

## Common Patterns

**Async Testing:**
```go
// Test goroutine-based operations with context
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()

// Run goroutine
done := make(chan error, 1)
go func() {
	done <- someAsyncFunc(ctx)
}()

select {
case err := <-done:
	if err != nil {
		t.Errorf("async failed: %v", err)
	}
case <-ctx.Done():
	t.Errorf("timeout")
}
```

**Error Testing:**
```go
// Test error classification
err := RunHash(&buf, []string{"/nonexistent"}, opts)

// Check if classified correctly
if !errors.Is(err, os.ErrNotExist) {
	t.Errorf("want ErrNotExist, got %v", err)
}

// Or check wrapped sentinel
var exitErr *cmderr.ExitError
if !errors.As(err, &exitErr) {
	t.Errorf("want ExitError")
}
```

**Temporary Directory:**
```go
// Automatic cleanup
tmpDir := t.TempDir()
file := filepath.Join(tmpDir, "test.txt")
os.WriteFile(file, []byte("content"), 0644)
// tmpDir auto-removed when test ends
```

**JSON Output Testing:**
```go
var buf bytes.Buffer
err := RunHash(&buf, args, HashOptions{OutputFormat: output.FormatJSON})
if err != nil {
	t.Fatal(err)
}

var result HashesResult
if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
	t.Fatalf("unmarshal: %v", err)
}

if result.Count != 2 {
	t.Errorf("got %d hashes, want 2", result.Count)
}
```

## Test Coverage

**Coverage Files:**
- 100+ test files in `internal/cli/*/`
- 40+ test files in `pkg/*/`
- All public APIs have unit tests
- Golden master tests (117 total) for CLI integration

**Key Areas Covered:**
- All encoding formats (base64, base32, base58, hex, url)
- Hash algorithms (MD5, SHA256, SHA512, CRC32, CRC64)
- Text operations (head, tail, sort, uniq, cut, tr, paste, fold, nl, column, join)
- Search (grep, rg with gitignore support, parallel walking)
- File operations (cp, mv, rm, ln, stat, find)
- JSON/YAML/XML operations (jq, yq, validation, conversion)
- Video download (format selection, HLS, YouTube signature decryption)
- Tree scanning (parallel, max-files, hash, compare)
- Error classification (cmderr sentinels)
- Command parsing (pipe, pipeline, platform-specific flags)
- Gitignore parsing (negation, directory-only, double-glob patterns)

**Gaps:**
- Some cloud integration commands (AWS, Kubernetes) have minimal tests
- Complex video extractor paths not fully covered
- Performance benchmarks exist but not comprehensive

---

*Testing analysis: 2026-04-11*
