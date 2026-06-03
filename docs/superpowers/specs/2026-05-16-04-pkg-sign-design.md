# Phase 04 — pkg/sign/ — Signing Primitive

**Status:** Planned
**Date:** 2026-05-16 (synthesized from ROADMAP — no phase directory yet)
**Requirements:** SIGN-01 through SIGN-09
**Depends on:** Phase 3
**ADR Gate:** Required before any code lands (2 ADRs)
**Plans:** TBD

---

## Design / Approach / Components

Ship `omni sign` / `omni verify` / `omni sign keygen` with a pure-Go, fail-closed, minisign-compatible signing path — the cryptographic primitive every later supply-chain phase reuses.

**Expected components:**
- `pkg/sign/` — stable public types: `Signer`, `Verifier`, `KeyPair`, `Signature`.
- `internal/cli/sign/` + `internal/cli/verify/` — CLI wrappers registered with `internal/cli/command/` Registry.
- `secret.Key` wrapper with redacting `String()` / `GoString()` / `LogValue()` — prevents key leakage in errors, logs, and panics.
- Ed25519 keypair generation: passphrase-protected, `0600`/`0644` permissions, minisign file format compatible.
- Detached signature files; `omni verify` fails closed on every failure mode.
- `omni_sigstore` build tag: enables `omni verify --bundle <path>` against Sigstore bundles; without tag returns `cmderr.ErrUnsupported`.
- Both commands usable inside `omni pipe` / `omni pipeline`.

**ADR gates (must exist before code):**
1. Key handling policy ADR — storage, rotation, dev-vs-release key separation, never-as-flag rule, `secret.Key` wrapper design.
2. Cosign-compat scope ADR — v1.0 is minisign-only by default; Sigstore bundle *verification* behind `omni_sigstore` build tag; no Rekor, no Fulcio, no OCI.

---

## Rationale & Decisions

| Decision | Rationale |
|----------|-----------|
| Minisign-only default | Keeps default binary lean; Fulcio/Rekor/OCI rejected as v1.0 scope |
| Sigstore behind build tag | Optional capability without inflating default binary |
| Private keys never as CLI flag values | Must use file path or env-var-pointing-to-file only; `secret.Key` wrapper enforces this |
| Fail-closed on every verify failure mode | Pitfall 2 — fail-open verify; all failure modes return `cmderr.ErrConflict` |
| Pitfalls addressed | 2 (fail-open verify), 3 (key handling), 9 (cosign scope), 17 (unmaintained deps) |

---

## Constraints & Assumptions

- Pure Go only — no exec, no system tools.
- `pkg/sign/` usable as standalone Go library (stable public surface after Phase 3 API triage).
- ADR gates must be written and reviewed before any implementation begins.
- Plan decomposition TBD at phase planning time.

---

## Testing & Acceptance

Success criteria (from ROADMAP):
1. `omni sign keygen` produces a passphrase-protected Ed25519 keypair with `0600`/`0644` permissions, compatible with minisign file format.
2. Sign + verify round-trip works; verify fails closed on: missing sig, wrong key, tampered payload, bad algorithm, expired key, unknown key ID — each locked by a golden-master negative test.
3. Private keys never appear in CLI flags, slog output, error messages, or panic traces.
4. `pkg/sign/` usable as standalone library with stable public types; `omni sign`/`omni verify` work inside `omni pipe`.
5. With `-tags omni_sigstore`: `omni verify --bundle <path>` works against Sigstore bundles. Without tag: returns `cmderr.ErrUnsupported` cleanly.
