"""Data types for golden master testing."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Optional


@dataclass
class TestCase:
    """A single golden master test definition."""
    name: str
    category: str
    args: list[str]
    stdin: Optional[str] = None
    fixture: Optional[str] = None
    normalizations: list[str] = field(default_factory=list)
    platform_specific: bool = False

    @property
    def golden_key(self) -> str:
        return f"{self.category}/{self.name}"

    @property
    def golden_filename(self) -> str:
        return self.name

    @property
    def golden_subdir(self) -> str:
        return self.category


@dataclass
class RunResult:
    """Result of running a single test case."""
    test_case: TestCase
    stdout: str
    stderr: str
    exit_code: int
    duration_ms: float


@dataclass
class CompareResult:
    """Result of comparing a test run against its golden master."""
    test_case: TestCase
    status: str  # match, mismatch, missing, error, new
    diff: str = ""
    error_msg: str = ""
