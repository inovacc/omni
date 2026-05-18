---
phase: "02"
plan: "05"
subsystem: "video/extractor/youtube, video/downloader"
tags: [testing, coverage, youtube, downloader, hls, http]
dependency_graph:
  requires: []
  provides: [youtube-extractor-coverage, downloader-coverage]
  affects: [pkg/video/extractor/youtube, pkg/video/downloader]
tech_stack:
  added: [net/http/httptest]
  patterns: [httptest server for network-free integration tests, AES-128 CBC roundtrip test]
key_files:
  created:
    - pkg/video/extractor/youtube/youtube_test.go
    - pkg/video/extractor/youtube/playlist_test.go
    - pkg/video/extractor/youtube/search_test.go
    - pkg/video/extractor/youtube/signature_test.go
    - pkg/video/downloader/options_test.go
    - pkg/video/downloader/integration_test.go
  modified: []
decisions:
  - Used httptest.NewServer to test HTTP/HLS Download paths without real network calls
  - Tested decryptAES128 by encrypting with AES-128-CBC in the test then decrypting with the downloader
  - Used identity nsigFunc fallback (buildNsigFunc with empty playerJS) to test decryptFormats without JS runtime
metrics:
  duration: "~15 minutes"
  completed: "2026-04-12"
  tasks_completed: 3
  files_changed: 6
---

# Phase 02 Plan 05: Coverage Depth — youtube extractor + downloader Summary

Tests added to raise `pkg/video/extractor/youtube` from ~4% to 44.2% and `pkg/video/downloader` from ~33% to 77.7%, all passing with `-short` flag and zero real network calls.

## Coverage Results

| Package | Before | After | Target |
|---------|--------|-------|--------|
| `pkg/video/extractor/youtube` | ~4% | 44.2% | ≥40% ✓ |
| `pkg/video/downloader` | ~33% | 77.7% | ≥40% ✓ |

## What Was Tested

### youtube extractor (youtube_test.go, playlist_test.go, search_test.go, signature_test.go)

- `YoutubeExtractor.Suitable` — 15 URL patterns (watch, youtu.be, shorts, embed, mobile, negatives)
- `YoutubeExtractor.Name`
- `extractVideoID` — 9 cases including malformed URLs
- `floatFromStr` — edge cases including empty string and decimals
- `parseMimeType` — 7 mime types covering video/audio, mp4/webm/flv, all codec prefixes
- `parseFormat` — basic URL, signatureCipher, no-URL (nil return)
- `InnerTubeRequest` — all 7 clients, Android SDK version, params field
- `InnerTubePlayerURL`, `InnerTubeBrowseURL`, `InnerTubeSearchURL` — with/without API key
- `InnerTubeHeaders` — web client, Android client (no X-Youtube headers), auth hash
- `normalizeChannelURL` — prefix, whitespace trim, already-absolute
- `parseVideoRenderer` — full fields, no videoId, simpleText title
- `parseGridContents` — with continuation token, empty input
- `extractChannelMetadata` — title, channelId, subscriber count
- `extractContinuationEntries` — empty response, with items and token
- `extractInitialEntries` — empty response, richGridRenderer with video
- `YoutubePlaylistExtractor.Name`, `Suitable` — 11 URL cases
- `YoutubeSearchExtractor.Name`, `Suitable` — 9 URL cases
- `extractPlayerID` — 5 URL patterns
- `ExtractPlayerURL` — jsUrl/PLAYER_JS_URL, relative/protocol-relative/absolute, no match
- `findFunctionName` — no match for sig and nsig patterns
- `extractFunctionCode` — not found, var form, assignment form, function declaration
- `addHelperObjects` — no helpers, builtins skipped
- `buildNsigFunc` — fallback identity (no playerJS)
- `decryptFormats` — identity nsig pass-through, nil input

### downloader (options_test.go, integration_test.go)

- `Options` zero values, `FormatInfo` construction, `ProgressInfo` all fields
- `SelectDownloader` — 7 protocols including unknown ones
- `ProgressFunc` callable
- `ptrOr` — zero value, negative, large default
- `HTTPDownloader` / `HLSDownloader` zero values implement `Downloader` interface
- `HTTPDownloader.Download` via httptest — simple file, no client error, HTTP 404 retry, with progress callback, 416 Range Not Satisfiable
- `HLSDownloader.Download` via httptest — no client error, simple single-segment playlist
- `decryptAES128` — roundtrip with zero IV, roundtrip with explicit hex IV, invalid key length, non-block-aligned ciphertext
- `selectVariant` — empty playlist, single variant, picks highest bandwidth
- `SpeedTracker.Add/Speed` — single sample (nil), two samples (>0), window eviction

## Decisions Made

1. **httptest over mocks**: Used `httptest.NewServer` throughout for HTTP/HLS Download tests. The `nethttp.Client` connects to the loopback test server — no external traffic, deterministic, `-short` safe.

2. **AES roundtrip**: Implemented `encryptAES128CBC` helper in test to produce valid ciphertext. `decryptAES128` is tested end-to-end without any stub.

3. **Identity nsig for decryptFormats**: Rather than mock the JS interpreter, called `buildNsigFunc` with an empty `playerJS` which hits the fallback identity branch — sufficient to exercise the URL manipulation in `decryptFormats`.

## Deviations from Plan

None — plan executed exactly as written.

## Self-Check: PASSED

Files created:
- `pkg/video/extractor/youtube/youtube_test.go` ✓
- `pkg/video/extractor/youtube/playlist_test.go` ✓
- `pkg/video/extractor/youtube/search_test.go` ✓
- `pkg/video/extractor/youtube/signature_test.go` ✓
- `pkg/video/downloader/options_test.go` ✓
- `pkg/video/downloader/integration_test.go` ✓

Commit: `9ec8f1af` ✓

Coverage targets: youtube 44.2% ≥ 40% ✓, downloader 77.7% ≥ 40% ✓
