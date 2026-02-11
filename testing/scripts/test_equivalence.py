#!/usr/bin/env python3
"""Reference-equivalence tests (public tools as golden behavior)."""

from __future__ import annotations

import difflib
import os
import shutil
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from helpers import Colors, OmniTester


@dataclass
class EquivalenceCase:
    name: str
    reference_cmd: list[str]
    omni_args: list[str]
    stdin: str | None = None
    platforms: set[str] | None = None  # {"linux", "darwin", "windows"}


def normalize_output(text: str) -> str:
    return text.replace("\r\n", "\n")


def diff_preview(expected: str, actual: str, max_lines: int = 24) -> str:
    diff = list(
        difflib.unified_diff(
            expected.splitlines(),
            actual.splitlines(),
            fromfile="reference",
            tofile="omni",
            lineterm="",
        )
    )
    if len(diff) <= max_lines:
        return "\n".join(diff)
    return "\n".join(diff[:max_lines] + [f"... ({len(diff) - max_lines} more lines)"])


def create_fixtures(root: Path) -> None:
    (root / "src" / "pkg").mkdir(parents=True, exist_ok=True)
    (root / "src" / "main.go").write_text(
        "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n",
        encoding="utf-8",
    )
    (root / "src" / "pkg" / "helper.go").write_text(
        "package pkg\n\n// TODO: Add helpers\nfunc Helper() string {\n\treturn \"helper\"\n}\n",
        encoding="utf-8",
    )
    (root / "README.md").write_text(
        "# Project\n\nhello world\nHello World\n",
        encoding="utf-8",
    )
    (root / "names.txt").write_text(
        "charlie\nalpha\nbravo\n",
        encoding="utf-8",
    )


def main() -> None:
    print("=== Testing reference equivalence (public tools as golden) ===")

    with OmniTester() as t:
        if not t.check_binary():
            sys.exit(1)

        fixtures = Path(t.temp_dir) / "equivalence"
        create_fixtures(fixtures)
        omni_bin = str(Path(t.binary).resolve())

        cases = [
            # Cross-platform: ripgrep parity
            EquivalenceCase(
                name="rg_exact_single_file",
                reference_cmd=["rg", "Hello", "src/main.go"],
                omni_args=["rg", "Hello", "src/main.go"],
                platforms={"linux", "darwin", "windows"},
            ),
            EquivalenceCase(
                name="rg_case_insensitive_single_file",
                reference_cmd=["rg", "-i", "hello", "README.md"],
                omni_args=["rg", "-i", "hello", "README.md"],
                platforms={"linux", "darwin", "windows"},
            ),
            EquivalenceCase(
                name="rg_fixed_string_single_file",
                reference_cmd=["rg", "-F", "Helper()", "src/pkg/helper.go"],
                omni_args=["rg", "-F", "Helper()", "src/pkg/helper.go"],
                platforms={"linux", "darwin", "windows"},
            ),
            # Linux/macOS only: GNU/BSD tool parity where command availability varies on Windows.
            EquivalenceCase(
                name="grep_exact_single_file",
                reference_cmd=["grep", "-n", "Hello", "README.md"],
                omni_args=["grep", "-n", "Hello", "README.md"],
                platforms={"linux", "darwin"},
            ),
            EquivalenceCase(
                name="sort_exact_stdin",
                reference_cmd=["sort"],
                omni_args=["sort"],
                stdin="charlie\nalpha\nbravo\n",
                platforms={"linux", "darwin"},
            ),
        ]

        strict = os.environ.get("OMNI_EQUIV_STRICT", "").lower() in {"1", "true", "yes"}
        platform = sys.platform
        if platform.startswith("win"):
            platform_name = "windows"
        elif platform == "darwin":
            platform_name = "darwin"
        else:
            platform_name = "linux"

        passed = 0
        failed = 0
        skipped = 0

        for case in cases:
            allowed = case.platforms or {"linux", "darwin", "windows"}
            if platform_name not in allowed:
                print(f"  {case.name} ... {Colors.YELLOW}SKIP{Colors.NC} (platform={platform_name})")
                skipped += 1
                continue

            ref_bin = case.reference_cmd[0]
            if shutil.which(ref_bin) is None:
                if strict:
                    print(f"  {case.name} ... {Colors.RED}FAIL{Colors.NC} (missing reference tool: {ref_bin})")
                    failed += 1
                else:
                    print(f"  {case.name} ... {Colors.YELLOW}SKIP{Colors.NC} (missing reference tool: {ref_bin})")
                    skipped += 1
                continue

            ref = subprocess.run(
                case.reference_cmd,
                cwd=fixtures,
                capture_output=True,
                text=True,
                input=case.stdin,
            )
            omni = subprocess.run(
                [omni_bin] + case.omni_args,
                cwd=fixtures,
                capture_output=True,
                text=True,
                input=case.stdin,
            )

            ref_stdout = normalize_output(ref.stdout)
            omni_stdout = normalize_output(omni.stdout)
            ref_stderr = normalize_output(ref.stderr)
            omni_stderr = normalize_output(omni.stderr)

            if ref.returncode != omni.returncode:
                print(
                    f"  {case.name} ... {Colors.RED}FAIL{Colors.NC} "
                    f"(exit: reference={ref.returncode}, omni={omni.returncode})"
                )
                failed += 1
                continue

            if ref_stdout != omni_stdout:
                print(f"  {case.name} ... {Colors.RED}FAIL{Colors.NC} (stdout mismatch)")
                print(diff_preview(ref_stdout, omni_stdout))
                failed += 1
                continue

            if ref_stderr != omni_stderr:
                print(f"  {case.name} ... {Colors.RED}FAIL{Colors.NC} (stderr mismatch)")
                print(diff_preview(ref_stderr, omni_stderr))
                failed += 1
                continue

            print(f"  {case.name} ... {Colors.GREEN}PASS{Colors.NC}")
            passed += 1

        print()
        print("=" * 40)
        print(f"Cases run:    {passed + failed + skipped}")
        print(f"Cases passed: {Colors.GREEN}{passed}{Colors.NC}")
        print(f"Cases failed: {Colors.RED}{failed}{Colors.NC}")
        print(f"Cases skipped:{Colors.YELLOW}{skipped}{Colors.NC}")
        print("=" * 40)

        sys.exit(1 if failed else 0)


if __name__ == "__main__":
    main()
