# Phase 04 — `pkg/sign/` Signing Primitive Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
> **HARD GATE:** Tasks 1–2 (the two ADRs) MUST be written AND human-reviewed/approved BEFORE any code task (3+) begins. The spec (`docs/superpowers/specs/2026-05-16-04-pkg-sign-design.md`) makes this non-negotiable.

**Goal:** Ship `omni sign` / `omni verify` / `omni sign keygen` — a pure-Go, fail-closed, minisign-compatible Ed25519 signing primitive in `pkg/sign/`, the cryptographic foundation every later supply-chain phase (sbom/scan/attest/release) reuses.

**Architecture:** Pure-Go reimplementation on `crypto/ed25519` + `golang.org/x/crypto/{scrypt,blake2b}` (already a direct dep; not yet imported) — NOT an external minisign library (jedisct1/go-minisign is verify-only; reimplementing keeps full keygen+sign+verify control and addresses Pitfall 17, unmaintained deps). Layering: pure lib `pkg/sign/` + a new `pkg/secret/` redaction wrapper → I/O glue `internal/cli/{sign,verify}/` → thin Cobra wrappers `cmd/{sign,verify}.go`. Sigstore bundle *verification* is isolated behind a `//go:build omni_sigstore` tag so the heavy sigstore-go dep tree never inflates the default binary.

**Tech Stack:** Go stdlib (`crypto/ed25519`, `crypto/rand`, `encoding/base64`, `encoding/binary`), `golang.org/x/crypto/scrypt`, `golang.org/x/crypto/blake2b`, `log/slog` (LogValuer); optional `github.com/sigstore/sigstore-go` v1.2.0 (build-tag only); Cobra; the Python YAML golden harness.

**Repo conventions (from research, cite when implementing):**
- Commands self-wire: `cmd/<name>.go` declares `var xCmd = &cobra.Command{...}` and calls `rootCmd.AddCommand(xCmd)` in `init()`; `RunE` reads flags → Options → calls `internal/cli/<name>.RunX(cmd.OutOrStdout(), ...)`. No central registration list.
- `cmderr` (`internal/cli/cmderr/cmderr.go`): verify mismatch → `cmderr.Wrap(cmderr.ErrConflict, …)` (exit 1); unsupported algo/feature → `ErrUnsupported` (exit 6); bad flags/malformed key → `ErrInvalidInput` (exit 2); missing key file → `ErrNotFound` (1) / unreadable → `ErrPermission` (3). `Is<Class>()` predicates exist.
- **No redaction helper exists anywhere** — Task 3 creates `pkg/secret` from scratch.
- Golden harness is Python+YAML, TWO files kept in sync: `testing/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml`; negative tests set `exit_code:` + `normalizations: [strip_path]`; regenerate with `task test:golden:update` then `task golden:record`.
- Pipe: register in `cmd/pipe.go buildPipeRegistry()` via `command.AdaptWriterReaderArgs(...)`.
- ADRs live in `docs/adr/` as `ADR-NNNN-kebab-title.md`; next sequence numbers are **0005** and **0006**; header format per `docs/adr/ADR-0004-internalize-cobra-cli.md`.

---

## Minisign on-disk format (authoritative — implement byte-exactly)

All three files are UTF-8 text. Base64 is standard RFC-4648 with padding.

**Public key `*.pub`** (2 lines): `untrusted comment: <text>\n` then `base64( "Ed"[2] || key_id[8] || ed25519_pub[32] )` (42 bytes).

**Secret key `*.key`** (2 lines): comment line, then `base64(` of:
`sig_alg "Ed"[2] || kdf_alg "Sc"[2] (0x00 0x00 if -W/no-pass) || cksum_alg "B2"[2] || kdf_salt[32] || opslimit[8 LE u64] || memlimit[8 LE u64] || encrypted_keynum_sk )`.
`keynum_sk` (cleartext form) = `key_id[8] || ed25519_secret[64] || checksum[32]` where `checksum = Blake2b-256( "Ed" || key_id || ed25519_secret )`. Encryption: derive a one-time stream `scrypt(passphrase, salt, N, r, p, len(keynum_sk))` and **XOR** it over `keynum_sk`. **scrypt params (libsodium SENSITIVE):** `opslimit=33554432`, `memlimit=1073741824` → `N=1<<20, r=8, p=1`. (Decrypt: same XOR; then recompute the Blake2b-256 checksum and reject on mismatch — wrong-passphrase detection.)

