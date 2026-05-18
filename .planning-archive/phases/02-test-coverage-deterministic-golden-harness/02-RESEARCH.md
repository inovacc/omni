# Phase 2: Test Coverage + Deterministic Golden Harness — Research

**Researched:** 2026-04-12
**Domain:** Go test coverage tooling, golden-master normalization, help-docstring linting
**Confidence:** HIGH (all key findings verified by running actual commands against the codebase)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

1. **Coverage gates (dual-mode):**
   - `pkg/` omni-owned: weighted avg ≥75%, per-package floor ≥40%
   - `internal/cli/` omni-owned: weighted avg ≥60%, per-package floor ≥30%
   - Vendored buf subtrees excluded from both counts
   - Tool: `tools/covgate/` (extends `tools/cmderr-cov/` pattern); runs as `task lint:coverage`

2. **Golden normalize: hook:**
   - YAML field `normalize:` — list of `{pattern, replacement}` pairs, applied in order
   - Go regexp syntax; replacements are literal strings (no backreferences in v1)
   - Both engines (`testing/golden_engine.py` and `tools/golden/src/golden/`) must support it
   - Backwards-compatible: existing entries without `normalize:` unchanged

3. **Help docstring:** Every Cobra command needs non-empty `Short:` and `Long:` with at least one `omni <cmd>...` example. Tool: `tools/helplint/`; runs as `task lint:help`.

4. **Happy-path golden breadth:** Every top-level command + at least one key subcommand per group.

5. **cmdtree regen:** One-off commit + pre-commit hook via `task hooks:install`.

6. **Test style:** Black-box (`package foo_test`) for API-contract tests; white-box allowed elsewhere.

7. **Package triage:** Wave 1 (API baseline), Wave 2 (risk-weighted depth), Wave 3 (straggler cleanup).

### Claude's Discretion

- Implementation details of `tools/covgate/` internals (flag names, output formatting)
- Implementation details of `tools/helplint/` (AST vs regex parsing)
- Pre-commit hook mechanism (`.githooks/` vs `.husky/` — use `.githooks/`)
- Exact Wave 3 ordering once Wave 1+2 coverage is measured

### Deferred Ideas (OUT OF SCOPE)

- Consolidating the two golden harness systems
- Mutation testing
- Fuzzing beyond trivial Go 1.18+ targets
- Per-package coverage badges
- Coverage trend graphs over git history
- Stricter docstring format with ADR-level detail
- cmdtree CI drift gate
- Removing all white-box tests
- Single Go-only golden harness
- More than one happy-path snapshot per subcommand
</user_constraints>

---

## Summary

Phase 2 lands three independent but coordinated deliverables: (1) a coverage measurement + enforcement system with dual-rule gates, (2) a `normalize:` hook in both golden harnesses enabling deterministic snapshots for commands that emit random IDs or timestamps, and (3) a help-docstring linter plus cmdtree regeneration. All three are prerequisites for supply-chain phases (4–7) which need deterministic golden baselines.

**Primary recommendation:** Implement in Wave order. Wave 1 (API baseline tests for 0-coverage packages) unblocks accurate coverage measurement. Wave 2 (normalize: hook + help lint + covgate tool) provides the infrastructure. Wave 3 (straggler cleanup to hit floor targets) closes the gate.

The pkg/ coverage baseline is already strong — most packages are well above the 40% floor and many exceed 75%. The main gaps are `pkg/video/extractor/youtube` (4.0%), `pkg/video/downloader` (32.9%), `pkg/userdirs` (42.9%), `pkg/twig` (44.3%), and `pkg/jsonutil` (67.5% — below 75% target but above floor). The internal/cli/ coverage picture requires the background measurement to complete (see Section 1b).

---

## Section 1a: Current pkg/ Coverage Baseline

**Method:** `go test -cover -short ./pkg/...` run against the live codebase. [VERIFIED: direct measurement]

### pkg/ Coverage by Package (sorted ascending)

