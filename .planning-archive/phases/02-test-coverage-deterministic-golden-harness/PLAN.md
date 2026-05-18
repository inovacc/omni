# Phase 2 — Test Coverage + Deterministic Golden Harness
# Master Plan Index

**Phase goal:** Raise test coverage to v1.0 targets AND land golden-master normalization hooks
(timestamps, random IDs) that supply-chain commands will require in Phases 4–7.

**Requirements:** POLISH-04, POLISH-05, POLISH-06, POLISH-07, POLISH-11, POLISH-12, POLISH-13, POLISH-14

**Wave structure:**
- Wave 1: Infrastructure tools (covgate, helplint, golden normalize hook) — run in parallel
- Wave 2: Coverage depth + breadth (pkg/ gaps, internal/cli/ baseline, golden entries) — after Wave 1
- Wave 3: Cleanup + final gate enforcement — after Wave 2

---

## Decision Coverage Matrix

| Decision | Plan | Task | Coverage |
|----------|------|------|----------|
| D-01: covgate dual-rule tool + task lint:coverage | 01 | 1-2 | Full |
| D-02: Golden normalize: YAML hook, both engines | 02 | 1-2 | Full |
| D-03: helplint tool + task lint:help + fix 3 gaps | 03 | 1-2 | Full |
| D-04: pkg/ Wave 1 API baseline (0-test packages) | 04 | 1-3 | Full |
| D-05: pkg/video/extractor/youtube 4% → ≥40% | 05 | 1 | Full |
| D-06: pkg/video/downloader 32.9% → ≥40% | 05 | 2 | Full |
| D-07: pkg/jsonutil 67.5% → ≥75% | 06 | 1 | Full |
| D-08: pkg/twig 44.3% → ≥75% | 06 | 2 | Full |
| D-09: internal/cli/ baseline measure + straggler fix | 07 | 1-2 | Full |
| D-10: Happy-path golden entries + normalize hooks | 08 | 1-3 | Full |
| D-11: cmdtree regen + pre-commit hook | 09 | 1-2 | Full |
| D-12: Final coverage gate run + wave 3 fixes | 10 | 1-2 | Full |

---

## Plans

---

### Plan 02-01: tools/covgate/ — Dual-Rule Coverage Gate

```yaml
plan_id: "02-01"
name: "tools/covgate/ — dual-rule coverage gate"
wave: 1
requirements: [POLISH-04, POLISH-05, POLISH-06]
depends_on: []
files_modified:
  - tools/covgate/main.go
  - Taskfile.yml
  - .github/workflows/test.yml
autonomous: true
must_haves:
  truths:
    - "task lint:coverage runs and exits non-zero when any pkg/ package is below 40%"
    - "task lint:coverage runs and exits non-zero when pkg/ weighted avg is below 75%"
    - "pkg/private/ subtree is excluded from all counts"
    - "CI job calls lint:coverage and fails the build on violation"
  artifacts:
    - path: "tools/covgate/main.go"
      provides: "Dual-rule coverage gate binary"
    - path: "Taskfile.yml"
      provides: "lint:coverage target"
  key_links:
    - from: "Taskfile.yml lint:coverage"
      to: "tools/covgate/main.go"
      via: "go run ./tools/covgate"
```

**Tasks:**

**Task 1: Implement tools/covgate/main.go**

Files: `tools/covgate/main.go`

Action: Create a stdlib-only Go program (no exec, pure profile parsing) that:
- Accepts flags: `-profile string` (coverage.out path), `-pkg-prefix string` (e.g. `pkg/`),
  `-avg-min float` (weighted average threshold), `-floor float` (per-package floor),
  `-exclude string` (comma-separated path substrings to skip, default `private`)
- Parses the Go coverage profile format directly (no `go tool cover` subprocess):
  Each non-header line: `github.com/inovacc/omni/pkg/foo/foo.go:25.45,27.2 2 1`
  Fields: `file:start,end stmts count`
- Aggregates per-package: sum `stmts` into `total_stmts`, sum `stmts` where `count > 0` into `covered_stmts`
- Skips packages where path contains any string from `-exclude` (e.g. `/private/`)
- Skips packages where path does NOT contain `-pkg-prefix`
- Computes weighted average: `sum(covered) / sum(total) * 100`
- Checks per-package floor: any package with `covered/total*100 < floor` is a violator
- Emits human-readable report to stdout: weighted avg, pass/fail, worst-5 offenders by coverage %
- Exit 0 if all rules pass, exit 1 on any violation, exit 2 on bad flags/I/O

Error handling: use `fmt.Fprintf(os.Stderr, ...)` + `os.Exit(2)` for I/O errors; `os.Exit(1)` for rule failures.

Follow the `tools/cmderr-cov/main.go` style: `//nolint` where needed, `defer func() { _ = f.Close() }()`.

Verify: `go run ./tools/covgate -help` prints usage without error.

Done: Binary compiles cleanly (`go build ./tools/covgate`), help flag works, runs against a synthetic two-line profile and exits 0/1 correctly.

---

**Task 2: Wire into Taskfile + CI**

Files: `Taskfile.yml`, `.github/workflows/test.yml`