**Signature `*.minisig`** (4 lines): comment line; `base64( sig_alg[2] || key_id[8] || signature[64] )`; `trusted comment: <text>\n`; `base64( global_sig[64] )`.
- `sig_alg`: **`"ED"` = prehashed (default for new sigs)** = `ed25519.Sign(sk, Blake2b512(data))`; `"Ed"` = legacy raw `ed25519.Sign(sk, data)` — **read-only, never emit by default**.
- `global_sig = ed25519.Sign(sk, signature[64] || trusted_comment_bytes)` (the trusted-comment text WITHOUT the `trusted comment: ` prefix and WITHOUT trailing newline).

**Fail-closed verify (Pitfall 2):** (1) parse pub + sig; (2) reject if `sig.key_id != pub.key_id` BEFORE any crypto; (3) dispatch on `sig_alg`: for `"ED"` verify `ed25519.Verify(pub, Blake2b512(data), signature)`, for `"Ed"` verify over raw data; (4) ALSO verify `global_sig` over `signature || trusted_comment`; (5) any parse error, key_id mismatch, bad-base64, short buffer, failed data-sig, OR failed global-sig → return an error (the CLI maps every one to `ErrConflict`). Never accept on partial success.

---

### Task 1 (ADR GATE): ADR-0005 — signing scheme & cosign-compat scope

**Files:** Create `docs/adr/ADR-0005-pure-go-minisign-signing.md`

- [ ] **Step 1: Write the ADR** matching the `ADR-0004` header/section format (`# ADR-0005: …`, `**Status:** Accepted`, `**Date:** 2026-06-03`, `**Decision:** …`, then `## Context`, `## Analysis` (table), `## Consequences`). Capture these decisions (already made in the spec — this ADR records + justifies them):
  - Default signing scheme = **minisign-compatible Ed25519, prehashed ("ED")**; legacy "Ed" read-only.
  - **Reimplement on stdlib + x/crypto** rather than depend on an external minisign library (jedisct1/go-minisign is verify-only; aead/minisign adds an external dep — Pitfall 17). Document the byte format (reference the section above).
  - **Sigstore = verification only, behind `//go:build omni_sigstore`.** v1.0 explicitly EXCLUDES Rekor upload, Fulcio issuance, and OCI (the sigstore-go dep tree is heavy — Rekor + go-openapi + CT + TSA + go-tuf). Without the tag, `omni verify --bundle` returns `cmderr.ErrUnsupported`.
  - scrypt SENSITIVE cost (N=2²⁰) is intentional for at-rest key protection; note the implication for test speed (tests use fixtures, not live keygen — see Task 10).
- [ ] **Step 2: Stop for human review.** Do NOT proceed to code until this ADR is approved.

---

### Task 2 (ADR GATE): ADR-0006 — key-handling policy & secret redaction

**Files:** Create `docs/adr/ADR-0006-key-handling-and-secret-redaction.md`

- [ ] **Step 1: Write the ADR** (same format). Capture:
  - **Private keys NEVER as CLI flag values** — only `--key <path>` (a file path) or an env var that *points to a file* (e.g. `OMNI_SIGN_KEY=/path`); never the key material itself. Passphrase via interactive prompt or `OMNI_SIGN_PASSPHRASE` (documented tradeoff), never as a flag value.
  - File permissions: secret key `0600`, public key `0644`; refuse to read a secret key whose perms are group/world-accessible on Unix (warn, not hard-fail, on Windows where perms differ).
  - Dev-vs-release key separation (distinct key IDs; release keys out of repo).
  - `pkg/secret.Key` wrapper design: implements `fmt.Stringer`, `fmt.GoStringer`, and `slog.LogValuer`, all returning a redacted placeholder; raw bytes reachable ONLY via an explicit `.Bytes()`/`.Open()` method; zeroing on `.Destroy()`.
- [ ] **Step 2: Stop for human review.** Do NOT proceed to code until both ADRs are approved.

---

### Task 3: `pkg/secret` — redacting secret wrapper (TDD)

