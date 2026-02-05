"""Test helpers for black-box testing omni binary."""

import os
import subprocess
import tempfile
import shutil
from pathlib import Path
from typing import Optional
from dataclasses import dataclass, field

# Colors for terminal output
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'  # No Color


@dataclass
class TestResult:
    name: str
    passed: bool
    message: str = ""


@dataclass
class TestSuite:
    name: str
    results: list[TestResult] = field(default_factory=list)

    @property
    def passed(self) -> int:
        return sum(1 for r in self.results if r.passed)

    @property
    def failed(self) -> int:
        return sum(1 for r in self.results if not r.passed)

    @property
    def total(self) -> int:
        return len(self.results)


class OmniTester:
    """Test harness for omni binary."""

    def __init__(self, binary_path: Optional[str] = None):
        self.binary = binary_path or os.environ.get("OMNI_BIN", "./omni")
        if os.name == "nt" and not self.binary.endswith(".exe"):
            self.binary += ".exe"
        self.temp_dir = tempfile.mkdtemp(prefix="omni_test_")
        self.suite = TestSuite(name="default")

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.cleanup()

    def cleanup(self):
        """Remove temporary test files."""
        if os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)

    def check_binary(self) -> bool:
        """Check if binary exists and is executable."""
        if not os.path.exists(self.binary):
            print(f"{Colors.RED}ERROR{Colors.NC}: Binary not found at {self.binary}")
            print("Build it first: go build -o omni .")
            return False
        return True

    def run(self, *args, stdin: Optional[str] = None, check: bool = False) -> subprocess.CompletedProcess:
        """Run omni with given arguments."""
        cmd = [self.binary] + list(args)
        return subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            input=stdin,
            check=check
        )

    def create_temp_file(self, content: str, filename: str = "testfile.txt") -> Path:
        """Create a temporary file with given content."""
        filepath = Path(self.temp_dir) / filename
        filepath.parent.mkdir(parents=True, exist_ok=True)
        filepath.write_text(content)
        return filepath

    def test(self, name: str):
        """Decorator for test functions."""
        def decorator(func):
            def wrapper():
                try:
                    func()
                    self.suite.results.append(TestResult(name=name, passed=True))
                    print(f"  {name} ... {Colors.GREEN}PASS{Colors.NC}")
                except AssertionError as e:
                    self.suite.results.append(TestResult(name=name, passed=False, message=str(e)))
                    print(f"  {name} ... {Colors.RED}FAIL{Colors.NC}")
                    print(f"    {e}")
                except Exception as e:
                    self.suite.results.append(TestResult(name=name, passed=False, message=str(e)))
                    print(f"  {name} ... {Colors.RED}ERROR{Colors.NC}")
                    print(f"    {e}")
            return wrapper
        return decorator

    def print_summary(self):
        """Print test summary."""
        print()
        print("=" * 40)
        print(f"Tests run:    {self.suite.total}")
        print(f"Tests passed: {Colors.GREEN}{self.suite.passed}{Colors.NC}")
        print(f"Tests failed: {Colors.RED}{self.suite.failed}{Colors.NC}")
        print("=" * 40)

    def exit_code(self) -> int:
        """Return exit code based on test results."""
        return 1 if self.suite.failed > 0 else 0


def assert_eq(actual, expected, message: str = ""):
    """Assert two values are equal."""
    if actual != expected:
        raise AssertionError(f"{message}\n  Expected: {expected}\n  Actual:   {actual}")


def assert_contains(haystack: str, needle: str, message: str = ""):
    """Assert string contains substring."""
    if needle not in haystack:
        raise AssertionError(f"{message}\n  Expected to contain: {needle}\n  Actual: {haystack[:200]}")


def assert_not_contains(haystack: str, needle: str, message: str = ""):
    """Assert string does not contain substring."""
    if needle in haystack:
        raise AssertionError(f"{message}\n  Expected NOT to contain: {needle}\n  Actual: {haystack[:200]}")


def assert_exit_code(result: subprocess.CompletedProcess, expected: int, message: str = ""):
    """Assert command exit code."""
    if result.returncode != expected:
        raise AssertionError(f"{message}\n  Expected exit code: {expected}\n  Actual: {result.returncode}\n  Stderr: {result.stderr}")


def assert_line_count(text: str, expected: int, message: str = ""):
    """Assert number of lines in text."""
    lines = text.strip().split('\n') if text.strip() else []
    actual = len(lines)
    if actual != expected:
        raise AssertionError(f"{message}\n  Expected {expected} lines, got {actual}")