Action:
- Add `lint:coverage` task to `Taskfile.yml` (after the existing `lint:cmderr-coverage` task):
  ```yaml
  lint:coverage:
    desc: "Enforce dual-rule coverage gates (avg + floor) for pkg/ and internal/cli/"
    cmds:
      - go test -coverprofile={{.COVERAGE_FILE | default "coverage.out"}} -short ./pkg/... ./internal/cli/...
      - go run ./tools/covgate -profile={{.COVERAGE_FILE | default "coverage.out"}} -pkg-prefix=github.com/inovacc/omni/pkg/ -avg-min=75 -floor=40
      - go run ./tools/covgate -profile={{.COVERAGE_FILE | default "coverage.out"}} -pkg-prefix=github.com/inovacc/omni/internal/cli/ -avg-min=60 -floor=30
  ```
- Add CI job in `.github/workflows/test.yml` as a new step named `coverage-gate` that runs after the test step:
  `run: task lint:coverage`
- Keep existing `lint:cmderr-coverage` target unchanged.

Verify: `grep -n "lint:coverage" Taskfile.yml` finds the new target. `grep -n "lint:coverage" .github/workflows/test.yml` finds the CI step.

Done: `task lint:coverage` runs end-to-end and reports pass or fail on the current codebase.

---

---

### Plan 02-02: Golden normalize: Hooks — Both Harness Engines

```yaml
plan_id: "02-02"
name: "Golden normalize: regex hook for both harness engines"
wave: 1
requirements: [POLISH-13]
depends_on: []
files_modified:
  - testing/golden_engine.py
  - tools/golden/src/golden/types.py
  - tools/golden/src/golden/normalize.py
  - tools/golden/src/golden/runner.py
  - docs/GOLDEN_MASTER_TESTING.md
autonomous: true
must_haves:
  truths:
    - "A golden test entry with normalize: [{pattern, replacement}] applies the regex substitution to stdout before diff"
    - "Existing entries without normalize: behave identically to before"
    - "Both engines produce byte-identical normalization results for the same input"
    - "docs/GOLDEN_MASTER_TESTING.md has a normalize: examples section"
  artifacts:
    - path: "testing/golden_engine.py"
      provides: "Engine 1 with normalize: support"
    - path: "tools/golden/src/golden/normalize.py"
      provides: "Engine 2 normalize function with regex rules"
  key_links:
    - from: "golden_tests.yaml normalize: field"
      to: "testing/golden_engine.py GoldenEngine.normalize()"
      via: "GoldenTestCase.normalize list"
```

**Tasks:**

**Task 1: Patch testing/golden_engine.py (Engine 1)**

Files: `testing/golden_engine.py`

Action:
1. In `GoldenTestCase` dataclass, add field after `normalizations`:
   `normalize: list[dict] = field(default_factory=list)  # {pattern, replacement} pairs`
2. In `GoldenEngine.normalize()` (or equivalent method), after applying named normalizers, add:
   ```python
   import re  # ensure import at top
   for rule in (normalize_rules or []):
       text = re.sub(rule["pattern"], rule["replacement"], text)
   ```
   Update method signature to accept `normalize_rules: list[dict] | None = None`.
3. In `load_registry()` (YAML loading), add `normalize=test.get("normalize", [])` to `GoldenTestCase(...)` constructor.
4. In `record()` and `compare()` callers of `normalize()`, pass `test.normalize` as the extra argument.

Do NOT change any existing `normalizations` (named-normalizer) behavior. Existing tests must still pass.

Verify: `python testing/golden_engine.py --list` runs without error. Manually trace: given a fake test entry with `normalize: [{pattern: "\\d+", replacement: "NUM"}]`, the `normalize()` method replaces digits in "abc123" with "NUM".

Done: All existing golden tests still pass via `task test:golden`. New field accepted without error.

---

**Task 2: Patch tools/golden/src/golden/ (Engine 2) + docs**

Files: `tools/golden/src/golden/types.py`, `tools/golden/src/golden/normalize.py`, `tools/golden/src/golden/runner.py`, `docs/GOLDEN_MASTER_TESTING.md`

Action:
1. In `tools/golden/src/golden/types.py`, add to `TestCase` dataclass:
   `normalize: list[dict] = field(default_factory=list)  # {pattern, replacement} pairs`
2. In `tools/golden/src/golden/normalize.py`, update the `normalize()` function:
   - Add `normalize_rules: list[dict] | None = None` parameter
   - After the named-normalizer loop, add the regex substitution loop (same logic as Task 1)
   - Ensure `import re` is present at the top
3. In `tools/golden/src/golden/runner.py`, update all calls to `normalize()` to pass `test_case.normalize`
4. Find the YAML-loading module (likely `discovery.py` or `config.py` in same directory) and add `normalize=test.get("normalize", [])` to the `TestCase(...)` constructor call.
5. In `docs/GOLDEN_MASTER_TESTING.md`, add a new section "## normalize: Hook" with 5 example YAML snippets:
   - Timestamp pattern: `'\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z'` → `'{TIMESTAMP}'`
   - UUID v4: `'[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}'` → `'{UUID}'`
   - ULID: `'[0-9A-Z]{26}'` → `'{ULID}'`
   - Temp path (Unix): `'/tmp/[^\s]+'` → `'{TMPPATH}'`
   - Random port: `':\d{4,5}'` → `':{PORT}'`

Verify: `python tools/golden/src/golden/run_tests.py --list` (or equivalent entry point) runs without error. The docs section renders correctly in markdown preview.

Done: Engine 2 accepts `normalize:` in YAML, applies regex substitution, existing tests unaffected.

---

---

### Plan 02-03: tools/helplint/ — Help Docstring Linter

