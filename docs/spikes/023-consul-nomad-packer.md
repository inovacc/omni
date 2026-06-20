# Spike 023: Consul / Nomad / Packer integration assessment

> **Status**: design note + read-only Consul/Nomad MVP shipped; Packer deferred.
> **Planned at**: commit `36af526f`, 2026-06-20. See `plans/023-consul-nomad-packer.md`.

## 1. No-exec restatement

omni's foundational design principle is **"No exec — never spawn external
processes."** All three of Consul, Nomad, and Packer must therefore be
implemented in **pure Go** (`net/http` + `encoding/json` against the tools' HTTP
APIs, or pure-Go libraries). Shelling out to the `consul`, `nomad`, or `packer`
binaries is permanently off the table. That single rule is what shapes every
decision below.

## 2. Per-tool API assessment

| Tool | Has read API? | Pure-Go read MVP feasible? | Verdict |
|------|---------------|----------------------------|---------|
| **Consul** | Yes — HTTP+JSON (`/v1/agent/members`, `/v1/kv/<key>`, `/v1/catalog/services`) | Yes, trivial via `net/http` | **Read MVP shipped** |
| **Nomad** | Yes — HTTP+JSON (`/v1/jobs`, `/v1/nodes`, `/v1/allocations`) | Yes, trivial via `net/http` | **Read MVP shipped** |
| **Packer** | **No useful read API** — HCL build CLI | Only `fmt` via HCL2; `validate` needs Packer schema/plugins | **Deferred** (see §6) |

### Consul

Consul agents expose a plain HTTP+JSON API on `:8500`. The read endpoints the
MVP needs require no agent/Raft/gossip access:

- `GET /v1/agent/members` → cluster members (name, addr, status enum).
- `GET /v1/kv/<key>` → KV entry; the `Value` is base64-encoded in the JSON.
- `GET /v1/catalog/services` → map of service name → tags.

### Nomad

Nomad servers expose a plain HTTP+JSON API on `:4646`. Read endpoints used:

- `GET /v1/jobs` → job stubs (ID, Name, Status).
- `GET /v1/nodes` → node stubs (ID, Name, Status, Drain).
- `GET /v1/allocations` → allocation stubs (ID, JobID, ClientStatus).

### Packer

`packer` is an HCL2 build/orchestration CLI. It has **no read API** to model — it
drives builder plugins and provisions images. The only operations conceptually
expressible in pure Go are `fmt` and `validate` over HCL2 templates, and both are
problematic (see §6).

## 3. Dependency recommendation (the maintainer decision)

The central question of this spike: **official clients vs. hand-rolled
`net/http`.**

| Option | Pros | Cons |
|--------|------|------|
| Official `consul/api` + `nomad/api` | Ergonomic; full surface; well-tested | **Each is a large dependency** that vendors much of its parent repo (pulls big chunks of Consul/Nomad internals, extra HashiCorp libs). Heavy for a read-only MVP. |
| **Hand-rolled `net/http` + `encoding/json`** (chosen) | **Zero new deps**; tiny surface (only the handful of read endpoints modeled); matches omni's stdlib-first value; keeps the static binary lean | You hand-model response structs; if the read surface grows large, this becomes maintenance work |

**Recommendation: hand-rolled `net/http` for the read MVP.** It adds **zero**
dependencies to `go.mod`/`go.sum`, models only the endpoints we need, and honors
omni's stdlib-first principle. The official clients remain a deliberate
**maintainer decision** to be revisited only if the read surface grows enough
that hand-modeled structs become a burden (tracked as a deferred follow-up).

**Counter-precedent — why vault differs:** `internal/cli/vault` *does* use the
official `hashicorp/vault/api` client. Vault was justified because its surface is
large (KV v2 versioning, multiple auth methods, token lifecycle, metadata) and
the client encapsulates non-trivial request shaping. Read-only Consul/Nomad
(three GET endpoints each, flat JSON) does **not** need that weight — a few
structs and one `http.Get` helper cover it. Adding two more large HashiCorp
clients to drag in the entire Consul/Nomad codebases for six read endpoints would
be a poor trade.

## 4. Command surface

Legend: ✅ shipped in this MVP · ⏸ deferred (write/mutating, gated behind an
explicit destructive flag per CLAUDE.md "safe defaults") · ❌ impossible
(no-exec).

### Consul

| Command | Endpoint | Method | v1 status |
|---------|----------|--------|-----------|
| `omni consul members` | `/v1/agent/members` | GET | ✅ read |
| `omni consul kv get <key>` | `/v1/kv/<key>` | GET | ✅ read |
| `omni consul services` | `/v1/catalog/services` | GET | ✅ read |
| `omni consul kv put <key> <value>` | `/v1/kv/<key>` | PUT | ⏸ write — deferred |
| `omni consul kv delete <key>` | `/v1/kv/<key>` | DELETE | ⏸ write — deferred |
| service register / deregister | `/v1/agent/service/*` | PUT | ⏸ write — deferred |

### Nomad

