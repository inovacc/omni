# Phase 1 Session Checkpoint

**Date:** 2026-04-11
**Reason:** Pausing long orchestrator session after 2 executor crashes + context pressure.

## Progress

**Wave 0 — Infrastructure: COMPLETE**
- Plan 01 ✓ `cmderr_test.go` expanded, 100% coverage → commit `697a9c08`
- Plan 02 ✓ `task lint:cmderr-coverage` + CI gate (pure-Go parser in `tools/cmderr-cov/`) → commits `690e1b74`, `0b14aa23`
- Plan 03 ✓ `MIGRATION-LEDGER.md` — **47 commands remain** (not 76), CLAUDE.md was off by 1 → commit `af423a57`

**Wave A — CI-critical: COMPLETE**
- Plan 04 ✓ `env`, `date` migrated. `sort` is a **no-op** — `internal/cli/sort/` is a library helper; real `omni sort` is `internal/cli/text/RunSort` which was already migrated. → commits `19361bfc`, `99a1cd77`
- Plan 05 ⚠ `kill` migrated (unix + windows platform split) — **executor crashed before committing**, recovered manually → commit `f8ffda6c`

**Wave B — data/format: PARTIAL**
- Plan 06 ✓ `cssfmt`, `htmlfmt`, `sqlfmt`, `xmlutil` migrated → commits `c58247c0`, `822a415d`, `e8329b4c`
- Plan 07 ⚠ `yamlutil`, `csvutil`, `json2struct`, `yaml2struct` migrated — **executor crashed mid-Task 2 before committing or adding goldens**, recovered manually → commits `377ba3ad`, `a86b83fb`
- Plan 08 — **NOT STARTED** (idgen subcommands: ksuid, nanoid, snowflake, ulid)

**Wave C — system/proc: NOT STARTED**
- Plans 09 (ps/pkill), 10 (df/du/free), 11 (lsof/ss/uname/yes)

**Wave D — tail: NOT STARTED**
- Plans 12 (cloud wrappers), 13 (scaffold/project/repo), 14 (dev tools incl. exec/tree/pipe), 15 (misc tail incl. video)

**Wave Z — cleanup: NOT STARTED**
- Plans 16 (risky matrices — ALSO backfill missing goldens from Plan 07 recovery), 17 (docs update), 18 (CI enforce)

## Commands migrated this session (11)

`env`, `date`, `kill`, `cssfmt`, `htmlfmt`, `sqlfmt`, `xmlutil`, `yamlutil`, `csvutil`, `json2struct`, `yaml2struct`

Plus 1 confirmed no-op: `sort` (see Plan 04 deviation)

## Known issues & carry-over work

1. **Missing goldens for json2struct + yaml2struct** — Plan 07 crashed before adding golden snapshots. Source code is classified correctly, tests pass, but POLISH-02 golden requirement is not satisfied for these two commands. **Plan 16 must backfill these.**

2. **Executor crash pattern** — 2 of 4 substantive opus executor subagents (Plans 05 and 07) hit the same failure mode: all source edits land cleanly, tests pass, but the agent crashes before committing and writing its final output (both returned truncated garbage strings). No work was lost because the filesystem state was recoverable, but orchestrator had to diagnose + commit manually each time. **Next session should consider:** sonnet executors, smaller plan granularity, or inline execution.

3. **Shared YAML contention** — `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`, and `EXIT-CODE-CHANGES.md` are all modified by every wave plan. Parallel execution risks interleaving. Sequential execution (current approach for Wave B) works but is slow.

4. **CLAUDE.md drift** — root CLAUDE.md says "Commands adopted (84): ..." but live count is 83 pre-migration. Plan 17 must refresh this.

5. **Pre-existing uncommitted dirty tree at session start** — rg + pipe + docs work was committed in 3 cleanup commits (`d2c694de`, `f98d63db`, `8bea021b`) before Phase 1 execution began. Not Phase 1 scope.

## Resume instructions

Next session:
1. `/clear` to reset context
2. `/gsd-execute-phase 1` — init tool will auto-detect which plans have committed work. If it doesn't skip completed plans automatically, manually resume at **Plan 08**.
3. Consider switching executor model to `sonnet` if crashes continue.
4. Sequential execution within waves is working — stay with it for Wave B remainder, C, D.
5. Plan 16 must backfill `json2struct` + `yaml2struct` goldens.

## Commit history (this session, Phase 1 scope)

```
a86b83fb refactor(json2struct,yaml2struct): adopt cmderr sentinels  [Plan 07 recovery]
377ba3ad refactor(yamlutil,csvutil): adopt cmderr sentinels          [Plan 07 Task 1]
e8329b4c test(golden): add cmderr_wave_b error-path snapshots        [Plan 06]
822a415d refactor(sqlfmt,xmlutil): adopt cmderr sentinels            [Plan 06]
c58247c0 refactor(cssfmt,htmlfmt): adopt cmderr sentinels            [Plan 06]
f8ffda6c refactor(kill): adopt cmderr sentinels w/ platform split    [Plan 05 recovery]
99a1cd77 test(golden): add cmderr_wave_a error-path snapshots        [Plan 04]
19361bfc refactor(date,env): adopt cmderr sentinels                  [Plan 04]
0b14aa23 build(phase1): wire cmderr-coverage-gate job into CI        [Plan 02]
690e1b74 build(phase1): add lint:cmderr-coverage Taskfile gate       [Plan 02]
af423a57 docs(phase1): migration audit ledger                        [Plan 03]
d4975585 docs(phase1): add EXIT-CODE-CHANGES.md tracker              [Plan 01]
697a9c08 test(cmderr): expand baseline to 100% coverage              [Plan 01]
```

## Observations for future phases

- **Opus executors on mechanical work may be overkill** — the work is straightforward pattern application; sonnet should be tried for batches 2+.
- **Plan sizing matters** — Plans that touch >4 commands + goldens + exit-code ledger + tests are the ones that crash. Consider splitting future wave plans so each plan covers at most 2–3 commands.
- **Shared-file risk was real** — the plan-checker warned about shared YAML writes; the warning came true as contention-adjacent behavior even under sequential execution (the YAMLs got updated in separate commits that merged cleanly by luck).