```yaml
plan_id: "02-03"
name: "tools/helplint/ — cobra help docstring linter + fix 3 known gaps"
wave: 1
requirements: [POLISH-11]
depends_on: []
files_modified:
  - tools/helplint/main.go
  - cmd/copy.go
  - cmd/move.go
  - cmd/remove.go
  - Taskfile.yml
  - .github/workflows/test.yml
autonomous: true
must_haves:
  truths:
    - "task lint:help exits non-zero listing cmd/copy.go, cmd/move.go, cmd/remove.go as violations BEFORE the fix"
    - "After fixing the 3 gaps, task lint:help exits 0"
    - "helplint checks Short: non-empty AND Long: contains at least one line starting with 'omni '"
  artifacts:
    - path: "tools/helplint/main.go"
      provides: "Help docstring linter binary"
    - path: "Taskfile.yml"
      provides: "lint:help target"
  key_links:
    - from: "Taskfile.yml lint:help"
      to: "tools/helplint/main.go"
      via: "go run ./tools/helplint"
```

**Tasks:**

**Task 1: Implement tools/helplint/main.go**

Files: `tools/helplint/main.go`

Action: Create a stdlib-only Go AST-based linter that:
- Accepts flags: `-dir string` (directory to walk, default `cmd/`), `-verbose bool`
- Uses `go/parser` + `go/ast` to parse each `*.go` file under the directory
- For each `cobra.Command` struct literal found (identified by the type `*cobra.Command` in composite literals or var declarations), extracts values of `Short` and `Long` fields from the AST
- Validates: `Short` is a non-empty string literal; `Long` is a non-empty string literal containing at least one occurrence of the substring `omni ` (space after omni, case-sensitive)
- Reports violations in format: `cmd/copy.go: command 'cp' missing Long: field with omni example`
- Exit 0 if no violations, exit 1 if any violations found, exit 2 on parse errors

Implementation approach: Walk AST for `ast.CompositeLit` nodes where `Type` resolves to `cobra.Command`. For each, find `ast.KeyValueExpr` where `Key` is an `ast.Ident` with name `Short` or `Long`. Extract the string value via `ast.BasicLit`.

Note: Some commands define `Long:` via a variable. For v1 of this tool, only detect literal string assignments — skip commands where the value is a variable reference (to avoid false positives). Document this limitation in the tool's help text.

Verify: `go run ./tools/helplint -dir cmd/` before the fix lists at least `copy.go`, `move.go`, `remove.go`. After fixing those three files (Task 2), it exits 0.

Done: Tool compiles, runs against `cmd/`, correctly identifies the 3 known violators.

---

**Task 2: Fix the 3 known gaps + wire into Taskfile + CI**

Files: `cmd/copy.go`, `cmd/move.go`, `cmd/remove.go`, `Taskfile.yml`, `.github/workflows/test.yml`

Action:
1. In `cmd/copy.go`: Add or fill `Long:` field on the Cobra command with a concrete example:
   ```go
   Long: `Copy SOURCE to DEST, or multiple SOURCE(s) to DIRECTORY.

   omni cp file.txt /tmp/backup.txt
   omni cp -r src/ dest/
   `,
   ```
2. In `cmd/move.go`: Add `Long:` with example:
   ```go
   Long: `Move (rename) SOURCE to DEST, or move multiple SOURCE(s) to DIRECTORY.

   omni mv old.txt new.txt
   omni mv *.log /var/logs/
   `,
   ```
3. In `cmd/remove.go`: Add `Long:` with example:
   ```go
   Long: `Remove (delete) files or directories.

   omni rm file.txt
   omni rm -r old_dir/
   `,
   ```
4. Add `lint:help` task to `Taskfile.yml`:
   ```yaml
   lint:help:
     desc: "Verify all Cobra commands have Short: and Long: with at least one omni example"
     cmds:
       - go run ./tools/helplint -dir cmd/
   ```
5. Add CI step in `.github/workflows/test.yml`:
   `run: task lint:help`

Verify: `go run ./tools/helplint -dir cmd/` exits 0 after edits. `grep -n "Long:" cmd/copy.go` shows a non-empty value containing "omni".

Done: All 3 gaps fixed; `task lint:help` exits 0; CI includes the check.

---

---

### Plan 02-04: Wave 1 API Baseline Tests for pkg/ Packages

```yaml
plan_id: "02-04"
name: "API baseline test files for pkg/ packages below coverage floor"
wave: 2
requirements: [POLISH-07]
depends_on: []
files_modified:
  - pkg/userdirs/userdirs_api_test.go
  - pkg/video/video_api_test.go
  - pkg/video/format/format_api_test.go
  - pkg/video/utils/utils_api_test.go
  - pkg/video/nethttp/nethttp_api_test.go
  - pkg/video/cache/cache_api_test.go
  - pkg/twig/builder/builder_api_test.go
autonomous: true
must_haves:
  truths:
    - "Every targeted pkg/ package has at least one *_api_test.go file using package foo_test (black-box)"
    - "go test ./pkg/... passes with no compilation errors"
    - "Each new test file calls at least one exported function and asserts a non-error result"
  artifacts:
    - path: "pkg/userdirs/userdirs_api_test.go"
      provides: "API contract baseline for userdirs"
    - path: "pkg/video/format/format_api_test.go"
      provides: "API contract baseline for video/format"
  key_links:
    - from: "pkg/userdirs/userdirs_api_test.go"
      to: "pkg/userdirs"
      via: "import github.com/inovacc/omni/pkg/userdirs"
```

**Tasks:**

**Task 1: Baseline tests for pkg/userdirs and pkg/twig/builder**

Files: `pkg/userdirs/userdirs_api_test.go`, `pkg/twig/builder/builder_api_test.go`

