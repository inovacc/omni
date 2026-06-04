# Phase 07 — `pkg/attest/` SLSA Provenance Attestation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
> **HARD GATE — SATISFIED (2026-06-03):** ADR-0009 is written and accepted (pins SLSA Build **L2**, an ADR-pinned `builder.id` allowlist enforced in code, no numeric `slsaLevel` field, pure-Go format with zero new deps). Code tasks (2+) proceed.

**Goal:** Ship `omni attest` / `omni attest verify` — a pure-Go generator and fail-closed verifier for DSSE-wrapped, in-toto-formatted **SLSA v1.0 provenance** attestations, signed with the Phase 04 `pkg/sign` Ed25519 primitive, claiming only the honest SLSA Build level (ADR-pinned, L2) that omni's GitHub-Actions release process actually achieves.

**Architecture:** Pure-Go reimplementation of the in-toto Statement v1 envelope, SLSA Provenance v1 predicate, and DSSE Pre-Authentication Encoding (PAE) on `encoding/json` + `encoding/base64` + `crypto/sha256` — NOT an external in-toto/sigstore SDK (those drag in the heavy Rekor/go-tuf/protobuf tree and bump go.mod via MVS — Phase 04 finding). The signature primitive is reused from `pkg/sign` (the minisign-compatible Ed25519 `Sign`/`Verify`), so the same key infrastructure (`omni sign keygen`) produces release keys. Layering: pure lib `pkg/attest/` (no cobra, no io.Writer-for-output) → I/O glue `internal/cli/attest/` → thin Cobra wrapper `cmd/attest.go`. The SLSA-schema-validation gate runs in CI (a `task` target), not in the binary.

**Tech Stack:** Go stdlib (`encoding/json`, `encoding/base64`, `crypto/sha256`, `strconv`, `time`, `os`); `github.com/inovacc/omni/pkg/sign` (Ed25519 Sign/Verify); `github.com/inovacc/omni/pkg/secret` (already present); `log/slog`; Cobra; the Python YAML golden harness. No new third-party deps in the default build path.

**Repo conventions (from research, cite when implementing):**
- Commands self-wire: `cmd/<name>.go` declares `var xCmd = &cobra.Command{...}` and calls `rootCmd.AddCommand(xCmd)` in `init()`; `RunE` reads flags → Options → calls `internal/cli/<name>.RunX(cmd.OutOrStdout(), ...)`. No central registration list. (Mirror `cmd/sign.go` exactly: a parent `attestCmd` + an `attestVerifyCmd` subcommand added via `attestCmd.AddCommand(...)`.)
- `cmderr` (`internal/cli/cmderr/cmderr.go`): verification failure (bad sig, key_id mismatch, digest mismatch) → `cmderr.Wrap(cmderr.ErrConflict, …)` (exit 1); malformed envelope/predicate/flags → `ErrInvalidInput` (exit 2); missing input file → `ErrNotFound` (1); unreadable → `ErrPermission` (3); SLSA-level overclaim refusal → `ErrInvalidInput` (2). Sentinels are `errors.New`; never compare with `==` (use `errors.Is`). `Wrap(sentinel, msg)` = `fmt.Errorf("%s: %w", msg, sentinel)`. `Is<Class>()` predicates exist.
- `pkg/sign` exports (confirmed present from Phase 04): `func Sign(data []byte, sk SecretKey, opts ...Option) ([]byte, error)` (returns a 4-line `.minisig` text blob), `func Verify(data, sig []byte, pub PublicKey) error`, `func ParsePublicKey(text []byte) (PublicKey, error)`, `func ParseSecretKey(text []byte, passphrase string, opts ...Option) (SecretKey, error)`, sentinels `sign.ErrVerification` / `sign.ErrMalformed`, `PublicKey{KeyID [8]byte; Pub ed25519.PublicKey}`, `WithScryptParams(n,r,p)`. **`Sign` returns the minisign text blob, which is what we place (base64-of-the-blob) into the DSSE `signatures[].sig` field** — keeping one signature format across the whole supply-chain toolchain.
- `pkg/secret.Key` exists; not needed here (no new secret material — passphrase handling is delegated to `pkg/sign`/the existing sign CLI conventions).
- Golden harness is Python+YAML, TWO files kept in sync: `testing/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml`; negative tests set `exit_code:` + `normalizations: ["strip_path"]`; committed fixtures live under `testing/golden/fixtures/<category>/` referenced by `fixtures_dir:`; regenerate with `task test:golden:update` then `task golden:record`; verify with `python testing/scripts/test_golden.py`. A `sign` fixture category already exists (`testing/golden/fixtures/sign/`: `test.key`, `test.pub`, `data.txt`, `data.txt.minisig`, `gen_fixtures.go`) — Task 8 reuses that keypair.
- Pipe: register in `cmd/pipe.go buildPipeRegistry()` via `command.AdaptWriterReaderArgs(...)`. `sign`/`verify` are already registered there (lines ~206/212) — `attest verify` follows the same shape but `attest` (generate) reads a file argument, not stdin, so it stays Cobra-only.
- ADRs live in `docs/adr/` as `ADR-NNNN-kebab-title.md`; **0001–0006 are used → the next number is `0007`**; header format per `docs/adr/ADR-0004-internalize-cobra-cli.md` (`# ADR-NNNN: Title`, `**Status:** Accepted`, `**Date:** 2026-06-03`, `**Decision:** …`, `## Context`, `## Analysis` (table), `## Consequences`).
- INVARIANTS: pure-Go, NO `os/exec`, no CGO; cross-platform via `//go:build` tags (never runtime `os ==`); `io.Writer`/`io.Reader`; deferred `Close`.

---

## Authoritative wire formats (implement byte-exactly)

These three formats are fixed external specs. Implement to the byte; unit tests pin them to published reference vectors.

### 1. DSSE Pre-Authentication Encoding (PAE) — v1.0.2

```
PAE(type, body) = "DSSEv1" + SP + LEN(type) + SP + type + SP + LEN(body) + SP + body
SP        = single ASCII space (0x20)
"DSSEv1"  = ASCII [0x44,0x53,0x53,0x45,0x76,0x31]
LEN(s)    = ASCII decimal byte-length of s, NO leading zeros (strconv.Itoa(len(s)))
```

**Reference vector (MUST pass in a unit test):** `type = "http://example.com/HelloWorld"` (len 29), `body = "hello world"` (len 11) →
`DSSEv1 29 http://example.com/HelloWorld 11 hello world`.
The signed bytes for an attestation are `PAE("application/vnd.in-toto+json", <statement-json-bytes>)`.

### 2. DSSE JSON Envelope — v1.0.2

```json
{
  "payload": "<base64-std(SERIALIZED_BODY)>",
  "payloadType": "application/vnd.in-toto+json",
  "signatures": [ { "keyid": "<hex key_id>", "sig": "<base64-std(SIGNATURE)>" } ]
}
```
- `payload` = `base64.StdEncoding.EncodeToString(statementJSON)`. `SERIALIZED_BODY` (the statement JSON) is signed via PAE, NOT the base64 string.
- `payloadType` is **always** `application/vnd.in-toto+json` (lowercase, fixed).
- `SIGNATURE` = the `pkg/sign.Sign` output (the minisign `.minisig` text blob), then base64-std encoded into `sig`.
- `keyid` = lowercase-hex of the 8-byte `PublicKey.KeyID`.
- **CRITICAL (DSSE rule):** the verifier MUST verify the SIGNATURE against `PAE(payloadType, SERIALIZED_BODY)` and then hand the SAME `SERIALIZED_BODY` to the application layer. It MUST NOT re-parse the envelope to re-extract the payload after verification.

