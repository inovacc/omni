# Phase 1 Migration Ledger

**Generated:** 2026-04-11
**Source:** live git state (`git grep 'cmderr\.' -- 'internal/cli/*/*.go'`), **not** RESEARCH.md's 2026-04-11 snapshot.
**Method:** for each subdirectory of `internal/cli/`, classify as migrated iff `git grep -lq 'cmderr\.'` returns a hit inside that directory (test files included — presence of the import/usage anywhere in the package is the signal).

## Totals

| Metric                                          | Count   |
| ----------------------------------------------- | ------- |
| Total subdirs under `internal/cli/`             | **138** |
| Excluded (non-command helpers)                  | **8**   |
| Effective command directories                   | **130** |
| Already migrated (has `cmderr.` refs)           | **83**  |
| Needs migration (no `cmderr.` refs, in scope)   | **47**  |
| Verify + top-up (cmderr present, coverage TBD)  | 0 confirmed today (flagged candidates below) |

**Note:** CLAUDE.md currently claims "84 commands adopted". Live count is **83**. The discrepancy is almost certainly drift in CLAUDE.md's hand-maintained list; a CLAUDE.md refresh is part of Wave Z. No correction required in this phase.

## Divergence from RESEARCH.md snapshot (2026-04-11)

RESEARCH.md's Open Question 2 asked whether `env`, `pipe`, and `buf` were partially migrated. **Live state: all three have ZERO `cmderr` references in their package directories.**

| Command | RESEARCH.md status            | Live status (today)   | Action                 |
| ------- | ----------------------------- | --------------------- | ---------------------- |
| `env`   | "partially migrated"          | **0 cmderr refs**     | Treat as full migration, Wave A |
| `pipe`  | "partially migrated"          | **0 cmderr refs**     | Treat as full migration, Wave A |
| `buf`   | "buf build/format/lint done"  | **0 cmderr refs**     | Treat as full migration, Wave B |
| `sort`  | listed adopted in CLAUDE.md   | **0 cmderr refs**     | **Re-migrate**, Wave A |

Likely explanation: CLAUDE.md's adopted-list was updated optimistically, or these commands were refactored (e.g., `sort` split out of `text`) without carrying cmderr wrapping across the split. Either way, the live-git truth is what Wave A–D must execute against.

No commands appear in RESEARCH.md's remaining-list that are already migrated per live grep (no false-positive work items).

## Excluded (non-command helpers) — 8 dirs

Per CONTEXT.md Decision 1 ("Don't chase commands that are test.go-only wrappers, timeutil helpers, or safepath libraries"). These are shared support code, not user-facing Cobra commands.

| Dir            | Reason                                                                  |
| -------------- | ----------------------------------------------------------------------- |
| `cmderr`       | The sentinel package itself                                             |
| `command`      | Unified Command interface + Registry + adapters                         |
| `input`        | Input helper library                                                    |
| `safepath`     | Path-safety helper library                                              |
| `timeutil`     | Time helper library                                                     |
| `scaffolding`  | Scaffolding sub-libs (cobra/handler/repository/testgen/mcp). **Note:** the `scaffold` subcommand wiring lives in `cmd/scaffold.go` + these sub-libs; its cmderr classification is handled by whichever Wave plan touches it (tracked in Wave D as `scaffold`). |
| `testcheck`    | Test-infrastructure helper                                              |
| `echo`         | Excluded: `cmd/echo.go` is the Cobra wrapper; the dir has no surface of its own beyond trivial write-to-writer — classified under Wave D "trivial no-classifiable-error" per CONTEXT Decision on yes/printf/seq. **Kept in scope — not truly excluded.** (Reclassified below — remove from this list.) |

**Corrected excluded count: 7** (`echo` moved into Wave D).

## Already Migrated (83 dirs, no action)

