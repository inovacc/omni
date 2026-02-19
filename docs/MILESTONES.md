# omni - Version Milestones

## Released

### v1.0.0 - Foundation
**Released:** Tagged in repository

**Goals:**
- [x] Core shell utilities (ls, cat, pwd, date, basename, dirname, realpath)
- [x] Cobra CLI framework
- [x] Cross-platform support (Linux, macOS, Windows)
- [x] Basic test suite

**Test Coverage:** 59.4% overall (pkg/ packages avg ~78%)

---

### v1.1.0 - File Operations
**Released:** Tagged in repository

**Goals:**
- [x] File manipulation: cp, mv, rm, mkdir, rmdir, touch, stat, ln, readlink
- [x] Permission management: chmod, chown
- [x] File discovery: find, which, file
- [x] Safe defaults for destructive operations

---

### v1.2.0 - Text Processing & Search
**Released:** Tagged in repository

**Goals:**
- [x] Text processing: grep, egrep, fgrep, head, tail, sort, uniq, wc, cut, tr, sed, awk
- [x] Additional text: nl, paste, tac, column, fold, join, shuf, split, rev, comm, cmp, strings
- [x] Ripgrep-style search (rg) with gitignore support and parallel walking
- [x] Diff engine with unified format output

---

### v1.3.0 - Data, Encoding & Security
**Released:** Tagged in repository

**Goals:**
- [x] Data processing: jq, yq, dotenv, json (tostruct, tocsv, fromcsv, toxml, fromxml)
- [x] Encoding: base64, base32, base58, hex, url, html encode/decode, xxd
- [x] Hashing: hash, md5sum, sha256sum, sha512sum, crc32sum, crc64sum
- [x] Security: encrypt (AES-256-GCM), decrypt, uuid, random, jwt decode
- [x] Archive: tar, zip, unzip, gzip, bzip2, xz (and decompressors)
- [x] Formatting: sql fmt/minify/validate, css fmt/minify/validate, html fmt/minify/validate

---

### v1.4.0 - Cloud, DevOps & Engines
**Released:** Tagged in repository

**Goals:**
- [x] Kubernetes integration via k8s.io/kubectl
- [x] Terraform CLI wrapper
- [x] AWS SDK integration (EC2, S3, IAM, STS, SSM)
- [x] Git hacks (12 shortcuts): quick-commit, branch-clean, undo, amend, log-graph, etc.
- [x] Kubectl hacks (17 shortcuts): kga, klf, keb, kpf, krr, kge, etc.
- [x] Pipe engine (Cobra dispatch with variable substitution)
- [x] Pipeline engine (streaming io.Pipe stages, 20 built-in transforms)
- [x] Video download engine (pure Go youtube-dl port)
- [x] Protobuf tooling (buf lint, format, compile, breaking, generate)
- [x] Project analyzer (info, deps, docs, git, health)
- [x] pkg/ library extraction (12 reusable packages)
- [x] SQLite and BBolt database management
- [x] Cobra CLI code generator
- [x] Black-box test suite (Python)

---

### v1.5.0 - Infrastructure & Analysis
**Released:** Tagged in repository (current)

**Goals:**
- [x] Unified `Command` interface contract (`internal/cli/command/`)
- [x] cmderr error sentinels adopted in 49+ commands (batches 1-5)
- [x] `repo analyze` command — LLM-optimized repository context generation
- [x] Exported project package functions for reuse (DetectProjectTypes, AnalyzeDeps, etc.)
- [x] Golden master tests expanded to 117 tests, 13 categories
- [x] buf format/lint upgraded with protocompile AST parser
- [x] ADR-0003: Keep protocompile as external dependency
- [x] Flatten buf module into omni root module

---

## Planned

### v2.0.0 - Production Ready
**Target:** TBD

**Goals:**
- [x] Unified Command interface contract
- [~] Consistent error model with exit codes (49+ of 160+ commands adopted)
- [ ] `--json` flag on all commands
- [ ] 80%+ overall test coverage
- [ ] CI coverage threshold enforcement
- [x] Golden master testing framework (117 tests, 13 categories)
- [ ] Full command reference documentation
- [ ] Automated releases with goreleaser
- [ ] Multi-platform, multi-arch builds

**Test Coverage Target:** 80%+

---

## Test Coverage Summary

**Current Total:** 59.4% (includes vendored buf packages) | **Omni-owned pkg/ avg:** ~78%

### Omni-owned pkg/ Packages

| Package | Coverage | Status |
|---------|----------|--------|
| pkg/encoding | 100.0% | Excellent |
| pkg/twig/models | 100.0% | Excellent |
| pkg/twig/expander | 98.1% | Excellent |
| pkg/video/m3u8 | 96.8% | Excellent |
| pkg/twig/comparer | 96.3% | Excellent |
| pkg/textutil/diff | 95.2% | Excellent |
| pkg/textutil | 93.7% | Excellent |
| pkg/video/jsinterp | 91.7% | Excellent |
| pkg/idgen | 90.3% | Excellent |
| pkg/hashutil | 88.5% | Good |
| pkg/cssfmt | 87.3% | Good |
| pkg/search/rg | 86.6% | Good |
| pkg/cryptutil | 85.3% | Good |
| pkg/figlet | 82.9% | Good |
| pkg/twig/scanner | 81.9% | Good |
| pkg/pipeline | 81.5% | Good |
| pkg/twig/formatter | 80.4% | Good |
| pkg/sqlfmt | 79.1% | Acceptable |
| pkg/twig/parser | 79.1% | Acceptable |
| pkg/search/grep | 77.9% | Acceptable |
| pkg/htmlfmt | 77.9% | Acceptable |
| pkg/video/cache | 73.3% | Acceptable |
| pkg/jsonutil | 67.5% | Needs improvement |
| pkg/video/nethttp | 61.8% | Needs improvement |
| pkg/twig/builder | 58.9% | Needs improvement |
| pkg/video/utils | 58.4% | Needs improvement |
| pkg/video/format | 50.0% | Needs improvement |
| pkg/video (root) | 46.0% | Needs improvement |
| pkg/twig | 44.3% | Needs improvement |
| pkg/userdirs | 42.9% | Needs improvement |
| pkg/video/extractor | 41.7% | Needs improvement |
| pkg/video/downloader | 34.6% | Needs improvement |
| pkg/video/extractor/youtube | 4.0% | Minimal |
| pkg/video/extractor/generic | 0.0% | No tests |
