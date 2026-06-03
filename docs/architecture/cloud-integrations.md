<!-- Extracted from CLAUDE.md leanness pass, 2026-05-24 -->
# Cloud & DevOps Integrations

### Kubernetes (kubectl / k)

Full kubectl integration via `k8s.io/kubectl` package using local source code.

```bash
omni kubectl get pods          # or: omni k get pods
omni k get pods -A
omni k describe node mynode
omni k logs -f mypod
omni k apply -f manifest.yaml
```

**Source:** `B:\shared\personal\repos\kubernetes\kubectl` (local replace directive)

### Terraform (terraform / tf)

Terraform CLI wrapper for infrastructure management.

```bash
omni tf init
omni tf plan -out=plan.tfplan
omni tf apply -auto-approve
omni tf destroy
omni tf state list
omni tf workspace select prod
```

**Commands:** init, plan, apply, destroy, validate, fmt, output, state (list/show/mv/rm), workspace (list/new/select/delete), import, taint, untaint, refresh, graph, console, providers, get, test, show, version

### Git Hacks

Shortcuts for common Git operations.

| Command | Alias | Description |
|---------|-------|-------------|
| `omni git quick-commit -m "msg"` | `omni gqc -m "msg"` | Stage all + commit |
| `omni git branch-clean` | `omni gbc` | Delete merged branches |
| `omni git undo` | - | Soft reset HEAD~1 |
| `omni git amend` | - | Amend --no-edit |
| `omni git stash-staged` | - | Stash staged only |
| `omni git log-graph` | `omni git lg` | Pretty log with graph |
| `omni git diff-words` | - | Word-level diff |
| `omni git blame-line` | - | Blame line range |
| `omni git status` | `omni git st` | Short status |
| `omni git push` | - | Push (--force-with-lease) |
| `omni git pull-rebase` | `omni git pr` | Pull --rebase |
| `omni git fetch-all` | `omni git fa` | Fetch --all --prune |

### Kubectl Hacks

Shortcuts for common Kubernetes operations.

| Command | Description |
|---------|-------------|
| `omni kga` | Get all resources (pods, svc, deploy, etc.) |
| `omni klf <pod>` | Follow logs with timestamps |
| `omni keb <pod>` | Exec bash into pod (falls back to sh) |
| `omni kpf <target> <local:remote>` | Port forward |
| `omni kdp <selector>` | Delete pods by selector |
| `omni krr <deployment>` | Rollout restart deployment |
| `omni kge` | Get events sorted by time |
| `omni ktp` | Top pods by resource |
| `omni ktn` | Top nodes by resource |
| `omni kcs [context]` | Switch/list contexts |
| `omni kns [namespace]` | Switch/list namespaces |
| `omni kwp` | Watch pods continuously |
| `omni kscale <deploy> <n>` | Scale deployment |
| `omni kdebug <pod>` | Debug with ephemeral container |
| `omni kdrain <node>` | Drain node for maintenance |
| `omni krun <name> --image=<img>` | Run one-off pod |
| `omni kconfig` | Show kubeconfig info |

---

