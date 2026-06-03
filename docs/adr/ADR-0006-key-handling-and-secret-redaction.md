# ADR-0006: Key-handling policy & secret redaction

**Status:** Accepted
**Date:** 2026-06-03
**Decision:** Private key material is never accepted as a CLI flag value, secret-key files are permission-checked, dev and release keys are separated, and all in-memory secret bytes are wrapped in a `pkg/secret.Key` type that redacts itself in every textual and structured-logging surface.

## Context

Phase 04 introduces `omni sign` / `omni verify` / `omni sign keygen` — a pure-Go, minisign-compatible Ed25519 signing primitive (`pkg/sign/`). This is the cryptographic foundation that every later supply-chain phase (sbom/scan/attest/release) reuses, so the way the project handles private keys and passphrases must be locked down once, here, before any code lands.

omni's foundational principles are relevant: no `os/exec`, stdlib-first, cross-platform via `//go:build` tags, fail-closed defaults. Secret handling must respect all of these. Two distinct leak classes need addressing:

1. **Ingress leaks** — secret material arriving through an observable channel (CLI flag values appear in shell history, process listings (`ps`/`/proc`), CI logs, and `--help` echoes; environment variables holding the *material* are nearly as exposed).
2. **Egress leaks** — secret material leaving through a logging/diagnostic channel (`slog` structured logs, `fmt.Sprintf("%v"/"%s"/"%#v", key)`, error strings, and panic dumps). Go has **no built-in redaction helper anywhere** in the codebase today, so an unwrapped `[]byte` or `ed25519.PrivateKey` would render its raw bytes in any of these.

The invariant from the phase plan is explicit: *secret key material NEVER in logs/errors/flags*. This ADR records the policy and the wrapper design that enforce it; `pkg/secret` (Task 3) and the CLI flag policy (Task 7) implement it.

## Analysis

### Decision summary

| Concern | Policy | Rationale |
|---------|--------|-----------|
| Private key ingress | `--key <path>` (file path only) or `OMNI_SIGN_KEY=<path>` (env var pointing *at a file*) — never inline key material | Flag values leak to shell history, `ps`, and CI logs; an env var holding a *path* is safe, an env var holding the *bytes* is not |
| Passphrase ingress | Interactive prompt (default) or `OMNI_SIGN_PASSPHRASE` env var (documented tradeoff) — never a flag value | Prompts never touch history/process args; the env var is an explicit, documented CI escape hatch, not the default |
| Secret-key file perms | Written `0600`; refuse (hard error) to *read* a secret key that is group/world-accessible on Unix | A `0600` key that becomes `0644` is the classic exposure; refusing to read forces the operator to fix it |
| Public-key file perms | Written `0644` | Public keys are not secret; world-readable is correct |
| Windows perms | Warn, do not hard-fail, when Unix-style perm bits are not meaningful | NTFS ACLs do not map to `rwx` triplets; a hard-fail would make the tool unusable on Windows without adding value |
| Dev vs release keys | Distinct key IDs; release private keys live **outside** the repo | A leaked dev key cannot forge release artifacts; key-ID separation makes provenance auditable |
| In-memory representation | `pkg/secret.Key` wrapper, raw bytes reachable only via an explicit accessor | Makes the safe path the easy path: anything that prints a `Key` gets a placeholder, not the bytes |

### Ingress: keys and passphrases are never flag values

A private key is supplied **only** as a filesystem path:

- `--key <path>` — the path to a `*.key` (secret) or `*.pub` (public) file.
- `OMNI_SIGN_KEY=<path>` — an environment variable that *points to a file*. The variable holds a path, never the key bytes.

The CLI must reject a `--key` value that looks like inline key material (e.g. begins with the `untrusted comment:` minisign header or decodes as a key body) rather than a path — this catches an operator who pastes a key where a path is expected, before it reaches shell history or a CI log.

A passphrase is supplied via:

- an interactive terminal prompt (default, with echo disabled), or
- `OMNI_SIGN_PASSPHRASE` — a documented tradeoff for non-interactive CI. This is *not* a flag, so it never appears in `ps`, `--help`, or shell history. CI users are responsible for sourcing it from a secret store and not echoing it.

