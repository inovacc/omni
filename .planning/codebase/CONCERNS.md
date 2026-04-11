# Codebase Concerns

**Analysis Date:** 2026-04-11

## Tech Debt

### Command Error Handling Migration (P1 - High Priority)

**Issue:** ~76 of 160+ commands still use raw `fmt.Errorf()` instead of `cmderr` sentinels for structured exit code classification.

**Files affected:** `internal/cli/{data,compress,system,flow,cloud}/` (data operations: jq, yq, json, yaml, csv, xml; compression: gzip, bzip2, xz; system: ps, lsof, uptime, free, df, du; flow: xargs, watch, yes, pipeline; cloud: aws, kubectl, terraform, vault)

**Current status:** 84 commands adopted (batches 1-8 complete as of Mar 2026). Remaining ~76 commands (mostly in data, compress, system, flow, cloud categories).

**Impact:** Without cmderr adoption:
- Exit codes are non-standard (falls back to OS defaults)
- Error types not classifiable for scripting/CI pipelines
- Users can't distinguish between file-not-found (1) vs permission-denied (3) vs invalid-input (2)
- Inconsistent behavior across omni commands

**Fix approach:**
1. Identify remaining commands by grep for `fmt.Errorf` patterns
2. Create cmderr batch 9: migrate data category (jq, yq, json, yaml, csv, xml)
3. Create cmderr batch 10: migrate compress category (gzip, bzip2, xz)
4. Create cmderr batch 11+: migrate system, flow, cloud categories
5. Add CI lint rule: forbid raw `fmt.Errorf` for new code

### Large Untested Packages

**Issue:** Video extractor packages have minimal or no test coverage.

**Files:**
- `pkg/video/extractor/youtube/youtube.go` (478 LOC) - **No dedicated tests**
- `pkg/video/extractor/youtube/channel.go` (407 LOC) - **No dedicated tests**
- `pkg/video/extractor/generic/generic.go` (210 LOC) - **No dedicated tests**
- `pkg/video/extractor/youtube/signature.go` (263 LOC) - **Depends on JS runtime (goja)**

**Impact:** 
- YouTube signature decryption changes silently fail
- InnerTube API changes break channel downloads
- Generic fallback extraction logic untested

**Fix approach:**
1. Add unit tests for signature parsing (mock JS responses)
2. Add tests for channel extraction (fixture-based YouTube responses)
3. Add tests for generic extractor fallback cases
4. Mock HTTP responses for all extractor tests (no live YouTube calls)

### Buf Package Vendoring Performance

**Issue:** `go test ./...` is slow due to vendored `github.com/bufbuild/protocompile` causing full recompilation.

**Files:** `vendor/github.com/bufbuild/protocompile/` (large Go package)

**Impact:** Local test runs and CI pipeline slow significantly. Developers avoid running full test suite.

**Fix approach:**
1. Use `go test -short` flag in development (skip buf tests)
2. Add separate `task test:buf` target for buf-specific tests
3. Consider extracting protocompile as external dependency (not vendored) if version lock allows

---

## Known Bugs

### YouTube Login Wall Blocking Downloads

**Symptoms:** Video download fails with "LOGIN_REQUIRED: Sign in to confirm you're not a bot" error on many videos.

**Files:** `pkg/video/extractor/youtube/innertube.go` (InnerTube API client with client config rotation)

**Trigger:** Downloading videos that require bot verification or are age-restricted.

**Error example:**
```
Error: YouTube [9tmsq-Gvx6g]: all InnerTube clients failed 
[ANDROID_VR: ERROR This video is unavailable; 
 WEB: LOGIN_REQUIRED Sign in to confirm you're not a bot; ...]
```

**Cause:** InnerTube API (used by YouTube internally) detects automated requests. Multiple client configs (ANDROID_VR, WEB, ANDROID, IOS, TVHTML5) rotate but bot detection is sophisticated.

**Current mitigation:**
- Code tries 5 different InnerTube client types to bypass detection
- ANDROID_VR client used first (less bot detection)
- Users can provide cookies via `--cookie-file`

**Improvement path:**
1. Add support for user-provided authentication tokens (OAuth, cookies)
2. Implement exponential backoff with randomized delays
3. Add proxy rotation support
4. Consider integrating yt-dlp's PoToken (proof-of-origin token) generation
5. Document workaround: use external yt-dlp for age-gated content