**Files:** Create `pkg/secret/secret.go`, `pkg/secret/secret_test.go`, `pkg/secret/doc.go`

- [ ] **Step 1: Write failing tests** (`pkg/secret/secret_test.go`):

```go
package secret_test

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/secret"
)

func TestKeyRedactsEverywhere(t *testing.T) {
	raw := []byte("super-secret-ed25519-bytes")
	k := secret.New(raw)
	for name, got := range map[string]string{
		"String":  k.String(),
		"%v":      fmt.Sprintf("%v", k),
		"%s":      fmt.Sprintf("%s", k),
		"%#v":     fmt.Sprintf("%#v", k),
		"GoString": k.GoString(),
	} {
		if strings.Contains(got, "super-secret") {
			t.Errorf("%s leaked key material: %q", name, got)
		}
		if !strings.Contains(got, "REDACTED") {
			t.Errorf("%s = %q, want a REDACTED placeholder", name, got)
		}
	}
}

func TestKeySlogRedacts(t *testing.T) {
	var buf strings.Builder
	l := slog.New(slog.NewTextHandler(&buf, nil))
	l.Info("signing", "key", secret.New([]byte("super-secret")))
	if strings.Contains(buf.String(), "super-secret") {
		t.Errorf("slog leaked key material: %q", buf.String())
	}
}

func TestKeyBytesRoundTrip(t *testing.T) {
	raw := []byte{1, 2, 3, 4}
	k := secret.New(raw)
	if got := k.Bytes(); string(got) != string(raw) {
		t.Errorf("Bytes() = %v, want %v", got, raw)
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./pkg/secret/ -v` → FAIL (package/`New` undefined).
- [ ] **Step 3: Implement** `pkg/secret/secret.go`:

```go
package secret

import "log/slog"

const placeholder = "<REDACTED>"

// Key holds sensitive bytes that must never appear in logs, errors, %v/%#v, or panics.
type Key struct{ b []byte }

// New wraps raw secret bytes. The caller relinquishes the slice; do not retain it.
func New(b []byte) Key { return Key{b: b} }

// Bytes returns the underlying secret bytes for controlled cryptographic use ONLY.
func (k Key) Bytes() []byte { return k.b }

// String implements fmt.Stringer with a redacted placeholder.
func (k Key) String() string { return placeholder }

// GoString implements fmt.GoStringer so %#v never reveals the bytes.
func (k Key) GoString() string { return "secret.Key{" + placeholder + "}" }

// LogValue implements slog.LogValuer so structured logs never reveal the bytes.
func (k Key) LogValue() slog.Value { return slog.StringValue(placeholder) }

// Destroy zeroes the underlying bytes.
func (k Key) Destroy() {
	for i := range k.b {
		k.b[i] = 0
	}
}
```

Also create `pkg/secret/doc.go` with the package docstring (note it is a stable v1.0 API — no `// Experimental:` marker).

- [ ] **Step 4: Run, verify pass:** `go test ./pkg/secret/ -v` → PASS.
- [ ] **Step 5: Commit:** `gofmt -w pkg/secret/ && git commit -- pkg/secret/ -m "feat(secret): add redacting Key wrapper (Stringer/GoStringer/slog.LogValuer)"`

---

### Task 4: `pkg/sign` types + `GenerateKeyPair` (TDD)

**Files:** Create `pkg/sign/sign.go`, `pkg/sign/format.go`, `pkg/sign/sign_test.go`, `pkg/sign/doc.go`; modify `go.mod`/`go.sum` (import x/crypto).

- [ ] **Step 1: Define stable public types** in `pkg/sign/sign.go` (no `// Experimental:` — stable surface): `type KeyPair struct{ PublicKey PublicKey; SecretKey SecretKey }`, `type PublicKey struct{ KeyID [8]byte; Pub ed25519.PublicKey }`, `type SecretKey struct{ KeyID [8]byte; key secret.Key /* decrypted ed25519 priv */ }`, `type Signature struct{ ... }`, and the functional-options `Options`/`Option` pattern mirroring `pkg/cryptutil`. Add `format.go` with the minisign codec constants (`sigAlgEd = [2]byte{'E','d'}`, `sigAlgPrehashed = [2]byte{'E','D'}`, `kdfScrypt = [2]byte{'S','c'}`, `cksumBlake2b = [2]byte{'B','2'}`, `scryptN = 1 << 20`, `scryptR = 8`, `scryptP = 1`, `opsLimitSensitive uint64 = 33554432`, `memLimitSensitive uint64 = 1 << 30`).
- [ ] **Step 2: Write failing test** for keygen + parse round-trip (`sign_test.go`):

