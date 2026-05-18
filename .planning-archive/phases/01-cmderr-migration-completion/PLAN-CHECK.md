# PLAN-CHECK -- Phase 1: cmderr Migration Completion

**Verdict:** PASS-WITH-NOTES
**Blockers:** 0
**Warnings:** 1
**Info:** 2
**Plans verified:** 18 (Wave 0: 3, Wave A: 2, Wave B: 3, Wave C: 3, Wave D: 4, Wave Z: 3)

---

## Coverage Summary

| Requirement | Plans | Status |
|-------------|-------|--------|
| POLISH-01 (every command returns classified cmderr) | 04,05,06,07,08,09,10,11,12,13,14,15,17 | Covered |
| POLISH-02 (every command has error-path golden test) | 06,07,08,09,10,11,12,13,14,15,16 | Covered |
| POLISH-03 (cmderr coverage gate >= 90% in CI) | 01,02,18 | Covered |

All three requirements have explicit plan-level coverage. POLISH-01 is the broadest -- 13 plans migrate commands. POLISH-02 is addressed via per-wave golden task additions and the Wave Z risky matrix (Plan 16). POLISH-03 is established in Wave 0 (Plans 01-02) and enforced in Wave Z (Plan 18).

---

## Plan Summary

| Plan | Tasks | Type | Wave | Deps | Requirements | Status |
|------|-------|------|------|------|--------------|--------|
| 01-cmderr-test-baseline | 3 | auto | 0 | -- | POLISH-03 | Valid |
| 02-taskfile-coverage-gate | 3 | auto | 0 | 01 | POLISH-03 | Valid |
| 03-migration-audit | 2 | auto | 0 | -- | POLISH-01 | Valid |
| 04-wave-a-core | 3 | auto | A | 01,02,03 | POLISH-01,02 | Valid |
| 05-wave-a-kill | 4 | auto | A | 01,02,03 | POLISH-01,02 | Valid (note) |
| 06-wave-b-formatters | 3 | auto | B | 04,05 | POLISH-01,02 | Valid |
| 07-wave-b-structs | 3 | auto | B | 04,05 | POLISH-01,02 | Valid |
| 08-wave-b-idgen | 3 | auto | B | 04,05 | POLISH-01,02 | Valid |
| 09-wave-c-proc | 3 | auto | C | 06,07,08 | POLISH-01,02 | Valid |
| 10-wave-c-diskmem | 3 | auto | C | 06,07,08 | POLISH-01,02 | Valid |
| 11-wave-c-host | 3 | auto | C | 06,07,08 | POLISH-01,02 | Valid |
| 12-wave-d-cloud-wrappers | 3 | auto | D | 09,10,11 | POLISH-01,02 | Valid |
| 13-wave-d-scaffold | 3 | auto | D | 09,10,11 | POLISH-01,02 | Valid |
| 14-wave-d-dev-tools | 4 | auto | D | 09,10,11 | POLISH-01,02 | Valid |
| 15-wave-d-misc-tail | 4 | auto | D | 09,10,11 | POLISH-01,02 | Valid |
| 16-wave-z-risky-matrices | 3 | auto | Z | 12,13,14,15 | POLISH-02 | Valid |
| 17-wave-z-docs | 3 | auto | Z | 12,13,14,15 | POLISH-01 | Valid |
| 18-wave-z-ci-enforce | 4 | auto+checkpoint | Z | 16,17 | POLISH-03 | Valid |

---

## Dimension Results

### Dimension 1: Requirement Coverage -- PASS

All three requirement IDs appear in at least one plan requirements frontmatter. No requirement is orphaned.

- POLISH-01: migration wave plans (04-15) and docs (17)
- POLISH-02: per-wave golden tasks and Plan 16 risky matrices
- POLISH-03: Plan 01 (baseline test), Plan 02 (Taskfile gate), Plan 18 (CI enforcement)

### Dimension 2: Task Completeness -- PASS

All tasks across all 18 plans have required structure: Files present, Action specific, Verify automated, Done measurable.

Plan 18 Task 3 is checkpoint:human-action -- enabling GitHub branch protection requires admin UI access. Checkpoint includes resume signal and CLI alternative via gh api. No field requirements apply to checkpoint tasks.

### Dimension 3: Dependency Correctness -- PASS

Dependency graph is acyclic:
- Wave 0: Plans 01 and 03 parallel (no deps). Plan 02 depends on 01.
- Wave A: Plans 04,05 depend on [01,02,03]. Parallel.
- Wave B: Plans 06,07,08 depend on [04,05]. Parallel.
- Wave C: Plans 09,10,11 depend on [06,07,08]. Parallel.
- Wave D: Plans 12,13,14,15 depend on [09,10,11]. Parallel.
- Wave Z: Plans 16,17 depend on [12,13,14,15]. Parallel. Plan 18 depends on [16,17].

No forward references. No cycles. All referenced plan IDs exist.

### Dimension 4: Key Links Planned -- PASS