### Binary File Detection Heuristic

**Symptoms:** Some text files with null bytes misclassified as binary; some binary files incorrectly scanned as text.

**Files:** `pkg/search/rg/rg.go` (IsBinary function), `pkg/video/utils/parse.go` (similar pattern)

**Trigger:** Files with embedded nulls or unusual encodings.

**Cause:** Uses simple null-byte heuristic (`bytes.Contains(data, []byte{0})`) instead of magic-number based detection or file command.

**Workaround:** Use `-g` glob pattern to explicitly exclude binary types.

**Fix approach:**
1. Add magic number detection (ELF, PE, Mach-O headers)
2. Check file extension against known binary types
3. Fall back to null-byte heuristic only for unknown extensions

---

## Security Considerations

### Unsafe Regex in Grep/Rg

**Risk:** ReDoS (Regular Expression Denial of Service) via malicious patterns could cause CPU spike or hang.

**Files:** 
- `pkg/search/grep/grep.go` (uses `regexp.Compile`)
- `internal/cli/rg/rg.go` (uses `regexp.Compile`)
- `internal/cli/sed/sed.go` (pattern substitution)

**Current mitigation:** Regex engine is Go's `regexp` (NFA-based, linear time), not PCRE (vulnerable to ReDos).

**Recommendations:**
1. Add regex compile timeout (context with timeout)
2. Log/warn on slow patterns
3. Add `--max-compile-time` flag for safety in CI/CD

### Environment Variable Injection in Commands

**Risk:** Commands that execute shell code (pipe, pipeline, watch, xargs) could be vulnerable to env var injection.

**Files:**
- `internal/cli/pipe/execute.go` (executes parsed commands)
- `internal/cli/xargs/xargs.go` (variable substitution)
- `internal/cli/watch/watch.go` (executes commands repeatedly)

**Current mitigation:** Commands use `Command` interface with typed args (no shell parsing), except where explicitly designed for pipes.

**Recommendations:**
1. Sanitize variable names in pipe substitution (reject special chars)
2. Document security model: "pipe does not use shell, only Command registry"
3. Add `--safe-mode` flag that disables variable substitution

### Cookie File Plaintext Storage

**Risk:** User cookies stored in plaintext on disk for video downloads.

**Files:** `pkg/video/nethttp/cookies.go` (LoadNetscapeCookies), `internal/cli/video/download.go` (--cookie-file option)

**Current mitigation:** Users opt-in with `--cookie-file` flag; documentation should warn about plaintext.

**Recommendations:**
1. Support encrypted cookie files (AES-256-GCM via existing cryptutil)
2. Add `--encrypt-cookies` flag to auto-encrypt on load
3. Document plaintext risk in help text

---

## Performance Bottlenecks

### Video Download Signature Decryption Overhead

**Problem:** YouTube signature decryption uses goja JS runtime, causing ~500ms+ latency per download.

**Files:**
- `pkg/video/extractor/youtube/signature.go` (JS player parsing and signature deobfuscation)
- `pkg/video/jsinterp/jsinterp.go` (goja JavaScript interpreter)

**Cause:** YouTube player JS is large (~200KB), parsed and executed for each video. No caching between sessions.

**Improvement path:**
1. Cache deobfuscated signature function across session (in-memory hash by player version)
2. Lazy-load JS only if signature required (not all formats need it)
3. Pre-compute known signature patterns (yt-dlp style)
4. Consider WASM-based JS execution for speed

### Parallel Rg Scanning Contention

**Problem:** Large directory trees with many files cause goroutine overhead to exceed benefits.

**Files:** `pkg/search/rg/rg.go` (searchDirParallel with configurable workers), `internal/cli/rg/rg.go` (Run with context)

**Current:** Default `-j` (workers) is auto (NumCPU), but large channel buffers can cause memory bloat.

**Improvement path:**
1. Profile with real-world large codebases (Linux kernel, Chromium)
2. Tune worker pool size dynamically based on file count and depth
3. Add memory-limit flag `--max-memory` to cap goroutines
4. Benchmark against native ripgrep on same machine

### Tree Scanning with MaxHashSize

**Problem:** Hashing large files (e.g., VM images, binaries) during tree scan slows down comparison.

