# Contributors

## Maintainers

This project is maintained by [Buf Technologies, Inc.](https://buf.build)

Top contributors by commit count:
- bufdev (751 commits)
- Doria Keung (260 commits)
- Edward McFarlane (196 commits)
- Philip K. Warren (142 commits)
- Joshua Humphries (91 commits)
- Stefan VanBuren (76 commits)
- Alex McKinney (59 commits)
- Julian Figueroa (53 commits)

## Contributing

### Setup

```bash
git clone https://github.com/bufbuild/buf.git
cd buf
make install
```

### Development Workflow

1. Create a feature branch
2. Make changes following the project conventions (see `CLAUDE.md`)
3. Run linters: `make lint`
4. Run tests: `make test`
5. Format code: `make gofmtmodtidy`
6. Submit a pull request

### Conventions

- **Commits:** Conventional commit messages (e.g., `feat:`, `fix:`, `refactor:`)
- **License:** Apache 2.0 headers on all Go files (2020-2025)
- **Linting:** No `//nolint` directives allowed; exceptions go in `.golangci.yml`
- **Dependencies:** Use `github.com/bufbuild/buf/internal/thread.Parallelize` instead of `errgroup`, `internal/pkg/protoencoding` instead of `proto.Marshal`, `internal/pkg/osext` instead of `os.Getwd`
- **Testing:** Table-driven tests preferred; run `make installtest` before full test suite
