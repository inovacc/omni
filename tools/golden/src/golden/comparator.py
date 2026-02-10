"""Compare test results against golden master baselines."""

from __future__ import annotations

import difflib
import json
from pathlib import Path

from golden.manifest import compute_sha256, load_manifest
from golden.types import CompareResult, RunResult


def compare_results(
    results: list[RunResult],
    golden_dir: Path,
) -> list[CompareResult]:
    """Compare each run result against its stored golden master."""
    manifest = load_manifest(golden_dir)
    comparisons: list[CompareResult] = []

    for result in results:
        tc = result.test_case
        cat_dir = golden_dir / tc.golden_subdir
        json_path = cat_dir / f"{tc.golden_filename}.json"
        stdout_path = cat_dir / f"{tc.golden_filename}.stdout"

        # No golden file exists
        if not json_path.exists():
            comparisons.append(CompareResult(
                test_case=tc, status="new",
                error_msg="No golden master. Run 'record' to create baseline.",
            ))
            continue

        # Load stored golden
        try:
            with open(json_path, "r", encoding="utf-8") as f:
                meta = json.load(f)
            expected_stdout = stdout_path.read_text(encoding="utf-8") if stdout_path.exists() else ""
            expected_exit = meta["exit_code"]
            expected_stderr = meta.get("stderr", "")
        except Exception as e:
            comparisons.append(CompareResult(
                test_case=tc, status="error",
                error_msg=f"Failed to load golden: {e}",
            ))
            continue

        # SHA-256 fast path: check manifest hash first
        entry = manifest.find_entry(tc.category, tc.name)
        actual_sha = compute_sha256(result.stdout)
        if entry and entry.sha256 == actual_sha and result.exit_code == expected_exit and result.stderr == expected_stderr:
            comparisons.append(CompareResult(test_case=tc, status="match"))
            continue

        # Detailed comparison
        if result.exit_code != expected_exit:
            comparisons.append(CompareResult(
                test_case=tc, status="mismatch",
                diff=f"Exit code: expected {expected_exit}, got {result.exit_code}",
            ))
            continue

        if result.stdout != expected_stdout:
            diff = _unified_diff(expected_stdout, result.stdout, tc.golden_key)
            comparisons.append(CompareResult(
                test_case=tc, status="mismatch", diff=diff,
            ))
            continue

        if result.stderr != expected_stderr:
            diff = _unified_diff(expected_stderr, result.stderr, f"{tc.golden_key} (stderr)")
            comparisons.append(CompareResult(
                test_case=tc, status="mismatch", diff=diff,
            ))
            continue

        comparisons.append(CompareResult(test_case=tc, status="match"))

    return comparisons


def _unified_diff(expected: str, actual: str, label: str) -> str:
    """Generate unified diff."""
    return "".join(difflib.unified_diff(
        expected.splitlines(keepends=True),
        actual.splitlines(keepends=True),
        fromfile=f"golden/{label}",
        tofile=f"actual/{label}",
        lineterm="",
    ))