### 3. in-toto Statement v1 + SLSA Provenance v1 predicate

Statement (`payload`, before base64):
```json
{
  "_type": "https://in-toto.io/Statement/v1",
  "subject": [ { "name": "<artifact-name>", "digest": { "sha256": "<lowercase-hex>" } } ],
  "predicateType": "https://slsa.dev/provenance/v1",
  "predicate": { ... SLSA Provenance v1 ... }
}
```
- `_type` is fixed `https://in-toto.io/Statement/v1`. `predicateType` is fixed `https://slsa.dev/provenance/v1`.
- Each `subject[]` MUST have a `digest`. We compute `sha256` over the artifact bytes; hex is lowercase.

SLSA Provenance v1 predicate:
```json
{
  "buildDefinition": {
    "buildType": "<TypeURI>",
    "externalParameters": { ... },
    "internalParameters": { ... },
    "resolvedDependencies": [ { "uri": "...", "digest": { "...": "..." } } ]
  },
  "runDetails": {
    "builder": { "id": "<TypeURI>", "version": { "<name>": "<version>" }, "builderDependencies": [ ... ] },
    "metadata": { "invocationId": "...", "startedOn": "<RFC3339>", "finishedOn": "<RFC3339>" },
    "byproducts": [ ... ]
  }
}
```
- REQUIRED for SLSA Build L1: `buildDefinition.buildType`, `buildDefinition.externalParameters`, `runDetails.builder.id`.
- `builder.id` is the **sole determiner of SLSA Build level** (per SLSA verifying-artifacts step 1). Therefore the honesty rule = "the `builder.id` we emit MUST be the exact ADR-pinned URI and nothing else". The attestation itself carries NO numeric level field; the level is implied by `builder.id`. This is why the ADR gate matters: it fixes the one URI string the generator is allowed to emit.

---

### Task 1 (ADR GATE): ADR-0009 — honest SLSA level & builder.id policy

**Files:** Create `docs/adr/ADR-0009-honest-slsa-level-and-builder-id.md`

- [ ] **Step 1: Write the ADR** matching the `ADR-0004` header/section format (`# ADR-0009: …`, `**Status:** Accepted`, `**Date:** 2026-06-03`, `**Decision:** …`, then `## Context`, `## Analysis` (table), `## Consequences`). Record + justify these decisions (the spec already made them — the ADR pins them):
  - **Pinned honest level = SLSA Build L2.** Rationale: omni releases run on GitHub Actions with OIDC-identified runners producing signed provenance — that is L2 (build platform generates+signs provenance, hosted). It is NOT L3 because omni does not run on an isolated/hermetic build service with non-falsifiable provenance; claiming L3 would be Pitfall 5 (SLSA overclaim).
  - **Pinned `builder.id` constant** = the exact URI the generator emits: `https://github.com/inovacc/omni/.github/workflows/release.yml@refs/heads/main` for the dogfooded release path, and a generic fallback `https://github.com/inovacc/omni/attest/local@v1` for non-CI/local generation (which represents a LOWER trust tier — local, unverified, effectively L1). Document that local-generated attestations MUST use the local builder.id, never the release one.
  - **No numeric level field is emitted.** Per SLSA v1.0 the level is implied by `builder.id`; the predicate carries no `slsaLevel`/`completeness` field. The honesty contract is therefore "emit only an ADR-listed builder.id".
  - **`buildType` constant** = `https://slsa-framework.github.io/github-actions-buildtypes/workflow/v1` when `--from-env` detects GitHub Actions; otherwise `https://github.com/inovacc/omni/attest/local-buildtype/v1`.
  - **CI schema gate** = before every release, CI validates the emitted predicate against the official SLSA Provenance v1 JSON schema; the build fails on any deviation (Task 9 wires the `task` target).
  - **Enforcement in code:** the generator refuses (`cmderr.ErrInvalidInput`) any `--builder-id` value not in the ADR allowlist set defined in `pkg/attest`. There is no flag to claim L3.
- [x] **Step 2: DONE (2026-06-03).** ADR-0009 written and accepted at `docs/adr/ADR-0009-honest-slsa-level-and-builder-id.md`; decisions above are pinned. Proceed to Task 2.

---

### Task 2: `pkg/attest` types + PAE (TDD)

**Files:** Create `pkg/attest/attest.go`, `pkg/attest/pae.go`, `pkg/attest/pae_test.go`, `pkg/attest/doc.go`.

- [ ] **Step 1: Define stable public types** in `pkg/attest/attest.go` (this is a v1.0 supply-chain primitive — STABLE surface, no `// Experimental:` marker in `doc.go`):

```go
package attest

// Fixed wire constants (see the in-toto/DSSE/SLSA specs).
const (
	StatementType   = "https://in-toto.io/Statement/v1"
	PredicateTypeSLSAProvenance = "https://slsa.dev/provenance/v1"
	PayloadTypeInToto = "application/vnd.in-toto+json"
)

// ResourceDescriptor is the in-toto subject/dependency descriptor (digest-bearing subset).
type ResourceDescriptor struct {
	Name   string            `json:"name,omitempty"`
	URI    string            `json:"uri,omitempty"`
	Digest map[string]string `json:"digest,omitempty"`
}

// Statement is the in-toto Statement v1 envelope payload.
type Statement struct {
	Type          string               `json:"_type"`
	Subject       []ResourceDescriptor `json:"subject"`
	PredicateType string               `json:"predicateType"`
	Predicate     any                  `json:"predicate"`
}

// Provenance is the SLSA Provenance v1 predicate.
type Provenance struct {
	BuildDefinition BuildDefinition `json:"buildDefinition"`
	RunDetails      RunDetails      `json:"runDetails"`
}

// BuildDefinition describes the build inputs.
type BuildDefinition struct {
	BuildType            string               `json:"buildType"`
	ExternalParameters   map[string]any       `json:"externalParameters"`
	InternalParameters   map[string]any       `json:"internalParameters,omitempty"`
	ResolvedDependencies []ResourceDescriptor `json:"resolvedDependencies,omitempty"`
}

// RunDetails describes a single execution of the build.
type RunDetails struct {
	Builder    Builder        `json:"builder"`
	Metadata   *BuildMetadata `json:"metadata,omitempty"`
	Byproducts []ResourceDescriptor `json:"byproducts,omitempty"`
}

// Builder identifies the trusted build platform. Builder.ID is the sole determiner of SLSA level.
type Builder struct {
	ID                  string               `json:"id"`
	Version             map[string]string    `json:"version,omitempty"`
	BuilderDependencies []ResourceDescriptor `json:"builderDependencies,omitempty"`
}

// BuildMetadata captures invocation timing.
type BuildMetadata struct {
	InvocationID string `json:"invocationId,omitempty"`
	StartedOn    string `json:"startedOn,omitempty"`
	FinishedOn   string `json:"finishedOn,omitempty"`
}

// Envelope is the DSSE JSON envelope.
type Envelope struct {
	Payload     string             `json:"payload"`
	PayloadType string             `json:"payloadType"`
	Signatures  []EnvelopeSignature `json:"signatures"`
}

// EnvelopeSignature is one signature over PAE(payloadType, payload-bytes).
type EnvelopeSignature struct {
	KeyID string `json:"keyid,omitempty"`
	Sig   string `json:"sig"`
}
```

- [ ] **Step 2: Write failing PAE test** (`pkg/attest/pae_test.go`) pinning the published reference vector:

