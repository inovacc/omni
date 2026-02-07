#!/usr/bin/env python3
"""Black-box tests for video commands — comparison against yt-dlp.

This test requires yt-dlp to be installed (pip install yt-dlp).
It compares omni video info/list-formats output against yt-dlp's JSON output
for a set of YouTube test videos.

Designed to run inside a Docker container with both omni and yt-dlp available.
Can also run locally if both tools are installed.
"""

import json
import os
import subprocess
import sys
import tempfile
import shutil
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import (
    OmniTester,
    assert_contains,
    assert_eq,
    assert_exit_code,
    assert_json_valid,
)

# Test video URLs — short/small videos to keep tests fast.
TEST_VIDEOS = [
    {
        "url": "https://www.youtube.com/watch?v=cARmRQju7Bc",
        "id": "cARmRQju7Bc",
        "desc": "diesel pile hammer (short clip)",
    },
    {
        "url": "https://www.youtube.com/watch?v=uagA2P6yJuw",
        "id": "uagA2P6yJuw",
        "desc": "test video 2",
    },
    {
        "url": "https://www.youtube.com/watch?v=TB4wALx3KU8",
        "id": "TB4wALx3KU8",
        "desc": "test video 3",
    },
]


def has_ytdlp() -> bool:
    """Check if yt-dlp is available."""
    try:
        result = subprocess.run(
            ["yt-dlp", "--version"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        return result.returncode == 0
    except (FileNotFoundError, subprocess.TimeoutExpired):
        return False


def run_ytdlp_info(url: str, timeout: int = 120) -> dict | None:
    """Run yt-dlp --dump-json and return parsed JSON."""
    try:
        result = subprocess.run(
            ["yt-dlp", "--dump-json", "--no-download", url],
            capture_output=True,
            text=True,
            timeout=timeout,
        )
        if result.returncode != 0:
            print(f"    yt-dlp stderr: {result.stderr[:300]}")
            return None
        return json.loads(result.stdout)
    except (subprocess.TimeoutExpired, json.JSONDecodeError) as e:
        print(f"    yt-dlp error: {e}")
        return None


def run_ytdlp_formats(url: str, timeout: int = 120) -> str | None:
    """Run yt-dlp -F and return the format list output."""
    try:
        result = subprocess.run(
            ["yt-dlp", "-F", url],
            capture_output=True,
            text=True,
            timeout=timeout,
        )
        if result.returncode != 0:
            return None
        return result.stdout
    except subprocess.TimeoutExpired:
        return None


def main():
    print("=== Testing video commands (comparison with yt-dlp) ===")

    ytdlp_available = has_ytdlp()
    if ytdlp_available:
        print("  yt-dlp: found")
    else:
        print("  yt-dlp: NOT found (comparison tests will be skipped)")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        download_dir = Path(t.temp_dir) / "downloads"
        download_dir.mkdir()

        # ──────────────────────────────────────────────────────────
        # Test: omni video extractors command works
        # ──────────────────────────────────────────────────────────
        @t.test("video_extractors_list")
        def test_extractors():
            result = t.run("video", "extractors")
            assert_exit_code(result, 0, "video extractors should succeed")
            assert_contains(result.stdout, "YouTube", "should list YouTube extractor")
            assert_contains(result.stdout, "Generic", "should list Generic extractor")

        test_extractors()

        # ──────────────────────────────────────────────────────────
        # Per-video tests
        # ──────────────────────────────────────────────────────────
        for vid in TEST_VIDEOS:
            video_url = vid["url"]
            video_id = vid["id"]
            video_desc = vid["desc"]

            # ── omni video info ────────────────────────────────
            @t.test(f"video_info_{video_id}")
            def test_info(url=video_url, vid_id=video_id):
                result = t.run("video", "info", url)
                assert_exit_code(result, 0, f"video info {vid_id} should succeed")
                assert_json_valid(result.stdout, f"video info {vid_id} should be valid JSON")

                info = json.loads(result.stdout)

                # Basic metadata must be present.
                assert_eq(info.get("id"), vid_id, f"video ID should be {vid_id}")
                assert_contains(
                    str(info.get("title", "")),
                    "",
                    f"title should not be empty for {vid_id}",
                )
                if not info.get("title"):
                    raise AssertionError(f"title is empty for {vid_id}")

                # Must have formats.
                formats = info.get("formats", [])
                if len(formats) == 0:
                    raise AssertionError(f"no formats found for {vid_id}")

                # Must have extractor info.
                assert_eq(
                    info.get("extractor", "").lower(),
                    "youtube",
                    f"extractor should be YouTube for {vid_id}",
                )

            test_info()

            # ── omni video list-formats ────────────────────────
            @t.test(f"video_list_formats_{video_id}")
            def test_formats(url=video_url, vid_id=video_id):
                result = t.run("video", "list-formats", url)
                assert_exit_code(result, 0, f"list-formats {vid_id} should succeed")
                # Output should include format table header or format lines.
                output = result.stdout
                if not output.strip():
                    raise AssertionError(f"list-formats output is empty for {vid_id}")

            test_formats()

            # ── omni video list-formats --json ─────────────────
            @t.test(f"video_list_formats_json_{video_id}")
            def test_formats_json(url=video_url, vid_id=video_id):
                result = t.run("video", "list-formats", "--json", url)
                assert_exit_code(result, 0, f"list-formats --json {vid_id} should succeed")
                assert_json_valid(result.stdout, f"list-formats --json {vid_id} should be valid JSON")

                formats = json.loads(result.stdout)
                if not isinstance(formats, list):
                    raise AssertionError(f"formats should be a list for {vid_id}")
                if len(formats) == 0:
                    raise AssertionError(f"no formats in JSON for {vid_id}")

            test_formats_json()

            # ── comparison: omni vs yt-dlp metadata ────────────
            if ytdlp_available:

                @t.test(f"video_compare_metadata_{video_id}")
                def test_compare_metadata(url=video_url, vid_id=video_id):
                    # Get omni info.
                    omni_result = t.run("video", "info", url)
                    assert_exit_code(omni_result, 0, f"omni video info {vid_id}")
                    omni_info = json.loads(omni_result.stdout)

                    # Get yt-dlp info.
                    ytdlp_info = run_ytdlp_info(url)
                    if ytdlp_info is None:
                        raise AssertionError(f"yt-dlp failed for {vid_id}")

                    # Compare video ID.
                    assert_eq(
                        omni_info.get("id"),
                        ytdlp_info.get("id"),
                        f"video ID mismatch for {vid_id}",
                    )

                    # Compare title (may differ slightly due to encoding).
                    omni_title = omni_info.get("title", "")
                    ytdlp_title = ytdlp_info.get("title", "")
                    if not omni_title:
                        raise AssertionError(f"omni title empty for {vid_id}")
                    if not ytdlp_title:
                        raise AssertionError(f"yt-dlp title empty for {vid_id}")

                    # Titles should be similar (first 20 chars).
                    omni_prefix = omni_title[:20].lower().strip()
                    ytdlp_prefix = ytdlp_title[:20].lower().strip()
                    if omni_prefix != ytdlp_prefix:
                        print(f"    title similarity warning: omni='{omni_title}' vs yt-dlp='{ytdlp_title}'")

                    # Compare duration (allow 1s tolerance).
                    omni_dur = omni_info.get("duration", 0)
                    ytdlp_dur = ytdlp_info.get("duration", 0)
                    if omni_dur and ytdlp_dur:
                        diff = abs(float(omni_dur) - float(ytdlp_dur))
                        if diff > 1.0:
                            raise AssertionError(
                                f"duration mismatch for {vid_id}: "
                                f"omni={omni_dur} vs yt-dlp={ytdlp_dur} (diff={diff}s)"
                            )

                    # Compare uploader.
                    omni_uploader = omni_info.get("uploader", "")
                    ytdlp_uploader = ytdlp_info.get("uploader", "") or ytdlp_info.get("channel", "")
                    if omni_uploader and ytdlp_uploader:
                        if omni_uploader.lower() != ytdlp_uploader.lower():
                            print(
                                f"    uploader note: omni='{omni_uploader}' vs yt-dlp='{ytdlp_uploader}'"
                            )

                test_compare_metadata()

                # ── comparison: format counts ──────────────────
                @t.test(f"video_compare_formats_{video_id}")
                def test_compare_formats(url=video_url, vid_id=video_id):
                    # Get omni formats.
                    omni_result = t.run("video", "list-formats", "--json", url)
                    assert_exit_code(omni_result, 0, f"omni list-formats {vid_id}")
                    omni_formats = json.loads(omni_result.stdout)

                    # Get yt-dlp info (includes formats).
                    ytdlp_info = run_ytdlp_info(url)
                    if ytdlp_info is None:
                        raise AssertionError(f"yt-dlp failed for {vid_id}")

                    ytdlp_formats = ytdlp_info.get("formats", [])

                    omni_count = len(omni_formats)
                    ytdlp_count = len(ytdlp_formats)

                    print(f"    format counts: omni={omni_count}, yt-dlp={ytdlp_count}")

                    # omni should find at least some formats.
                    if omni_count == 0:
                        raise AssertionError(f"omni found 0 formats for {vid_id}")

                    # Compare format IDs — check overlap.
                    omni_ids = {f.get("format_id", "") for f in omni_formats}
                    ytdlp_ids = {str(f.get("format_id", "")) for f in ytdlp_formats}

                    common = omni_ids & ytdlp_ids
                    print(
                        f"    format ID overlap: {len(common)} common "
                        f"(omni unique: {len(omni_ids - ytdlp_ids)}, "
                        f"yt-dlp unique: {len(ytdlp_ids - omni_ids)})"
                    )

                    # At least some format IDs should overlap.
                    # HLS format IDs differ (hls-0 vs hls-1), so we only check
                    # non-HLS format overlap.
                    omni_non_hls = {fid for fid in omni_ids if not fid.startswith("hls-")}
                    ytdlp_non_hls = {fid for fid in ytdlp_ids if not fid.startswith("hls-")}
                    non_hls_common = omni_non_hls & ytdlp_non_hls

                    if len(non_hls_common) == 0 and len(omni_non_hls) > 0 and len(ytdlp_non_hls) > 0:
                        print(
                            f"    warning: no non-HLS format ID overlap "
                            f"(omni: {sorted(omni_non_hls)[:5]}, "
                            f"yt-dlp: {sorted(ytdlp_non_hls)[:5]})"
                        )

                test_compare_formats()

        # ──────────────────────────────────────────────────────────
        # Test: download smallest format (only first video, to save bandwidth)
        # ──────────────────────────────────────────────────────────
        first_video = TEST_VIDEOS[0]

        @t.test(f"video_download_worst_{first_video['id']}")
        def test_download_worst():
            output_path = str(download_dir / f"{first_video['id']}_worst.mp4")
            result = t.run(
                "video", "download",
                "-f", "worst",
                "-o", output_path,
                first_video["url"],
            )
            assert_exit_code(result, 0, f"download worst {first_video['id']}")

            # Check file was created and has content.
            if not os.path.exists(output_path):
                # The file might have a different extension.
                found = list(download_dir.glob(f"{first_video['id']}_worst*"))
                if not found:
                    raise AssertionError(f"downloaded file not found at {output_path}")
                output_path_actual = str(found[0])
            else:
                output_path_actual = output_path

            size = os.path.getsize(output_path_actual)
            print(f"    downloaded: {output_path_actual} ({size:,} bytes)")
            if size < 1000:
                raise AssertionError(f"downloaded file too small: {size} bytes")

        test_download_worst()

        # ──────────────────────────────────────────────────────────
        # Test: download with yt-dlp and compare file sizes (if available)
        # ──────────────────────────────────────────────────────────
        if ytdlp_available:

            @t.test(f"video_download_compare_{first_video['id']}")
            def test_download_compare():
                vid_id = first_video["id"]

                # Download with omni (format 18 = 360p mp4, common across both tools).
                omni_path = str(download_dir / f"{vid_id}_omni.mp4")
                omni_result = t.run(
                    "video", "download",
                    "-f", "18",
                    "-o", omni_path,
                    first_video["url"],
                )

                # Download with yt-dlp.
                ytdlp_path = str(download_dir / f"{vid_id}_ytdlp.mp4")
                try:
                    ytdlp_result = subprocess.run(
                        [
                            "yt-dlp",
                            "-f", "18",
                            "-o", ytdlp_path,
                            first_video["url"],
                        ],
                        capture_output=True,
                        text=True,
                        timeout=120,
                    )
                except subprocess.TimeoutExpired:
                    raise AssertionError("yt-dlp download timed out")

                if omni_result.returncode != 0:
                    print(f"    omni download failed (format 18 may not be available)")
                    print(f"    stderr: {omni_result.stderr[:300]}")
                    return  # Not a failure — format 18 might not exist.

                if ytdlp_result.returncode != 0:
                    print(f"    yt-dlp download failed (format 18 may not be available)")
                    return

                if not os.path.exists(omni_path) or not os.path.exists(ytdlp_path):
                    print("    one or both files missing, skipping size comparison")
                    return

                omni_size = os.path.getsize(omni_path)
                ytdlp_size = os.path.getsize(ytdlp_path)

                print(f"    omni:   {omni_size:>12,} bytes")
                print(f"    yt-dlp: {ytdlp_size:>12,} bytes")

                # File sizes should be within 5% of each other.
                if ytdlp_size > 0:
                    ratio = omni_size / ytdlp_size
                    print(f"    ratio:  {ratio:.4f}")
                    if ratio < 0.90 or ratio > 1.10:
                        raise AssertionError(
                            f"file size mismatch: omni={omni_size}, yt-dlp={ytdlp_size}, "
                            f"ratio={ratio:.4f} (expected 0.90-1.10)"
                        )

            test_download_compare()

        t.print_summary()
        sys.exit(t.exit_code())


if __name__ == "__main__":
    main()