Action:
- Both test files use `package userdirs_test` / `package builder_test` (black-box style per D-07).
- Import the package by full module path: `github.com/inovacc/omni/pkg/userdirs`.
- `pkg/userdirs`: Read `userdirs.go` to find exported types/functions. Create one test that calls the primary entry point (likely `GetUserDirs()` or similar), asserts result is non-nil or non-error. If it returns a struct, assert at least one field is non-empty.
- `pkg/twig/builder`: Read `builder.go` to find exported API. Create one test that constructs the builder, calls a build or render method with minimal input, asserts non-error.
- Use `t.Helper()` pattern; table-driven if there are multiple cases, otherwise a simple function test.

Verify: `go test ./pkg/userdirs/... ./pkg/twig/builder/...` passes.

Done: Both packages have `_api_test.go` files; `go test` passes; coverage increases from baseline.

---

**Task 2: Baseline tests for pkg/video sub-packages (format, utils, nethttp, cache)**

Files: `pkg/video/format/format_api_test.go`, `pkg/video/utils/utils_api_test.go`, `pkg/video/nethttp/nethttp_api_test.go`, `pkg/video/cache/cache_api_test.go`

Action: For each package, read the main `.go` file to identify the primary exported type/function. Write a black-box test (`package foo_test`) that:
- Constructs the primary type with safe zero values or minimal config
- Calls the primary non-network entry point
- Asserts non-error or expected value

Specific guidance per package:
- `pkg/video/format`: Test `SelectFormat()` or `SortFormats()` with a small slice of fake `Format` structs. Assert the returned format has expected properties.
- `pkg/video/utils`: Call at least 2 exported utility functions (sanitize filename, URL join, or similar). Assert deterministic outputs.
- `pkg/video/nethttp`: Test `NewClient()` construction; assert client is non-nil. Do NOT make network calls.
- `pkg/video/cache`: Test `New()` or `NewCache()` with a temp directory. Assert cache is non-nil.

No network calls in any test. Use `t.TempDir()` for filesystem tests.

Verify: `go test ./pkg/video/format/... ./pkg/video/utils/... ./pkg/video/nethttp/... ./pkg/video/cache/...` passes.

Done: 4 new test files; all packages show increased coverage; `go test` passes.

---

**Task 3: Baseline test for pkg/video root package**

Files: `pkg/video/video_api_test.go`

Action: Read `pkg/video/video.go` to find exported functions (likely `NewClient()`, `Extract()`, or `Download()`). Write `package video_test` baseline:
- Test `NewClient()` with default options — assert client is non-nil.
- Test a pure-Go utility function in the package (not `Extract`/`Download` which require network).
- If the only entry points are network-dependent, create a test that constructs a `Client` with a custom `http.Client` using `httptest.NewServer` serving a minimal fake response, and call `Extract()` with the test server URL — assert error handling works (e.g., unsupported URL returns a typed error).

No real network calls. Use `httptest` from stdlib if network simulation is needed.

Verify: `go test -short ./pkg/video` passes.

Done: `pkg/video_api_test.go` exists; `go test -short` passes; coverage increases from 46.0%.

---

---

### Plan 02-05: Coverage Depth — pkg/video/extractor/youtube + pkg/video/downloader

```yaml
plan_id: "02-05"
name: "Coverage depth: youtube extractor (4%) + downloader (32.9%) → ≥40%"
wave: 2
requirements: [POLISH-04]
depends_on: []
files_modified:
  - pkg/video/extractor/youtube/youtube_test.go
  - pkg/video/downloader/downloader_test.go
  - pkg/video/downloader/http_test.go
autonomous: true
must_haves:
  truths:
    - "pkg/video/extractor/youtube coverage reaches ≥40%"
    - "pkg/video/downloader coverage reaches ≥40%"
    - "No test makes real network calls (all tests pass with -short flag)"
  artifacts:
    - path: "pkg/video/extractor/youtube/youtube_test.go"
      provides: "YouTube extractor tests (URL matching, pure-Go logic)"
    - path: "pkg/video/downloader/http_test.go"
      provides: "HTTP downloader unit tests"
  key_links:
    - from: "pkg/video/extractor/youtube/youtube_test.go"
      to: "pkg/video/extractor/youtube"
      via: "package youtube_test"
```

**Tasks:**

**Task 1: Tests for pkg/video/extractor/youtube**

Files: `pkg/video/extractor/youtube/youtube_test.go`

Action: Read `pkg/video/extractor/youtube/` directory files to understand the exported surface. Focus on pure-Go testable functions:
1. URL matching — the extractor's `Match(url string) bool` method. Test with 10+ YouTube URL patterns: `https://www.youtube.com/watch?v=...`, `https://youtu.be/...`, `https://youtube.com/shorts/...`, playlist URLs, channel URLs, and non-YouTube URLs (assert false).
2. Video ID extraction — if there's a `extractVideoID(url)` helper or equivalent, test it with the same URL patterns.
3. Duration/view count parsing — from `channel_test.go` (already referenced in CLAUDE.md), there are `parseDuration` and `parseViewCount` functions. Add tests for edge cases: empty string, malformed input, valid ISO8601 duration.
4. Metadata struct construction — if there's a `NewExtractor()` or `Register()` that returns an interface, call it and assert non-nil.

Use `package youtube_test` for exported API tests; use `package youtube` (white-box) only if internals must be accessed for duration/view count parsing (those may be unexported).

Verify: `go test -short ./pkg/video/extractor/youtube/...` passes. `go test -cover -short ./pkg/video/extractor/youtube/...` shows ≥40%.

Done: Coverage ≥40%; all tests pass with `-short`; no network calls.