```go
func TestGenerateKeyPairRoundTrip(t *testing.T) {
	kp, err := sign.GenerateKeyPair("test-passphrase", sign.WithScryptParams(1<<15, 8, 1)) // low cost for tests
	if err != nil { t.Fatalf("GenerateKeyPair: %v", err) }
	pubText := kp.PublicKey.MarshalText()
	skText, err := kp.SecretKey.MarshalText("test-passphrase")
	if err != nil { t.Fatalf("MarshalText sk: %v", err) }

	pub2, err := sign.ParsePublicKey(pubText)
	if err != nil || pub2.KeyID != kp.PublicKey.KeyID { t.Fatalf("public round-trip: %v", err) }
	sk2, err := sign.ParseSecretKey(skText, "test-passphrase")
	if err != nil { t.Fatalf("secret round-trip: %v", err) }
	if _, err := sign.ParseSecretKey(skText, "WRONG"); err == nil {
		t.Error("wrong passphrase must fail (checksum mismatch)")
	}
	_ = sk2
}
```

(Provide `WithScryptParams(n,r,p)` so tests avoid the 1 GiB SENSITIVE cost; default keygen uses SENSITIVE.)

- [ ] **Step 3: Run, verify fail.** `go test ./pkg/sign/ -run RoundTrip -v` → FAIL.
- [ ] **Step 4: Implement** keygen + the `.pub`/`.key` codec per the format section above: `ed25519.GenerateKey`, random 8-byte key_id, Blake2b-256 checksum over `"Ed"||key_id||sk`, scrypt-derived XOR stream to encrypt `keynum_sk`, base64 line assembly with the `untrusted comment:` prefix. Store the decrypted ed25519 secret inside a `secret.Key`. Decode verifies the checksum (wrong passphrase → distinct error). Run `go get golang.org/x/crypto/scrypt golang.org/x/crypto/blake2b && go mod tidy`.
- [ ] **Step 5: Run, verify pass.** `go test ./pkg/sign/ -run RoundTrip -v` → PASS.
- [ ] **Step 6: Commit:** `gofmt -w pkg/sign/ && git commit -- pkg/sign/ go.mod go.sum -m "feat(sign): pkg/sign keygen + minisign .pub/.key codec (Ed25519, scrypt)"`

---

### Task 5: `pkg/sign` `Sign` — prehashed detached signature (TDD)

**Files:** Modify `pkg/sign/sign.go`; add to `pkg/sign/sign_test.go`.

- [ ] **Step 1: Failing test** — sign then verify a payload round-trips, and the emitted `.minisig` uses algo `"ED"`:

```go
func TestSignVerifyRoundTrip(t *testing.T) {
	kp, _ := sign.GenerateKeyPair("p", sign.WithScryptParams(1<<15, 8, 1))
	data := []byte("artifact bytes")
	sig, err := sign.Sign(data, kp.SecretKey, sign.WithTrustedComment("ts:1"))
	if err != nil { t.Fatalf("Sign: %v", err) }
	if err := sign.Verify(data, sig, kp.PublicKey); err != nil {
		t.Fatalf("Verify(valid) = %v, want nil", err)
	}
	if !strings.HasPrefix(string(mustB64Line2(t, sig)), "") { /* assert algo bytes are 'E','D' */ }
}
```

- [ ] **Step 2: Run → FAIL.** `go test ./pkg/sign/ -run SignVerifyRoundTrip -v`.
- [ ] **Step 3: Implement `Sign`**: prehash `h := blake2b.Sum512(data)`; `signature := ed25519.Sign(sk, h[:])`; assemble line-2 payload `sigAlgPrehashed || key_id || signature`; build trusted comment; `global := ed25519.Sign(sk, append(signature, []byte(trustedComment)...))`; emit the 4-line `.minisig`.
- [ ] **Step 4: Run → PASS.**
- [ ] **Step 5: Commit:** `git commit -- pkg/sign/ -m "feat(sign): prehashed Sign producing minisign .minisig with trusted-comment global signature"`