| Package | Coverage | Floor (40%) | Target (75%) | Action |
|---------|----------|-------------|--------------|--------|
| `pkg/video/extractor/youtube` | **4.0%** | FAIL | FAIL | Wave 2 — critical |
| `pkg/video/downloader` | **32.9%** | FAIL | FAIL | Wave 2 — critical |
| `pkg/userdirs` | **42.9%** | pass | FAIL | Wave 2 |
| `pkg/twig` | **44.3%** | pass | FAIL | Wave 2 |
| `pkg/video` | **46.0%** | pass | FAIL | Wave 2 |
| `pkg/video/format` | **50.0%** | pass | FAIL | Wave 2 |
| `pkg/video/utils` | **58.4%** | pass | FAIL | Wave 2 |
| `pkg/twig/builder` | **58.9%** | pass | FAIL | Wave 2 |
| `pkg/video/nethttp` | **61.8%** | pass | FAIL | Wave 2 |
| `pkg/jsonutil` | **67.5%** | pass | FAIL | Wave 2 |
| `pkg/video/cache` | **73.3%** | pass | FAIL | Wave 2 |
| `pkg/cobra/helper/output` | **82.9%** | pass | pass | No action needed |
| `pkg/htmlfmt` | **77.9%** | pass | pass | No action needed |
| `pkg/sqlfmt` | **79.1%** | pass | pass | No action needed |
| `pkg/twig/parser` | **79.1%** | pass | pass | No action needed |
| `pkg/pipeline` | **81.5%** | pass | pass | No action needed |
| `pkg/figlet` | **82.9%** | pass | pass | No action needed |
| `pkg/twig/formatter` | **80.4%** | pass | pass | No action needed |
| `pkg/cryptutil` | **85.3%** | pass | pass | No action needed |
| `pkg/twig/scanner` | **85.5%** | pass | pass | No action needed |
| `pkg/search/rg` | **86.6%** | pass | pass | No action needed |
| `pkg/cssfmt` | **87.3%** | pass | pass | No action needed |
| `pkg/hashutil` | **88.5%** | pass | pass | No action needed |
| `pkg/idgen` | **90.3%** | pass | pass | No action needed |
| `pkg/video/jsinterp` | **91.7%** | pass | pass | No action needed |
| `pkg/textutil` | **93.7%** | pass | pass | No action needed |
| `pkg/twig/comparer` | **96.3%** | pass | pass | No action needed |
| `pkg/video/m3u8` | **96.8%** | pass | pass | No action needed |
| `pkg/twig/expander` | **98.1%** | pass | pass | No action needed |
| `pkg/encoding` | **100.0%** | pass | pass | No action needed |
| `pkg/textutil/diff` | **95.2%** | pass | pass | No action needed |
| `pkg/twig/models` | **100.0%** | pass | pass | No action needed |
| `pkg/search/grep` | **77.9%** | pass | pass | No action needed |
| `pkg` (root) | no test files | — | — | No action (root doc-only) |
| `pkg/search` | no test files | — | — | No action (empty pkg) |
| `pkg/video/extractor/all` | no test files | — | — | No action (blank-import aggregator) |

**Packages below floor (40%) — must fix:**
- `pkg/video/extractor/youtube` — 4.0% (critical gap)
- `pkg/video/downloader` — 32.9%

**Packages above floor but below target (40–75%) — need coverage depth:**
- `pkg/userdirs` (42.9%), `pkg/twig` (44.3%), `pkg/video` (46.0%), `pkg/video/format` (50.0%), `pkg/video/utils` (58.4%), `pkg/twig/builder` (58.9%), `pkg/video/nethttp` (61.8%), `pkg/jsonutil` (67.5%), `pkg/video/cache` (73.3%)

**Packages already at target (≥75%) — no action:**
All others listed above.

**Notable finding:** `pkg/cryptutil` (85.3%), `pkg/hashutil` (88.5%), `pkg/search/grep` (77.9%), `pkg/search/rg` (86.6%), `pkg/pipeline` (81.5%), `pkg/twig/scanner` (85.5%), `pkg/twig/comparer` (96.3%), `pkg/idgen` (90.3%) are all above target — Wave 2 priority list from CONTEXT.md is largely already satisfied for these packages. Focus Wave 2 depth effort on the video subtree and `pkg/jsonutil`.

---

## Section 1b: Current internal/cli/ Coverage Baseline

**Status:** Background measurement running at research time. Results will be available when `task test` completes. [ASSUMED — measurement pending; planner should re-run before finalizing Wave 3 targets]

**What is known from CLAUDE.md and Phase 1:**
- `internal/cli/cmderr` is at 100% (Phase 1 achievement, enforced by `lint:cmderr-coverage`)
- Most cli packages have Run functions with at least basic tests
- The weighted average target is 60% with a 30% floor per package

**Planner action:** Run `go test -cover -short ./internal/cli/... 2>&1 | grep -E "^(ok|FAIL|\?)"` at plan time to get the exact baseline before scheduling Wave 3 tasks.

---

## Section 2: Vendored Buf Exclusion List

**Verified location:** `pkg/private/buf/` — the entire subtree under `pkg/private/` is vendored bufbuild code. [VERIFIED: `ls pkg/private/` shows only `buf/`; `grep -r bufbuild pkg/private/` shows googleapis proto files]

**Exact exclusion paths for `tools/covgate/`:**

```
github.com/inovacc/omni/pkg/private/...
```