- Plan 01: cmderr_test.go exercises real sentinels via internal/cli/cmderr package.
- Plan 02: Taskfile.yml lint:cmderr-coverage calls go test + go tool cover -- wiring explicit.
- Plan 02 -> Plan 18: CI workflow references job name from Plan 02; Plan 18 Task 1 stabilizes it.
- Plans 04-15: specify internal/cli/<cmd>/<cmd>.go boundary and cmderr.Wrap pattern. pkg/* preserved per Decision 2.
- Plans 04-15 -> registries: each writes to both testing/golden/ and tools/golden/ registries.
- Plan 16: risky matrix exit_code assertions link snapshots to cmderr sentinels.

### Dimension 5: Scope Sanity -- PASS

Max tasks per plan: 4 (Plans 05, 14, 15, 18) -- within the 4-task warning threshold. No plan reaches 5.

Plan 15 is largest (12 commands, 15 files) but uses Pattern 5 (write errors only) -- simplest migration class. Acceptable within context budget.

### Dimension 6: Verification Derivation -- PASS

All must_haves.truths are user-observable:
- Every command returns classified cmderr sentinel -- verifiable via exit code
- PR that drops coverage below 90% fails -- observable via CI gate
- Both golden registries contain full matrix -- verifiable with task test:golden

Artifacts map to truths. Key_links connect artifacts to functionality.

### Dimension 7: Context Compliance -- PASS

All 5 locked decisions implemented:
- Decision 1 (Wave ordering): Plans follow Wave 0->A->B->C->D->Z. Runtime scope from Plan 03.
- Decision 2 (pkg/ boundary): Every migration plan states pkg/* stays raw-error. No plan touches pkg/*.
- Decision 3 (golden breadth): Plans 04-15 include minimum one error snapshot. Plan 16 has 35-entry rich matrices.
- Decision 4 (coverage gate): Plan 01 adds cmderr_test.go, Plan 02 adds lint:cmderr-coverage.
- Decision 5 (platform sentinel): Plan 05 specifies Windows parity mapping for kill.

Deferred ideas excluded: EXIT-CODES.md (Phase 3), cmderr.Is helpers (Phase 3), golangci-lint rule (Phase 2), cross-command matrix (Phase 2+) -- none in any plan.

No contradictions. No deferred ideas in scope.

### Dimension 7b: Scope Reduction Detection -- PASS

Scanned all task actions for scope-reduction language. No instances found. Plan 05 Task 4 platform-gated is accurate technical scoping (Windows supports INT, KILL, TERM only), not reduction.

### Dimension 8: Nyquist Compliance -- PASS

All auto tasks include an automated verify command. No watch-mode flags.

Verify commands: Wave 0 (go test -race, grep), Wave A-D (go test -race per command), Golden tasks (task test:golden --filter), Wave Z (task lint:cmderr-coverage, task test:golden).

No window of 3 consecutive tasks without automated verification.

### Dimension 9: Cross-Plan Data Contracts -- WARNING

Issue: Concurrent writes to shared files during parallel wave execution.

Plans 12, 13, 14, 15 (Wave D, parallel) all list in files_modified:
  testing/golden/golden_tests.yaml
  tools/golden/golden_tests.yaml
  EXIT-CODE-CHANGES.md

Same pattern in Waves B (Plans 06-08) and C (Plans 09-11). PLAN-INDEX.md promises no shared files in parallel plans but these three are shared.

Mitigating factor: gsd-execute-phase runs plans sequentially within a wave (parallelization is human-review-level, not file-write-level). Risk does not materialize if this holds.

WARNING, not blocker.

### Dimension 10: CLAUDE.md Compliance -- PASS

- No exec rule: migration wraps at internal/cli boundary. No plan adds os/exec calls.
- io.Writer pattern: all migration patterns use Run(w io.Writer, ...) signature.
- Error wrapping: fmt.Errorf fallback, cmderr.Wrap for classified errors.
- errors.Is usage: errors.Is(err, os.ErrNotExist) specified -- correct.
- Table-driven tests: Plan 01 adds cmderr_test.go with table-driven tests.
- Both registries: all golden entries update both testing/golden/ and tools/golden/.

### Dimension 11: Research Resolution -- PASS

All open questions from RESEARCH.md resolved via CONTEXT.md decisions. Plan 03 handles dynamic remaining-command count.

---

## Issues

```yaml
issues:
  - plan: null
    dimension: cross_plan_data_contracts
    severity: warning
    description: >
      Waves B, C, D each have 3-4 parallel plans all listing
      testing/golden/golden_tests.yaml, tools/golden/golden_tests.yaml,
      EXIT-CODE-CHANGES.md in files_modified. PLAN-INDEX.md states
      no shared files in parallel plans but these are shared.
      Concurrent writes risk YAML corruption.
    fix_hint: >
      Confirm executor serializes within-wave writes (expected default).
      If truly parallel, designate one plan per wave as the appender.

  - plan: "03"
    dimension: requirement_coverage
    severity: info
    description: >
      Plan 03 produces MIGRATION-LEDGER.md at runtime from live git state.
      Phase 1 command coverage cannot be statically verified before
      Plan 03 executes. This is intentional dynamic scoping.
    fix_hint: >
      No action required. Re-check MIGRATION-LEDGER.md after Plan 03
      executes to confirm full coverage of remaining commands.

  - plan: "05"
    task: 4
    dimension: task_completeness
    severity: info
    description: >
      Plan 05 Task 4 adds platform-gated golden snapshots for Windows-only
      kill signal behavior. The verify command (task test:golden) may skip
      platform-gated entries on Linux CI without explicit harness support.
    fix_hint: >
      Acceptable if golden harness supports platform gating (per RESEARCH.md).
      Optionally add a platform: field to make the gate explicit.
```

---

## Recommendation

**0 blockers. Plans are ready to execute.**

The single WARNING (shared file writes in parallel waves) is mitigated by executor sequential behavior and does not block execution. The two INFO items require no action.

Run `/gsd-execute-phase 1` starting with Wave 0.

**Execution order:**
1. Wave 0: Plans 01 and 03 (parallel) -> Plan 02 (after 01)
2. Wave A: Plans 04 and 05 (parallel)
3. Wave B: Plans 06, 07, 08 (parallel)
4. Wave C: Plans 09, 10, 11 (parallel)
5. Wave D: Plans 12, 13, 14, 15 (parallel)
6. Wave Z: Plans 16 and 17 (parallel) -> Plan 18 (after both)