| Command | Endpoint | Method | v1 status |
|---------|----------|--------|-----------|
| `omni nomad job list` | `/v1/jobs` | GET | ✅ read |
| `omni nomad node list` | `/v1/nodes` | GET | ✅ read |
| `omni nomad alloc list` | `/v1/allocations` | GET | ✅ read |
| `omni nomad job run/stop` | `/v1/jobs`, `/v1/job/<id>` | POST/DELETE | ⏸ write — deferred |
| node drain / eligibility | `/v1/node/<id>/drain` | POST | ⏸ write — deferred |
| alloc signal / restart | `/v1/client/allocation/*` | POST | ⏸ write — deferred |

### Packer

| Command | v1 status |
|---------|-----------|
| `packer build` | ❌ impossible (no-exec) — permanently out of scope |
| `packer validate` | ⏸ deferred — not truly pure-Go (needs Packer schema/plugins) |
| `packer fmt` | ⏸ deferred — only pure-Go candidate but heavy dep, low value |

## 5. Auth / env model

Both clients default each `Options` field to the corresponding env var when the
field is empty, mirroring `internal/cli/vault`. A `--token` flag is accepted but
the value is also read from env so tokens never need to appear in shell history.

| Tool | Address env (default) | Token env | Token header | Namespace env | Other |
|------|-----------------------|-----------|--------------|---------------|-------|
| **Consul** | `CONSUL_HTTP_ADDR` (`http://127.0.0.1:8500`) | `CONSUL_HTTP_TOKEN` | `X-Consul-Token` | `CONSUL_NAMESPACE` | `CONSUL_HTTP_SSL`, `CONSUL_CACERT` (not yet wired) |
| **Nomad** | `NOMAD_ADDR` (`http://127.0.0.1:4646`) | `NOMAD_TOKEN` | `X-Nomad-Token` | `NOMAD_NAMESPACE` | `NOMAD_REGION`, `NOMAD_CACERT` (region passed as query; CACERT not yet wired) |

Namespace (Consul) and namespace+region (Nomad) are sent as the `ns` / `region`
query parameters that the respective APIs expect.

## 6. Packer: deferred

The recommendation is to **defer Packer entirely**. Reasons, recorded so nobody
re-litigates this:

1. **No-exec kills `build` outright.** `packer build` drives builder/provisioner
   plugins and provisions real images — it is inherently a subprocess-spawning
   build operation. It cannot exist under omni's non-negotiable no-exec rule.
   This is **permanent**, not a "for now."
2. **`validate` is not truly doable in pure Go.** `packer validate` checks an HCL2
   template *against Packer's own template/plugin schema and semantics*. The
   `github.com/hashicorp/hcl/v2` library gives you HCL **syntax** (parse, format,
   diagnostics) but not Packer's schema — reimplementing that schema/plugin
   resolution would be a large, brittle effort that tracks Packer's internals.
3. **`fmt` alone is the only genuinely pure-Go candidate, and it is not worth the
   dependency weight.** A pure-Go `packer fmt` would need `hashicorp/hcl/v2`
   (`hclsyntax` formatter), which is a **heavy new dependency** — it drags in
   `zclconf/go-cty`, `apparentlymart/go-textseg`, `agext/levenshtein`,
   `mitchellh/go-wordwrap`, and more. `github.com/hashicorp/hcl/v2` is **not**
   currently in `go.mod`/`go.sum` (only HCL **v1**, indirect via vault). Pulling
   HCL2 for a single low-value formatter that also overlaps omni's existing
   formatters is a bad trade.
4. **This is the lowest-urgency P1.** Consul and Nomad deliver the
   HashiCorp-ecosystem value the maintainer/CI audience wants; Packer does not.
   Spending the dependency budget here is not justified.

**Backlog annotation:** the Packer items in `docs/BACKLOG.md` stay **unchecked**,
annotated `(deferred — see docs/spikes/023)`. Revisit only if a pure-Go HCL2
`fmt` ever becomes worth the dep, or if the no-exec rule is relaxed (currently
non-negotiable, which would also reopen `build`/`validate`).

## 7. Implementation notes (what shipped)

- `internal/cli/consul/consul.go` — hand-rolled `net/http` client; `Options`
  with env defaults; `Members()`, `KVGet()`, `Services()`; `classifyConsulError`
  mapping 401/403→`ErrPermission`, 404→`ErrNotFound`, 400→`ErrInvalidInput`,
  else→`ErrIO`. All printers take `io.Writer`. JSON via `--json`.
- `internal/cli/nomad/nomad.go` — same shape; `JobList()`, `NodeList()`,
  `AllocList()`; `classifyNomadError` with the identical status→sentinel map.
- `cmd/consul.go`, `cmd/nomad.go` — thin Cobra wrappers; env-documented `Long`;
  `--address`/`--token`/`--namespace` flags; `--json` honored via the root
  persistent flag.
- Tests are `httptest`-based (no live cluster, no exec) and table-driven; they
  assert the cmderr sentinel for 404/401/5xx via `errors.Is`.
- **Zero new dependencies.** Nothing added under `pkg/` (API freeze untouched).
- Golden snapshot cases are **deferred** for this spike MVP (unit + httptest
  coverage only) — a tracked follow-up per the plan's maintenance notes.