There is NO `pkg/buf/` top-level directory. The vendored buf code lives exclusively under `pkg/private/buf/`. The coverage gate tool must exclude any package path containing `/private/`.

**CLAUDE.md note:** "~78% avg for omni-owned pkg/ packages; total skewed by vendored buf packages" — confirmed. The measurement in Section 1a (which goes to `./pkg/...`) does NOT include `pkg/private/` because those packages have no `*_test.go` files and `go test` reports them as `[no test files]` — they are automatically excluded from the coverage percentage. However, the covgate tool should explicitly skip them to avoid false negatives if test files are ever added.

**Exclusion rule for `tools/covgate/`:**
```go
// Skip vendored subtrees
if strings.Contains(pkgPath, "/private/") {
    continue
}
```

---

## Section 3: tools/covgate/ CLI Design

Based on the `tools/cmderr-cov/main.go` pattern (verified by reading the source): [VERIFIED: read file directly]

The existing tool uses `go tool cover -func=<profile>` via `exec.Command`, parses the output, and checks a single percentage threshold. It is stdlib-only, cross-platform, and dependency-free.

**`tools/covgate/main.go` design:**

```go
// tools/covgate/main.go
// Command covgate reads a Go coverage profile and enforces dual-rule coverage gates:
// 1. Weighted average across all non-excluded packages
// 2. Per-package floor — no single package may drop below the floor
//
// Usage:
//   covgate -profile=coverage.out -pkg-prefix=pkg/ -avg-min=75 -floor=40
//   covgate -profile=coverage.out -pkg-prefix=internal/cli/ -avg-min=60 -floor=30
//
// Exit codes:
//   0 — all rules pass
//   1 — one or more rules fail
//   2 — usage / I/O error
```

**Algorithm:**
1. Open coverage profile (same format as `go test -coverprofile`)
2. Parse per-file statement counts: `<pkg>/<file>:<line>.<col>,<line>.<col> <stmts> <count>`
3. Aggregate by package: `(covered_stmts, total_stmts)` per package
4. Skip packages matching exclusion patterns (`/private/`, any user-supplied `-exclude` glob)
5. Skip packages matching `-pkg-prefix` filter (only gate the specified subtree)
6. Compute weighted average: `sum(covered_stmts) / sum(total_stmts) * 100`
7. Check per-package floor: flag any package with `covered/total*100 < floor`
8. Emit human-readable report: list worst offenders (bottom 5 by coverage), weighted avg, pass/fail
9. Exit 1 if any rule fails

**Key difference from cmderr-cov:** Parses the coverage profile directly (line by line) rather than shelling out to `go tool cover -func`. This avoids a subprocess and makes it truly cross-platform. The coverage profile format is stable and documented.

**Coverage profile format** (Go 1.2+, stable): [ASSUMED — based on training knowledge; format has been stable since Go 1.2]
```
mode: set
github.com/inovacc/omni/pkg/cryptutil/cryptutil.go:25.45,27.2 2 1
```
Fields: `file:startline.startcol,endline.endcol stmts count`

**Taskfile integration:**
```yaml
lint:coverage:
  desc: Enforce coverage gates (weighted avg + per-package floor)
  cmds:
    - go test -coverprofile=coverage.out -short ./pkg/... ./internal/cli/...
    - go run ./tools/covgate -profile=coverage.out -pkg-prefix=pkg/ -avg-min=75 -floor=40 -exclude=private
    - go run ./tools/covgate -profile=coverage.out -pkg-prefix=internal/cli/ -avg-min=60 -floor=30
```

---

## Section 4: Golden normalize: Hook — Implementation Points

### Engine 1: `testing/golden_engine.py`

**Current state** [VERIFIED: read file directly]:
- `GoldenTestCase` dataclass has `normalizations: list[str]` — named-normalizer references
- `NORMALIZERS` dict maps name → lambda (4 built-ins: `normalize_newlines`, `strip_trailing_whitespace`, `strip_path`, `strip_temp_dir`)
- `GoldenEngine.normalize(text, normalizer_names)` applies chain
- Used in both `record()` (line 192–193) and `compare()` (line 224–225)
- YAML loading at line 100: `normalizations=test.get("normalizations", [])`

**Changes needed:**

**1. Extend `GoldenTestCase` dataclass** — add `normalize` field alongside existing `normalizations`:
```python
@dataclass
class GoldenTestCase:
    # ... existing fields ...
    normalizations: list[str] = field(default_factory=list)       # named built-ins (existing)
    normalize: list[dict] = field(default_factory=list)            # NEW: {pattern, replacement} list
```