```
archive awk banner base basename bbolt bzip2 caseconv cat chmod chown cmp
column comm copy cron crypt curl cut dd diff dirname dotenv exist file find
fold fs grep gzip hash head hexenc htmlenc id join jq jsonfmt jwt lint ln
loc ls mkdir nl note pager paste path pipeline printf pwd random readlink
realpath rev rg rm sed seq shuf sleep split sqlite stat strings tac tail
text tomlutil tr uptime urlenc uuid watch wc which whoami xargs xmlfmt xxd
xz yq
```

Per CONTEXT Decision "Golden test breadth — one error snapshot per command (minimum)", each of these still needs **at least one error-path golden snapshot** if it doesn't already have one. That's Plan 20 (golden snapshot wave), not a re-migration.

Rich-surface commands already migrated that need a **3–5 snapshot error matrix** per CONTEXT.md: `find`, `sed`, `awk`, `dd`, `grep`, `curl`, `tar` (in `archive/`), `jq`, `diff`.

## Needs Migration — Wave Assignment (47 dirs)

Wave ordering follows CONTEXT.md Decision "Migration Order — Risk-weighted" (A = CI-critical; B = data/format/encoding; C = system/proc/info; D = everything else). Wave sizes should be 3–5 plans; counts below inform whether re-split is needed.

### Wave A — CI-critical, user-facing, highest blast radius (7 dirs)

| Dir      | Notes |
| -------- | ----- |
| `env`    | Was claimed adopted in CLAUDE.md; live grep: 0 refs. Full migration. |
| `pipe`   | Same as `env`. `execute.go`/`parse.go`/`substitute.go` all need classification. |
| `sort`   | Was claimed adopted in CLAUDE.md; live grep: 0 refs. Standalone package now (split from `text`). |
| `date`   | Date formatting errors → `ErrInvalidInput`. |
| `df`     | Filesystem stat errors → `ErrIO`/`ErrPermission`. Platform-split. |
| `du`     | Directory-walk errors → `ErrIO`/`ErrPermission`/`ErrNotFound`. |
| `tree`   | Large surface — already has parallel scanner + compare; classify walk errors + MaxFiles/MaxHashSize validation. |

**Wave A size fits one plan** (small enough, but `tree` is complex; if needed, split into A1 = `env pipe sort date` / A2 = `df du tree`). Final split decided in Plan 04–05 authorship.

### Wave B — data/format/encoding (10 dirs)

| Dir            | Notes |
| -------------- | ----- |
| `buf`          | Was claimed adopted; live grep: 0 refs. Touches build/format/lint/generate/proto. Rich-surface → needs error matrix, not just one snapshot. |
| `sqlfmt`       | SQL format/minify/validate → `ErrInvalidInput` on parse error. |
| `cssfmt`       | Same pattern as sqlfmt. |
| `htmlfmt`      | Same pattern. |
| `xmlutil`      | XML validate/tojson/fromjson. |
| `yamlutil`     | YAML validate/tostruct. |
| `csvutil`      | CSV round-trips. |
| `json2struct`  | JSON → Go struct codegen; parse failures → `ErrInvalidInput`. |
| `yaml2struct`  | YAML → Go struct codegen. |
| `brdoc`        | Brazilian document formatter/validator (CPF/CNPJ/etc.) → `ErrInvalidInput`. |

**Wave B size: 1–2 plans** (10 commands, mostly mechanical; `buf` justifies its own plan given multi-file surface).

### Wave C — system/proc/info (16 dirs)

| Dir         | Notes |
| ----------- | ----- |
| `ps`        | Process listing; platform-split. |
| `kill`      | Signal send; platform-split. Already tested. |
| `free`      | Memory stats; platform-split. |
| `uname`     | Trivial — write-only → `ErrIO` on stdout failure only. |
| `aicontext` | Project introspection; likely I/O errors. |
| `arch`      | Trivial — `ErrIO` only. |
| `cloud`     | Cloud-provider detect. |
| `gh`        | GitHub hacks — network + subprocess-free? Verify. |
| `git`       | Git hacks wrapper. |
| `kubectl`   | kubectl passthrough — may already rely on kubectl's exit codes; investigate before wrapping. |
| `kubehacks` | kubectl shortcuts. |
| `lsof`      | Open-files lister; platform-split. |
| `pkill`     | Pattern-based kill; platform-split. |
| `ss`        | Socket stats; platform-split. |
| `task`      | Taskfile runner wrapper. |
| `vault`     | Vault interactions — validate network vs input errors. |

