# Golden Master Testing

Golden master (snapshot) testing captures exact command outputs as baselines and detects regressions by comparing future runs against them.

## Motivation

omni has 148+ CLI commands. When command output changes, golden master tests detect whether the change was intentional or a regression, without writing individual assertions for each output format.

## How It Works

```
YAML Registry → Discovery → Execution → Normalization → Comparison → Report
                                                              ↓
                                                    SHA-256 fast path
```

1. **Discovery**: Parse `golden_tests.yaml` for test definitions (args, stdin, fixtures)
2. **Execution**: Run omni binary with each test case's arguments
3. **Normalization**: Apply normalizers (strip paths, temp dirs, CRLF)
4. **Comparison**: SHA-256 fast path, then unified diff on mismatch
5. **Report**: Color-coded PASS/FAIL with diff output

## Test Coverage

| Category | Tests | Input Type |
|----------|-------|-----------|
| encoding | 13 | stdin |
| hashing | 8 | stdin + file |
| text | 15 | stdin |
| text_with_files | 10 | file |
| data | 9 | stdin + file |
| format | 8 | stdin |
| utils | 9 | args + stdin |
| security | 1 | stdin |
| xxd | 3 | file + stdin |
| strings | 1 | file |
| case_conv | 4 | args |
| **Total** | **81** | |

## Quick Start

```bash
# Compare current output against baselines
task golden:compare

# Record baselines after intentional changes
task golden:record

# List all test cases
task golden:list
```

## Commands

### record
Run all test cases and save output as golden master files.
```bash
task golden:record
```

### compare
Run all test cases and diff against stored baselines.
```bash
task golden:compare
```

### list
Show all registered test cases without execution.
```bash
task golden:list
```

### update
Re-record test cases matching a pattern.
```bash
cd tools/golden && PYTHONPATH=src python -m golden --pattern base64 update
```

### map
Generate a test map for analysis.
```bash
task golden:map:table    # Human-readable table
task golden:map          # JSON output to test-map.json
```

## Filtering

```bash
# By category
cd tools/golden && PYTHONPATH=src python -m golden --category encoding compare

# By name pattern
cd tools/golden && PYTHONPATH=src python -m golden --pattern sort compare
```

## Docker

Reproducible testing without local Go/Python:

```bash
task golden:docker:build    # Build image (multi-stage: Go build + Python runner)
task golden:docker:record   # Record baselines in container
task golden:docker:compare  # Compare in container
```

## Two Systems

omni has two golden test systems that share the same YAML registry:

| System | Location | Purpose |
|--------|----------|---------|
| **testing/golden/** | `testing/scripts/test_golden.py` | Lightweight, auto-discovered by `run_all.py`, integrated with existing black-box tests |
| **tools/golden/** | `tools/golden/src/golden/` | Full-featured: SHA-256 manifest, parallel execution, test mapping, Docker support |

Both use `golden_tests.yaml` as the source of truth.

## Directory Layout

```
tools/golden/
├── src/golden/
│   ├── cli.py          # CLI entry point (record, compare, list, update, map)
│   ├── config.py       # Project-specific constants
│   ├── types.py        # TestCase, RunResult, CompareResult
│   ├── discovery.py    # YAML registry parser
│   ├── runner.py       # Binary execution with ProcessPoolExecutor
│   ├── normalize.py    # Output normalizers
│   ├── manifest.py     # SHA-256 manifest (JSON with hashes)
│   ├── recorder.py     # Golden master file writer
│   ├── comparator.py   # SHA-256 fast path + unified diff
│   ├── report.py       # Color-coded terminal output
│   ├── mapper.py       # Test map (JSON/table)
│   └── incremental.py  # Change detection
├── golden_tests.yaml   # Test registry (shared with testing/golden/)
├── golden_masters/     # Recorded baselines (gitignored)
├── Dockerfile          # Multi-stage build
├── pyproject.toml      # Python project metadata
└── tests/              # Unit tests for the framework
```

## Workflow

### After code changes
```bash
task golden:compare     # Check for regressions
```

### After intentional output changes
```bash
task golden:record      # Regenerate baselines
```

### Adding tests for new commands
1. Add entry to `tools/golden/golden_tests.yaml` (and `testing/golden/golden_tests.yaml`)
2. Run `task golden:record` to generate baseline
3. Run `task golden:compare` to verify
4. Commit YAML and snapshot files

### Before PRs
```bash
task golden:compare     # Ensure no regressions
```

## Taskfile Reference

| Task | Description |
|------|-------------|
| `task golden:record` | Record golden master baselines |
| `task golden:compare` | Compare against baselines |
| `task golden:list` | List test cases |
| `task golden:map` | Generate test map (JSON) |
| `task golden:map:table` | Generate test map (table) |
| `task golden:docker:build` | Build Docker image |
| `task golden:docker:record` | Record via Docker |
| `task golden:docker:compare` | Compare via Docker |
| `task test:golden` | Lightweight verify (testing/golden/) |
| `task test:golden:update` | Lightweight update (testing/golden/) |
| `task test:golden:check` | Verify snapshots exist (CI) |

## Design Decisions

- **YAML registry**: Declarative test definitions allow adding tests without code changes
- **SHA-256 fast path**: Skip diff computation when hashes match (manifest.json)
- **Normalizers**: Strip non-deterministic paths and temp dirs before comparison
- **Sidecar files**: `.stdout` plaintext files enable readable git diffs
- **No `jq keys` test**: Excluded because Go map iteration order is non-deterministic
- **Parallel execution**: `ProcessPoolExecutor` with configurable workers
- **Two systems**: Lightweight (testing/) for CI auto-discovery + full-featured (tools/) for development
