"""Test execution: run the omni binary for each test case."""

from __future__ import annotations

import subprocess
import tempfile
import time
from concurrent.futures import ProcessPoolExecutor, as_completed
from pathlib import Path
from typing import Optional

from golden.normalize import normalize
from golden.types import RunResult, TestCase


def run_test_case(
    test_case: TestCase,
    binary: str,
    timeout: int = 30,
    temp_dir: Optional[str] = None,
) -> RunResult:
    """Run a single test case and capture output."""
    args = list(test_case.args)

    # Handle {file} placeholder
    temp_file = None
    if test_case.fixture is not None:
        td = temp_dir or tempfile.gettempdir()
        temp_file = Path(td) / f"{test_case.name}.txt"
        temp_file.write_text(test_case.fixture, encoding="utf-8")
        args = [str(temp_file) if a == "{file}" else a for a in args]

    cmd = [binary] + args
    start = time.monotonic()
    try:
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            input=test_case.stdin,
            timeout=timeout,
        )
        duration_ms = (time.monotonic() - start) * 1000
        stdout = normalize(result.stdout, test_case.normalizations)
        stderr = normalize(result.stderr, test_case.normalizations)
        return RunResult(
            test_case=test_case,
            stdout=stdout,
            stderr=stderr,
            exit_code=result.returncode,
            duration_ms=duration_ms,
        )
    except subprocess.TimeoutExpired:
        duration_ms = (time.monotonic() - start) * 1000
        return RunResult(
            test_case=test_case,
            stdout="",
            stderr=f"TIMEOUT: command exceeded {timeout}s",
            exit_code=-1,
            duration_ms=duration_ms,
        )
    finally:
        if temp_file and temp_file.exists():
            temp_file.unlink()


def _run_wrapper(args: tuple) -> RunResult:
    """Wrapper for ProcessPoolExecutor (top-level function for pickling)."""
    tc, binary, timeout, temp_dir = args
    return run_test_case(tc, binary, timeout, temp_dir)


def run_all(
    test_cases: list[TestCase],
    binary: str,
    timeout: int = 30,
    workers: int = 1,
    verbose: bool = False,
) -> list[RunResult]:
    """Run all test cases, optionally in parallel."""
    temp_dir = tempfile.mkdtemp(prefix="omni_golden_")
    results: list[RunResult] = []

    if workers <= 1:
        for tc in test_cases:
            result = run_test_case(tc, binary, timeout, temp_dir)
            results.append(result)
            if verbose:
                print(f"  ran {tc.golden_key} ({result.duration_ms:.0f}ms, exit={result.exit_code})")
    else:
        tasks = [(tc, binary, timeout, temp_dir) for tc in test_cases]
        with ProcessPoolExecutor(max_workers=workers) as pool:
            futures = {pool.submit(_run_wrapper, t): t[0] for t in tasks}
            for future in as_completed(futures):
                result = future.result()
                results.append(result)
                if verbose:
                    print(f"  ran {result.test_case.golden_key} ({result.duration_ms:.0f}ms)")

    # Sort results to match input order
    order = {tc.golden_key: i for i, tc in enumerate(test_cases)}
    results.sort(key=lambda r: order.get(r.test_case.golden_key, 0))

    # Cleanup temp dir
    import shutil
    shutil.rmtree(temp_dir, ignore_errors=True)

    return results
