"""Golden master testing engine for omni CLI.

Captures exact command outputs as snapshots and verifies them on subsequent runs.
Detects regressions vs intentional changes by comparing stdout, stderr, and exit codes.
"""

import difflib
import json
import os
import re
import subprocess
import tempfile
import shutil
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

import yaml


@dataclass
class GoldenTestCase:
    """A single golden master test definition."""
    name: str
    category: str
    args: list[str]
    stdin: Optional[str] = None
    fixture: Optional[str] = None  # file content for {file} placeholder
    normalizations: list[str] = field(default_factory=list)
    platform_specific: bool = False


@dataclass
class GoldenSnapshot:
    """Stored golden master snapshot."""
    exit_code: int
    stdout: str
    stderr: str


@dataclass
class GoldenResult:
    """Result of comparing a test run against its snapshot."""
    name: str
    category: str
    passed: bool
    message: str = ""
    diff: str = ""
    is_new: bool = False


# Built-in normalizers: name -> callable(str) -> str
NORMALIZERS = {
    "normalize_newlines": lambda text: text.replace("\r\n", "\n"),
    "strip_trailing_whitespace": lambda text: "\n".join(
        line.rstrip() for line in text.split("\n")
    ),
    "strip_path": lambda text: re.sub(
        r"[A-Za-z]:\\[^\s\"']+|/[^\s\"']+(?:/[^\s\"']+)+",
        "<PATH>",
        text,
    ),
    "strip_temp_dir": lambda text: re.sub(
        r"(?:[A-Za-z]:\\|/)(?:tmp|temp|Temp)[/\\][^\s\"']*",
        "<TMPDIR>",
        text,
        flags=re.IGNORECASE,
    ),
}


