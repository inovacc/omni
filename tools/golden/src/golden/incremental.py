"""Incremental testing: skip unchanged test cases."""

from __future__ import annotations

import hashlib
from pathlib import Path

from golden.manifest import Manifest, load_manifest
from golden.types import TestCase


def compute_input_hash(tc: TestCase) -> str:
    """Hash the test case definition (args + stdin + fixture) for change detection."""
    h = hashlib.sha256()
    h.update("|".join(tc.args).encode("utf-8"))
    if tc.stdin is not None:
        h.update(b"\x00stdin\x00")
        h.update(tc.stdin.encode("utf-8"))
    if tc.fixture is not None:
        h.update(b"\x00fixture\x00")
        h.update(tc.fixture.encode("utf-8"))
    h.update(b"\x00norms\x00")
    h.update("|".join(tc.normalizations).encode("utf-8"))
    return h.hexdigest()


def filter_changed_cases(
    cases: list[TestCase],
    golden_dir: Path,
) -> list[TestCase]:
    """Return only test cases that have changed or have no golden master.

    A test case is considered unchanged if its golden master exists and
    the manifest records the same input hash.
    """
    manifest = load_manifest(golden_dir)
    changed: list[TestCase] = []

    for tc in cases:
        entry = manifest.find_entry(tc.category, tc.name)
        if entry is None:
            # No golden master exists, must run
            changed.append(tc)
            continue

        # Check if golden file still exists on disk
        golden_path = golden_dir / tc.golden_subdir / f"{tc.golden_filename}.json"
        if not golden_path.exists():
            changed.append(tc)
            continue

        # For now, always include (input hash tracking is future enhancement)
        # In a full implementation, compare compute_input_hash(tc) against stored hash
        changed.append(tc)

    return changed
