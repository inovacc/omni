<!-- Extracted from CLAUDE.md leanness pass, 2026-05-24 -->
# Testing

### Run Tests

```bash
# Run all tests with race detection and coverage
go test -race -cover ./...

# Run specific package tests
go test -race -v ./internal/cli/pipe/...

# Run specific test
go test -race ./internal/cli/... -run TestSubstituteVariables -v

# Generate coverage report
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests
task test:integration
```

### Test Coverage

Current coverage: ~25.8% overall (~78% avg for omni-owned pkg/ packages; total skewed by vendored buf packages)

### Test Files

| File | Tests |
|------|-------|
| `internal/cli/hash/hash_test.go` | SHA256, SHA512, MD5, CRC32, CRC64 hashing |
| `internal/cli/base/base_test.go` | Base64, Base32, Base58 encoding |
| `internal/cli/uuid/uuid_test.go` | UUID generation |
| `internal/cli/random/random_test.go` | Random string, hex, password, int |
| `internal/cli/fs/fs_test.go` | mkdir, cp, touch, stat, ln, readlink |
| `internal/cli/text/text_test.go` | tr, cut, nl, uniq, paste, tac, fold, column, join |
| `internal/cli/kill/kill_test.go` | Signal listing, PID validation |
| `internal/cli/jq/jq_test.go` | JSON querying |
| `internal/cli/yq/yq_test.go` | YAML querying |
| `internal/cli/diff/diff_test.go` | File comparison |
| `internal/cli/crypt/crypt_test.go` | Encrypt/decrypt |
| `internal/cli/path/path_test.go` | dirname, basename, realpath, clean, abs |
| `internal/cli/pipe/pipe_test.go` | Command parsing, variable substitution |
| `internal/cli/xxd/xxd_test.go` | Hex dump, reverse, plain, include, bits modes |
| `internal/cli/rg/rg_test.go` | Ripgrep search, parallel walking, streaming JSON, gitignore integration |
| `internal/cli/rg/gitignore_test.go` | Gitignore pattern parsing, negation, directory-only, double globs |
| `pkg/idgen/idgen_test.go` | UUID v4/v7, ULID, KSUID, Nanoid, Snowflake generation |
| `pkg/hashutil/hashutil_test.go` | HashFile, HashReader, HashString, HashBytes, CRC32, CRC64 |
| `pkg/jsonutil/jsonutil_test.go` | Query, QueryString, QueryReader, ApplyFilter |
| `pkg/encoding/encoding_test.go` | Base64/32/58 encode/decode, WrapString |
| `pkg/cryptutil/cryptutil_test.go` | Encrypt/decrypt roundtrip, key generation/derivation |
| `pkg/sqlfmt/sqlfmt_test.go` | Format, Minify, Validate, Tokenize, IsKeyword |
| `pkg/cssfmt/cssfmt_test.go` | Format, Minify, Validate, RemoveComments, ParseDeclarations |
| `pkg/htmlfmt/htmlfmt_test.go` | Format, Minify, Validate, IsSelfClosing, CollapseWhitespace |
| `pkg/textutil/textutil_test.go` | Sort, SortLines, UniqueConsecutive, Uniq, TrimLines |
| `pkg/textutil/diff/diff_test.go` | ComputeDiff, FormatUnified, CompareJSON, CompareJSONBytes |
| `pkg/search/grep/grep_test.go` | Search, SearchWithOptions, CompilePattern |
| `pkg/search/rg/rg_test.go` | ParsePattern, GitignoreSet, MatchesFileType, MatchesGlob, IsBinary |
| `pkg/twig/scanner/scanner_test.go` | Directory scanning, MaxFiles, MaxHashSize, parallel scanning, progress callback |
| `pkg/twig/formatter/formatter_test.go` | ASCII tree, JSON, flattened hash, NDJSON streaming |
| `pkg/twig/comparer/comparer_test.go` | Tree comparison: added, removed, modified, moved, detect-moves |
| `pkg/video/format/format_test.go` | Format sorting, HasVideo, HasAudio, resolution |
| `pkg/video/format/selector_test.go` | Format selector parsing ("best", "worst", filter expressions) |
| `pkg/video/m3u8/parser_test.go` | M3U8 manifest parsing (master/media playlists, segments, keys) |
| `pkg/video/utils/*_test.go` | Sanitize, HTML, URL, parse, traverse tests |
| `pkg/video/extractor/registry_test.go` | Extractor registration and matching |
| `pkg/video/extractor/youtube/channel_test.go` | Channel URL matching, duration/view count parsing |
| `internal/cli/video/channel_test.go` | Channel DB init, schema, incremental tracking, upsert |
| `testing/scripts/test_video.py` | Black-box: omni video vs yt-dlp comparison (Docker) |
| `pkg/pipeline/pipeline_test.go` | Orchestrator, multi-stage, context cancel, head drain |
| `pkg/pipeline/parse_test.go` | CLI string parser, sed expressions, flag combinations |
| `internal/cli/pipeline/pipeline_test.go` | CLI wrapper integration tests |
| `internal/cli/exist/exist_test.go` | File, dir, path, command, env, process, port existence checks, JSON/quiet modes |
| `internal/cli/project/project_test.go` | Project detection, deps parsing, health scoring, output formatting |
| `internal/cli/command/command_test.go` | Command interface, Registry (register, get, names, concurrency), adapters |
| `internal/cli/repo/repo_test.go` | Repo analyze: path resolution, remote detection, tree, key files, entry points, architecture, sections, JSON/Markdown output |
| `pkg/video/downloader/progress_test.go` | SpeedTracker, FormatBytes, FormatSpeed, FormatETA, FormatPercent |
| `pkg/video/downloader/fragment_test.go` | SaveFragmentState/LoadFragmentState roundtrip, RemoveFragmentState, AppendToFile |
| `pkg/video/downloader/downloader_test.go` | SelectDownloader type assertions for all protocols |
| `pkg/video/nethttp/cookies_test.go` | LoadNetscapeCookies parsing, roundtrip, malformed lines, nonexistent file |
| `pkg/video/nethttp/sapisidhash_test.go` | ComputeSAPISIDHash format, ExtractSAPISID from cookie jar |
| `pkg/video/extractor/helpers_test.go` | SearchRegex, ParseJSON, ParseM3U8Formats |
| `pkg/video/options_test.go` | applyOptions defaults, With* option composition |