There is deliberately **no `--passphrase` and no `--key-bytes`/`--secret` flag**. Adding one later would be a security regression and would require the breaking-change/deprecation protocol.

### Egress: the `pkg/secret.Key` wrapper

`pkg/secret.Key` is a small value type wrapping `[]byte` whose entire job is to be impossible to accidentally print:

- `String() string` (`fmt.Stringer`) → returns `<REDACTED>`. Covers `%v` and `%s`.
- `GoString() string` (`fmt.GoStringer`) → returns `secret.Key{<REDACTED>}`. Covers `%#v`.
- `LogValue() slog.Value` (`slog.LogValuer`) → returns `slog.StringValue("<REDACTED>")`. Covers structured logging, the highest-volume egress path.
- Raw bytes are reachable **only** via an explicit `Bytes()` accessor (and may also be exposed via an `Open()`-style method) intended for controlled cryptographic use at the call site — the friction is intentional.
- `Destroy()` zeroes the underlying bytes so a key can be scrubbed after use. (Best-effort: Go's GC may copy the slice; this is defense-in-depth, not a hard guarantee.)

`sign.SecretKey` stores its decrypted Ed25519 private key inside a `secret.Key`, so the secret never exists as a bare, printable `[]byte`/`ed25519.PrivateKey` on the `SecretKey` struct. Error values constructed in `pkg/sign` and `internal/cli/{sign,verify}` must never interpolate raw key bytes; they reference the key by *path* or by *key ID* (which is public) only.

`pkg/secret` is a **stable v1.0 API** — no `// Experimental:` marker. It depends on `log/slog` and nothing else, keeping it importable from anywhere in the tree without a dependency cycle.

### Why not an external secret-handling library

The wrapper is ~30 lines over stdlib (`log/slog` + manual zeroing). An external library would add a transitive dependency for a trivial, security-critical surface we want full control over — the same Pitfall-17 reasoning that ADR-0005 applies to the signing scheme itself. Stdlib-first wins.

## Decision

**Adopt the table above as the binding key-handling policy** and implement the `pkg/secret.Key` wrapper to enforce egress redaction:

1. Private keys enter only as file paths (`--key`, `OMNI_SIGN_KEY`); never inline. The CLI rejects values that look like key material.
2. Passphrases enter only via interactive prompt or `OMNI_SIGN_PASSPHRASE`; never a flag.
3. Secret keys are written `0600` and **refused on read** if group/world-accessible on Unix; on Windows the perm check warns instead of failing.
4. Public keys are written `0644`.
5. Dev and release keys use distinct key IDs; release private keys are kept out of the repository.
6. All in-memory secret bytes live inside `pkg/secret.Key`, which redacts itself for `fmt.Stringer`, `fmt.GoStringer`, and `slog.LogValuer`, exposes raw bytes only through an explicit accessor, and supports `Destroy()` zeroing.

## Consequences

- **`pkg/secret` (Task 3)** implements `Key` exactly as described; its tests assert that the raw bytes never appear in `String`/`%v`/`%s`/`%#v`/`GoString`/`slog` output.
- **`pkg/sign` (Tasks 4–6)** stores the decrypted secret inside `secret.Key`; no API returns or logs raw private bytes.
- **The CLI glue (Task 7)** enforces the ingress policy: path-only `--key`, env passphrase, perm checks (`cmderr.ErrPermission` on a too-open key on Unix, warning on Windows), and rejection of inline-key-material `--key` values (`cmderr.ErrInvalidInput`).
- **Cross-platform:** the perm check splits by build tag (`_unix.go` / `_windows.go`) — Unix hard-fails on group/world bits, Windows warns. No silent runtime OS branching.
- **CI tradeoff documented:** `OMNI_SIGN_PASSPHRASE` is the supported non-interactive path; users must source it from a secret store. This is a conscious, recorded tradeoff, not an oversight.
- **Future-proofing:** introducing any flag that carries key bytes or a passphrase would be a security regression governed by the breaking-change/deprecation protocol — effectively prohibited for v1.0.
- **Best-effort zeroing:** `Destroy()` reduces, but cannot fully eliminate, the window in which secret bytes reside in process memory (GC copies, swap). This is acceptable defense-in-depth, not a formal guarantee, and is noted so future readers do not over-trust it.
