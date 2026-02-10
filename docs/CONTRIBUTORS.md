# Contributors

## Maintainers

| Name | GitHub | Role |
|------|--------|------|
| inovacc | [@inovacc](https://github.com/inovacc) | Creator & Maintainer |

---

## Contributing

We welcome contributions! Here's how to get started.

### Prerequisites

- Go 1.25+
- [Task](https://taskfile.dev/) (task runner)
- [golangci-lint](https://golangci-lint.run/) (linting)

### Setup

```bash
git clone https://github.com/inovacc/omni.git
cd omni
go mod download
task build
```

### Development Workflow

```bash
# Run tests
task test

# Run linter with auto-fix
task lint

# Run all quality checks (fmt, vet, lint, test)
task check

# Build binary
task build

# Run black-box tests
task test:blackbox
```

### Adding a New Command

1. Create `internal/cli/newcmd/newcmd.go` with Options struct and Run function
2. Create `internal/cli/newcmd/newcmd_test.go` with table-driven tests
3. Create `cmd/newcmd.go` with thin Cobra wrapper
4. Register the command in `init()` via `rootCmd.AddCommand()`
5. Add black-box tests to `testing/scripts/`
6. Update CLAUDE.md command categories

### Code Conventions

- **io.Writer pattern**: All Run functions accept `io.Writer` for testability
- **Options struct**: All command options collected in a struct
- **Thin cmd/ layer**: cmd/ files only parse flags and delegate to internal/cli/
- **pkg/ for libraries**: Reusable logic lives in pkg/ with functional options
- **Table-driven tests**: Use subtests with test case structs
- **Error wrapping**: Always `fmt.Errorf("command: %w", err)`
- **Muted returns**: `_, _ = fmt.Fprintln(w, output)`
- **Defers**: `defer func() { _ = file.Close() }()`

### Commit Convention

Follow conventional commits:
```
feat(scope): add new feature
fix(scope): fix bug description
refactor(scope): restructure code
docs: update documentation
test: add or update tests
```

### Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure `task check` passes
5. Submit a PR with clear description

### Platform-Specific Code

If your change needs OS-specific behavior:
- Create `_unix.go` and `_windows.go` variants
- Use `//go:build unix` and `//go:build windows` tags
- Ask before splitting by OS if unsure

---

## Code of Conduct

Be respectful and constructive. Focus on the code, not the person.
