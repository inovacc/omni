# Contributing to omni

Thanks for your interest. **omni** is a cross-platform, Go-native replacement for
common shell utilities, built primarily for the author's CI/CD pipelines. Broader
adoption is welcome, but design decisions optimize for **determinism and a
dependency-light single static binary** — not general-purpose breadth.

## Non-negotiable invariants

Every change must preserve these (they are the project's core value):

1. **No `os/exec`** in utility implementations. The binary spawns no external
   processes. (A small set of sanctioned exec wrappers — `exec`, `forloop`, `task`,
   `terraform`, git/gh hacks, `repo`, `buf generate` — is documented in
   `docs/architecture/patterns.md`; do not add new ones without an ADR.)
2. **Stdlib-first, pure-Go, no CGO.** New third-party dependencies are a high bar.
   A build tag does **not** keep a heavy dep out of `go.mod` (MVS pulls it in
   anyway) — isolate heavy/optional deps into a separate `contrib/` module
   (see `contrib/sigstore-verify`). See ADR-0007.
3. **Cross-platform** via `//go:build` tags (`_unix.go` / `_windows.go` / `_darwin.go`),
   never runtime `if runtime.GOOS == ...` branches.
4. **Errors** use `internal/cli/cmderr` sentinels (`ErrNotFound`=1, `ErrInvalidInput`=2,
   `ErrPermission`=3, `ErrIO`=4, `ErrTimeout`=5, `ErrUnsupported`=6, `ErrConflict`=1)
   via `cmderr.Wrap` + `errors.Is`/`As` (never `==`). See `docs/EXIT-CODES.md`.
5. **`io.Writer` / `io.Reader`** for all I/O (testability). `cmd/` stays thin;
   logic lives in `internal/cli/<cmd>/` and reusable `pkg/<domain>/`.

## Development

```bash
go run . <cmd> --flags          # run (never build-then-run for dev)
go test -race -cover ./...      # unit tests
task lint                       # golangci-lint (govet); must be 0 issues
task test:golden                # golden-master snapshots
task freeze:check               # frozen pkg/* public API matches docs/API-FREEZE.md
```

(`task build` stamps a deterministic `dev` version; release builds stamp the tag.)

## Adding a command

1. `internal/cli/<cmd>/<cmd>.go` (Options struct + `Run(w io.Writer, ...) error`) + tests.
2. `cmd/<cmd>.go` (thin Cobra wrapper, self-wired in `init()`).
3. Golden cases in **both** `testing/golden/golden_tests.yaml` and
   `tools/golden/golden_tests.yaml`; record with `task test:golden:update` +
   `task golden:record`. Fixtures that are hashed/signed go under
   `testing/golden/fixtures/` (`-text`-locked via `.gitattributes`).
4. Update `docs/COMMANDS.md` + `CHANGELOG.md`.

## Conventions

- **Commits:** Conventional Commits (`feat:`, `fix:`, `docs:`, `test:`, `build:`,
  `refactor:`, `chore:`). No AI attribution lines.
- **Breaking changes** to a frozen `pkg/*` package follow the 30-day deprecation
  protocol (`// Deprecated: … Will be removed after YYYY-MM-DD.`, slog warning,
  `docs/BACKLOG.md` `DEPRECATION` tag, separate cleanup commit). `// Experimental:`
  packages are exempt.
- **Architecture decisions** that change an invariant or add a dependency need an
  ADR in `docs/adr/` (see existing ADR-0001…0010).
- **Security:** no committed secrets; parameterized queries; never log/flag key
  material (passphrases via env only).

## License

BSD 3-Clause (see `LICENSE`). By contributing you agree your contributions are
licensed under the same terms.