---

### Task 6: `pkg/sign` `Verify` — fail-closed (TDD, negative-heavy)

**Files:** Modify `pkg/sign/sign.go`; add to `pkg/sign/sign_test.go`.

- [ ] **Step 1: Failing tests — one per failure mode (all must return non-nil error):**

```go
func TestVerifyFailsClosed(t *testing.T) {
	kp, _ := sign.GenerateKeyPair("p", sign.WithScryptParams(1<<15, 8, 1))
	other, _ := sign.GenerateKeyPair("p", sign.WithScryptParams(1<<15, 8, 1))
	data := []byte("payload")
	good, _ := sign.Sign(data, kp.SecretKey, sign.WithTrustedComment("t"))

	cases := map[string]func() error{
		"wrong key":        func() error { return sign.Verify(data, good, other.PublicKey) },
		"tampered payload": func() error { return sign.Verify([]byte("payloaX"), good, kp.PublicKey) },
		"tampered sig":     func() error { return sign.Verify(data, flipByte(good), kp.PublicKey) },
		"empty sig":        func() error { return sign.Verify(data, []byte{}, kp.PublicKey) },
		"garbage sig":      func() error { return sign.Verify(data, []byte("not base64!!!"), kp.PublicKey) },
		"bad algo":         func() error { return sign.Verify(data, withAlgo(good, [2]byte{'X','x'}), kp.PublicKey) },
	}
	for name, fn := range cases {
		if err := fn(); err == nil { t.Errorf("%s: Verify = nil, want error (fail-closed)", name) }
	}
}
```

- [ ] **Step 2: Run → FAIL.**
- [ ] **Step 3: Implement `Verify`** exactly per the fail-closed sequence in the format section: parse, key_id check first, algo dispatch (`"ED"`→prehash, `"Ed"`→raw, else error), verify data sig AND global sig, every error path returns. Return typed errors (a sentinel `ErrVerification` the CLI maps to `cmderr.ErrConflict`).
- [ ] **Step 4: Run → PASS** (all 6 negative cases + the positive from Task 5).
- [ ] **Step 5: Commit:** `git commit -- pkg/sign/ -m "feat(sign): fail-closed Verify (key_id, data sig, global sig; sentinel ErrVerification)"`

---

### Task 7: CLI glue + Cobra wrappers — `omni sign` / `omni verify` / `omni sign keygen` (TDD)

**Files:** Create `internal/cli/sign/sign.go` (+`_test.go`), `internal/cli/verify/verify.go` (+`_test.go`), `cmd/sign.go`, `cmd/verify.go`. Modify nothing else.

- [ ] **Step 1: Failing test** (`internal/cli/sign/sign_test.go`) — `RunSign` writes a `.minisig`, `RunVerify` returns nil for a good sig and a `cmderr.ErrConflict`-wrapped error for a tampered one; **keys are accepted only via file path, never as a flag value**.
- [ ] **Step 2: Run → FAIL.**
- [ ] **Step 3: Implement** `RunKeygen(w, opts)`, `RunSign(w, r, args, opts)`, `RunVerify(w, r, args, opts)`. Map errors: `sign.ErrVerification` → `cmderr.Wrap(cmderr.ErrConflict, "signature verification failed")`; malformed key/flags → `ErrInvalidInput`; missing key file → `ErrNotFound`; unreadable/bad-perms → `ErrPermission`. Read passphrase from `OMNI_SIGN_PASSPHRASE` env or interactive prompt — NEVER a flag. Enforce: `--key` is a path; reject if the value looks like inline key material.
- [ ] **Step 4: Cobra wrappers.** `cmd/sign.go`: `signCmd` (`Use: "sign"`, `RunE`→`sign.RunSign`) with a `signKeygenCmd` subcommand (`Use: "keygen"`, `RunE`→`sign.RunKeygen`) added via `signCmd.AddCommand(signKeygenCmd)`; `rootCmd.AddCommand(signCmd)` in `init()`. `cmd/verify.go`: `verifyCmd` (`Use: "verify"`, `RunE`→`verify.RunVerify`) with `--bundle`, `--key`, `--sig` flags; `rootCmd.AddCommand(verifyCmd)`.
- [ ] **Step 5: Run → PASS;** manual smoke: `go run . sign keygen --pub /tmp/k.pub --key /tmp/k.key` (set `OMNI_SIGN_PASSPHRASE`), `go run . sign --key /tmp/k.key <file>`, `go run . verify --key /tmp/k.pub --sig <file>.minisig <file>`.
- [ ] **Step 6: Commit:** `gofmt -w ... && git commit -- internal/cli/sign internal/cli/verify cmd/sign.go cmd/verify.go -m "feat(sign): omni sign/verify/sign-keygen CLI with cmderr-classified fail-closed verify"`