```go
package attest

import (
	"bytes"
	"testing"
)

func TestPAEReferenceVector(t *testing.T) {
	// From the DSSE protocol spec test vectors.
	got := PAE("http://example.com/HelloWorld", []byte("hello world"))
	want := []byte("DSSEv1 29 http://example.com/HelloWorld 11 hello world")
	if !bytes.Equal(got, want) {
		t.Fatalf("PAE mismatch:\n got=%q\nwant=%q", got, want)
	}
}

func TestPAEEmptyBody(t *testing.T) {
	got := PAE("t", []byte{})
	want := []byte("DSSEv1 1 t 0 ")
	if !bytes.Equal(got, want) {
		t.Fatalf("PAE empty body:\n got=%q\nwant=%q", got, want)
	}
}
```

- [ ] **Step 3: Run, verify fail:** `go test ./pkg/attest/ -run TestPAE -v` → FAIL (`PAE` undefined).
- [ ] **Step 4: Implement** `pkg/attest/pae.go`:

```go
package attest

import "strconv"

// PAE computes the DSSE Pre-Authentication Encoding (v1.0.2):
//
//	"DSSEv1" SP LEN(type) SP type SP LEN(body) SP body
//
// where SP is a single ASCII space and LEN is the base-10 byte length with no
// leading zeros. The result is the exact byte sequence to sign/verify.
func PAE(payloadType string, body []byte) []byte {
	out := make([]byte, 0, 6+len(payloadType)+len(body)+24)
	out = append(out, "DSSEv1"...)
	out = append(out, ' ')
	out = append(out, strconv.Itoa(len(payloadType))...)
	out = append(out, ' ')
	out = append(out, payloadType...)
	out = append(out, ' ')
	out = append(out, strconv.Itoa(len(body))...)
	out = append(out, ' ')
	out = append(out, body...)
	return out
}
```

Also create `pkg/attest/doc.go` (package docstring; STABLE v1.0 surface — no `// Experimental:` marker).

- [ ] **Step 5: Run, verify pass:** `go test ./pkg/attest/ -run TestPAE -v` → PASS.
- [ ] **Step 6: Commit:** `gofmt -w pkg/attest/ && git commit -- pkg/attest/ -m "feat(attest): pkg/attest in-toto/SLSA/DSSE types + PAE (reference-vector pinned)"`

---

### Task 3: `pkg/attest` Statement + Provenance builders (TDD)

**Files:** Modify `pkg/attest/attest.go`; create `pkg/attest/build.go`, `pkg/attest/build_test.go`.

- [ ] **Step 1: Define the builder API + allowlist** to put in `pkg/attest/build.go`. Subject digests are computed with sha256; the builder-id allowlist enforces ADR-0009.

- [ ] **Step 2: Write failing tests** (`pkg/attest/build_test.go`):

```go
package attest

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestSubjectFromBytes(t *testing.T) {
	data := []byte("artifact bytes")
	rd := SubjectFromBytes("app.tar.gz", data)
	sum := sha256.Sum256(data)
	if rd.Name != "app.tar.gz" {
		t.Fatalf("name = %q", rd.Name)
	}
	if rd.Digest["sha256"] != hex.EncodeToString(sum[:]) {
		t.Fatalf("digest = %q", rd.Digest["sha256"])
	}
}

func TestNewStatementSLSA(t *testing.T) {
	prov := Provenance{
		BuildDefinition: BuildDefinition{
			BuildType:          BuildTypeLocal,
			ExternalParameters: map[string]any{"artifact": "app.tar.gz"},
		},
		RunDetails: RunDetails{Builder: Builder{ID: BuilderIDLocal}},
	}
	st := NewStatement([]ResourceDescriptor{SubjectFromBytes("app.tar.gz", []byte("x"))}, prov)
	if st.Type != StatementType || st.PredicateType != PredicateTypeSLSAProvenance {
		t.Fatalf("fixed type fields wrong: %+v", st)
	}
}

func TestValidateBuilderIDRejectsOverclaim(t *testing.T) {
	if err := ValidateBuilderID("https://slsa.dev/some-l3-platform"); err == nil {
		t.Fatal("ValidateBuilderID must reject a builder.id not in the ADR-0009 allowlist")
	}
	if err := ValidateBuilderID(BuilderIDLocal); err != nil {
		t.Fatalf("ValidateBuilderID(local) = %v, want nil", err)
	}
	if err := ValidateBuilderID(BuilderIDRelease); err != nil {
		t.Fatalf("ValidateBuilderID(release) = %v, want nil", err)
	}
}
```

- [ ] **Step 3: Run, verify fail:** `go test ./pkg/attest/ -run "Subject|NewStatement|ValidateBuilder" -v` → FAIL.
- [ ] **Step 4: Implement** `pkg/attest/build.go` (constants pinned to ADR-0009; `ErrOverclaim` is a typed sentinel the CLI maps to `cmderr.ErrInvalidInput`):

```go
package attest

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// Builder-id / build-type constants pinned by ADR-0009. The generator may emit
// ONLY these builder IDs; anything else is an SLSA overclaim and is refused.
const (
	BuilderIDRelease = "https://github.com/inovacc/omni/.github/workflows/release.yml@refs/heads/main"
	BuilderIDLocal   = "https://github.com/inovacc/omni/attest/local@v1"
	BuildTypeGHA     = "https://slsa-framework.github.io/github-actions-buildtypes/workflow/v1"
	BuildTypeLocal   = "https://github.com/inovacc/omni/attest/local-buildtype/v1"
)

// ErrOverclaim is returned when a requested builder.id is not in the ADR-0009
// allowlist (i.e. would claim a higher SLSA level than omni honestly achieves).
// The CLI maps it to cmderr.ErrInvalidInput.
var ErrOverclaim = errors.New("attest: builder.id not permitted by ADR-0009 (SLSA overclaim)")

// allowedBuilderIDs is the ADR-0009 allowlist.
var allowedBuilderIDs = map[string]bool{
	BuilderIDRelease: true,
	BuilderIDLocal:   true,
}

// ValidateBuilderID returns ErrOverclaim unless id is an ADR-0009-pinned value.
func ValidateBuilderID(id string) error {
	if !allowedBuilderIDs[id] {
		return fmt.Errorf("%w: %q", ErrOverclaim, id)
	}
	return nil
}

// SubjectFromBytes builds an in-toto subject descriptor with a sha256 digest.
func SubjectFromBytes(name string, data []byte) ResourceDescriptor {
	sum := sha256.Sum256(data)
	return ResourceDescriptor{Name: name, Digest: map[string]string{"sha256": hex.EncodeToString(sum[:])}}
}

// NewStatement wraps subjects + a SLSA provenance predicate into a Statement.
func NewStatement(subject []ResourceDescriptor, prov Provenance) Statement {
	return Statement{
		Type:          StatementType,
		Subject:       subject,
		PredicateType: PredicateTypeSLSAProvenance,
		Predicate:     prov,
	}
}
```

- [ ] **Step 5: Run, verify pass:** `go test ./pkg/attest/ -run "Subject|NewStatement|ValidateBuilder" -v` → PASS.
- [ ] **Step 6: Commit:** `gofmt -w pkg/attest/ && git commit -- pkg/attest/ -m "feat(attest): SLSA statement/provenance builders + ADR-0009 builder.id allowlist (no overclaim)"`

---

### Task 4: `pkg/attest` envelope Sign (TDD) — reuse `pkg/sign`

**Files:** Create `pkg/attest/envelope.go`, `pkg/attest/envelope_test.go`; modify `go.mod`/`go.sum` only if `pkg/sign` import pulls anything new (it should not — sign is in-repo).

- [ ] **Step 1: Define the signer signature.** To keep `pkg/attest` decoupled from a concrete key type AND reuse `pkg/sign`, define a tiny function-typed signer so tests can inject a stub and the CLI can wire `pkg/sign`:

```go
// Signer signs the DSSE PAE bytes and returns (signatureBytes, keyidHex).
type Signer func(pae []byte) (sig []byte, keyidHex string, err error)
```

