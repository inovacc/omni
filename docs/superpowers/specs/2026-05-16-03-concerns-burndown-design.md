# Phase 03 — CONCERNS Burn-down, Windows Parity, pkg/ API Audit, Docs

**Status:** Complete (2026-06-03) — security/robustness half delivered by the hardening sweep (`docs/quality/HARDENING.md`); the non-security remainder (Windows `id` helpers, `cmderr.Is*`, `EXIT-CODES.md`, `pkg/*` Experimental triage, ISSUES→empty, no-exec TODO reconciliation) executed per `docs/superpowers/plans/2026-06-03-phase-03-concerns-burndown.md` on branch `harden/audit-fixes`. Deferred: crypto-02 envelope + darwin machine-id (tracked in `docs/BACKLOG.md`).
**Date:** 2026-05-16 (synthesized from ROADMAP — no phase directory yet)
**Requirements:** POLISH-08, POLISH-09, POLISH-10, POLISH-15, POLISH-16, POLISH-17
**Depends on:** Phase 2
**Plans:** `docs/superpowers/plans/2026-06-03-phase-03-concerns-burndown.md`

---

## Design / Approach / Components

Close the tech-debt backlog, resolve or explicitly document Windows parity gaps, and triage the `pkg/*` public surface before the API freeze so experimental code can still move.

**Expected work streams:**
- Resolve or explicitly defer (with written rationale) every Critical and High item in `.planning/codebase/CONCERNS.md`.
- Fix or return `cmderr.ErrUnsupported` for Windows parity gaps in `ps`, `df`, `free`, `kill`, `uptime`, `id`.
- Triage `docs/ISSUES.md` to empty — every bug either fixed, backlogged, or closed with rationale.
- Triage every exported `pkg/*` symbol into stable / experimental / internal buckets.
  - Experimental symbols carry `// Experimental:` godoc.
  - Internal symbols moved or unexported.
- Audit and stabilize (or restrict) `pkg/video/` public API.
- Adopt CLAUDE.md breaking-change protocol for every future `pkg/*` public-API change.
- Generate `docs/EXIT-CODES.md` (deferred from Phase 1).
- Add `cmderr.Is<Class>()` convenience helpers (deferred from Phase 1).

---

## Rationale & Decisions

- Tech debt must be resolved before supply-chain phases — inconsistent exit-code semantics on security commands (Pitfall 13) is the forcing function.
- `pkg/*` API triage before the freeze prevents accidental stabilization of experimental symbols.
- Windows users must never see a silent runtime branch or a cryptic syscall error — explicit `ErrUnsupported` is the contract.
- Deferred items from Phase 1 (`EXIT-CODES.md`, `Is<Class>()` helpers, custom golangci-lint rule) land here.

---

## Constraints & Assumptions

- ADR gates: none required for Phase 3 entry.
- Must complete before Phase 4 — supply-chain work cannot start with open tech debt.
- Pitfalls addressed: 7 (pkg/ surface sprawl), 8 (Windows gaps), 16 (undocumented exit codes).
- Plan decomposition TBD at phase planning time.

---

## Testing & Acceptance

Success criteria (from ROADMAP):
1. Every Critical and High CONCERNS.md item resolved or explicitly deferred in `docs/BACKLOG.md` — no silent "we'll figure it out" items remain.
2. Windows parity gaps in `ps`, `df`, `free`, `kill`, `uptime`, `id` resolved or returning `cmderr.ErrUnsupported` with a clear message.
3. `docs/ISSUES.md` triaged to empty.
4. Every currently-exported `pkg/*` symbol triaged into stable/experimental/internal; experimental symbols carry `// Experimental:` godoc.
5. `pkg/video/` public API audited and either stabilized or moved behind a smaller frozen surface; CLAUDE.md breaking-change protocol adopted for all future `pkg/*` changes.
