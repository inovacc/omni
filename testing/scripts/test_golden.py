#!/usr/bin/env python3
"""Golden master tests for omni CLI.

Modes:
  (default)         Verify all golden tests against stored snapshots
  --update [cat]    Regenerate golden files (all or specific category)
  --filter <sub>    Run only tests whose name contains substring
  --list            List all registered test cases
  --check           Verify all golden snapshot files exist (CI pre-flight)
  --verbose         Show full diffs on failure
"""

import argparse
import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from golden_engine import GoldenEngine

# Colors for terminal output
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'


def main():
    parser = argparse.ArgumentParser(description="Golden master tests for omni CLI")
    parser.add_argument("--update", nargs="?", const="__all__", metavar="CATEGORY",
                        help="Regenerate golden files (all or specific category)")
    parser.add_argument("--filter", metavar="SUBSTRING",
                        help="Run only tests whose name contains substring")
    parser.add_argument("--list", action="store_true",
                        help="List all registered test cases")
    parser.add_argument("--check", action="store_true",
                        help="Verify all golden snapshot files exist")
    parser.add_argument("--verbose", action="store_true",
                        help="Show full diffs on failure")
    args = parser.parse_args()

    # Paths
    project_root = Path(__file__).parent.parent.parent
    golden_base = Path(__file__).parent.parent / "golden"
    omni_bin = os.environ.get("OMNI_BIN", str(project_root / "bin" / "omni"))

    engine = GoldenEngine(binary_path=omni_bin, golden_base=str(golden_base))

    try:
        tests = engine.load_registry()
    except FileNotFoundError:
        print(f"{Colors.RED}ERROR{Colors.NC}: Registry not found at {engine.registry_file}")
        sys.exit(1)

    # --list mode
    if args.list:
        print(f"{Colors.CYAN}=== Golden Test Registry ==={Colors.NC}")
        current_cat = ""
        for t in tests:
            if t.category != current_cat:
                current_cat = t.category
                print(f"\n  {Colors.YELLOW}{current_cat}{Colors.NC}:")
            stdin_tag = " (stdin)" if t.stdin is not None else ""
            file_tag = " (file)" if t.fixture is not None else ""
            print(f"    {t.name}{stdin_tag}{file_tag}")
        print(f"\nTotal: {len(tests)} tests")
        sys.exit(0)

    # Apply filters
    if args.filter:
        tests = [t for t in tests if args.filter in t.name]
    if args.update and args.update != "__all__":
        tests = [t for t in tests if t.category == args.update]

    if not tests:
        print(f"{Colors.YELLOW}No tests matched filter.{Colors.NC}")
        sys.exit(0)

    # --check mode: verify all snapshot files exist
    if args.check:
        print(f"{Colors.CYAN}=== Golden Snapshot Check ==={Colors.NC}")
        missing = 0
        for t in tests:
            json_path = engine.snapshot_dir / t.category / f"{t.name}.json"
            if not json_path.exists():
                print(f"  {Colors.RED}MISSING{Colors.NC}: {t.category}/{t.name}")
                missing += 1
        if missing:
            print(f"\n{Colors.RED}{missing} snapshot(s) missing. Run --update to generate.{Colors.NC}")
            sys.exit(1)
        else:
            print(f"\n{Colors.GREEN}All {len(tests)} snapshots present.{Colors.NC}")
            sys.exit(0)

    # --update mode
    if args.update:
        scope = "all" if args.update == "__all__" else args.update
        print(f"{Colors.CYAN}=== Updating Golden Snapshots ({scope}) ==={Colors.NC}")
        print()
        updated = 0
        errors = 0
        for t in tests:
            result = engine.update(t)
            if result.passed:
                print(f"  {t.category}/{t.name} ... {Colors.GREEN}SAVED{Colors.NC}  {result.message}")
                updated += 1
            else:
                print(f"  {t.category}/{t.name} ... {Colors.RED}ERROR{Colors.NC}  {result.message}")
                errors += 1
        print()
        print(f"Updated: {Colors.GREEN}{updated}{Colors.NC}, Errors: {Colors.RED}{errors}{Colors.NC}")
        sys.exit(1 if errors else 0)

    # Default: verify mode
    print(f"=== Golden master tests ===")

    passed = 0
    failed = 0
    new_tests = 0

    for t in tests:
        result = engine.compare(t)
        if result.passed:
            print(f"  {t.category}/{t.name} ... {Colors.GREEN}PASS{Colors.NC}")
            passed += 1
        elif result.is_new:
            print(f"  {t.category}/{t.name} ... {Colors.YELLOW}NEW{Colors.NC} (no snapshot)")
            new_tests += 1
            failed += 1
        else:
            print(f"  {t.category}/{t.name} ... {Colors.RED}FAIL{Colors.NC}")
            print(f"    {result.message}")
            if args.verbose and result.diff:
                for line in result.diff.split("\n"):
                    if line.startswith("+") and not line.startswith("+++"):
                        print(f"    {Colors.GREEN}{line}{Colors.NC}")
                    elif line.startswith("-") and not line.startswith("---"):
                        print(f"    {Colors.RED}{line}{Colors.NC}")
                    else:
                        print(f"    {line}")
            elif result.diff:
                diff_lines = result.diff.strip().split("\n")
                shown = diff_lines[:8]
                for line in shown:
                    print(f"    {line}")
                if len(diff_lines) > 8:
                    print(f"    ... ({len(diff_lines) - 8} more lines, use --verbose)")
            failed += 1

    # Summary
    print()
    print("=" * 40)
    print(f"Tests run:    {passed + failed}")
    print(f"Tests passed: {Colors.GREEN}{passed}{Colors.NC}")
    print(f"Tests failed: {Colors.RED}{failed}{Colors.NC}")
    if new_tests:
        print(f"New tests:    {Colors.YELLOW}{new_tests}{Colors.NC} (run --update to generate snapshots)")
    print("=" * 40)

    engine.cleanup()
    sys.exit(1 if failed else 0)


if __name__ == "__main__":
    main()