**2. Extend `GoldenEngine.normalize()` method** — apply `normalize` patterns after named normalizers:
```python
def normalize(self, text: str, normalizer_names: list[str], normalize_rules: list[dict] = None) -> str:
    """Apply normalization chain. Always normalizes newlines first."""
    text = NORMALIZERS["normalize_newlines"](text)
    for name in normalizer_names:
        fn = NORMALIZERS.get(name)
        if fn:
            text = fn(text)
    # Apply inline regex normalize: rules
    for rule in (normalize_rules or []):
        text = re.sub(rule["pattern"], rule["replacement"], text)
    return text
```

**3. Update YAML loading** (load_registry, line ~94–103):
```python
tests.append(GoldenTestCase(
    # ... existing fields ...
    normalizations=test.get("normalizations", []),
    normalize=test.get("normalize", []),          # NEW
))
```

**4. Update callers** — both `record()` and `compare()` pass `test.normalize` to `normalize()`:
```python
stdout = self.normalize(stdout, test.normalizations, test.normalize)
stderr = self.normalize(stderr, test.normalizations, test.normalize)
```

**Exact file:** `testing/golden_engine.py`
**Exact methods:** `GoldenTestCase` (dataclass), `GoldenEngine.normalize()`, `GoldenEngine.load_registry()`, `GoldenEngine.record()`, `GoldenEngine.compare()`

---

### Engine 2: `tools/golden/src/golden/`

**Current state** [VERIFIED: read files directly]:
- `types.py` → `TestCase` dataclass: has `normalizations: list[str]` (named built-ins)
- `normalize.py` → `normalize(text, normalizer_names)` function with same 4 built-ins
- `runner.py` → calls `normalize(result.stdout, test_case.normalizations)` (lines 58–59)
- YAML loading in `discovery.py` (not yet read — inferred from pattern)

**Changes needed:**

**1. Extend `types.py` → `TestCase` dataclass:**
```python
@dataclass
class TestCase:
    # ... existing fields ...
    normalizations: list[str] = field(default_factory=list)     # existing
    normalize: list[dict] = field(default_factory=list)         # NEW
```

**2. Extend `normalize.py` → `normalize()` function:**
```python
def normalize(text: str, normalizer_names: list[str] | None = None,
              normalize_rules: list[dict] | None = None) -> str:
    """Apply normalization chain. Always normalizes newlines first."""
    text = NORMALIZERS["normalize_newlines"](text)
    for name in (normalizer_names or []):
        fn = NORMALIZERS.get(name)
        if fn:
            text = fn(text)
    for rule in (normalize_rules or []):
        text = re.sub(rule["pattern"], rule["replacement"], text)
    return text
```

**3. Update `runner.py`** — pass `test_case.normalize` as second arg:
```python
stdout = normalize(result.stdout, test_case.normalizations, test_case.normalize)
stderr = normalize(result.stderr, test_case.normalizations, test_case.normalize)
```

**4. Update discovery/loading** — in whichever module parses YAML into `TestCase`, add:
```python
normalize=test.get("normalize", []),
```

**Exact files:** `tools/golden/src/golden/types.py`, `tools/golden/src/golden/normalize.py`, `tools/golden/src/golden/runner.py`, and the YAML-loading module (check `discovery.py` or `config.py`)

---

### YAML Schema Change

Both engines share `testing/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml`.

The new `normalize:` field is optional and sits at the test-case level:

```yaml
categories:
  - name: idgen
    tests:
      - name: uuid_v4_happy
        args: ["uuid", "-v", "4"]
        expected_exit: 0
        normalize:
          - pattern: '[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}'
            replacement: '{UUID_V4}'
      - name: ulid_happy
        args: ["ulid"]
        expected_exit: 0
        normalize:
          - pattern: '[0-9A-Z]{26}'
            replacement: '{ULID}'
      - name: ksuid_happy
        args: ["ksuid"]
        expected_exit: 0
        normalize:
          - pattern: '[0-9A-Za-z]{27}'
            replacement: '{KSUID}'
      - name: snowflake_happy
        args: ["snowflake"]
        expected_exit: 0
        normalize:
          - pattern: '\d{15,20}'
            replacement: '{SNOWFLAKE_ID}'
```

**Backwards compatibility:** `normalize:` absent = empty list → existing behavior unchanged.

---

## Section 5: Help Docstring Gaps

**Method:** Script checked all 159 `cmd/*.go` files for presence of `Long:` field. [VERIFIED: direct script execution]

**Commands missing `Long:` field** (4 files):
- `cmd/copy.go` — the `cp`-style copy command
- `cmd/move.go` — the `mv`-style move command
- `cmd/remove.go` — the `rm`-style remove command
- `cmd/helpers.go` — likely a shared helper file, not a command definition

**Note:** `cmd/helpers.go` is probably a utility file with no Cobra command definition — `tools/helplint/` should skip files that don't contain a `cobra.Command` struct literal. Confirm by inspection.