### Golden Master Tests

Characterization tests that capture exact command outputs as snapshots and detect regressions.

**MUST run on every feature/code change:**

```bash
# Verify outputs match snapshots
task test:golden

# After intentional output changes, regenerate snapshots
task test:golden:update

# Update specific category only
task test:golden:update -- --update encoding

# CI pre-flight: verify all snapshot files exist
task test:golden:check

# List all registered tests
python testing/scripts/test_golden.py --list

# Run with verbose diffs
python testing/scripts/test_golden.py --verbose
```

**Adding golden tests for new commands:**
1. Add test case to `testing/golden/golden_tests.yaml` AND `tools/golden/golden_tests.yaml`
2. Run `task test:golden:update` and `task golden:record` to generate snapshots
3. Review the generated `.stdout` files
4. Commit both the YAML entry and snapshot files

**Full-featured system (tools/golden/):**
```bash
task golden:compare             # Compare with SHA-256 fast path
task golden:record              # Record baselines with manifest
task golden:list                # List test cases
task golden:map:table           # Test map as table
task golden:docker:build        # Build Docker image
task golden:docker:compare      # Compare in container
task golden:docker:record       # Record in container
```

**Lightweight system (testing/golden/):**
```bash
task docker:test:golden          # Verify in Linux container
task docker:test:golden:update   # Regenerate and persist to host via volume mount
```

**Structure:**
- `testing/golden_engine.py` — Lightweight engine (auto-discovered by `run_all.py`)
- `testing/golden/golden_tests.yaml` — Declarative test registry (117 tests, 13 categories)
- `testing/golden/snapshots/` — JSON metadata + .stdout sidecars
- `tools/golden/src/golden/` — Full engine (11 modules: manifest, parallel, map, Docker)
- `tools/golden/golden_tests.yaml` — Shared registry (same tests)
- `tools/golden/golden_masters/` — Baselines with SHA-256 manifest (gitignored)
- See `docs/GOLDEN_MASTER_TESTING.md` for full documentation

### Test Pattern

```go
func TestRun(t *testing.T) {
    var buf bytes.Buffer
    err := ls.Run(&buf, []string{"."}, ls.Options{})
    if err != nil {
        t.Fatal(err)
    }
    // Assert on buf.String()
}
```

### Linter Directives

Use `//nolint` for intentional patterns:

```go
//nolint:dupword // intentional duplicate content for testing uniq
content := "a\na\na\nb\nb\n"
```

---

