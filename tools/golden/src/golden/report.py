"""Color-coded terminal reporting for golden master results."""

from __future__ import annotations

from golden.types import CompareResult


class Colors:
    RED = "\033[0;31m"
    GREEN = "\033[0;32m"
    YELLOW = "\033[1;33m"
    CYAN = "\033[0;36m"
    NC = "\033[0m"


STATUS_LABELS = {
    "match": (Colors.GREEN, "PASS"),
    "mismatch": (Colors.RED, "FAIL"),
    "new": (Colors.YELLOW, "NEW"),
    "missing": (Colors.YELLOW, "MISS"),
    "error": (Colors.RED, "ERR"),
}


def print_report(results: list[CompareResult], verbose: bool = False) -> int:
    """Print results and return exit code (0=pass, 1=mismatch, 2=error)."""
    counts = {"match": 0, "mismatch": 0, "new": 0, "missing": 0, "error": 0}

    for r in results:
        color, label = STATUS_LABELS.get(r.status, (Colors.NC, r.status.upper()))
        print(f"  {r.test_case.golden_key} ... {color}{label}{Colors.NC}")
        counts[r.status] = counts.get(r.status, 0) + 1

        if r.status == "mismatch" and r.diff:
            if verbose:
                for line in r.diff.split("\n"):
                    if line.startswith("+") and not line.startswith("+++"):
                        print(f"    {Colors.GREEN}{line}{Colors.NC}")
                    elif line.startswith("-") and not line.startswith("---"):
                        print(f"    {Colors.RED}{line}{Colors.NC}")
                    else:
                        print(f"    {line}")
            else:
                diff_lines = r.diff.strip().split("\n")
                for line in diff_lines[:6]:
                    print(f"    {line}")
                if len(diff_lines) > 6:
                    print(f"    ... ({len(diff_lines) - 6} more lines, use --verbose)")

        if r.status in ("new", "error") and r.error_msg:
            print(f"    {r.error_msg}")

    # Summary
    total = sum(counts.values())
    print()
    print("=" * 40)
    print(f"Tests run:    {total}")
    print(f"Passed:       {Colors.GREEN}{counts['match']}{Colors.NC}")
    if counts["mismatch"]:
        print(f"Mismatched:   {Colors.RED}{counts['mismatch']}{Colors.NC}")
    if counts["new"]:
        print(f"New:          {Colors.YELLOW}{counts['new']}{Colors.NC} (run 'record' to create baselines)")
    if counts["error"]:
        print(f"Errors:       {Colors.RED}{counts['error']}{Colors.NC}")
    print("=" * 40)

    if counts["error"]:
        return 2
    if counts["mismatch"] or counts["new"]:
        return 1
    return 0
