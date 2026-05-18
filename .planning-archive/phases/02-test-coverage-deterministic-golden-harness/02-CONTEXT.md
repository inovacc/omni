# Phase 2 Context — Test Coverage + Deterministic Golden Harness

**Phase:** 2
**Goal:** Raise test coverage to v1.0 targets AND land golden-master normalization hooks (timestamps, random IDs) that supply-chain commands will require in Phases 4–7.
**Requirements:** POLISH-04, POLISH-05, POLISH-06, POLISH-07, POLISH-11, POLISH-12, POLISH-13, POLISH-14
**Depends on:** Phase 1 (complete)

## Decisions

### Coverage measurement — weighted average + per-package floor

Coverage gates are **dual-mode**: a weighted average across the corpus AND a hard floor per package. Neither can be gamed by one well-tested big package.

**pkg/ (omni-owned only):**
- Weighted average (by statements/SLOC) ≥ **75%**
- Per-package floor ≥ **40%** — no single omni-owned `pkg/*` package may drop below 40%
- Vendored `buf` subtrees (`pkg/buf/`, anything with `github.com/bufbuild/` vendored underneath) are **excluded** from both counts

**internal/cli/ (omni-owned only):**
- Weighted average (by SLOC) ≥ **60%**
- Per-package floor ≥ **30%**
- Excluded: pure helpers with no runtime code (`cmderr` itself is already 100% from Phase 1 and is exempt from the floor by virtue of being above it)

