"""Normalize output for deterministic comparison."""

from __future__ import annotations

import re

# Built-in normalizer implementations
NORMALIZERS: dict[str, callable] = {
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
    # Normalizes the wall-clock-relative DB age in a stale-DB error
    # ("generated 3695h57m10s ago" -> "generated <AGE> ago") so the
    # scan_stale_db golden is deterministic across runs.
    "strip_db_age": lambda text: re.sub(
        r"generated \S+ ago",
        "generated <AGE> ago",
        text,
    ),
}


def normalize(text: str, normalizer_names: list[str] | None = None) -> str:
    """Apply normalization chain. Always normalizes newlines first."""
    text = NORMALIZERS["normalize_newlines"](text)
    for name in (normalizer_names or []):
        fn = NORMALIZERS.get(name)
        if fn:
            text = fn(text)
    return text


def apply_normalize_rules(text: str, rules: list) -> str:
    """Apply a list of {pattern, replacement} regex rules to text."""
    for rule in rules:
        text = re.sub(rule["pattern"], rule["replacement"], text)
    return text