- [ ] **Step 2: Write failing test** (`pkg/attest/envelope_test.go`) — Sign produces a well-formed envelope whose payload base64-decodes to the statement JSON and whose signer was called with the exact PAE bytes:

```go
package attest

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestSignEnvelope(t *testing.T) {
	st := NewStatement(
		[]ResourceDescriptor{SubjectFromBytes("app", []byte("x"))},
		Provenance{
			BuildDefinition: BuildDefinition{BuildType: BuildTypeLocal, ExternalParameters: map[string]any{}},
			RunDetails:      RunDetails{Builder: Builder{ID: BuilderIDLocal}},
		},
	)
	var gotPAE []byte
	signer := func(pae []byte) ([]byte, string, error) {
		gotPAE = append([]byte(nil), pae...)
		return []byte("SIGBYTES"), "deadbeef", nil
	}
	env, err := SignStatement(st, signer)
	if err != nil {
		t.Fatalf("SignStatement: %v", err)
	}
	if env.PayloadType != PayloadTypeInToto {
		t.Fatalf("payloadType = %q", env.PayloadType)
	}
	body, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil {
		t.Fatalf("payload not std-base64: %v", err)
	}
	var rt Statement
	if err := json.Unmarshal(body, &rt); err != nil {
		t.Fatalf("payload not statement JSON: %v", err)
	}
	wantPAE := PAE(PayloadTypeInToto, body)
	if string(gotPAE) != string(wantPAE) {
		t.Fatalf("signer received wrong PAE:\n got=%q\nwant=%q", gotPAE, wantPAE)
	}
	if len(env.Signatures) != 1 || env.Signatures[0].KeyID != "deadbeef" {
		t.Fatalf("signatures = %+v", env.Signatures)
	}
	if env.Signatures[0].Sig != base64.StdEncoding.EncodeToString([]byte("SIGBYTES")) {
		t.Fatalf("sig not std-base64 of signer output: %q", env.Signatures[0].Sig)
	}
}
```

- [ ] **Step 3: Run, verify fail:** `go test ./pkg/attest/ -run TestSignEnvelope -v` → FAIL (`SignStatement`/`Signer` undefined).
- [ ] **Step 4: Implement** `pkg/attest/envelope.go`. Marshal the statement deterministically (default `json.Marshal` field order = struct order, which is stable), PAE it, sign, base64-std both payload and signature:

```go
package attest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// MarshalStatement serializes a Statement to its canonical JSON SERIALIZED_BODY.
func MarshalStatement(st Statement) ([]byte, error) {
	b, err := json.Marshal(st)
	if err != nil {
		return nil, fmt.Errorf("attest: marshal statement: %w", err)
	}
	return b, nil
}

// SignStatement marshals st, computes the DSSE PAE over the in-toto payload
// type, invokes signer, and returns a DSSE JSON envelope. The SAME bytes that
// are signed are base64-encoded into Payload (DSSE binding requirement).
func SignStatement(st Statement, signer Signer) (Envelope, error) {
	body, err := MarshalStatement(st)
	if err != nil {
		return Envelope{}, err
	}
	pae := PAE(PayloadTypeInToto, body)
	sig, keyid, err := signer(pae)
	if err != nil {
		return Envelope{}, fmt.Errorf("attest: sign: %w", err)
	}
	return Envelope{
		Payload:     base64.StdEncoding.EncodeToString(body),
		PayloadType: PayloadTypeInToto,
		Signatures:  []EnvelopeSignature{{KeyID: keyid, Sig: base64.StdEncoding.EncodeToString(sig)}},
	}, nil
}
```

- [ ] **Step 5: Run, verify pass:** `go test ./pkg/attest/ -run TestSignEnvelope -v` → PASS.
- [ ] **Step 6: Commit:** `gofmt -w pkg/attest/ && git commit -- pkg/attest/ -m "feat(attest): SignStatement → DSSE envelope (PAE-bound payload, injected Signer)"`

---

### Task 5: `pkg/attest` fail-closed Verify (TDD, negative-heavy)

**Files:** Modify `pkg/attest/envelope.go`; add to `pkg/attest/envelope_test.go`.

- [ ] **Step 1: Define the verifier contract + sentinel.** A `Verifier` re-derives PAE from the decoded payload and checks the signature; `VerifyEnvelope` returns the parsed `Statement` ONLY on success (DSSE rule: hand back the bytes that were verified, never re-parse). Add sentinel:

```go
// ErrVerification is returned by VerifyEnvelope for any fail-closed condition
// (bad base64, wrong payloadType, no signatures, signature mismatch, malformed
// statement). The CLI maps it to cmderr.ErrConflict.
var ErrVerification = errors.New("attest: envelope verification failed")
```

- [ ] **Step 2: Write failing tests** — one per failure mode, all MUST return non-nil error; plus the positive round-trip:

```go
func TestVerifyEnvelopeFailsClosed(t *testing.T) {
	st := NewStatement(
		[]ResourceDescriptor{SubjectFromBytes("app", []byte("x"))},
		Provenance{
			BuildDefinition: BuildDefinition{BuildType: BuildTypeLocal, ExternalParameters: map[string]any{}},
			RunDetails:      RunDetails{Builder: Builder{ID: BuilderIDLocal}},
		},
	)
	// A verifier that accepts ONLY this exact byte string.
	accept := func(pae, sig []byte, keyid string) error {
		if string(sig) == "GOOD" {
			return nil
		}
		return ErrVerification
	}
	signer := func(pae []byte) ([]byte, string, error) { return []byte("GOOD"), "kid", nil }
	good, err := SignStatement(st, signer)
	if err != nil {
		t.Fatalf("SignStatement: %v", err)
	}

	// Positive.
	if _, err := VerifyEnvelope(good, accept); err != nil {
		t.Fatalf("VerifyEnvelope(valid) = %v, want nil", err)
	}

	bad := func(mut func(*Envelope)) Envelope { e := good; mut(&e); return e }
	cases := map[string]Envelope{
		"wrong payloadType": bad(func(e *Envelope) { e.PayloadType = "application/json" }),
		"no signatures":     bad(func(e *Envelope) { e.Signatures = nil }),
		"bad payload b64":   bad(func(e *Envelope) { e.Payload = "!!!not base64!!!" }),
		"bad sig b64":       bad(func(e *Envelope) { e.Signatures = []EnvelopeSignature{{Sig: "!!!"}} }),
		"tampered payload":  bad(func(e *Envelope) { e.Payload = base64.StdEncoding.EncodeToString([]byte(`{"_type":"x"}`)) }),
		"sig rejected":      bad(func(e *Envelope) { e.Signatures = []EnvelopeSignature{{Sig: base64.StdEncoding.EncodeToString([]byte("BAD"))}} }),
	}
	for name, env := range cases {
		if _, err := VerifyEnvelope(env, accept); err == nil {
			t.Errorf("%s: VerifyEnvelope = nil, want error (fail-closed)", name)
		}
	}
}
```

- [ ] **Step 3: Run, verify fail:** `go test ./pkg/attest/ -run TestVerifyEnvelopeFailsClosed -v` → FAIL.
- [ ] **Step 4: Implement** `VerifyEnvelope` (and the `Verifier` type) in `pkg/attest/envelope.go`, fail-closed at every step:

