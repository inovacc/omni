"""Test case discovery from YAML registry."""

from __future__ import annotations

from pathlib import Path
from typing import Optional

import yaml

from golden.types import TestCase


def discover_test_cases(
    registry_path: Path,
    category: Optional[str] = None,
    pattern: Optional[str] = None,
) -> list[TestCase]:
    """Load and filter test cases from the YAML registry."""
    with open(registry_path, "r", encoding="utf-8") as f:
        data = yaml.safe_load(f)

    cases: list[TestCase] = []
    for cat_entry in data.get("categories", []):
        cat_name = cat_entry["name"]
        if category and cat_name != category:
            continue
        for test in cat_entry.get("tests", []):
            tc = TestCase(
                name=test["name"],
                category=cat_name,
                args=test["args"],
                stdin=test.get("stdin"),
                fixture=test.get("fixture"),
                normalizations=test.get("normalizations", []),
                platform_specific=test.get("platform_specific", False),
            )
            if pattern and pattern not in tc.name:
                continue
            cases.append(tc)

    return cases
