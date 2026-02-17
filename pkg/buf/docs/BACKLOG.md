# Backlog

## P2 - Medium Priority

### Tech debt: TODO/FIXME cleanup
- **Status:** Triaged
- **Effort:** Medium
- **Details:** 139 files contain TODO/FIXME/HACK comments. Triaged items below.

~~**controller.go — functionOptions refactor:**~~
- Completed. Split `FunctionOption` into domain-specific `WorkspaceOption`, `ImageOption`, and `MessageOption` types with embedded composition.

~~**controller.go — flag name plumbing:**~~
- Completed. Flag names (`pathFlagName`, `excludePathFlagName`, `errorFormatFlagName`) are now plumbed through option constructors and used in error messages.

**controller.go — allowNotExist behavior (intentional):**
- Lines 1245, 1459: `allowNotExist` causes silent skipping when `--path` doesn't match. This is intentional — consistent with glob semantics where non-matching patterns are silently skipped. No change needed, but worth a clarifying comment if the code is touched.

**ref_parser.go — input type heuristic:**
- Line 834: "terrible heuristic" for detecting input type. Works in practice but fragile. Defer until it causes real issues.

**plugin_config.go — validation gaps:**
- Lines 48, 128: Validation improvements for plugin config. Low urgency.

**config_ignore_yaml.go — output format coupling:**
- Line 34: "this is messed" — the function generates v1 `buf.yaml` `ignore_only` YAML by string concatenation instead of using a proper YAML library or config model. Works but tightly couples output format to string building. Low urgency, would only matter if config format changes.

### Package extraction candidates
- **Status:** Evaluated (Phase 4 complete)
- **Effort:** Medium–Large per package
- **Details:** Extraction priority based on isolation analysis:

| Priority | Package | Internal Deps | Dependents | Notes |
|----------|---------|--------------|------------|-------|
| ~~P2~~ | ~~`thread`~~ | ~~0~~ | ~~9~~ | **DONE** — Extracted to `.local-deps/thread/` as `github.com/bufbuild/buf/internal/thread`. Established pattern for future extractions. |
| P3 | `dag` | 2 (`syserror`, `xslices`) | 4 | Deferred: `syserror` (130 importers) and `xslices` (146 importers) are foundational — co-extracting is disproportionate. No repo precedent for extracting internal utilities (existing pattern is BSR proto inlining only). |
| P3 | `protoencoding` | 2 (`protodescriptor`, `protoyaml`) | 17 | After pattern established with simpler packages. |
| P3 | `storage` | 9 | 36 | Too coupled currently. Defer until dependency count reduced. |

Note: All `buf.build` BSR proto dependencies and `github.com/bufbuild/protocompile` have been fully inlined into the main module (`internal/gen/proto/ext/` and `internal/protocompile/`). No `buf.build` imports or `replace` directives remain in `go.mod`.

## P3 - Low Priority

### Test helper consolidation
- **Status:** Assessed — minimal action needed
- **Effort:** Small
- **Details:** 8 test helper packages evaluated. All are well-scoped and domain-specific:
  - `slogtestext` (21 imports) — test logger, widely used, keep as-is
  - `bufmoduletesting` (9 imports) — mock module system, buf-specific
  - `appcmdtesting` (7 imports) — CLI test runner, well-designed
  - `prototesting` (6 imports) — protoc integration
  - `buftesting` (5 imports) — protoc execution, googleapis caching
  - `xtesting` (4 imports) — `SkipIfShort()`, `GetTestTimeout()`
  - `bufanalysistesting` (2 imports) — annotation builders
  - `bufimagetesting` (1 import) — image file builders
  - ~~`iotesting` (0 imports) — removed, was unused~~

### Windows path handling
- **Status:** Known limitation
- **Effort:** Medium
- **Details:** Several packages use `normalpath` for cross-platform path handling, but some test files have Windows-specific variants (`_windows_test.go`). The `xfilepath` package notes `os.Getwd` usage that should use `osext`.

## Done

### Remove unused iotesting package
- Removed. Package had zero imports across the codebase.

### Remove bandeps dependency enforcement
- Completed. Go's built-in `internal/` visibility rules now enforce the 2-layer architecture.

### Fix .pb.go rawDesc handling in import rewriting
- Completed. Proto files regenerated with correct `internal/` `go_package` paths.

### Regenerate proto code with updated go_package
- Completed. All rawDesc descriptors now use `internal/gen/proto/go`.

### Improve error messages for offline mode
- Completed. Offline delegate errors already include actionable guidance.

### Split functionOptions into domain-specific option types
- Completed. `FunctionOption` replaced with `WorkspaceOption`, `ImageOption`, `MessageOption`. `imageOptions` embeds `workspaceOptions` for shared fields.

### Plumb flag names through controller options
- Completed. `WithTargetPaths`/`WithImageTargetPaths` accept flag names. `WithFileAnnotationErrorFormat` accepts flag name. Error messages now use caller-provided flag names instead of hardcoded strings.
