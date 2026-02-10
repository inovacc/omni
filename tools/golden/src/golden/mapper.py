"""Test map: categorized, enriched view of all test cases."""

from __future__ import annotations

import json
from collections import Counter
from dataclasses import dataclass

from golden.types import TestCase


@dataclass
class TestMapEntry:
    """Enriched test case for mapping."""
    name: str
    category: str
    has_stdin: bool
    has_fixture: bool
    normalizations: list[str]
    platform_specific: bool


def build_test_map(cases: list[TestCase]) -> list[TestMapEntry]:
    """Convert test cases to map entries."""
    return [
        TestMapEntry(
            name=tc.name,
            category=tc.category,
            has_stdin=tc.stdin is not None,
            has_fixture=tc.fixture is not None,
            normalizations=tc.normalizations,
            platform_specific=tc.platform_specific,
        )
        for tc in cases
    ]


def format_json(entries: list[TestMapEntry]) -> str:
    """JSON output with summary."""
    by_category = Counter(e.category for e in entries)
    data = {
        "summary": {
            "total": len(entries),
            "by_category": dict(sorted(by_category.items())),
        },
        "tests": [
            {
                "name": e.name,
                "category": e.category,
                "has_stdin": e.has_stdin,
                "has_fixture": e.has_fixture,
                "normalizations": e.normalizations,
                "platform_specific": e.platform_specific,
            }
            for e in entries
        ],
    }
    return json.dumps(data, indent=2)


def format_table(entries: list[TestMapEntry]) -> str:
    """Human-readable table."""
    lines = []
    lines.append(f"{'Name':<35} {'Category':<18} {'Input':<10} {'Normalizations'}")
    lines.append("-" * 85)
    for e in entries:
        input_type = "stdin" if e.has_stdin else ("file" if e.has_fixture else "args")
        norms = ", ".join(e.normalizations) if e.normalizations else "-"
        plat = " [platform]" if e.platform_specific else ""
        lines.append(f"{e.name:<35} {e.category:<18} {input_type:<10} {norms}{plat}")

    by_cat = Counter(e.category for e in entries)
    lines.append("")
    lines.append(f"Total: {len(entries)} tests across {len(by_cat)} categories")
    for cat, count in sorted(by_cat.items()):
        lines.append(f"  {cat}: {count}")

    return "\n".join(lines)