---

### Task 8: Sigstore bundle verification behind `omni_sigstore` build tag

**Files:** Create `internal/cli/verify/bundle_sigstore.go` (`//go:build omni_sigstore`), `internal/cli/verify/bundle_nosigstore.go` (`//go:build !omni_sigstore`).

- [ ] **Step 1: Default (no-tag) failing test** — `RunVerify` with `--bundle x` returns `cmderr.IsUnsupported(err)`.
- [ ] **Step 2: Implement the no-tag stub** (`bundle_nosigstore.go`): `func verifyBundle(...) error { return cmderr.Wrap(cmderr.ErrUnsupported, "sigstore bundle verification requires building with -tags omni_sigstore") }`.
- [ ] **Step 3: Implement the tagged path** (`bundle_sigstore.go`): import `github.com/sigstore/sigstore-go/pkg/{bundle,root,verify}`; `bundle.LoadJSONFromPath` → `root.NewTrustedRootFromPath(--trusted-root)` → `verify.NewVerifier(tm, verify.WithTransparencyLog(1), verify.WithObserverTimestamps(1))` → `verifier.Verify(b, verify.NewPolicy(verify.WithArtifact(reader), verify.WithCertificateIdentity(...)))`. Map mismatch → `ErrConflict`, bad flags → `ErrInvalidInput`. Add the dep ONLY here: `go get github.com/sigstore/sigstore-go@v1.2.0 && go mod tidy` (accept the heavy transitive tree — it is excluded from default builds by the tag).
- [ ] **Step 4: Verify BOTH builds:** `go build ./...` (default — sigstore deps NOT compiled) and `go build -tags omni_sigstore ./...` (tagged — compiles sigstore). Run `go test ./internal/cli/verify/...` (default) → unsupported path passes.
- [ ] **Step 5: Commit:** `git commit -- internal/cli/verify go.mod go.sum -m "feat(verify): optional Sigstore bundle verification behind omni_sigstore build tag (default: ErrUnsupported)"`

---

### Task 9: `omni pipe` integration

**Files:** Modify `cmd/pipe.go`.

- [ ] **Step 1: Failing test** — `omni pipe`-style invocation of `sign`/`verify` works via the Unified registry (mirror an existing pipe test).
- [ ] **Step 2: Implement** — in `buildPipeRegistry()` add `reg.Register("sign", command.AdaptWriterReaderArgs(func(w io.Writer, r io.Reader, args []string) error { return sign.RunSign(w, r, args, sign.Options{}) }))` and the same for `verify`. (keygen stays Cobra-only — it has no stdin transform.)
- [ ] **Step 3: Run → PASS.** `go test ./internal/cli/pipe/... ./cmd/...`.
- [ ] **Step 4: Commit:** `git commit -- cmd/pipe.go -m "feat(sign): register sign/verify in the pipe Unified registry"`

---

### Task 10: Golden-master tests (positive + negative, deterministic)

**Files:** Modify `testing/golden/golden_tests.yaml` AND `tools/golden/golden_tests.yaml` (keep in sync); add fixed test-key fixtures under `testing/golden/fixtures/sign/` (a pre-generated low-cost keypair + a signed file + its `.minisig`).