**Files:** `pkg/twig/scanner/scanner.go` (MaxHashSize option), `pkg/twig/comparer/comparer.go` (move detection via hash)

**Current:** Default MaxHashSize is unlimited. Move detection requires hashing all files.

**Improvement path:**
1. Skip hash for files larger than `--max-hash-size` (default 50MB)
2. Use file size + mtime instead of hash for large files (acceptable false positives)
3. Add `--fast-compare` flag to disable move detection

---

## Fragile Areas

### YouTube InnerTube API Brittle

**Files:**
- `pkg/video/extractor/youtube/innertube.go` (client rotation, endpoint hardcoding)
- `pkg/video/extractor/youtube/youtube.go` (videoDetails parsing, captions extraction)
- `pkg/video/extractor/youtube/search.go` (search API via InnerTube)
- `pkg/video/extractor/youtube/channel.go` (channel tab continuation tokens)

**Why fragile:**
- YouTube changes InnerTube API response structure frequently (breaking existing parsers)
- Client configs require updates when YouTube player JS changes
- Continuation tokens (pagination) are opaque, structure changes break channel scraping
- No API contract or documentation (reverse-engineered)

**Safe modification approach:**
1. Add integration test fixtures (saved JSON responses) before modifying parsers
2. Update client configs from yt-dlp upstream when they change
3. Add version detection for player JS to cache signature functions per version
4. Test coverage: Add 20+ fixtures for different video types (standard, age-gated, livestream, premiere)

**Test coverage gaps:**
- No unit tests for innertube.go (0%)
- No tests for various error conditions (unavailable, login required, blocked by country)
- No tests for pagination/continuation tokens in channels

### Sed Pattern Matching Incomplete

**Files:** `internal/cli/sed/sed.go` (basic substitution), `internal/cli/sed/sed_test.go` (limited test cases)

**Why fragile:**
- Only supports basic `s/pattern/replacement/flags` syntax
- No multi-line patterns, hold space, branching, labels, or advanced features
- Regex differences from GNU sed may surprise users

**Safe modification approach:**
1. Document limitations clearly in help text
2. Require explicit flag `--advanced` to enable future features
3. Add tests for GNU sed compatibility (test against `sed` binary on each platform)
4. Add warning when parsing sed expressions with advanced syntax

---

## Scaling Limits

### HLS Segment Download Parallelism

**Resource:** Network bandwidth and goroutine overhead during HLS stream download.

**Current capacity:** Fixed worker pool (default 4 goroutines per stream segment download).

**Limit:** CPU and memory plateau around 50-100 concurrent segments. Larger streams fail to scale beyond network I/O limit.

**Scaling path:**
1. Make worker count tunable via `--hls-workers` flag
2. Add adaptive worker pool (start at 4, scale up to NumCPU or bandwidth limits)
3. Implement per-segment timeout and retry with backoff
4. Monitor and log segment download latency

### SQLite Pure-Go Performance

**Resource:** Database query throughput under load.

**Current capacity:** modernc.org/sqlite is slower than cgo-based sqlite3 (no WAL mode optimization).

**Limit:** >1000 queries/second becomes noticeable (10-100ms queries on typical hardware).

**Scaling path:**
1. Add connection pooling for concurrent queries
2. Use batch inserts for bulk operations (channel incremental download)
3. Consider cgo-based sqlite3 as optional faster backend
4. Profile with real workload (video channel with >10k videos)

---

## Dependencies at Risk

### Goja JS Runtime (YouTube Signature)

**Risk:** goja is pure Go JS interpreter, slower than V8 but safer (no FFI/C bindings). YouTube signature obfuscation changes require manual updates to deobfuscation patterns.

**Impact:** If goja has bugs with newer JS features, signature decryption breaks silently.

**Migration plan:**
1. Pre-compute known signature patterns (maintain mapping of player version → function)
2. Cache deobfuscated function across sessions
3. Monitor yt-dlp upstream for signature changes
4. Consider switching to wasmer/wasmtime if goja becomes unmaintained

### Bufbuild Protocompile (Vendored)

**Risk:** Vendored dependency makes updates difficult. If protocompile has security issue, patch requires full revendoring.

**Impact:** Slow builds, dependency update friction, security patch delays.

