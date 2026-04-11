# Phase 1 Plan Index — cmderr Migration Completion

**Phase goal:** Every command in `cmd/` + `internal/cli/` returns classified `cmderr` sentinels with correct exit codes.
**Requirements:** POLISH-01, POLISH-02, POLISH-03
**Plans:** 18 total across 5 waves (0 → A → B → C → D → Z)

## Wave structure

Waves run strictly sequential: Wave 0 must land before A, A before B, etc. Plans **within** a wave run in parallel (no shared `files_modified`).

## Plans

### Wave 0 — Infrastructure

| Plan | Goal | Commands | Depends |
|------|------|----------|---------|
| 01-cmderr-test-baseline | Add `cmderr_test.go` with ≥90% coverage; create EXIT-CODE-CHANGES.md tracker | (none — touches cmderr package) | — |
| 02-taskfile-coverage-gate | Add `task lint:cmderr-coverage` + wire into CI workflow | (none) | 01 |
| 03-migration-audit | Recompute authoritative work list (MIGRATION-LEDGER.md) from live git state | (none — produces ledger) | — |

### Wave A — CI-critical commands

| Plan | Goal | Commands | Depends |
|------|------|----------|---------|
| 04-wave-a-core | Migrate sort, env (verify+topup), date | sort, env, date | 01, 02, 03 |
| 05-wave-a-kill | Migrate kill with Windows parity + 3-snapshot risky matrix | kill | 01, 02, 03 |

### Wave B — Data / format / encoding

| Plan | Goal | Commands | Depends |
|------|------|----------|---------|
| 06-wave-b-formatters | Migrate cssfmt, htmlfmt, sqlfmt, xmlutil | cssfmt, htmlfmt, sqlfmt, xmlutil | 04, 05 |
| 07-wave-b-structs | Migrate yamlutil, json2struct, yaml2struct, csvutil | yamlutil, json2struct, yaml2struct, csvutil | 04, 05 |
| 08-wave-b-idgen | Migrate ksuid, ulid, nanoid, snowflake | ksuid, ulid, nanoid, snowflake | 04, 05 |

### Wave C — System / proc / info

| Plan | Goal | Commands | Depends |
|------|------|----------|---------|
| 09-wave-c-proc | Migrate ps, pkill (POLISH-09 preview for ps) | ps, pkill | 06, 07, 08 |
| 10-wave-c-diskmem | Migrate df, du, free (POLISH-09 preview for df, free) | df, du, free | 06, 07, 08 |
| 11-wave-c-host | Migrate lsof, ss, uname, yes (Patterns 5 + 6) | lsof, ss, uname, yes | 06, 07, 08 |

### Wave D — Tail

| Plan | Goal | Commands | Depends |
|------|------|----------|---------|
| 12-wave-d-cloud-wrappers | Migrate git, gh, kubectl, kubehacks, terraform, aws, cloud, vault | git, gh, kubectl, kubehacks, terraform, aws, cloud, vault | 09, 10, 11 |
| 13-wave-d-scaffold | Migrate scaffold, project, repo | scaffolding, project, repo | 09, 10, 11 |
| 14-wave-d-dev-tools | Migrate tree (Risky), pipe (Risky), exec (Risky), buf (verify+topup), task, lint, loc, testcheck | tree, pipe, exec, buf, task, lint, loc, testcheck | 09, 10, 11 |
| 15-wave-d-misc-tail | Migrate banner, brdoc, cron, forloop, echo, arch, aicontext, tagfixer, timecmd, note, pager, video | banner, brdoc, cron, forloop, echo, arch, aicontext, tagfixer, timecmd, note, pager, video | 09, 10, 11 |

### Wave Z — Cleanup + enforcement

| Plan | Goal | Commands | Depends |
|------|------|----------|---------|
| 16-wave-z-risky-matrices | Audit + add rich error matrices for 9 risky already-migrated commands | find, sed, awk, dd, grep, curl, tar, jq, diff (YAML + source if bugs found) | 12, 13, 14, 15 |
| 17-wave-z-docs | Update CLAUDE.md to 100%, finalize EXIT-CODE-CHANGES.md, sync docs/BACKLOG.md | (docs only) | 12, 13, 14, 15 |
| 18-wave-z-ci-enforce | Promote coverage gate to required CI check; human checkpoint for branch protection | (none) | 16, 17 |

## Parallelization summary

- Wave 0: 3 plans parallel (01, 02, 03). Plan 02 sequences after 01.
- Wave A: 2 plans parallel (04, 05).
- Wave B: 3 plans parallel (06, 07, 08).
- Wave C: 3 plans parallel (09, 10, 11).
- Wave D: 4 plans parallel (12, 13, 14, 15).
- Wave Z: Plans 16 and 17 parallel; Plan 18 sequences after both.

## Non-goals (out of scope for Phase 1)

- Any `pkg/*` modifications (CONTEXT Decision 2 — CLI boundary only)
- Any `cmd/*` modifications beyond what waves already touch (root.go wiring already in place)
- Raising omni-wide coverage (POLISH-04/05 → Phase 2)
- Full Windows parity implementation (POLISH-09 → Phase 3)
- Generating `docs/EXIT-CODES.md` (Deferred → Phase 3)
- Consolidating the two golden registries (post-1.0 cleanup)
- Adding new `cmderr` sentinels (RESEARCH.md §"Gaps Flagged" confirms none needed)

## Key artifacts produced

- `internal/cli/cmderr/cmderr_test.go` — baseline coverage ≥90%
- `Taskfile.yml` — `lint:cmderr-coverage` target
- `.github/workflows/test.yml` — required CI gate
- `.github/branch-protection.md` — required-checks documentation
- `.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md` — authoritative work list
- `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md` — exit-code transition log (→ v1.0 release notes)
- `testing/golden/golden_tests.yaml` + `tools/golden/golden_tests.yaml` — ~65–75 new error-path snapshots + 35 risky-matrix entries
- Updated `CLAUDE.md` reflecting 100% cmderr adoption