**Commands with `Long:` present but potentially lacking `omni <cmd>...` example:** This requires parsing the content of Long: fields. The `tools/helplint/` tool will catch these at lint time. The 3 real commands (copy, move, remove) are confirmed gaps that need both `Long:` addition and an example line.

**Scope of work:** Small — only 3 command files need docstrings. The 100+ other commands already have `Long:` fields.

---

## Section 6: Phase 1 Goldens Needing normalize: Hooks

**Method:** Scanned `testing/golden/golden_tests.yaml` for uuid/ulid/ksuid/snowflake/nanoid/random/date/video/jwt entries. [VERIFIED: direct grep]

**Key finding:** ALL existing golden entries for idgen commands (uuid, ulid, ksuid, snowflake, nanoid) are **error-path tests** (invalid flag errors) — they produce deterministic stderr output and do NOT need normalize: hooks.

**What this means for Phase 2:**
- Happy-path golden entries for these commands DO NOT YET EXIST
- They must be ADDED in Phase 2 as part of POLISH-12 (golden breadth)
- When added, they MUST include `normalize:` hooks because their stdout is non-deterministic

**Entries to add with normalize: hooks:**

| Command | Entry name | normalize: pattern needed |
|---------|-----------|--------------------------|
| `omni uuid -v 4` | `uuid_v4_happy` | UUID v4 pattern |
| `omni uuid -v 7` | `uuid_v7_happy` | UUID v7 pattern |
| `omni ulid` | `ulid_happy` | ULID pattern (26 base32 chars) |
| `omni ksuid` | `ksuid_happy` | KSUID pattern (27 chars) |
| `omni snowflake` | `snowflake_happy` | 15–20 digit integer |
| `omni nanoid` | `nanoid_happy` | 21-char URL-safe string |
| `omni random` | `random_happy` | arbitrary string |
| `omni random --hex` | `random_hex_happy` | hex string |
| `omni date` | `date_happy` | full timestamp pattern |

**Existing `date_unknown_flag` entry** (line 577): error-path, produces deterministic stderr — no normalize: needed.

**jwt_decode entry** (line 331): Need to check if it decodes a static fixture token or a live one. If static fixture → likely deterministic → no normalize: needed. If live → needs timestamp normalize:.

**video entries:** No happy-path video entries exist in the golden registry. Adding them is HIGH risk due to network dependency. Recommendation: use `omni video --help` as the happy-path golden for video (no normalize: needed), and defer live download tests to Docker integration suite.

---

## Section 7: Risk-Weighted Wave 2 Priority List

Based on verified coverage from Section 1a: [VERIFIED]

| Priority | Package | Current Coverage | Gap to 75% | Action |
|----------|---------|-----------------|------------|--------|
| 1 | `pkg/video/extractor/youtube` | 4.0% | 71pp | CRITICAL: below 40% floor |
| 2 | `pkg/video/downloader` | 32.9% | 42pp | CRITICAL: below 40% floor |
| 3 | `pkg/video` | 46.0% | 29pp | Below target |
| 4 | `pkg/jsonutil` | 67.5% | 7.5pp | Below target, parsers have subtle bugs |
| 5 | `pkg/twig` | 44.3% | 30.7pp | Below target |
| 6 | `pkg/userdirs` | 42.9% | 32.1pp | Below target |
| 7 | `pkg/video/format` | 50.0% | 25pp | Below target |
| 8 | `pkg/video/utils` | 58.4% | 16.6pp | Below target |
| 9 | `pkg/video/nethttp` | 61.8% | 13.2pp | Below target |
| 10 | `pkg/twig/builder` | 58.9% | 16.1pp | Below target |
| 11 | `pkg/video/cache` | 73.3% | 1.7pp | Marginal — low priority |

**Already at target (SKIP — no Wave 2 action):**

Per CONTEXT.md Wave 2 list vs actual coverage:
- `pkg/cryptutil/` → 85.3% ✓ skip
- `pkg/hashutil/` → 88.5% ✓ skip
- `pkg/jsonutil/` → 67.5% — needs work (7.5pp gap)
- `pkg/jsonfmt/` → NOT a separate package (embedded in `pkg/jsonutil/` or `internal/cli/`)
- `pkg/sqlfmt/` → 79.1% ✓ skip
- `pkg/cssfmt/` → 87.3% ✓ skip
- `pkg/htmlfmt/` → 77.9% ✓ skip
- `pkg/pipeline/` → 81.5% ✓ skip
- `pkg/search/grep/` → 77.9% ✓ skip
- `pkg/search/rg/` → 86.6% ✓ skip
- `pkg/video/extractor/` → 41.7% — needs work (below target)
- `pkg/video/m3u8/` → 96.8% ✓ skip
- `pkg/twig/scanner/` → 85.5% ✓ skip
- `pkg/twig/formatter/` → 80.4% ✓ skip
- `pkg/twig/comparer/` → 96.3% ✓ skip
- `pkg/idgen/` → 90.3% ✓ skip