**Migration plan:**
1. Move to go.mod dependency (unvendor) if version stability allows
2. Add separate build target for buf operations (skip in CI unless testing buf)
3. Contribute patches upstream to protocompile if needed

---

## Missing Critical Features

### Error Recovery for Pipe/Pipeline

**Problem:** Single command failure in pipe chain stops entire pipeline (no continue-on-error support).

**Files:** `internal/cli/pipe/execute.go` (command execution), `pkg/pipeline/pipeline.go` (stage orchestration)

**Blocks:** Error recovery patterns, fault-tolerant data processing

**Implementation path:**
1. Add `--continue-on-error` flag to pipe command
2. Propagate error context through pipeline stages
3. Allow conditional branching: `if-error` commands

### Context Cancellation in Long-Running Ops

**Problem:** Commands like `video channel` (which downloads 1000s of videos incrementally) don't support Ctrl+C gracefully in all cases.

**Files:**
- `internal/cli/video/channel.go` (channel incremental download)
- `internal/cli/video/download.go` (video download with progress)
- `pkg/video/video.go` (extract/download dispatcher)

**Blocks:** Graceful shutdown, cleanup of partial files, resumption

**Fix approach:**
1. Thread context.Context through all download functions
2. Ensure cleanup on context cancellation (remove .part files)
3. Add `--resume` flag to continue interrupted downloads by fragment state
4. Test with `timeout` or SIGINT injection

---

## Test Coverage Gaps

### pkg/video/extractor/generic - No Tests

**What's not tested:** Fallback extraction for non-YouTube sites (direct video URLs, og:video tags, `<video>` elements).

**Files:** `pkg/video/extractor/generic/generic.go` (210 LOC)

**Risk:** Generic extraction silently fails on HTML structure changes.

**Priority:** P2

**Approach:**
1. Create fixtures with common HTML patterns (og:video, `<video>` tags, Twitter embeds)
2. Add 15+ test cases covering major site patterns
3. Test fallback to direct URL extraction

### pkg/video/extractor/youtube - Minimal Tests (4%)

**What's not tested:** InnerTube API response parsing, signature extraction, channel pagination, playlist extraction.

**Files:**
- `pkg/video/extractor/youtube/youtube.go` (478 LOC) - **0% coverage**
- `pkg/video/extractor/youtube/innertube.go` (204 LOC) - **0% coverage**
- `pkg/video/extractor/youtube/channel.go` (407 LOC) - **0% coverage**
- `pkg/video/extractor/youtube/search.go` (147 LOC) - **0% coverage**

**Risk:** YouTube API changes break silently. Playlist/channel extraction untested edge cases fail in production.

**Priority:** P1

**Approach:**
1. Create 30+ JSON fixtures from real YouTube API responses
2. Add test cases for: standard video, age-gated, live, premiere, shorts, playlists, channels, error responses
3. Mock nethttp.Client in tests (no live API calls)
4. Test continuation token handling for pagination

### Platform-Specific Edge Cases

**What's not tested:** Windows-specific `ln -s`, permission model differences, signal handling on Windows.

**Files:**
- `internal/cli/ln/ln_windows.go` (symlink handling)
- `internal/cli/kill/kill_windows.go` (signal handling)
- `internal/cli/chmod/chmod_windows.go` (ACL mapping)

**Risk:** Windows users encounter untested code paths.

**Priority:** P2

**Approach:**
1. Add platform-specific test suites (build tags: `//go:build windows`)
2. Mock OS APIs for cross-platform testing
3. Test symlink creation with/without admin privileges
4. Test signal enumeration on Windows

---

## Build & CI Issues

### No Coverage Threshold Enforcement

**Issue:** Coverage is ~25.8% overall (~78% for omni-owned packages), but no CI check enforces minimum.

**Impact:** Coverage can regress unnoticed.

**Fix:** Add CI step checking coverage >= 80% for omni packages, allow overrides for vendored packages.

### Slow Test Suite

**Issue:** Full `go test ./...` includes buf vendored package compilation (~30-60s).

**Files:** `vendor/github.com/bufbuild/protocompile/`

**Fix:** Split test targets:
- `task test` - Omni packages only (<10s)
- `task test:all` - Including buf (<60s)
- `task test:buf` - buf package only

---

*Concerns audit: 2026-04-11*
