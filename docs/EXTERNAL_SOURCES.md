# External Source Repositories

This document tracks external repositories used as source code for omni features.
Monitor these repos for updates, security patches, and new features.

---

## Active Integrations

### Kubernetes

| Component | Local Path | Module Path | Feature |
|-----------|-----------|-------------|---------|
| kubectl | `B:\shared\personal\repos\kubernetes\kubectl` | `k8s.io/kubectl` | `omni kubectl` / `omni k` |
| client-go | `B:\shared\personal\repos\kubernetes\client-go` | `k8s.io/client-go` | Kubernetes API client |

**go.mod replace directives:**
```go
replace (
    k8s.io/kubectl => B:\shared\personal\repos\kubernetes\kubectl
    k8s.io/client-go => B:\shared\personal\repos\kubernetes\client-go
)
```

**Upstream repos:**
- https://github.com/kubernetes/kubectl
- https://github.com/kubernetes/client-go

**Key files to monitor:**
- `kubectl/pkg/cmd/cmd.go` - Main command registration
- `kubectl/pkg/cmd/get/get.go` - Get command implementation
- `kubectl/pkg/cmd/apply/apply.go` - Apply command implementation
- `client-go/tools/clientcmd/` - Kubeconfig handling

---

## Planned Integrations

### HashiCorp Terraform

| Component | Local Path | Feature |
|-----------|-----------|---------|
| terraform | `B:\shared\personal\repos\hashicorp\terraform` | `omni terraform` |

**Upstream repo:** https://github.com/hashicorp/terraform

**Key directories:**
- `internal/command/` - CLI command implementations
- `internal/configs/` - Configuration parsing
- `internal/terraform/` - Core engine
- `internal/states/` - State management

**Commands to implement:**
- init, plan, apply, destroy, validate, fmt, output, state, workspace

---

### HashiCorp Vault

| Component | Local Path | Feature |
|-----------|-----------|---------|
| vault | `B:\shared\personal\repos\hashicorp\vault` | `omni vault` |

**Upstream repo:** https://github.com/hashicorp/vault

**Key directories:**
- `command/` - CLI commands
- `api/` - API client
- `vault/` - Core vault logic

**Commands to implement:**
- login, read, write, list, kv get, kv put

---

### HashiCorp Consul

| Component | Local Path | Feature |
|-----------|-----------|---------|
| consul | `B:\shared\personal\repos\hashicorp\consul` | `omni consul` |

**Upstream repo:** https://github.com/hashicorp/consul

**Key directories:**
- `command/` - CLI commands
- `api/` - API client

**Commands to implement:**
- members, kv get, kv put, services

---

### HashiCorp Nomad

| Component | Local Path | Feature |
|-----------|-----------|---------|
| nomad | `B:\shared\personal\repos\hashicorp\nomad` | `omni nomad` |

**Upstream repo:** https://github.com/hashicorp/nomad

**Key directories:**
- `command/` - CLI commands
- `api/` - API client

**Commands to implement:**
- job run, job stop, node status, alloc status

---

### HashiCorp Packer

| Component | Local Path | Feature |
|-----------|-----------|---------|
| packer | `B:\shared\personal\repos\hashicorp\packer` | `omni packer` |

**Upstream repo:** https://github.com/hashicorp/packer

**Key directories:**
- `command/` - CLI commands
- `packer/` - Core packer logic

**Commands to implement:**
- build, validate, fmt, inspect

---

## Monitoring Commands

### Check local repo status
```bash
# Kubernetes repos
cd B:\shared\personal\repos\kubernetes\kubectl && git log -1 --oneline
cd B:\shared\personal\repos\kubernetes\client-go && git log -1 --oneline

# HashiCorp repos
cd B:\shared\personal\repos\hashicorp\terraform && git log -1 --oneline
cd B:\shared\personal\repos\hashicorp\vault && git log -1 --oneline
cd B:\shared\personal\repos\hashicorp\consul && git log -1 --oneline
cd B:\shared\personal\repos\hashicorp\nomad && git log -1 --oneline
cd B:\shared\personal\repos\hashicorp\packer && git log -1 --oneline
```

### Sync with upstream
```bash
# Example for kubectl
cd B:\shared\personal\repos\kubernetes\kubectl
git fetch origin
git log HEAD..origin/master --oneline  # See new commits
git pull origin master                  # Apply updates
```

### Check go.mod versions
```bash
cd B:\shared\personal\GolandProjects\omni
grep "k8s.io" go.mod
grep "hashicorp" go.mod
```

---

## Version Tracking

| Integration | Local Commit | Last Synced | Notes |
|-------------|--------------|-------------|-------|
| kubectl | `d0855f97` | 2026-02-04 | Initial integration |
| client-go | `aa31c74d1` | 2026-02-04 | Initial integration |
| terraform | - | - | Planned |
| vault | - | - | Planned |
| consul | - | - | Planned |
| nomad | - | - | Planned |
| packer | - | - | Planned |

---

## Update Checklist

When syncing external repos:

1. [ ] Check upstream releases/changelog for breaking changes
2. [ ] Pull latest changes to local repo
3. [ ] Run `go mod tidy` in omni
4. [ ] Build and test: `go build ./... && go test ./...`
5. [ ] Test affected commands manually
6. [ ] Update version tracking table above
7. [ ] Commit changes with sync note

---

## Security Monitoring

Subscribe to security advisories:
- https://github.com/kubernetes/kubernetes/security/advisories
- https://github.com/hashicorp/terraform/security/advisories
- https://github.com/hashicorp/vault/security/advisories

Check for CVEs:
```bash
govulncheck ./...
```
