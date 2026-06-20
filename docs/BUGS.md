# Bug Tracker

> Last updated: 2026-06-04

## Open

### BUG-0001: `omni video dl` fails on bot-gated YouTube videos (InnerTube LOGIN_REQUIRED)

**Severity:** Medium
**Status:** Resolved (Won't-fix): video feature removed in plan 015
**Resolution:** The entire `omni video` feature (commands, `internal/cli/video`, `pkg/video`, and the `goja` dependency) was removed in plan 015 to keep the no-exec invariant absolute. This bug is therefore moot — there is no longer a video downloader to bot-gate. Entry retained for historical record.
**Affected:** `omni video dl` (`pkg/video`, InnerTube client rotation) — removed in plan 015
**Platform:** All

**Description:**
Some YouTube videos fail to download with every InnerTube client exhausted, e.g.
`YouTube [<id>]: all InnerTube clients failed [ANDROID_VR: ERROR This video is unavailable;
WEB: LOGIN_REQUIRED Sign in to confirm you're not a bot; ... TVHTML5: LOGIN_REQUIRED ...]`.
YouTube's anti-bot / login wall rejects the anonymous InnerTube clients omni rotates through.

**Steps to Reproduce:**
1. `omni video dl "https://www.youtube.com/watch?v=3NzCBIcIqD0"`
2. (and `...?v=9tmsq-Gvx6g`)

**Expected Behavior:** Download succeeds, or fails with a clear "this video requires sign-in / is bot-gated" message distinguishing it from a genuine error.

**Actual Behavior:** All InnerTube clients fail (`LOGIN_REQUIRED` / `ERROR This video is unavailable`); exits non-zero.

**Workaround:** None for bot-gated videos without authenticated cookies. Some videos work; this is environment/video-specific.

**Root cause / scope:** External fragility, not a code regression — consistent with the known limitation in `docs/BACKLOG.md` ("YouTube signature decryption is fragile; player-JS changes require updates"). A real fix needs authenticated-cookie support (an omni design question — credential handling) or accepting bot-gated videos as out of scope. Pre-v1.0 this is a documented limitation, not a release blocker (video is not part of the supply-chain core value).

**Fixed in:** —

---

## Template

When reporting bugs, include:

```
### BUG-XXXX: [Title]

**Severity:** Critical / High / Medium / Low
**Status:** Open / In Progress / Fixed / Won't Fix
**Affected:** [command or package]
**Platform:** All / Linux / macOS / Windows

**Description:**
[What happens]

**Steps to Reproduce:**
1. [Step 1]
2. [Step 2]

**Expected Behavior:**
[What should happen]

**Actual Behavior:**
[What actually happens]

**Workaround:**
[If any]

**Fixed in:** [version or commit]
```