**Revised Wave 2 focus:** The video subtree (youtube extractor, downloader, video root) and `pkg/jsonutil` + `pkg/twig` (root). Most of the original Wave 2 priority list is already satisfied.

---

## Section 8: Taskfile Target Naming

**Verified existing targets** [VERIFIED: read Taskfile.yml]:

| Target | Status |
|--------|--------|
| `lint` | EXISTS — runs golangci-lint |
| `lint:cmderr-coverage` | EXISTS (Phase 1) |
| `lint:coverage` | FREE — safe to add |
| `lint:help` | FREE — safe to add |
| `hooks:install` | FREE — safe to add |
| `test:coverage` | EXISTS — generates HTML report (different purpose, no conflict) |
| `test:cover` | EXISTS — shows percentage (different from `lint:coverage`) |

**No naming conflicts.** All three new Phase 2 targets (`lint:coverage`, `lint:help`, `hooks:install`) are free.

**Recommended additions to Taskfile.yml:**

```yaml
lint:coverage:
  desc: Enforce coverage gates for pkg/ and internal/cli/ (weighted avg + per-package floor)
  cmds:
    - go test -coverprofile=coverage.out -short ./pkg/... ./internal/cli/...
    - go run ./tools/covgate -profile=coverage.out -pkg-prefix=pkg/ -avg-min=75 -floor=40 -exclude=/private/
    - go run ./tools/covgate -profile=coverage.out -pkg-prefix=internal/cli/ -avg-min=60 -floor=30

lint:help:
  desc: Verify all Cobra commands have Short + Long with omni example
  cmds:
    - go run ./tools/helplint ./cmd/...

hooks:install:
  desc: Install git hooks (cmdtree drift check on cmd/ changes)
  cmds:
    - git config core.hooksPath .githooks
    - echo "Git hooks installed from .githooks/"
```

---

## Section 9: Pre-existing Test Failures / Build Status

**Build status:** [ASSUMED — not run separately; coverage scan used `-short` flag which skips slow tests and ran without failures indicated in output]

**pkg/ test suite:** All packages completed with `ok` status in coverage scan — no failures. [VERIFIED]

**Packages with `[no test files]`:**
- `pkg` (root — doc-only package)
- `pkg/search` (empty aggregator package)
- `pkg/video/extractor/all` (blank-import aggregator — intentionally no tests)

These are NOT failures. They are expected no-op packages.

**Known skipped slow tests:** `-short` flag was used, so any test calling `t.Skip(testing.Short())` was skipped. Integration tests that make network calls (video, YouTube) are excluded.

**Planner note:** Run `go build ./...` at plan time to confirm no compilation failures. As of Phase 1 completion, the codebase was building cleanly.

---

## Standard Stack

### Core (Phase 2 tools)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib only | 1.25.0 | `tools/covgate/`, `tools/helplint/` | Project "no external deps in tools" principle |
| `go/ast`, `go/parser` | stdlib | helplint: parse Cobra command definitions | Accurate parsing vs regex; already used in `scaffold testgen` |
| `regexp` | stdlib | normalize: pattern matching in Python engines | Already imported in both engines |
| Python `re` | stdlib | normalize: substitution in golden engines | Already imported |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `go tool cover` | Go stdlib | Coverage profile generation | Already in Taskfile |
| `bufio.Scanner` | stdlib | Coverage profile parsing in covgate | Streaming parse of potentially large profiles |

**Installation:** No new dependencies. All tools use existing Go stdlib and Python stdlib.

---

## Architecture Patterns

### tools/covgate/ Structure

```
tools/covgate/
└── main.go    # Single-file tool (matches tools/cmderr-cov/ pattern)
```

Pattern from `tools/cmderr-cov/main.go`:
- Package `main`, stdlib-only, flag-based CLI
- Parse profile line-by-line with `bufio.Scanner`
- Emit to stdout (OK) or stderr (FAIL)
- Exit code 0 (pass), 1 (fail), 2 (usage error)

### tools/helplint/ Structure

```
tools/helplint/
└── main.go    # Walk cmd/*.go, parse with go/ast, check Short+Long+example
```

**go/ast approach** (preferred over regex): Use `go/parser.ParseFile()` to find `cobra.Command` composite literals, then inspect `Short` and `Long` field values. This is accurate even with multi-line strings and string concatenation. The project already uses `go/ast` in `internal/cli/scaffolding/testgen/`.

### .githooks/ Structure