**Measurement tool:**
- Extend `tools/cmderr-cov/` (Phase 1's pure-Go coverage parser) into `tools/covgate/` (or similar) that reads `coverage.out`, applies the two-rule gate, and emits a human-readable report naming the worst offenders.
- The tool enforces both the weighted average and the per-package floor in one pass.
- Runs as `task lint:coverage` in Taskfile.yml; CI job calls it.

**CI enforcement:** Hard fail on any rule violation. Not a soft warning. This follows Phase 1's `cmderr-coverage-gate` precedent.

### Golden normalization hook — regex substitution in YAML

The golden-master harness gets a new optional field per test entry:

```yaml
- name: sbom_output_deterministic
  cmd: ["omni", "sbom", "."]
  expected_exit: 0
  normalize:
    - pattern: '\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z'
      replacement: '{TIMESTAMP}'
    - pattern: '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'
      replacement: '{UUID}'
```

**Rules:**
- `normalize:` is a list of `{pattern, replacement}` pairs applied IN ORDER to the captured stdout/stderr before diff.
- Patterns are Go `regexp` syntax. Replacements are literal strings (no capture backreferences in v1 — keep it simple).
- Normalization applies to both stdout and stderr equally unless a future `target:` field is added (deferred).
- Both harness systems (`testing/golden/` Python engine AND `tools/golden/` full engine) must support the same `normalize:` field with byte-identical results. Golden engines in both locations share the YAML registry, so they must share the normalize semantics.

**Backwards compat:** Existing entries without `normalize:` behave exactly as before. No forced migration.

**Documentation:** Add a section to `docs/GOLDEN_MASTER_TESTING.md` with 3–5 example entries showing common patterns (timestamp, UUID, ULID, temp file path, random port).

### Help docstring — loose template, example required

Every command's Cobra `Long:` field must contain at least one concrete `omni <cmd> ...` example line. Not a strict section structure.

**Lint rule** (implemented in Phase 2):
- A small Go tool at `tools/helplint/` walks `cmd/**/*.go`, parses Cobra command definitions, and verifies every command has:
  - Non-empty `Short:` (1-line summary)
  - Non-empty `Long:` containing at least one line starting with `  omni ` or `omni ` (concrete example)
- Runs as `task lint:help` in Taskfile.yml; CI job calls it.
- Failing commands listed by name so they can be fixed in a follow-up commit.

**No strict USAGE/EXAMPLES/FLAGS/EXIT CODES section headers.** Cobra renders the help page well enough with just Short + Long + native flag output.

### Happy-path snapshot breadth — top-level + key subcommands

POLISH-12 requires "every command registered in both golden registries with at least one happy-path snapshot". "Every command" interpretation:

- **Always:** Every top-level command (`omni aws`, `omni kubectl`, `omni sbom`, etc.)
- **Also:** At least one **key subcommand per group** where the top-level is a pure router
  - `omni aws` → plus `omni aws s3 ls` (the most-used subcommand)
  - `omni kubectl` → plus `omni k get pods`
  - `omni git` → plus `omni git st`
  - `omni video` → plus `omni video info <url>` (with a short stable test URL or `--help` equivalent)
  - `omni buf` → plus `omni buf format` and `omni buf lint`
- **Not required:** Exhaustive coverage of every leaf subcommand. Those can be added in later phases if specific subcommand behavior regresses.

**What counts as happy-path:**
- Exits 0
- Produces some non-empty stdout
- Doesn't depend on unstable external state (network, clock, random beyond what `normalize:` handles)

### cmdtree regeneration — one-off commit + pre-commit hook

POLISH-14 requires `omni cmdtree` output regenerated and committed.

**Immediate:** One-off commit near the end of Phase 2 that regenerates the cmdtree output (and `omni aicontext` equivalents per CLAUDE.md) and checks in the result.

**Ongoing drift prevention:** Add a pre-commit git hook (in `.githooks/pre-commit` or similar, installed via `task hooks:install`) that runs `omni cmdtree` if any file under `cmd/` changed in the staged set, and fails the commit if the output differs from the committed version.

**Not in scope:** CI gate for cmdtree drift. Pre-commit hook is enough; CI duplicate is nice-to-have and deferred.

### Test style — black-box for API-contract tests, white-box allowed elsewhere

Per POLISH-07, every `pkg/*` package needs at least one baseline test establishing its public API contract.

**Convention:**
- **API-contract tests** (new Phase 2 baseline tests) use **`package pkg_test`** (black-box). They import the package by path and test what external Go consumers can actually call. This validates the API surface at the same time.
- **Implementation-detail tests** (existing or new deeper tests) may stay **`package pkg`** (white-box) when they need access to unexported helpers.
- A single package may have both styles — `foo.go` + `foo_test.go` (white-box) + `foo_api_test.go` (black-box).
- Follows Go stdlib convention.

**Enforcement:** No lint rule — this is a convention, not a gate. Reviewers flag black-box vs white-box mismatches during PR review.

### Package triage — risk-weighted with a Wave 1 API-contract pass

Execution order across the ~80 omni-owned packages:

**Wave 1 — API contract baseline (POLISH-07):**
- Every `pkg/*` package that currently has zero test files gets a minimal black-box baseline test: construct the primary type, call the primary entry point once with a representative input, assert non-error. This is the "API contract" test.
- Target: ≥1 test file per package. Coverage may still be low (20–40%) but the baseline exists.
- Estimated ~15–25 packages need this.

**Wave 2 — Risk-weighted depth push:**
Priority order based on `.planning/codebase/CONCERNS.md` + security-sensitive code:
1. `pkg/cryptutil/` — crypto is never "good enough"
2. `pkg/hashutil/` — security-adjacent
3. `pkg/jsonutil/`, `pkg/jsonfmt/`, `pkg/sqlfmt/`, `pkg/cssfmt/`, `pkg/htmlfmt/` — parsers have subtle bugs
4. `pkg/pipeline/` (stages) — streaming correctness
5. `pkg/search/grep/`, `pkg/search/rg/` — gitignore + pattern matching
6. `pkg/video/extractor/` + `pkg/video/m3u8/` — user-facing download correctness
7. `pkg/twig/scanner/` + `pkg/twig/formatter/` + `pkg/twig/comparer/` — tree handling
8. `pkg/idgen/` — determinism guarantees

**Wave 3 — Lowest-coverage stragglers:**
After Wave 1 + 2, run the coverage gate. Any package below the per-package floor (40% for pkg/, 30% for internal/cli/) gets targeted tests until it clears the floor.

**internal/cli/ side:** Not a Wave 1 baseline push (those packages all have Run functions with tests or will after Phase 1). Only a Wave 3 straggler push to reach the ≥60% weighted average and ≥30% floor.

## Specifics the researcher / planner should know

- **Phase 1 produced a 100% cmderr coverage baseline.** Don't regress it.
- **The golden harness has two implementations** (`testing/golden_engine.py` and `tools/golden/src/golden/`) sharing `golden_tests.yaml`. Both must learn the `normalize:` field. Do not consolidate them — that's post-1.0.
- **Phase 1 Plan 16 added 74 new golden entries** but did NOT run `task test:golden:update`. Many of those entries may need `normalize:` hooks because they capture timestamps or random IDs (video, uuid, ksuid, ulid, etc.). This phase should regenerate the `.stdout` baselines after normalize lands.
- **Vendored buf subtrees skew current coverage numbers** (CLAUDE.md notes ~78% omni-owned avg vs ~25.8% including vendored). Researcher must identify exact exclusion paths.
- **`omni cmdtree` exists** (see CLAUDE.md "Tooling" section). Planner uses it as-is; no need to rewrite.
- **`omni aicontext` exists** and regenerates docs. Both it and `omni cmdtree` should be regenerated in the same commit.
- **Taskfile.yml already has `lint:cmderr-coverage`** (Phase 1). Phase 2 adds `lint:coverage` (general gate), `lint:help` (docstring lint). Keep them as separate targets so CI can skip individual ones if needed.
- **Do not touch pkg/buf/** internal vendored code for tests. If buf tooling has bugs, push them to backlog. Buf consolidation is post-1.0.
- **Black-box tests need to import by full module path** (`github.com/inovacc/omni/pkg/foo`). Verify `go.mod` supports this (it does — standard Go module layout).
- **Pre-commit hook installation** must be opt-in (`task hooks:install`), not automatic. Some developers don't want auto-hooks.

## Open questions for the researcher

- **Exact current coverage per package.** Run `go test -cover ./pkg/... ./internal/cli/...` and identify:
  - Packages at 0% (Wave 1 targets)
  - Packages below the floor (40% pkg/, 30% internal/cli/)
  - Packages already clearing the target (no work needed)
- **Which packages are vendored buf trees** — compute exact exclusion list for the coverage gate.
- **Which top-level commands lack Long/Short** — run `omni cmdtree` or parse Cobra definitions to find help-docstring gaps.
- **Which Phase 1 Plan 16 goldens need normalize hooks** — scan for timestamps, UUIDs, ULIDs, random IDs, temp paths in the `cmderr_wave_z` entries.
- **Taskfile target naming conflicts** — verify `lint:coverage`, `lint:help`, `hooks:install` don't collide with existing targets.

## Deferred ideas (not scope creep, noted for backlog)

- **Consolidate the two golden harness systems** — post-1.0 cleanup per PROJECT.md. Not Phase 2's job.
- **Mutation testing** (e.g., `go-mutesting`) — maybe Phase 3 or post-1.0; too speculative for Phase 2.
- **Fuzzing** beyond the trivial Go 1.18+ fuzz targets — defer; some packages (parsers especially) would benefit but it's a separate effort.
- **Per-package coverage badge in README** — nice-to-have, defer.
- **Coverage trend graph over git history** — defer to post-1.0 observability work.
- **Stricter docstring format with ADR-level detail per command** — rejected as over-engineering.
- **cmdtree CI drift gate** — redundant with pre-commit hook; defer.
- **Removing white-box tests in favor of pure black-box** — rejected; stdlib convention is mixed and that's fine.
- **Replace the two-engine golden system with a single Go-only harness** — post-1.0.
- **More than one happy-path snapshot per subcommand** — reviewable addition later; not required for v1.0.

## Ready for

`/gsd-plan-phase 2` — researcher computes exact per-package coverage baseline, identifies vendored buf exclusions, finds commands lacking help docstrings, and flags which existing goldens need `normalize:` hooks. Planner decomposes into waves (Wave 1: API baselines, Wave 2: coverage depth + normalization hook + help lint + cmdtree regen, Wave 3: straggler cleanup).