---

**Task 2: Tests for pkg/video/downloader**

Files: `pkg/video/downloader/http_test.go`, `pkg/video/downloader/downloader_test.go`

Action: Read existing `pkg/video/downloader/downloader_test.go` (already has some tests per CLAUDE.md). The existing tests cover `SelectDownloader` type assertions. Add to `http_test.go`:
1. Test `httptest.NewServer` scenario: serve a small response body (100 bytes). Call the HTTP downloader's download logic with the test server URL to a `t.TempDir()` file. Assert file exists with expected content.
2. Test resume logic: create a partial `.part` file, serve the remaining bytes with a `Range` header responder in `httptest.Server`. Assert the final file has correct full content.
3. Test error paths: server returns 404 → assert `cmderr.ErrNotFound` or equivalent error type; server returns 500 → assert generic error.
4. Test `FormatBytes`, `FormatSpeed`, `FormatETA`, `FormatPercent` from the progress subpackage (already tested per CLAUDE.md — skip if coverage already ≥40%).

Use `package downloader_test` (black-box). Do NOT make real network calls.

Verify: `go test -cover -short ./pkg/video/downloader/...` shows ≥40%.

Done: Combined coverage ≥40%; resume test passes; error paths covered.

---

---

### Plan 02-06: Coverage Depth — pkg/jsonutil + pkg/twig root

```yaml
plan_id: "02-06"
name: "Coverage depth: pkg/jsonutil (67.5%→≥75%) + pkg/twig root (44.3%→≥75%)"
wave: 2
requirements: [POLISH-04]
depends_on: []
files_modified:
  - pkg/jsonutil/jsonutil_test.go
  - pkg/twig/twig_test.go
autonomous: true
must_haves:
  truths:
    - "pkg/jsonutil coverage reaches ≥75%"
    - "pkg/twig (root package) coverage reaches ≥75%"
    - "go test ./pkg/jsonutil/... ./pkg/twig/... passes"
  artifacts:
    - path: "pkg/jsonutil/jsonutil_test.go"
      provides: "Extended jsonutil tests covering uncovered paths"
    - path: "pkg/twig/twig_test.go"
      provides: "twig root package tests"
  key_links:
    - from: "pkg/jsonutil/jsonutil_test.go"
      to: "pkg/jsonutil"
      via: "package jsonutil_test"
```

**Tasks:**

**Task 1: Extend pkg/jsonutil tests to ≥75%**

Files: `pkg/jsonutil/jsonutil_test.go`

Action: Read the existing `pkg/jsonutil/jsonutil_test.go` to see what's already tested (Query, QueryString, QueryReader, ApplyFilter per CLAUDE.md). Then read `pkg/jsonutil/jsonutil.go` and any other `.go` files to find uncovered paths. Add tests targeting:
1. Error paths in `Query()`: invalid JSON input → assert error; valid JSON with non-existent key path → assert nil or empty result (depending on API).
2. `QueryReader()` with an `io.Reader` wrapping a malformed JSON string → assert error propagation.
3. `ApplyFilter()` with multiple filter expressions if supported.
4. Any streaming or array-traversal functions not yet tested.
5. Edge cases: empty JSON `{}`, JSON array root `[]`, deeply nested paths.

Extend the existing test file — do NOT replace it. Add new `TestXxx` functions.

Verify: `go test -cover ./pkg/jsonutil/...` shows ≥75%.

Done: Coverage ≥75%; all existing tests still pass; new tests cover error + edge paths.

---

**Task 2: Add tests for pkg/twig root package**

Files: `pkg/twig/twig_test.go`

Action: Read `pkg/twig/` directory (root package files, not sub-packages) to understand what's exported at the root level (likely `Tree`, `Generate`, `Options`, or orchestration functions that wire `scanner` + `formatter`). Create `pkg/twig/twig_test.go` with `package twig_test`:
1. Call the primary entry point (e.g., `Generate(dir, options)`) with `t.TempDir()` as input directory containing 3–5 synthetic files. Assert non-error and non-empty output.
2. Test with `Options{JSON: true}` — assert output is valid JSON.
3. Test with a non-existent directory — assert error.
4. Test the compare entry point if exported at root level.
5. If the root package is just a re-export facade, test each re-exported function once.

Verify: `go test -cover ./pkg/twig` shows ≥75%.

Done: Coverage ≥75%; new test file compiles; all tests pass.

---

---

### Plan 02-07: internal/cli/ Coverage Baseline + Straggler Fix

```yaml
plan_id: "02-07"
name: "internal/cli/ coverage measurement + straggler packages to ≥30% floor"
wave: 2
requirements: [POLISH-05, POLISH-06]
depends_on: ["02-01"]
files_modified:
  - internal/cli/note/note_test.go
  - internal/cli/tomlutil/tomlutil_test.go
  - internal/cli/xmlfmt/xmlfmt_test.go
  - internal/cli/jsonfmt/jsonfmt_test.go
autonomous: true
must_haves:
  truths:
    - "go test -cover ./internal/cli/... produces coverage numbers for all packages"
    - "No internal/cli/ package falls below 30% floor after this plan"
    - "The weighted average for internal/cli/ is measured and reported"
  artifacts:
    - path: "internal/cli/note/note_test.go"
      provides: "Baseline tests for note package"
    - path: "internal/cli/tomlutil/tomlutil_test.go"
      provides: "Baseline tests for tomlutil package"
  key_links:
    - from: "internal/cli/ test files"
      to: "pkg/ libraries"
      via: "Run() function delegation"
```