**Wave C size: 2–3 plans** (`kubectl`/`git`/`gh`/`task` have subcommand trees; they likely each merit their own plan or a shared "wrappers" plan).

### Wave D — everything else, tail (14 dirs)

| Dir          | Notes |
| ------------ | ----- |
| `echo`       | Trivial write-only → `ErrIO`. |
| `yes`        | Trivial write-only → `ErrIO`. |
| `exec`       | Subprocess runner — needs careful error classification (spawn errors vs child exit codes). |
| `forloop`    | Loop helper — argument parse → `ErrInvalidInput`. |
| `tagfixer`   | Go struct-tag fixer; parse errors → `ErrInvalidInput`. |
| `project`    | Project introspection (info/deps/docs/git/health). |
| `repo`       | Repo analyzer — already has test file; check surface. |
| `video`      | Video downloader — network-heavy; likely `ErrIO`/`ErrNotFound`/`ErrUnsupported`. Rich surface; may need error matrix. |
| `ksuid`      | ID generator — write-only. |
| `nanoid`     | ID generator — write-only. |
| `snowflake`  | ID generator — write-only. |
| `ulid`       | ID generator — write-only. |
| `timecmd`    | Time duration formatter. |
| `aws`        | AWS wrapper (in CONTEXT but not `cloud`). |
| `terraform`  | Terraform wrapper. |

That's 15 — recount: echo, yes, exec, forloop, tagfixer, project, repo, video, ksuid, nanoid, snowflake, ulid, timecmd, aws, terraform = **15 dirs**.

**Wave D size: 2–3 plans** (5 ID generators + trivial write-only commands can be batched; wrappers like `aws`/`terraform`/`video` each warrant individual attention).

## Wave sizing summary

| Wave | Dirs | Suggested plan count |
| ---- | ---- | -------------------- |
| A    | 7    | 1–2 plans            |
| B    | 10   | 1–2 plans            |
| C    | 16   | 2–3 plans            |
| D    | 15   | 2–3 plans            |
| **Total** | **48** | **6–10 plans**  |

**Grand total:** 47 needs-migration + **1 re-tally**: `scaffold` (the user-facing command, wired in `cmd/scaffold.go`, backed by `internal/cli/scaffolding/` sub-libs) is **not** counted as a separate `internal/cli/` dir but **does** need error-classification work inside `cmd/scaffold.go`. Track it as a footnote for Wave D; does not change dir counts.

**Re-split recommendation:** The initial Phase 1 plan deck reserves Wave A for 2 plans (04, 05). With only 7 dirs, 1 plan may suffice — combine into Plan 04. Waves B/C/D absorb the rest. Final plan allocation is decided by the `/gsd-plan-phase` re-entry after this ledger lands.

## Verify + Top-up (0 confirmed)

No directories today have a partial cmderr state (some refs + some raw errors) that would warrant a top-up-only pass. If a Wave A–D plan discovers partial coverage inside an "already migrated" directory at execution time, that discovery becomes a Rule 1 deviation on the relevant plan, not a new Wave 0 line-item.

## Sources of truth

- **Authoritative:** this file (`MIGRATION-LEDGER.md`).
- **Superseded:** RESEARCH.md's Open Question 2 (resolved above), CLAUDE.md's "84 commands adopted" line (drift; update in Wave Z).
- **Wave A–D plans (04–19):** must reference this ledger in their frontmatter `context:` block, not RESEARCH.md's snapshot list.
