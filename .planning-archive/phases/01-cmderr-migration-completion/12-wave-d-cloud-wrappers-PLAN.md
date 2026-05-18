---
phase: 01-cmderr-migration-completion
plan: 12
type: execute
wave: D
depends_on: [09, 10, 11]
files_modified:
  - internal/cli/git/git.go
  - internal/cli/gh/gh.go
  - internal/cli/kubectl/kubectl.go
  - internal/cli/kubehacks/kubehacks.go
  - internal/cli/terraform/terraform.go
  - internal/cli/aws/aws.go
  - internal/cli/cloud/cloud.go
  - internal/cli/vault/vault.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "git, gh, kubectl, kubehacks, terraform, aws, cloud, vault classify their error returns via Pattern 2/6"
    - "Each command has ≥1 error-path golden"
    - "No command accidentally spawns external processes (flag pre-existing violations per CLAUDE.md 'no exec' rule)"
  artifacts:
    - path: "internal/cli/git/git.go"
      provides: "cmderr classification"
      contains: "cmderr."
    - path: "internal/cli/kubectl/kubectl.go"
      provides: "cmderr classification"
      contains: "cmderr."
    - path: "internal/cli/terraform/terraform.go"
      provides: "cmderr classification"
      contains: "cmderr."
  key_links:
    - from: "cloud wrapper commands"
      to: "cmderr sentinels"
      via: "Pattern 2/6 classification at CLI boundary"
      pattern: "cmderr\\.Wrap"
---

# Plan 12 — Wave D: Cloud / DevOps wrappers

## Goal

Migrate 8 cloud/devops wrapper commands: `git`, `gh`, `kubectl`, `kubehacks`, `terraform`, `aws`, `cloud`, `vault`. Most delegate to in-tree Go libraries (per CLAUDE.md `k8s.io/kubectl` local replace); classification is mechanical. Any exec-spawning code discovered is flagged as a pre-existing bug (Risky Commands note about `exec`).

## Wave

Wave D.

## Requirements covered

POLISH-01, POLISH-02.

## Depends on

Plans 09, 10, 11 (Wave C done).

## Parallelizable with

Plans 13, 14, 15.

## Commands touched

- `internal/cli/git/`, `internal/cli/gh/`, `internal/cli/kubectl/`, `internal/cli/kubehacks/`, `internal/cli/terraform/`, `internal/cli/aws/`, `internal/cli/cloud/`, `internal/cli/vault/`

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
@internal/cli/find/find.go
@internal/cli/head/head.go

## Tasks

### Task 1: Migrate git + gh + kubectl + kubehacks

**Files:**
- `internal/cli/git/git.go`, `git_test.go`
- `internal/cli/gh/gh.go`, `gh_test.go`
- `internal/cli/kubectl/kubectl.go`, `kubectl_test.go`
- `internal/cli/kubehacks/kubehacks.go`, `kubehacks_test.go`

**Action:**
- `git` hacks: file-not-a-repo → `ErrNotFound`, bad ref → `ErrInvalidInput`, auth issues → `ErrPermission`.
- `gh` wrapper: API errors → `ErrIO`, 401/403 → `ErrPermission`, 404 → `ErrNotFound`, invalid flag → `ErrInvalidInput`.
- `kubectl`: delegates to `k8s.io/kubectl`; wrap whatever it returns as `fmt.Errorf("kubectl: %w", err)` passthrough unless the error matches a cmderr sentinel already (Pitfall 3 — avoid double-wrap).
- `kubehacks`: shortcuts; classify invalid context/namespace → `ErrInvalidInput`, API errors → `ErrIO`.
- **Flag any code that calls `os/exec.Command` or `syscall.Exec`** and record in `EXIT-CODE-CHANGES.md` as a pre-existing CLAUDE.md violation for backlog.

**Verify:**
```
<automated>go test -race ./internal/cli/git/... ./internal/cli/gh/... ./internal/cli/kubectl/... ./internal/cli/kubehacks/...</automated>
```

**Done:** Classified; exec-spawn audit captured.

### Task 2: Migrate terraform + aws + cloud + vault

**Files:** identical pattern for each.

**Action:**
- `terraform`: tfstate file-not-found → `ErrNotFound`; validate failures → `ErrInvalidInput`; apply errors → `ErrIO`.
- `aws`: API errors → `ErrIO`, 403 → `ErrPermission`, 404 → `ErrNotFound`, invalid flag → `ErrInvalidInput`.
- `cloud` (multi-cloud aggregator): delegate-classify (pass-through per Pitfall 3).
- `vault`: auth failures → `ErrPermission`, not-found → `ErrNotFound`, invalid path → `ErrInvalidInput`.

**Verify:**
```
<automated>go test -race ./internal/cli/terraform/... ./internal/cli/aws/... ./internal/cli/cloud/... ./internal/cli/vault/...</automated>
```

**Done:** Classified.

### Task 3: Golden error snapshots (one per command = 8)

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

```yaml
- name: git_not_a_repo
  args: ["git", "status"]
  fixtures_dir: "empty"
  exit_code: 1
- name: gh_invalid_subcommand
  args: ["gh", "not-a-real-subcommand"]
  exit_code: 2
- name: kubectl_invalid_flag
  args: ["kubectl", "get", "--definitely-not-a-flag"]
  exit_code: 2
- name: kubehacks_invalid_context
  args: ["kubehacks", "ctx", "not-a-real-context-xyz"]
  exit_code: 2
- name: terraform_no_tfstate
  args: ["terraform", "state", "list"]
  fixtures_dir: "empty"
  exit_code: 1
- name: aws_invalid_flag
  args: ["aws", "--definitely-not-a-flag"]
  exit_code: 2
- name: cloud_invalid_provider
  args: ["cloud", "--provider", "notacloud"]
  exit_code: 2
- name: vault_path_not_found
  args: ["vault", "read", "nonexistent/path/xyz"]
  exit_code: 1
```

**Verify:**
```
<automated>task test:golden -- --filter 'git_|gh_|kubectl_|kubehacks_|terraform_|aws_|cloud_|vault_'</automated>
```

**Done:** Snapshots green (some may require `fixtures_dir` support; skip with comment if harness can't).

### Task 4: Log exit-code changes + exec-spawn audit

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

## Golden test additions

8 snapshots listed above.

## Verification

```bash
go test -race ./internal/cli/git/... ./internal/cli/gh/... ./internal/cli/kubectl/... ./internal/cli/kubehacks/... ./internal/cli/terraform/... ./internal/cli/aws/... ./internal/cli/cloud/... ./internal/cli/vault/...
task test:golden -- --filter 'git_|gh_|kubectl_|kubehacks_|terraform_|aws_|cloud_|vault_'
task lint:cmderr-coverage
```

## Out of scope

- Actually fixing any pre-existing exec-spawn violations (defer to backlog)
- scaffolding/project/repo (Plan 13)
- dev-tools (Plan 14); misc-tail (Plan 15)