**Tasks:**

**Task 1: Measure internal/cli/ coverage and identify stragglers**

Files: (measurement only — no new files, produces data for Task 2)

Action: Run `go test -cover -short ./internal/cli/...` and capture the per-package coverage output. Parse the output to identify:
- Any package showing `[no test files]` — these need baseline tests
- Any package showing coverage below 30% — these need straggler tests
- Compute the weighted average across all non-`cmderr` packages

Write findings as a comment block at the top of `internal/cli/note/note_test.go` (or a standalone `coverage-findings.txt` in the phase dir) summarizing which packages need attention.

The following packages are pre-identified from git status as recently modified but may lack tests:
- `internal/cli/note/` — check if test exists
- `internal/cli/tomlutil/` — check if test exists
- `internal/cli/xmlfmt/` — check if test exists
- `internal/cli/jsonfmt/` — check if test exists
- `internal/cli/jwt/` — check if test exists
- `internal/cli/pwd/` — check if test exists
- `internal/cli/random/` — already has tests per CLAUDE.md
- `internal/cli/uuid/` — already has tests per CLAUDE.md

Verify: `go test -cover -short ./internal/cli/...` completes without build errors (test failures are OK at this stage).

Done: Coverage numbers recorded; straggler list confirmed.

---

**Task 2: Add baseline tests for straggler internal/cli/ packages**

Files: `internal/cli/note/note_test.go`, `internal/cli/tomlutil/tomlutil_test.go`, `internal/cli/xmlfmt/xmlfmt_test.go`, `internal/cli/jsonfmt/jsonfmt_test.go`

Action: For each package identified as below 30% or missing tests in Task 1, create a minimal test file:
- Use `package <name>_test` (black-box) or `package <name>` if unexported helpers are needed
- Call `Run(w, args, opts)` with representative inputs and assert non-error + non-empty output
- Test at least one error path (invalid input)

Specific guidance:
- `note`: Test `Run()` with a note text argument — assert it writes to the writer without error.
- `tomlutil`: Test `Run()` with valid TOML input — assert formatted output. Test with invalid TOML — assert error.
- `xmlfmt`: Test `Run()` with valid XML — assert formatted output. Test with malformed XML — assert error.
- `jsonfmt`: Test `Run()` with valid JSON — assert formatted output. Test with invalid JSON — assert error.
- `jwt`: If no test exists, test `RunDecode()` with a known JWT string (any public JWT, e.g., `eyJhbGciOiJub25lIn0.eyJzdWIiOiJ0ZXN0In0.`) — assert no panic and output contains the decoded payload.
- `pwd`: Test `Run()` — assert it writes a non-empty path to the writer.

Verify: `go test -cover -short ./internal/cli/...` shows all targeted packages at ≥30%. `task lint:coverage` passes after these tests are added.

Done: No internal/cli/ package below 30% floor; weighted average ≥60%; `task lint:coverage` exits 0.

---

---

### Plan 02-08: Happy-Path Golden Entries with normalize: Hooks

```yaml
plan_id: "02-08"
name: "Happy-path golden entries for all top-level commands + normalize: hooks"
wave: 2
requirements: [POLISH-12, POLISH-13]
depends_on: ["02-02"]
files_modified:
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
autonomous: true
must_haves:
  truths:
    - "Every top-level command has at least one entry in both golden YAML registries"
    - "uuid, ulid, ksuid, snowflake, nanoid entries use normalize: to replace random output"
    - "date command entries use normalize: to replace timestamp output"
    - "jwt decode entry uses normalize: for any timestamp fields"
    - "Regenerated .stdout snapshots exist after task test:golden:update"
  artifacts:
    - path: "testing/golden/golden_tests.yaml"
      provides: "Expanded golden registry with normalize: entries"
    - path: "tools/golden/golden_tests.yaml"
      provides: "Identical expanded registry (synced)"
  key_links:
    - from: "testing/golden/golden_tests.yaml normalize: fields"
      to: "testing/golden_engine.py GoldenEngine.normalize()"
      via: "GoldenTestCase.normalize"
```

**Tasks:**

**Task 1: Add normalize: hooks to existing idgen golden entries**

Files: `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

Action: Find all existing entries in the `idgen` and related categories that capture random output (uuid, ulid, ksuid, nanoid, snowflake). For each, add `normalize:` rules:

```yaml
# uuid v4
normalize:
  - pattern: '[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}'
    replacement: '{UUID_V4}'

# uuid v7
normalize:
  - pattern: '[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}'
    replacement: '{UUID_V7}'

# ulid
normalize:
  - pattern: '[0-9A-Z]{26}'
    replacement: '{ULID}'

# ksuid
normalize:
  - pattern: '[0-9A-Za-z]{27}'
    replacement: '{KSUID}'

# nanoid (default 21 chars)
normalize:
  - pattern: '[A-Za-z0-9_-]{21}'
    replacement: '{NANOID}'

# snowflake (large integer)
normalize:
  - pattern: '\b\d{15,19}\b'
    replacement: '{SNOWFLAKE}'
```

Also add normalize: to `date` command entries:
```yaml
normalize:
  - pattern: '\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})'
    replacement: '{TIMESTAMP}'
  - pattern: '\w{3} \w{3}  ?\d{1,2} \d{2}:\d{2}:\d{2} \w+ \d{4}'
    replacement: '{DATE_STRING}'
```

Both YAML files must be byte-identical for all new/modified entries. Edit both files in the same task.

Verify: YAML is valid (no syntax errors). `python testing/golden_engine.py --list` shows the updated entries.

Done: All idgen + date entries have normalize: hooks; both YAML files updated identically.

---

**Task 2: Add missing top-level command happy-path entries**

Files: `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