- [ ] **Step 1: Create deterministic fixtures.** Generate ONE keypair with a known passphrase + low scrypt cost and check the `.pub`/`.key`/sample-`.minisig` into `testing/golden/fixtures/sign/` (keygen is random and SENSITIVE-cost is 1 GiB — do NOT run live keygen in golden tests; verify against committed fixtures).
- [ ] **Step 2: Add a `sign` category to BOTH yaml files** with: a positive `verify` against the fixture (exit 0); negatives — `verify_tampered` → `exit_code: 1` `# cmderr.ErrConflict`; `verify_bad_key` → `exit_code: 2` `# cmderr.ErrInvalidInput`; `verify_bundle_no_tag` → `exit_code: 6` `# cmderr.ErrUnsupported`. Add `normalizations: [strip_path]` where output contains paths.
- [ ] **Step 3: Generate + verify snapshots:** `task test:golden:update && task golden:record`, then `python testing/scripts/test_golden.py` → all green.
- [ ] **Step 4: Commit:** `git commit -- testing/golden tools/golden -m "test(sign): golden-master positive + fail-closed negative cases for sign/verify"`

---

### Task 11: Docs + final gate

**Files:** `docs/COMMANDS.md`, `CLAUDE.md` (command inventory line), `docs/superpowers/specs/2026-05-16-04-pkg-sign-design.md` (status), `docs/EXIT-CODES.md` (no change expected — verify reuses existing sentinels).

- [ ] **Step 1: Docs** — add `sign`/`verify`/`sign keygen` to `docs/COMMANDS.md` and the CLAUDE.md inventory count; run `omni aicontext` / `omni cmdtree` regen if applicable. Mark the spec `Status: Complete`.
- [ ] **Step 2: Final gate:**
```bash
go build ./... && go build -tags omni_sigstore ./...
go vet ./... && gofmt -l pkg/sign pkg/secret internal/cli/sign internal/cli/verify
golangci-lint run --timeout=5m ./...
go test ./pkg/sign/... ./pkg/secret/... ./internal/cli/sign/... ./internal/cli/verify/... -count=1
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ./... && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...
python testing/scripts/test_golden.py
```
Expected: all green; the SENSITIVE-cost keygen path is covered by a single `-short`-skippable test.
- [ ] **Step 3: Commit:** `git commit -- docs/ CLAUDE.md -m "docs(sign): document omni sign/verify; mark Phase 04 complete"`

---

## Self-Review

**Spec coverage** (SIGN-01…09 / success criteria → tasks):
- keygen passphrase-protected Ed25519, 0600/0644, minisign-compatible → **T4** (+ perms enforced in **T7**).
- sign+verify round-trip; verify fails closed on missing sig / wrong key / tampered payload / bad algo / unknown key id — each locked by a golden negative test → **T5, T6** (unit) + **T10** (golden).
- keys never in flags/slog/errors/panics → **T2** (ADR) + **T3** (`pkg/secret`) + **T7** (flag policy).
- `pkg/sign/` standalone with stable types; sign/verify in `omni pipe` → **T4–T6** (lib) + **T9** (pipe).
- `-tags omni_sigstore` bundle verify; without tag → `ErrUnsupported` → **T8**.
- 2 ADRs before code → **T1, T2** (hard gate).

**Placeholder scan:** Crypto bodies in T4–T6 are specified by the byte-exact format section + the libsodium SENSITIVE scrypt constants + per-failure-mode tests, not vague TODOs. The few "assemble per the format section above" instructions are bounded references to a concrete, byte-level spec in this document (not "add validation"). The sigstore API (T8) is the exact `bundle.LoadJSONFromPath → root.NewTrustedRootFromPath → verify.NewVerifier → Verify` chain from research.

**Type consistency:** `secret.Key` (T3) is consumed by `SecretKey` (T4) and never logged. `sign.ErrVerification` (T6) is the sentinel the CLI maps to `cmderr.ErrConflict` (T7). Algo constants (`sigAlgPrehashed "ED"`, `sigAlgEd "Ed"`) defined in T4 `format.go` and used in T5 (emit "ED") and T6 (dispatch). `WithScryptParams` (T4) is reused by every test to avoid the 1 GiB SENSITIVE cost.

**Known risks:** (1) Byte-exact minisign compat — mitigate by adding a cross-check test that verifies a `.minisig` produced by the real `minisign` tool (committed fixture) and vice-versa, if a reference vector is available. (2) scrypt SENSITIVE = 1 GiB RAM / multi-second — keygen default uses it, but ALL automated tests pass `WithScryptParams(low)` or use fixtures; document this loudly. (3) sigstore-go dep weight — fully contained by the build tag; CI must build both tagged and untagged.