class GoldenEngine:
    """Core engine for golden master testing."""

    def __init__(self, binary_path: str, golden_base: str):
        self.binary = binary_path
        if os.name == "nt" and not self.binary.endswith(".exe"):
            self.binary += ".exe"
        self.golden_base = Path(golden_base)
        self.snapshot_dir = self.golden_base / "snapshots"
        self.registry_file = self.golden_base / "golden_tests.yaml"
        self.temp_dir = tempfile.mkdtemp(prefix="omni_golden_")

    def load_registry(self) -> list[GoldenTestCase]:
        """Parse YAML registry into test cases."""
        with open(self.registry_file, "r", encoding="utf-8") as f:
            data = yaml.safe_load(f)

        tests = []
        for category_entry in data.get("categories", []):
            category = category_entry["name"]
            for test in category_entry.get("tests", []):
                tests.append(GoldenTestCase(
                    name=test["name"],
                    category=category,
                    args=test["args"],
                    stdin=test.get("stdin"),
                    fixture=test.get("fixture"),
                    normalizations=test.get("normalizations", []),
                    platform_specific=test.get("platform_specific", False),
                ))
        return tests

    def run_command(self, test: GoldenTestCase) -> tuple[int, str, str]:
        """Execute the command for a test case, returning (exit_code, stdout, stderr)."""
        args = list(test.args)

        # Handle {file} placeholder by creating temp file from fixture
        temp_file = None
        if test.fixture is not None:
            temp_file = Path(self.temp_dir) / f"{test.name}.txt"
            temp_file.write_text(test.fixture, encoding="utf-8")
            args = [str(temp_file) if a == "{file}" else a for a in args]

        cmd = [self.binary] + args
        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                input=test.stdin,
                timeout=30,
            )
            return result.returncode, result.stdout, result.stderr
        except subprocess.TimeoutExpired:
            return -1, "", "TIMEOUT: command exceeded 30s"
        finally:
            if temp_file and temp_file.exists():
                temp_file.unlink()

    def normalize(self, text: str, normalizer_names: list[str]) -> str:
        """Apply normalization chain to text. Always normalizes newlines."""
        # Always normalize newlines first
        text = NORMALIZERS["normalize_newlines"](text)
        for name in normalizer_names:
            if name in NORMALIZERS:
                text = NORMALIZERS[name](text)
        return text

    def _snapshot_dir_for(self, test: GoldenTestCase) -> Path:
        """Get snapshot directory for a test category."""
        return self.snapshot_dir / test.category

    def _snapshot_json_path(self, test: GoldenTestCase) -> Path:
        return self._snapshot_dir_for(test) / f"{test.name}.json"

    def _snapshot_stdout_path(self, test: GoldenTestCase) -> Path:
        return self._snapshot_dir_for(test) / f"{test.name}.stdout"

    def load_snapshot(self, test: GoldenTestCase) -> Optional[GoldenSnapshot]:
        """Load stored snapshot for a test case."""
        json_path = self._snapshot_json_path(test)
        stdout_path = self._snapshot_stdout_path(test)

        if not json_path.exists():
            return None

        with open(json_path, "r", encoding="utf-8") as f:
            meta = json.load(f)

        stdout = ""
        if stdout_path.exists():
            stdout = stdout_path.read_text(encoding="utf-8")

        return GoldenSnapshot(
            exit_code=meta["exit_code"],
            stdout=stdout,
            stderr=meta.get("stderr", ""),
        )

    def save_snapshot(self, test: GoldenTestCase, exit_code: int, stdout: str, stderr: str):
        """Write snapshot files (JSON metadata + .stdout sidecar)."""
        cat_dir = self._snapshot_dir_for(test)
        cat_dir.mkdir(parents=True, exist_ok=True)

        # Normalize before saving
        stdout = self.normalize(stdout, test.normalizations)
        stderr = self.normalize(stderr, test.normalizations)

        # JSON metadata
        meta = {
            "exit_code": exit_code,
            "stdout_file": f"{test.name}.stdout",
            "stderr": stderr,
        }
        json_path = self._snapshot_json_path(test)
        with open(json_path, "w", encoding="utf-8", newline="\n") as f:
            json.dump(meta, f, indent=2)
            f.write("\n")

        # Plain-text stdout sidecar for readable git diffs
        stdout_path = self._snapshot_stdout_path(test)
        with open(stdout_path, "w", encoding="utf-8", newline="\n") as f:
            f.write(stdout)

    def compare(self, test: GoldenTestCase) -> GoldenResult:
        """Run command, normalize, and diff against stored snapshot."""
        snapshot = self.load_snapshot(test)
        if snapshot is None:
            return GoldenResult(
                name=test.name,
                category=test.category,
                passed=False,
                message="No snapshot found. Run with --update to generate.",
                is_new=True,
            )

        exit_code, stdout, stderr = self.run_command(test)
        stdout = self.normalize(stdout, test.normalizations)
        stderr = self.normalize(stderr, test.normalizations)

        # Compare exit code
        if exit_code != snapshot.exit_code:
            return GoldenResult(
                name=test.name,
                category=test.category,
                passed=False,
                message=f"Exit code mismatch: expected {snapshot.exit_code}, got {exit_code}",
            )

        # Compare stdout
        if stdout != snapshot.stdout:
            diff = self.generate_diff(snapshot.stdout, stdout, test.name)
            return GoldenResult(
                name=test.name,
                category=test.category,
                passed=False,
                message="stdout differs from snapshot",
                diff=diff,
            )

        # Compare stderr
        if stderr != snapshot.stderr:
            diff = self.generate_diff(snapshot.stderr, stderr, f"{test.name} (stderr)")
            return GoldenResult(
                name=test.name,
                category=test.category,
                passed=False,
                message="stderr differs from snapshot",
                diff=diff,
            )

        return GoldenResult(
            name=test.name,
            category=test.category,
            passed=True,
        )

    def update(self, test: GoldenTestCase) -> GoldenResult:
        """Run command and unconditionally save as new golden file."""
        exit_code, stdout, stderr = self.run_command(test)
        self.save_snapshot(test, exit_code, stdout, stderr)
        return GoldenResult(
            name=test.name,
            category=test.category,
            passed=True,
            message=f"Snapshot saved (exit_code={exit_code})",
            is_new=True,
        )

    def generate_diff(self, expected: str, actual: str, label: str = "") -> str:
        """Generate unified diff between expected and actual."""
        expected_lines = expected.splitlines(keepends=True)
        actual_lines = actual.splitlines(keepends=True)
        diff = difflib.unified_diff(
            expected_lines,
            actual_lines,
            fromfile=f"snapshot/{label}",
            tofile=f"actual/{label}",
            lineterm="",
        )
        return "".join(diff)

    def cleanup(self):
        """Remove temp directory."""
        if os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)