Action: Run `omni cmdtree` (or parse `cmd/` directory) to get the full list of top-level commands. Compare against existing golden registry entries. Add a happy-path entry for every top-level command that lacks one.

For commands that are pure routers (aws, kubectl/k, git, video, buf, scaffold), also add one key subcommand entry:
- `aws`: add `omni aws --version` or `omni aws help` (exits 0, no credentials needed)
- `kubectl`/`k`: add `omni kubectl version --client` (exits 0, no cluster needed)
- `git`: add `omni git --help` or `omni gqc --help`
- `video`: add `omni video --help` (not `video info <url>` — no network)
- `buf`: add `omni buf --help`
- `scaffold`: add `omni scaffold --help`

For commands that produce non-deterministic output (random, uuid, date, jwt), ensure `normalize:` is set (done in Task 1 for idgen; apply same pattern here for `random` command).

Entry format:
```yaml
- name: cp_happy
  args: ["cp", "--help"]
  expected_exit: 0
```

For `--help` entries, no normalize: needed (help text is deterministic).

Both YAML files updated identically.

Verify: Every top-level command in `cmd/` has at least one entry. Run `python testing/golden_engine.py --list | wc -l` and confirm count increased.

Done: All top-level commands have ≥1 golden entry; both YAML files consistent.

---

**Task 3: Regenerate .stdout snapshots**

Files: `testing/golden/snapshots/` (multiple `.stdout` files updated)

Action: Run `task test:golden:update` to regenerate all snapshot `.stdout` files. This will:
- Run each command in the registry
- Apply normalize: rules to the captured output
- Write normalized output to the snapshot files

After regeneration, review the new snapshot files:
- Confirm idgen snapshots contain `{UUID_V4}`, `{ULID}`, `{KSUID}` placeholders (not raw random values)
- Confirm date snapshots contain `{TIMESTAMP}` placeholders
- Confirm all new `--help` entries produced non-empty snapshots

If any command fails (exits non-zero when expected_exit is 0), fix the golden entry (wrong args, or add `expected_exit: 1` if `--help` returns 1 in that command).

Verify: `task test:golden` passes after regeneration (no diff failures).

Done: All snapshots regenerated; `task test:golden` exits 0; normalize: hooks produce stable placeholders.

---

---

### Plan 02-09: cmdtree Regen + Pre-Commit Hook

```yaml
plan_id: "02-09"
name: "cmdtree regen one-off commit + hooks:install pre-commit hook"
wave: 2
requirements: [POLISH-14]
depends_on: []
files_modified:
  - .githooks/pre-commit
  - Taskfile.yml
autonomous: true
must_haves:
  truths:
    - "omni cmdtree output is regenerated and committed"
    - "omni aicontext output is regenerated and committed"
    - "task hooks:install installs the .githooks/pre-commit hook"
    - "The pre-commit hook runs omni cmdtree if cmd/ files are staged and fails if output changed"
  artifacts:
    - path: ".githooks/pre-commit"
      provides: "Pre-commit hook for cmdtree drift detection"
    - path: "Taskfile.yml"
      provides: "hooks:install target"
  key_links:
    - from: ".githooks/pre-commit"
      to: "omni cmdtree"
      via: "shell invocation of omni binary"
```

**Tasks:**

**Task 1: Regenerate cmdtree + aicontext outputs**

Files: `docs/cmdtree.md`, updated CLAUDE.md (via `omni aicontext`)

Action:
1. Run `omni cmdtree > docs/cmdtree.md` (creating the file if it does not yet exist — no existing committed file found in docs/).
2. Run `omni aicontext` to regenerate the AI context documentation (updates CLAUDE.md and related docs).
3. Stage both `docs/cmdtree.md` and any files updated by `omni aicontext` for commit.
4. Commit with message: `docs: regenerate cmdtree and aicontext for Phase 2 command set`

Verify: `git diff HEAD -- docs/cmdtree.md` shows the cmdtree content. The file contains all top-level commands from `cmd/`.

Done: cmdtree output committed; aicontext regenerated; git log shows the commit.

---

**Task 2: Create .githooks/pre-commit + task hooks:install**

Files: `.githooks/pre-commit`, `Taskfile.yml`

Action:
1. Create `.githooks/` directory if it does not exist.
2. Write `.githooks/pre-commit`:
   ```bash
   #!/usr/bin/env bash
   # Pre-commit hook: detect cmdtree drift when cmd/ files are staged.
   # Install via: task hooks:install

   set -euo pipefail

   # Check if any cmd/ files are staged
   if ! git diff --cached --name-only | grep -q '^cmd/'; then
     exit 0  # No cmd/ changes — skip check
   fi

   # Regenerate cmdtree and compare to committed version
   # NOTE: uses raw grep/git rather than omni — intentional exception for hook portability
   # (omni may not be in PATH at hook time; raw shell commands are always available)
   CURRENT_TREE=$(omni cmdtree 2>/dev/null)
   COMMITTED_TREE=$(git show HEAD:docs/cmdtree.md 2>/dev/null || echo "")

   # On first install docs/cmdtree.md may not be committed yet — skip check in that case
   if [ -z "$COMMITTED_TREE" ]; then
     echo "INFO: docs/cmdtree.md not yet committed — skipping drift check. Run: omni cmdtree > docs/cmdtree.md && git add docs/cmdtree.md"
     exit 0
   fi

   if [ "$CURRENT_TREE" != "$COMMITTED_TREE" ]; then
     echo "ERROR: omni cmdtree output has changed but docs/cmdtree.md is not staged."
     echo "Run: omni cmdtree > docs/cmdtree.md && git add docs/cmdtree.md"
     exit 1
   fi
   ```