```go
// Verifier checks that sig is a valid signature over pae for the key hinted by
// keyid. It returns nil ONLY on a cryptographically valid signature.
type Verifier func(pae, sig []byte, keyid string) error

// VerifyEnvelope verifies a DSSE envelope fail-closed and returns the parsed
// Statement on success. It (1) rejects a non-in-toto payloadType, (2) requires
// at least one signature, (3) base64-decodes payload + sig, (4) re-derives
// PAE(payloadType, payload) and checks the signature via verify, (5) parses the
// decoded payload into a Statement and returns it. ANY failure → ErrVerification.
// Per the DSSE binding rule, the returned Statement is parsed from the SAME
// bytes that were verified; the envelope is never re-parsed afterward.
func VerifyEnvelope(env Envelope, verify Verifier) (Statement, error) {
	if env.PayloadType != PayloadTypeInToto {
		return Statement{}, fmt.Errorf("%w: unexpected payloadType %q", ErrVerification, env.PayloadType)
	}
	if len(env.Signatures) == 0 {
		return Statement{}, fmt.Errorf("%w: no signatures", ErrVerification)
	}
	body, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil {
		return Statement{}, fmt.Errorf("%w: payload base64: %v", ErrVerification, err)
	}
	pae := PAE(env.PayloadType, body)
	ok := false
	for _, s := range env.Signatures {
		sig, err := base64.StdEncoding.DecodeString(s.Sig)
		if err != nil {
			continue
		}
		if verify(pae, sig, s.KeyID) == nil {
			ok = true
			break
		}
	}
	if !ok {
		return Statement{}, fmt.Errorf("%w: no valid signature", ErrVerification)
	}
	var st Statement
	if err := json.Unmarshal(body, &st); err != nil {
		return Statement{}, fmt.Errorf("%w: malformed statement: %v", ErrVerification, err)
	}
	if st.Type != StatementType {
		return Statement{}, fmt.Errorf("%w: unexpected _type %q", ErrVerification, st.Type)
	}
	return st, nil
}
```