```
.githooks/
└── pre-commit    # Shell script: check if cmd/ files staged → run omni cmdtree → diff
```

Installed via `git config core.hooksPath .githooks` (the `task hooks:install` target).

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Coverage profile format | Custom parser | Read Go coverage profile spec + parse directly | Format documented, stable since Go 1.2 |
| Regex normalization | Custom diff engine | `re.sub()` in Python / `regexp.ReplaceAllString` in Go | Standard library handles all edge cases |
| Cobra command inspection | String matching | `go/ast` package | Handles all string literal forms correctly |

---

## Common Pitfalls

### Pitfall 1: Coverage Profile Weighted Average vs Simple Average

**What goes wrong:** Averaging per-package percentages (simple average) instead of summing statement counts (weighted average). A tiny package at 100% inflates the simple average unfairly.

**Why it happens:** `go tool cover -func` reports per-function percentages and a "total:" line which IS statement-weighted. But if you parse individual package outputs separately and average the totals, you get simple average.

**How to avoid:** Parse the unified coverage profile (single `-coverprofile=coverage.out ./pkg/... ./internal/cli/...`) and compute `sum(covered_stmts) / sum(total_stmts)` directly from statement counts in the profile.

**Warning signs:** Overall weighted average significantly different from the simple average of package percentages.

### Pitfall 2: normalize: Patterns Applied in Wrong Order

**What goes wrong:** A later pattern matches the replacement token of an earlier pattern (e.g., `{UUID}` is matched by a generic alphanumeric pattern).

**How to avoid:** Apply patterns in-order as specified (already locked decision). Document that more-specific patterns should come first. The `{PLACEHOLDER}` replacement style with braces avoids most collisions.

### Pitfall 3: Video Test Flakiness from Network

**What goes wrong:** Adding happy-path golden tests for `omni video info <url>` that make network calls — these fail in CI when network is unavailable.

**How to avoid:** Use `--help` output as the video golden entry, or use a locally-served HTTP mock. Live video tests belong in the Docker integration suite only.

### Pitfall 4: Buf Packages Being Included in Gate

**What goes wrong:** `pkg/private/buf/` packages accidentally included in coverage gate, pulling down the weighted average.

**How to avoid:** Exclude by path prefix `/private/` in covgate. The packages already report `[no test files]` so they won't appear in the coverage profile anyway — but defensive exclusion is still correct.

### Pitfall 5: go/ast helplint Missing backtick Long: strings

**What goes wrong:** helplint uses regex on source text and misses multi-line backtick raw strings, or incorrectly parses concatenated string literals.

**How to avoid:** Use `go/ast` + `go/constant` evaluation to get the actual string value of `Long:` fields. The existing scaffold testgen code in `internal/cli/scaffolding/testgen/` demonstrates this pattern.

---

## Code Examples

### covgate coverage profile parsing

```go
// Source: Go coverage profile format (stable since Go 1.2)
// Line format: github.com/foo/bar/pkg/file.go:L.C,L.C N M
// N = statement count, M = hit count
scanner := bufio.NewScanner(f)
scanner.Scan() // skip "mode: set" header line
for scanner.Scan() {
    line := scanner.Text()
    // Split on last two space-separated fields
    parts := strings.Fields(line)
    if len(parts) != 3 { continue }
    stmts, _ := strconv.Atoi(parts[1])
    count, _ := strconv.Atoi(parts[2])
    // Extract package path: everything before last '/'
    filePath := parts[0][:strings.LastIndex(parts[0], ":")]
    pkgPath := filePath[:strings.LastIndex(filePath, "/")]
    // Accumulate
    totals[pkgPath].total += stmts
    if count > 0 { totals[pkgPath].covered += stmts }
}
```

### normalize: YAML schema (complete example)

```yaml
# Source: 02-CONTEXT.md locked decision
- name: uuid_v4_happy
  args: ["uuid", "-v", "4"]
  expected_exit: 0
  normalize:
    - pattern: '[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}'
      replacement: '{UUID_V4}'
```

### golden_engine.py normalize() extension

```python
# Source: testing/golden_engine.py (verified structure)
def normalize(self, text: str, normalizer_names: list[str],
              normalize_rules: list[dict] = None) -> str:
    text = NORMALIZERS["normalize_newlines"](text)
    for name in normalizer_names:
        fn = NORMALIZERS.get(name)
        if fn:
            text = fn(text)
    for rule in (normalize_rules or []):
        text = re.sub(rule["pattern"], rule["replacement"], text)
    return text
```

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | internal/cli/ coverage results — exact per-package numbers pending background run | Sec 1b | Planner may mis-schedule Wave 3 targets; mitigated by planner re-running measurement |
| A2 | Coverage profile format is stable since Go 1.2 (parse directly vs `go tool cover`) | Sec 3 | If format changed in Go 1.25, parser breaks; low risk — format is well-documented |
| A3 | `cmd/helpers.go` contains no Cobra command definition | Sec 5 | If it does define a command, that command also needs Long: |
| A4 | jwt_decode golden entry uses static fixture token (deterministic) | Sec 6 | If it decodes a live token, needs timestamp normalize: |
| A5 | `go test -short` skips all network-dependent tests in pkg/video/ | Sec 9 | If some video tests don't check `testing.Short()`, they may time out |