3. Make executable: `chmod +x .githooks/pre-commit`
4. Add `hooks:install` task to `Taskfile.yml`:
   ```yaml
   hooks:install:
     desc: "Install git hooks from .githooks/ (opt-in)"
     cmds:
       - git config core.hooksPath .githooks
       - echo "Git hooks installed from .githooks/"
   ```

Verify: `bash .githooks/pre-commit` exits 0 when no cmd/ files staged. `grep -n "hooks:install" Taskfile.yml` finds the target.

Done: Hook file exists and is executable; `task hooks:install` wires it via `git config core.hooksPath`.

---

---

### Plan 02-10: Final Coverage Gate + Wave 3 Straggler Cleanup

```yaml
plan_id: "02-10"
name: "Final coverage gate run + wave 3 straggler cleanup"
wave: 3
requirements: [POLISH-04, POLISH-05, POLISH-06]
depends_on: ["02-01", "02-04", "02-05", "02-06", "02-07"]
files_modified:
  - coverage.out
  - pkg/video/video_api_test.go
autonomous: true
must_haves:
  truths:
    - "task lint:coverage exits 0 (all gates pass)"
    - "pkg/ weighted average ≥75%"
    - "internal/cli/ weighted average ≥60%"
    - "No pkg/ package below 40% floor"
    - "No internal/cli/ package below 30% floor"
  artifacts:
    - path: "coverage.out"
      provides: "Final coverage profile used by CI"
  key_links:
    - from: "Taskfile.yml lint:coverage"
      to: "tools/covgate/main.go"
      via: "go run"
```

**Tasks:**

**Task 1: Run full coverage gate and identify remaining stragglers**

Files: (measurement only)

Action:
1. Run `go test -coverprofile=coverage.out -short ./pkg/... ./internal/cli/...`
2. Run `go run ./tools/covgate -profile=coverage.out -pkg-prefix=github.com/inovacc/omni/pkg/ -avg-min=75 -floor=40`
3. Run `go run ./tools/covgate -profile=coverage.out -pkg-prefix=github.com/inovacc/omni/internal/cli/ -avg-min=60 -floor=30`
4. Record any packages still failing either rule.

If all gates pass → this plan is complete after Task 1.

If gates fail → proceed to Task 2.

Verify: Capture output of both covgate runs and compare against thresholds.

Done: Gates measured; straggler list for Task 2 confirmed (or empty → done).

---

**Task 2: Targeted straggler tests to clear all floors**

Files: Any `*_test.go` files in failing packages (determined by Task 1)

Action: For each package still below its floor threshold after Plans 02-04 through 02-07:
- Read the package's `.go` files to find the highest-statement-count uncovered functions
- Add `*_api_test.go` or extend existing `*_test.go` with 1–3 tests targeting those functions
- Focus on the cheapest coverage wins: exported functions with simple I/O contracts

Expected stragglers based on research (but confirm with Task 1 output):
- `pkg/video` root (46.0%) — Plan 02-04 Task 3 should have improved this; if still below 75%, add tests for `DownloadInfo()` with a fake HTTP server
- `pkg/userdirs` (42.9%) — Plan 02-04 Task 1 should cover; if still below 75%, expand to test all exported functions
- Any `internal/cli/` package the Task 2 of Plan 02-07 missed

Re-run `task lint:coverage` after each batch of tests. Stop when all gates pass.

Verify: `task lint:coverage` exits 0.

Done: `task lint:coverage` exits 0; all coverage thresholds met; CI will pass.

---

---

## Dependency Graph

```
Wave 1 (parallel):
  02-01  tools/covgate/           [POLISH-04, 05, 06]
  02-02  golden normalize: hook   [POLISH-13]
  02-03  tools/helplint/          [POLISH-11]

Wave 2 (after Wave 1 infrastructure):
  02-04  pkg/ API baselines       [POLISH-07]           — no deps
  02-05  youtube + downloader     [POLISH-04]           — no deps
  02-06  jsonutil + twig root     [POLISH-04]           — no deps
  02-07  internal/cli/ baseline   [POLISH-05, 06]       — needs 02-01 for gate tool
  02-08  golden entries + regen   [POLISH-12, 13]       — needs 02-02 for normalize hook
  02-09  cmdtree + hook           [POLISH-14]           — no deps

Wave 3 (after Wave 2):
  02-10  final gate + stragglers  [POLISH-04, 05, 06]   — needs 02-01..02-07
```

## Success Criteria

Phase 2 is complete when:
- [ ] `task lint:coverage` exits 0 (pkg/ ≥75% avg, ≥40% floor; cli/ ≥60% avg, ≥30% floor)
- [ ] `task lint:help` exits 0 (all commands have Long: with omni example)
- [ ] `task test:golden` exits 0 (all snapshots match, normalize: hooks applied)
- [ ] `tools/covgate/main.go` exists and compiles
- [ ] `tools/helplint/main.go` exists and compiles
- [ ] Both golden engines accept `normalize:` field without error
- [ ] `docs/GOLDEN_MASTER_TESTING.md` has normalize: examples section
- [ ] `task hooks:install` wires `.githooks/pre-commit`
- [ ] cmdtree and aicontext outputs committed
- [ ] CI `test.yml` includes `lint:coverage` and `lint:help` steps