(`errors` is already imported by `build.go`; add `"errors"` to `envelope.go`'s import block for the sentinel.)

- [ ] **Step 5: Run, verify pass:** `go test ./pkg/attest/ -v` → PASS (all negatives + positive + Tasks 2–4).
- [ ] **Step 6: Commit:** `gofmt -w pkg/attest/ && git commit -- pkg/attest/ -m "feat(attest): fail-closed VerifyEnvelope (PAE re-derivation, DSSE binding; sentinel ErrVerification)"`

---

### Task 6: CLI glue + Cobra wrapper — `omni attest` / `omni attest verify` (TDD)

**Files:** Create `internal/cli/attest/attest.go` (+`_test.go`), `cmd/attest.go`. The CLI bridges `pkg/sign` (the concrete Signer/Verifier) into the `pkg/attest` function-typed hooks.

- [ ] **Step 1: Failing test** (`internal/cli/attest/attest_test.go`) — generate against the committed `testing/golden/fixtures/sign/test.key` keypair (low-cost scrypt) then verify the emitted envelope returns nil; a tampered envelope returns a `cmderr.ErrConflict`-classified error; a `--builder-id` not in the allowlist returns `cmderr.ErrInvalidInput`:

```go
package attest_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/internal/cli/attest"
	"github.com/inovacc/omni/internal/cli/cmderr"
)

func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunAttestThenVerifyRoundTrip(t *testing.T) {
	const fixtures = "../../../testing/golden/fixtures/sign"
	t.Setenv("OMNI_SIGN_PASSPHRASE", "test-passphrase") // the fixture key's passphrase
	artifact := writeTemp(t, "app.tar.gz", []byte("artifact-bytes"))
	out := writeTemp(t, "app.intoto.jsonl", nil)

	var w bytes.Buffer
	gen := attest.GenOptions{
		KeyPath:       filepath.Join(fixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		BuilderID:     "", // empty → defaults to local builder.id
		OutPath:       out,
	}
	if err := attest.RunAttest(&w, gen); err != nil {
		t.Fatalf("RunAttest: %v", err)
	}

	ver := attest.VerifyOptions{
		KeyPath:      filepath.Join(fixtures, "test.pub"),
		EnvelopePath: out,
		ArtifactPath: artifact,
	}
	if err := attest.RunVerify(&w, ver); err != nil {
		t.Fatalf("RunVerify(valid) = %v, want nil", err)
	}
}

func TestRunAttestRejectsOverclaim(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", "test-passphrase")
	artifact := writeTemp(t, "app", []byte("x"))
	var w bytes.Buffer
	err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       "../../../testing/golden/fixtures/sign/test.key",
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		BuilderID:     "https://slsa.dev/fake-l3-platform",
		OutPath:       writeTemp(t, "x.jsonl", nil),
	})
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("overclaim builder.id: err = %v, want cmderr.ErrInvalidInput", err)
	}
}

func TestRunVerifyTamperedFailsConflict(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", "test-passphrase")
	artifact := writeTemp(t, "app", []byte("x"))
	out := writeTemp(t, "x.jsonl", nil)
	var w bytes.Buffer
	if err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath: "../../../testing/golden/fixtures/sign/test.key", ArtifactPath: artifact,
		PredicateType: "slsa-provenance", OutPath: out,
	}); err != nil {
		t.Fatalf("RunAttest: %v", err)
	}
	env, _ := os.ReadFile(out)
	tampered := bytes.Replace(env, []byte(`"payload"`), []byte(`"PAYLOAD"`), 1)
	bad := writeTemp(t, "bad.jsonl", tampered)
	err := attest.RunVerify(&w, attest.VerifyOptions{
		KeyPath: "../../../testing/golden/fixtures/sign/test.pub", EnvelopePath: bad, ArtifactPath: artifact,
	})
	if !cmderr.IsConflict(err) && !cmderr.IsInvalidInput(err) {
		t.Fatalf("tampered envelope: err = %v, want Conflict or InvalidInput", err)
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./internal/cli/attest/ -v` → FAIL (package undefined).
- [ ] **Step 3: Implement** `internal/cli/attest/attest.go`. Options structs + `RunAttest` / `RunVerify`. The Signer adapter calls `sign.Sign` and derives the hex keyid from the secret key's public half via the parsed key; the Verifier adapter calls `sign.Verify`. Map errors per the cmderr table:

```go
package attest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/attest"
	"github.com/inovacc/omni/pkg/sign"
)

// GenOptions configures `omni attest` (generate).
type GenOptions struct {
	KeyPath       string // path to the *.key secret key (file path only, never key material)
	ArtifactPath  string // artifact whose sha256 becomes the subject digest
	PredicateType string // only "slsa-provenance" supported
	PredicatePath string // optional pre-built predicate JSON; if empty, build from flags/env
	BuilderID     string // empty → local builder.id; must be ADR-0009-allowed otherwise
	FromEnv       bool   // populate provenance from GITHUB_* env vars
	OutPath       string // output envelope path; empty → stdout
}

// VerifyOptions configures `omni attest verify`.
type VerifyOptions struct {
	KeyPath      string // path to the *.pub public key
	EnvelopePath string // path to the DSSE envelope JSON
	ArtifactPath string // optional: if set, verify subject sha256 matches this artifact
}

func readFile(path string, notFound, perm error) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, cmderr.Wrap(notFound, fmt.Sprintf("read %s", path))
		}
		if os.IsPermission(err) {
			return nil, cmderr.Wrap(perm, fmt.Sprintf("read %s", path))
		}
		return nil, cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("read %s", path))
	}
	return b, nil
}

// RunAttest generates and writes a signed DSSE/SLSA provenance attestation.
func RunAttest(w io.Writer, opts GenOptions) error {
	if opts.PredicateType != "slsa-provenance" {
		return cmderr.Wrap(cmderr.ErrUnsupported, "only --predicate-type slsa-provenance is supported")
	}
	keyText, err := readFile(opts.KeyPath, cmderr.ErrNotFound, cmderr.ErrPermission)
	if err != nil {
		return err
	}
	sk, err := sign.ParseSecretKey(keyText, os.Getenv("OMNI_SIGN_PASSPHRASE"))
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "parse secret key")
	}
	artifact, err := readFile(opts.ArtifactPath, cmderr.ErrNotFound, cmderr.ErrPermission)
	if err != nil {
		return err
	}

	prov, subjName, err := buildProvenance(opts, opts.ArtifactPath)
	if err != nil {
		return err // already classified
	}
	st := attest.NewStatement(
		[]attest.ResourceDescriptor{attest.SubjectFromBytes(subjName, artifact)},
		prov,
	)

	signer := func(pae []byte) ([]byte, string, error) {
		sig, e := sign.Sign(pae, sk)
		if e != nil {
			return nil, "", e
		}
		return sig, hex.EncodeToString(sk.KeyID[:]), nil
	}
	env, err := attest.SignStatement(st, signer)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrIO, "sign statement")
	}
	out, _ := json.MarshalIndent(env, "", "  ")
	if opts.OutPath == "" {
		_, _ = w.Write(out)
		_, _ = w.Write([]byte("\n"))
		return nil
	}
	if err := os.WriteFile(opts.OutPath, append(out, '\n'), 0o644); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, "write envelope")
	}
	return nil
}

// RunVerify verifies a DSSE/SLSA envelope fail-closed.
func RunVerify(w io.Writer, opts VerifyOptions) error {
	pubText, err := readFile(opts.KeyPath, cmderr.ErrNotFound, cmderr.ErrPermission)
	if err != nil {
		return err
	}
	pub, err := sign.ParsePublicKey(pubText)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "parse public key")
	}
	envText, err := readFile(opts.EnvelopePath, cmderr.ErrNotFound, cmderr.ErrPermission)
	if err != nil {
		return err
	}
	var env attest.Envelope
	if err := json.Unmarshal(envText, &env); err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "parse envelope JSON")
	}
	verifier := func(pae, sig []byte, _ string) error { return sign.Verify(pae, sig, pub) }
	st, err := attest.VerifyEnvelope(env, verifier)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrConflict, "attestation verification failed")
	}
	// Optional artifact binding: subject sha256 must match the provided artifact.
	if opts.ArtifactPath != "" {
		artifact, err := readFile(opts.ArtifactPath, cmderr.ErrNotFound, cmderr.ErrPermission)
		if err != nil {
			return err
		}
		want := attest.SubjectFromBytes("", artifact).Digest["sha256"]
		matched := false
		for _, s := range st.Subject {
			if s.Digest["sha256"] == want {
				matched = true
				break
			}
		}
		if !matched {
			return cmderr.Wrap(cmderr.ErrConflict, "artifact digest does not match any subject")
		}
	}
	_, _ = fmt.Fprintf(w, "OK: %d subject(s), predicateType %s\n", len(st.Subject), st.PredicateType)
	return nil
}
```

- [ ] **Step 4: Implement `buildProvenance`** in `internal/cli/attest/provenance.go` (file `internal/cli/attest/provenance.go`). It honors `--predicate <file>` (use it verbatim), else builds from flags/env, and ALWAYS validates the builder.id against the ADR-0009 allowlist:

```go
package attest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/attest"
)

// buildProvenance constructs the SLSA Provenance v1 predicate and returns it
// plus the subject name. It enforces the ADR-0009 builder.id allowlist.
func buildProvenance(opts GenOptions, artifactPath string) (attest.Provenance, string, error) {
	subjName := filepath.Base(artifactPath)

	// Explicit predicate file: trust its shape but still validate builder.id.
	if opts.PredicatePath != "" {
		raw, err := os.ReadFile(opts.PredicatePath)
		if err != nil {
			return attest.Provenance{}, "", cmderr.Wrap(cmderr.ErrNotFound, "read predicate")
		}
		var prov attest.Provenance
		if err := json.Unmarshal(raw, &prov); err != nil {
			return attest.Provenance{}, "", cmderr.Wrap(cmderr.ErrInvalidInput, "parse predicate JSON")
		}
		if err := attest.ValidateBuilderID(prov.RunDetails.Builder.ID); err != nil {
			return attest.Provenance{}, "", cmderr.Wrap(cmderr.ErrInvalidInput, "predicate builder.id")
		}
		return prov, subjName, nil
	}

	builderID := opts.BuilderID
	buildType := attest.BuildTypeLocal
	ext := map[string]any{"artifact": subjName}
	var meta *attest.BuildMetadata

	if opts.FromEnv && os.Getenv("GITHUB_ACTIONS") == "true" {
		builderID = attest.BuilderIDRelease
		buildType = attest.BuildTypeGHA
		ext = map[string]any{
			"workflow":   os.Getenv("GITHUB_WORKFLOW"),
			"repository": os.Getenv("GITHUB_REPOSITORY"),
			"ref":        os.Getenv("GITHUB_REF"),
			"sha":        os.Getenv("GITHUB_SHA"),
		}
		meta = &attest.BuildMetadata{
			InvocationID: os.Getenv("GITHUB_RUN_ID"),
			StartedOn:    time.Now().UTC().Format(time.RFC3339),
		}
	}
	if builderID == "" {
		builderID = attest.BuilderIDLocal
	}
	if err := attest.ValidateBuilderID(builderID); err != nil {
		return attest.Provenance{}, "", cmderr.Wrap(cmderr.ErrInvalidInput, "builder.id (SLSA overclaim refused)")
	}
	return attest.Provenance{
		BuildDefinition: attest.BuildDefinition{BuildType: buildType, ExternalParameters: ext},
		RunDetails:      attest.RunDetails{Builder: attest.Builder{ID: builderID}, Metadata: meta},
	}, subjName, nil
}
```

(Note: `time` output makes `--from-env` non-deterministic; golden tests in Task 8 never exercise `--from-env`, only fixed-input local generation, and `startedOn` is only set on the env path. Local generation produces byte-identical output for identical inputs.)

- [ ] **Step 5: Cobra wrapper** `cmd/attest.go` — mirror `cmd/sign.go`'s parent+subcommand shape:

```go
package cmd

import (
	"github.com/inovacc/omni/internal/cli/attest"
	"github.com/spf13/cobra"
)

var attestCmd = &cobra.Command{
	Use:   "attest [OPTION]... --artifact FILE",
	Short: "Generate a signed SLSA provenance attestation (DSSE/in-toto)",
	Long: `Generate an in-toto Statement v1 with a SLSA Provenance v1 predicate,
wrapped in a DSSE envelope and signed with a minisign-compatible Ed25519 secret
key (see 'omni sign keygen'). The passphrase is read from OMNI_SIGN_PASSPHRASE
or an interactive prompt — never a flag.

The claimed SLSA level is fixed by ADR-0009 via builder.id; omni refuses to emit
any builder.id outside the ADR allowlist (no SLSA overclaim).

  -k, --key FILE          secret key file (*.key)
  -a, --artifact FILE     artifact to attest (its sha256 is the subject digest)
      --predicate-type T  predicate type (only: slsa-provenance)
      --predicate FILE    use a pre-built predicate JSON instead of building one
      --builder-id URI    builder.id (must be ADR-0009-allowed; default: local)
      --from-env          populate provenance from GITHUB_* env vars (release path)
  -o, --out FILE          output envelope path (default: stdout)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := attest.GenOptions{}
		opts.KeyPath, _ = cmd.Flags().GetString("key")
		opts.ArtifactPath, _ = cmd.Flags().GetString("artifact")
		opts.PredicateType, _ = cmd.Flags().GetString("predicate-type")
		opts.PredicatePath, _ = cmd.Flags().GetString("predicate")
		opts.BuilderID, _ = cmd.Flags().GetString("builder-id")
		opts.FromEnv, _ = cmd.Flags().GetBool("from-env")
		opts.OutPath, _ = cmd.Flags().GetString("out")
		return attest.RunAttest(cmd.OutOrStdout(), opts)
	},
}