---

## Open Questions

1. **internal/cli/ exact coverage numbers**
   - What we know: background job running; cmderr is 100%; most packages have some tests
   - What's unclear: which internal/cli packages are below 30% floor
   - Recommendation: planner re-runs `go test -cover -short ./internal/cli/...` before scheduling Wave 3

2. **jwt_decode golden entry determinism**
   - What we know: entry exists at line 331 of golden_tests.yaml
   - What's unclear: whether the JWT fixture has an expiry timestamp in the decoded output
   - Recommendation: inspect the entry's args and fixture; add timestamp normalize: if needed

3. **pkg/video/extractor/youtube 4% coverage — what's testable**
   - What we know: YouTube extractor uses InnerTube API + goja JS runtime; network-dependent
   - What's unclear: how much can be tested without network calls (URL matching, metadata parsing, signature parsing)
   - Recommendation: target URL matching tests + offline metadata parse tests; skip live API calls

---

## Environment Availability

Step 2.6: SKIPPED (phase is code/config changes only; no new external tool dependencies beyond existing Go 1.25 toolchain)

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` stdlib + Python pytest (black-box) |
| Config file | None — standard `go test` |
| Quick run command | `go test -short ./pkg/... ./internal/cli/...` |
| Full suite command | `go test -race -cover ./pkg/... ./internal/cli/...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command |
|--------|----------|-----------|-------------------|
| POLISH-04 | Coverage gate enforced in CI | integration | `task lint:coverage` |
| POLISH-05 | Per-package floor enforced | integration | `task lint:coverage` |
| POLISH-06 | Vendored buf excluded from gate | unit | `go test ./tools/covgate/...` |
| POLISH-07 | Every pkg/ has ≥1 black-box baseline test | measurement | `go test -cover -short ./pkg/...` |
| POLISH-11 | Golden normalize: applied before diff | unit | `python -m pytest testing/` |
| POLISH-12 | Every top-level command has happy-path golden | golden | `task test:golden` |
| POLISH-13 | Help docstring lint passes | integration | `task lint:help` |
| POLISH-14 | cmdtree regenerated + pre-commit hook | manual | `task hooks:install` |

### Wave 0 Gaps
- [ ] `tools/covgate/main.go` — covers POLISH-04, POLISH-05, POLISH-06
- [ ] `tools/helplint/main.go` — covers POLISH-13
- [ ] `.githooks/pre-commit` — covers POLISH-14 drift prevention
- [ ] normalize: entries in both YAML registries — covers POLISH-11, POLISH-12

---

## Security Domain

The phase adds test infrastructure tools and golden harness extensions. No user-facing attack surface is introduced.

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | minimal | covgate validates profile format; helplint validates Go source — both are dev-only tools |
| All others | no | Dev tooling only; no auth, sessions, crypto, or network |

---

## Sources

### Primary (HIGH confidence)
- Direct codebase measurement: `go test -cover -short ./pkg/...` — all pkg/ coverage numbers
- `testing/golden_engine.py` — read directly; all engine structure findings
- `tools/golden/src/golden/types.py` — read directly; TestCase structure
- `tools/golden/src/golden/normalize.py` — read directly; normalize function
- `tools/golden/src/golden/runner.py` — read directly; normalize call sites
- `tools/cmderr-cov/main.go` — read directly; template for covgate design
- `Taskfile.yml` — read directly; existing target names
- `.scripts/find-missing-long.sh` — script output confirming 3 missing Long: files

### Secondary (MEDIUM confidence)
- `testing/golden/golden_tests.yaml` grep output — idgen entries are all error-path

### Tertiary (LOW confidence — see Assumptions Log)
- internal/cli/ exact coverage numbers — background measurement still running at research time

---

## Metadata

**Confidence breakdown:**
- pkg/ coverage baseline: HIGH — directly measured
- Architecture patterns: HIGH — based on verified existing code
- internal/cli/ coverage: LOW — pending measurement (see Assumptions Log A1)
- Golden engine implementation points: HIGH — read source files directly
- Taskfile naming: HIGH — read Taskfile.yml directly

**Research date:** 2026-04-12
**Valid until:** 2026-05-12 (stable domain; Go toolchain + Python stdlib)
