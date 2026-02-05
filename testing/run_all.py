#!/usr/bin/env python3
"""Main test runner for all black-box tests."""

import sys
import subprocess
import os
from pathlib import Path

# Colors
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'


def main():
    script_dir = Path(__file__).parent
    project_root = script_dir.parent
    os.chdir(project_root)

    # Configuration
    omni_bin = os.environ.get("OMNI_BIN", "./omni")
    if os.name == "nt" and not omni_bin.endswith(".exe"):
        omni_bin += ".exe"

    print(f"{Colors.CYAN}================================{Colors.NC}")
    print(f"{Colors.CYAN}  Omni Black-Box Test Suite{Colors.NC}")
    print(f"{Colors.CYAN}================================{Colors.NC}")
    print()

    # Check if binary exists, suggest build if not
    if not os.path.exists(omni_bin):
        print(f"{Colors.YELLOW}Binary not found at {omni_bin}{Colors.NC}")
        print("Building...")
        result = subprocess.run(["go", "build", "-o", "omni", "."], capture_output=True, text=True)
        if result.returncode != 0:
            print(f"{Colors.RED}Build failed:{Colors.NC}")
            print(result.stderr)
            sys.exit(1)
        print(f"{Colors.GREEN}Build complete{Colors.NC}")

    # Show version
    try:
        result = subprocess.run([omni_bin, "version"], capture_output=True, text=True)
        print(f"Testing: {result.stdout.strip()}")
    except Exception:
        print("Testing: unknown version")
    print()

    # Track results
    total_suites = 0
    passed_suites = 0
    failed_suites = 0

    # Find and run all test scripts
    test_scripts = sorted(script_dir.glob("scripts/test_*.py"))

    for script in test_scripts:
        name = script.stem
        total_suites += 1

        print(f"{Colors.CYAN}Running: {name}{Colors.NC}")
        print("-" * 40)

        env = os.environ.copy()
        env["OMNI_BIN"] = omni_bin

        result = subprocess.run(
            [sys.executable, str(script)],
            env=env,
            cwd=project_root
        )

        if result.returncode == 0:
            passed_suites += 1
        else:
            failed_suites += 1
            print(f"{Colors.RED}Suite {name} FAILED{Colors.NC}")

        print()

    # Final summary
    print(f"{Colors.CYAN}================================{Colors.NC}")
    print(f"{Colors.CYAN}  Final Summary{Colors.NC}")
    print(f"{Colors.CYAN}================================{Colors.NC}")
    print(f"Test suites run:    {total_suites}")
    print(f"Suites passed:      {Colors.GREEN}{passed_suites}{Colors.NC}")
    print(f"Suites failed:      {Colors.RED}{failed_suites}{Colors.NC}")

    if failed_suites > 0:
        print(f"\n{Colors.RED}TESTS FAILED{Colors.NC}")
        sys.exit(1)
    else:
        print(f"\n{Colors.GREEN}ALL TESTS PASSED{Colors.NC}")
        sys.exit(0)


if __name__ == "__main__":
    main()
