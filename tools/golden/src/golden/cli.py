"""CLI entry point for golden master testing."""

from __future__ import annotations

import argparse
import os
import sys
from pathlib import Path

from golden.comparator import compare_results
from golden.config import DEFAULT_TIMEOUT, DEFAULT_WORKERS, default_binary, default_golden_dir, default_registry
from golden.discovery import discover_test_cases
from golden.mapper import build_test_map, format_json, format_table
from golden.recorder import record_golden
from golden.report import Colors, print_report
from golden.runner import run_all


def main() -> int:
    parser = argparse.ArgumentParser(
        prog="golden",
        description="Golden master testing for omni CLI",
    )
    parser.add_argument("--binary", default=None, help="Path to omni binary")
    parser.add_argument("--golden-dir", default=None, help="Golden masters directory")
    parser.add_argument("--registry", default=None, help="Path to golden_tests.yaml")
    parser.add_argument("--workers", type=int, default=1, help=f"Parallel workers (default: 1)")
    parser.add_argument("--timeout", type=int, default=DEFAULT_TIMEOUT, help=f"Per-test timeout (default: {DEFAULT_TIMEOUT}s)")
    parser.add_argument("--verbose", "-v", action="store_true", help="Verbose output")
    parser.add_argument("--category", default=None, help="Filter by category")
    parser.add_argument("--pattern", default=None, help="Filter by name substring")
    parser.add_argument("--incremental", action="store_true", help="Only run changed test cases")

    sub = parser.add_subparsers(dest="command")

    sub.add_parser("record", help="Record golden master baselines")
    sub.add_parser("compare", help="Compare against golden masters")
    sub.add_parser("list", help="List test cases")

    update_parser = sub.add_parser("update", help="Re-record matching test cases")
    update_parser.add_argument("update_pattern", nargs="?", default=None, help="Name pattern to update")

    map_parser = sub.add_parser("map", help="Generate test map")
    map_parser.add_argument("--format", choices=["json", "table"], default="table", dest="map_format")
    map_parser.add_argument("--output", "-o", default=None, help="Output file (default: stdout)")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return 1

    # Resolve paths
    binary = args.binary or os.environ.get("OMNI_BIN") or default_binary()
    if os.name == "nt" and not binary.endswith(".exe"):
        binary += ".exe"
    golden_dir = Path(args.golden_dir) if args.golden_dir else default_golden_dir()
    registry = Path(args.registry) if args.registry else default_registry()

    # Discover test cases
    try:
        cases = discover_test_cases(registry, category=args.category, pattern=args.pattern)
    except FileNotFoundError:
        print(f"{Colors.RED}ERROR{Colors.NC}: Registry not found at {registry}")
        return 1

    if args.command == "update" and args.update_pattern:
        cases = [tc for tc in cases if args.update_pattern in tc.name]

    if not cases:
        print(f"{Colors.YELLOW}No test cases found.{Colors.NC}")
        return 0

    # Incremental filtering
    if args.incremental and args.command in ("record", "compare"):
        from golden.incremental import filter_changed_cases
        original = len(cases)
        cases = filter_changed_cases(cases, golden_dir)
        if args.verbose:
            print(f"Incremental: {len(cases)}/{original} cases to run")

    # === LIST ===
    if args.command == "list":
        print(f"{Colors.CYAN}=== Golden Test Registry ==={Colors.NC}")
        current_cat = ""
        for tc in cases:
            if tc.category != current_cat:
                current_cat = tc.category
                print(f"\n  {Colors.YELLOW}{current_cat}{Colors.NC}:")
            tags = []
            if tc.stdin is not None:
                tags.append("stdin")
            if tc.fixture is not None:
                tags.append("file")
            tag_str = f" ({', '.join(tags)})" if tags else ""
            print(f"    {tc.name}{tag_str}")
        print(f"\nTotal: {len(cases)} tests")
        return 0

    # === MAP ===
    if args.command == "map":
        entries = build_test_map(cases)
        if args.map_format == "json":
            output = format_json(entries)
        else:
            output = format_table(entries)
        if args.output:
            Path(args.output).write_text(output + "\n", encoding="utf-8")
            print(f"Test map written to {args.output}")
        else:
            print(output)
        return 0

    # Check binary exists
    if not os.path.exists(binary):
        print(f"{Colors.RED}ERROR{Colors.NC}: Binary not found at {binary}")
        print("Build it first: task build")
        return 1

    # === RECORD / UPDATE ===
    if args.command in ("record", "update"):
        scope = args.category or "all"
        print(f"{Colors.CYAN}=== Recording Golden Masters ({scope}, {len(cases)} tests) ==={Colors.NC}")
        results = run_all(cases, binary, args.timeout, args.workers, args.verbose)
        manifest = record_golden(results, golden_dir, verbose=args.verbose)
        print(f"\n{Colors.GREEN}Recorded {len(results)} golden masters.{Colors.NC}")
        print(f"Manifest: {golden_dir / 'manifest.json'}")
        return 0

    # === COMPARE ===
    if args.command == "compare":
        print(f"=== Golden Master Tests ({len(cases)} tests) ===")
        results = run_all(cases, binary, args.timeout, args.workers, args.verbose)
        comparisons = compare_results(results, golden_dir)
        return print_report(comparisons, verbose=args.verbose)

    return 1
