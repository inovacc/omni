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
}


def normalize(text: str, normalizer_names: list[str] | None = None) -> str:
    """Apply normalization chain. Always normalizes newlines first."""
    text = NORMALIZERS["normalize_newlines"](text)
    for name in (normalizer_names or []):
        fn = NORMALIZERS.get(name)
        if fn:
            text = fn(text)
    return text
