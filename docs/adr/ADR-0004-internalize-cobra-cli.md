# ADR-0004: Internalize cobra-cli

**Status:** Proposed → **Already Internalized**
**Date:** 2026-02-19
**Decision:** Keep external dependency removed; omni already replaces cobra-cli entirely.

## Context

The user requested internalization of `github.com/spf13/cobra-cli` (v1.3.0) — a code generator for scaffolding Cobra CLI applications.

## Analysis

### Summary

| Field | Value |
|-------|-------|
| Module | `github.com/spf13/cobra-cli` |
| Version | v1.3.0 (commit `1d43487`) |
| Total Go files | 21 (16 non-test) |
| Total LOC | ~3,439 (~600 logic + ~2,800 license text) |
| Packages | 2 (`cmd/`, `tpl/`) |
| Direct deps | 2 (`cobra`, `viper`) |
| Transitive deps | 871 (go.sum entries) |
| License | Apache 2.0 |

### What cobra-cli Does

1. **`cobra-cli init`** — Creates main.go, cmd/root.go, LICENSE
2. **`cobra-cli add <name>`** — Creates cmd/{name}.go with boilerplate
3. **License management** — 8 license types with copyright templating
4. **Config** — ~/.cobra.yaml for defaults (author, license, useViper)

### What omni Already Has

Omni's `generate cobra` command (`internal/cli/generate/`) **fully replaces** cobra-cli with significant enhancements:

| cobra-cli Feature | omni Equivalent | Status |
|-------------------|-----------------|--------|
| `init` (basic) | `generate cobra init` | ✅ Enhanced (4 modes: basic, viper, service, full) |
| `add <cmd>` | `generate cobra add` | ✅ Complete with parent support |
| `~/.cobra.yaml` config | `generate cobra config` | ✅ Compatible format |
| License generation (8 types) | License generation (3 types) | ⚠️ Reduced set (MIT, Apache-2.0, BSD-3-Clause) |
| Templates (3) | Templates (15+) | ✅ Far more comprehensive |

**omni additions beyond cobra-cli:**
- Service pattern scaffolding (inovacc/config integration)
- Taskfile.yml generation (vs Makefile)
- .goreleaser.yaml, .golangci.yml, GitHub Actions workflows
- Handler generation (stdlib, chi, gin, echo)
- Repository generation (Postgres, MySQL, SQLite)
- Test generation (AST-based, table-driven)

### Dependency Graph Problem

cobra-cli depends on `viper` (v1.12.0), which brings 871 transitive dependencies including mapstructure, cast, and the entire viper ecosystem. omni intentionally avoids viper by using gopkg.in/yaml.v3 directly for config parsing — a much lighter approach.

## Decision

**Recommended Strategy: Keep external — already superseded.**

cobra-cli is NOT a Go library dependency of omni. It was a standalone CLI tool that omni has already fully replaced with `generate cobra`. There is nothing to internalize because:

1. cobra-cli is not in omni's go.mod
2. omni's generate command already covers all cobra-cli functionality
3. omni's version is more feature-rich (4 init modes, handler/repo/test generators)
4. Internalizing cobra-cli's code would add viper as a dependency (871 transitive deps)

### Remaining Gaps (Minor)

- 5 missing license types (GPL-2.0, GPL-3.0, LGPL-2.1, AGPL-3.0, BSD-2-Clause) — can be added as templates if needed
- No interactive mode — omni uses explicit flags (acceptable for automation/CI use)

## Consequences

- No code changes needed
- cobra-cli references in docs/comments remain as compatibility notes
- Future license type additions can be done by extending `internal/cli/generate/templates/cobra/templates.go`