var attestVerifyCmd = &cobra.Command{
	Use:   "verify [OPTION]... ENVELOPE",
	Short: "Verify a SLSA provenance attestation fail-closed",
	Long: `Verify a DSSE/in-toto SLSA provenance envelope against a public key.
Every failure mode (bad signature, malformed envelope, digest mismatch) exits
non-zero with a classified error. With --artifact, also binds the envelope to a
specific artifact by sha256.

  -k, --key FILE       public key file (*.pub)
  -a, --artifact FILE  optional artifact to bind by sha256 to a subject`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := attest.VerifyOptions{}
		opts.KeyPath, _ = cmd.Flags().GetString("key")
		opts.ArtifactPath, _ = cmd.Flags().GetString("artifact")
		if len(args) == 1 {
			opts.EnvelopePath = args[0]
		} else {
			opts.EnvelopePath, _ = cmd.Flags().GetString("envelope")
		}
		return attest.RunVerify(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(attestCmd)
	attestCmd.AddCommand(attestVerifyCmd)

	attestCmd.Flags().StringP("key", "k", "", "secret key file (*.key)")
	attestCmd.Flags().StringP("artifact", "a", "", "artifact to attest")
	attestCmd.Flags().String("predicate-type", "slsa-provenance", "predicate type (only: slsa-provenance)")
	attestCmd.Flags().String("predicate", "", "pre-built predicate JSON file")
	attestCmd.Flags().String("builder-id", "", "builder.id (ADR-0009-allowed; default: local)")
	attestCmd.Flags().Bool("from-env", false, "populate provenance from GITHUB_* env vars")
	attestCmd.Flags().StringP("out", "o", "", "output envelope path (default: stdout)")

	attestVerifyCmd.Flags().StringP("key", "k", "", "public key file (*.pub)")
	attestVerifyCmd.Flags().StringP("artifact", "a", "", "artifact to bind by sha256")
	attestVerifyCmd.Flags().String("envelope", "", "envelope path (alternative to positional arg)")
}
```

- [ ] **Step 6: Run, verify pass:** `go test ./internal/cli/attest/... ./pkg/attest/... -v` → PASS. Manual smoke (reuse the sign fixture key):
```bash
OMNI_SIGN_PASSPHRASE=test-passphrase go run . attest \
  --key testing/golden/fixtures/sign/test.key \
  --artifact testing/golden/fixtures/sign/data.txt \
  --out /tmp/app.intoto.jsonl
go run . attest verify --key testing/golden/fixtures/sign/test.pub \
  --artifact testing/golden/fixtures/sign/data.txt /tmp/app.intoto.jsonl
```
- [ ] **Step 7: Commit:** `gofmt -w internal/cli/attest cmd/attest.go && git commit -- internal/cli/attest cmd/attest.go -m "feat(attest): omni attest / attest verify CLI (pkg/sign signer, ADR-0009 enforcement, fail-closed verify)"`

---

### Task 7: `omni pipe` integration (attest verify)

**Files:** Modify `cmd/pipe.go`.

- [ ] **Step 1: Failing test** — a `pipe`-style invocation of `attest verify` works via the Unified registry (mirror the existing `verify` pipe test in the pipe test file). The registered command reads an envelope from stdin and verifies against a key path passed in args; since pipe's `AdaptWriterReaderArgs` gives `(w, r, args)`, the adapter reads the envelope from `r` and the pubkey path from `args[0]`.
- [ ] **Step 2: Implement** — in `buildPipeRegistry()` (after the existing `verify` registration ~line 212) add:

```go
reg.Register("attest-verify", command.AdaptWriterReaderArgs(
	func(w io.Writer, r io.Reader, args []string) error {
		return attest.RunVerifyReader(w, r, args)
	}))
```

and add a thin `RunVerifyReader(w io.Writer, r io.Reader, args []string) error` to `internal/cli/attest/attest.go` that reads the envelope from `r`, takes `args[0]` as the pubkey path (`cmderr.ErrInvalidInput` if missing), and reuses the same verifier logic as `RunVerify` (factor the core into an unexported `verifyEnvelopeBytes(pubText, envText []byte, artifact []byte) (attest.Statement, error)` shared by both). `attest` (generate) stays Cobra-only — it consumes a file path, not a stdin stream. Import `io` in `cmd/pipe.go` if not already present (it is, given existing adapters).
- [ ] **Step 3: Run, verify pass:** `go test ./internal/cli/pipe/... ./cmd/... ./internal/cli/attest/...`.
- [ ] **Step 4: Commit:** `git commit -- cmd/pipe.go internal/cli/attest -m "feat(attest): register attest-verify in the pipe Unified registry"`

---

### Task 8: Golden-master tests (positive + negative, deterministic)

**Files:** Modify `testing/golden/golden_tests.yaml` AND `tools/golden/golden_tests.yaml` (keep in sync); add fixtures under `testing/golden/fixtures/attest/` (a committed, pre-signed envelope + the artifact it attests). Reuse the existing `testing/golden/fixtures/sign/` keypair (`test.key`/`test.pub`, passphrase `test-passphrase`).

- [ ] **Step 1: Create deterministic fixtures.** Add a `gen_fixtures.go` helper under `testing/golden/fixtures/attest/` (mirroring `testing/golden/fixtures/sign/gen_fixtures.go`) that, given the sign fixture key + a fixed artifact (`testing/golden/fixtures/attest/artifact.bin` with fixed bytes, e.g. `omni attest fixture v1\n`), produces `testing/golden/fixtures/attest/app.intoto.jsonl` using `BuilderIDLocal` (no `--from-env`, no timestamps → byte-deterministic). Commit `artifact.bin`, `app.intoto.jsonl`, and a `tampered.jsonl` (one byte of the base64 payload flipped). Do NOT generate live in golden runs.
- [ ] **Step 2: Add an `attest` category to BOTH yaml files** with these cases (annotate each negative with its sentinel):
  - `attest_verify_ok`: `attest verify --key {fixtures}/attest/test.pub --artifact {fixtures}/attest/artifact.bin {fixtures}/attest/app.intoto.jsonl` → exit 0; `normalizations: ["strip_path"]`. (Use the `fixtures_dir:` mechanism as the existing `sign` cases do; copy `test.pub` into the attest fixture dir or reference the sign dir per the harness's `fixtures_dir` convention.)
  - `attest_verify_tampered`: against `tampered.jsonl` → `exit_code: 1` `# cmderr.ErrConflict`.
  - `attest_verify_wrong_key`: against `{fixtures}/sign/wrong.pub` → `exit_code: 1` `# cmderr.ErrConflict`.
  - `attest_verify_bad_envelope`: envelope = `artifact.bin` (not JSON) → `exit_code: 2` `# cmderr.ErrInvalidInput`.
  - `attest_overclaim`: `attest --key {fixtures}/sign/test.key --artifact {fixtures}/attest/artifact.bin --builder-id https://slsa.dev/fake-l3 --out -` (with `OMNI_SIGN_PASSPHRASE`) → `exit_code: 2` `# cmderr.ErrInvalidInput (SLSA overclaim refused)`.
  - `attest_unsupported_predicate`: `--predicate-type spdx` → `exit_code: 6` `# cmderr.ErrUnsupported`.
  Add `normalizations: ["strip_path"]` to every case whose output contains paths.
- [ ] **Step 3: Generate + verify snapshots:** `task test:golden:update && task golden:record`, then `python testing/scripts/test_golden.py` → all green.
- [ ] **Step 4: Commit:** `git commit -- testing/golden tools/golden -m "test(attest): golden-master positive + fail-closed/overclaim negative cases"`

---

### Task 9: CI SLSA-schema validation gate + docs + final gate

**Files:** Modify `Taskfile.yml` (add `attest:validate-schema`), `docs/COMMANDS.md`, `CLAUDE.md` (command inventory count), `docs/superpowers/specs/2026-05-16-07-pkg-attest-design.md` (status); add the official SLSA Provenance v1 JSON schema under `testing/schemas/slsa-provenance-v1.schema.json`.

- [ ] **Step 1: Commit the SLSA Provenance v1 JSON schema** to `testing/schemas/slsa-provenance-v1.schema.json` (the official schema from the in-toto attestation / SLSA repo). Add a `task attest:validate-schema` target that: (a) generates an envelope from the attest fixture (local builder.id), (b) base64-decodes `.payload`, (c) extracts `.predicate`, and (d) validates it against the committed schema using the existing Python golden harness's JSON tooling (pure-Python `jsonschema`, no new Go dep; it runs in CI only, never in the binary). The target fails if the predicate deviates or if `builder.id` is not the ADR-0009 release/local value. This is the spec's "validated against official SLSA v1.0 JSON schema in CI before every release" requirement.
- [ ] **Step 2: Docs** — add `attest` / `attest verify` to `docs/COMMANDS.md`, bump the CLAUDE.md inventory count, and regenerate `omni cmdtree` / `omni aicontext` output if applicable. Mark the spec `Status: Complete`. Add an `docs/EXTERNAL_SOURCES.md` entry for the in-toto/DSSE/SLSA specs and the SLSA schema file (attribution).
- [ ] **Step 3: Final gate:**
```bash
go build ./...
go vet ./... && gofmt -l pkg/attest internal/cli/attest cmd/attest.go
golangci-lint run --timeout=5m ./...
go test ./pkg/attest/... ./internal/cli/attest/... ./internal/cli/pipe/... ./cmd/... -count=1
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ./... && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...
task attest:validate-schema
python testing/scripts/test_golden.py
```
Expected: all green; the predicate validates against the SLSA v1.0 schema; the emitted `builder.id` is exactly the ADR-0009 value.
- [ ] **Step 4: Commit:** `git commit -- Taskfile.yml testing/schemas docs/ CLAUDE.md -m "ci(attest): SLSA v1.0 schema validation gate; docs; mark Phase 07 complete"`

---

## Self-Review

**Spec coverage** (ATTEST success criteria → tasks):

| Requirement (spec) | Task(s) |
|---|---|
| ADR pinning the honest SLSA level before any code | **T1** (ADR-0009, hard gate) |
| `omni attest --predicate-type slsa-provenance --predicate <file> --artifact <path>` → in-toto Statement + SLSA v1.0 predicate in DSSE envelope | **T2** (types/PAE) + **T3** (statement/predicate) + **T4** (SignStatement) + **T6** (CLI, `--predicate`/`--artifact`) |
| PAE format validated against in-toto reference test vectors | **T2** (`TestPAEReferenceVector` pins `DSSEv1 29 … 11 hello world`) |
| `omni attest verify <envelope>` fail-closed; every error mode non-zero with cmderr classification | **T5** (negative-heavy `VerifyEnvelope`) + **T6** (cmderr mapping) + **T8** (golden negatives) |
| `--from-env` auto-populates from `GITHUB_RUN_ID`/`GITHUB_WORKFLOW`/`GITHUB_SHA`; `builder.id` from OIDC/env, not hardcoded-arbitrary | **T6** (`buildProvenance` env path → `BuilderIDRelease` + GHA externalParameters) |
| Claimed level == ADR-pinned level; no overclaim | **T3** (`ValidateBuilderID` allowlist + `ErrOverclaim`) + **T6** (refusal → `ErrInvalidInput`) + **T9** (CI schema/builder.id gate) |
| SLSA predicate validated against official v1.0 JSON schema in CI before release | **T9** (`task attest:validate-schema` + committed schema) |
| `pkg/attest` reusable standalone library; CLI thin | **T2–T5** (lib, no cobra) + **T6** (glue) ; pipe **T7** |
| Pure-Go, no exec, reuse Phase 04 sign | **T4/T6** (inject `pkg/sign.Sign`/`Verify`; no new third-party deps) |

**Placeholder scan:** Every code step contains complete, compilable Go or a byte-exact format spec. The three wire formats are pinned to published specs in the "Authoritative wire formats" section (PAE reference vector `DSSEv1 29 http://example.com/HelloWorld 11 hello world`; DSSE envelope shape; in-toto Statement v1 + SLSA Provenance v1 field tables). No "TBD"/"add validation"/"handle edge cases"/"similar to Task N". The only deliberately deferred artifact is the verbatim SLSA JSON schema file (T9 Step 1) — it is an external published document copied in, not logic to invent.

**Type consistency:** `attest.Signer`/`attest.Verifier` (T4/T5) are the seams the CLI fills with `pkg/sign.Sign`/`Verify` (T6). `attest.ErrVerification` (T5) → `cmderr.ErrConflict` (T6 `RunVerify`). `attest.ErrOverclaim` (T3) → `cmderr.ErrInvalidInput` (T6 `buildProvenance`). `PayloadTypeInToto`/`StatementType`/`PredicateTypeSLSAProvenance` constants (T2) are emitted by `SignStatement` (T4) and re-checked by `VerifyEnvelope` (T5). `BuilderIDRelease`/`BuilderIDLocal`/`BuildTypeGHA`/`BuildTypeLocal` (T3) are the ADR-0009-pinned constants used by `buildProvenance` (T6) and `ValidateBuilderID` (T3). `Statement`/`Provenance`/`BuildDefinition`/`RunDetails`/`Builder`/`BuildMetadata`/`Envelope`/`EnvelopeSignature`/`ResourceDescriptor` are all defined in T2 before first use. `sk.KeyID[:]` (T6 signer) exists on `sign.SecretKey` (confirmed from Phase 04 source: `SecretKey{KeyID [8]byte; …}`).

**Known risks:**
1. **`sign.Sign` signs raw bytes, not a prehash chosen by us** — `pkg/sign` prehashes internally (Blake2b-512) for the `"ED"` scheme. We pass the full PAE bytes to `sign.Sign`; that is correct (the signer is opaque to DSSE, which only requires `verify(PAE) == sign(PAE)`). The minisig blob in `sig` is non-standard-DSSE (most tooling expects a raw Ed25519 signature), so omni-produced envelopes are verifiable by `omni attest verify` but NOT by generic cosign/sigstore verifiers. ADR-0009 must state this interop limitation explicitly (omni-signed attestations are self-consistent, not cosign-interoperable in v1.0). Mitigation noted; a future phase can add a raw-Ed25519 DSSE signer behind a flag.
2. **JSON field ordering / canonicalization** — `encoding/json` emits struct fields in declaration order deterministically, so identical inputs yield byte-identical payloads (golden-safe). We do NOT claim JCS canonicalization; if a future consumer needs JCS, add it behind a flag. Golden fixtures use local builder.id + no timestamps to stay deterministic.
3. **`--from-env` non-determinism** (`startedOn` timestamp) — intentionally excluded from golden tests; only the fixed-input local path is golden-tested. Documented in T6 Step 4.
4. **SLSA L2 honesty** — the binary cannot itself prove L2; the ADR + CI schema/builder.id gate (T9) are the enforcement. The code-level guarantee is narrower but real: omni refuses to emit any builder.id outside the allowlist, so it can never silently overclaim L3.
