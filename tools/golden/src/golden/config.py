"""Configuration constants for omni golden master testing."""

from __future__ import annotations

from pathlib import Path

# Test categories (matching golden_tests.yaml)
CATEGORIES = {
    "encoding", "hashing", "text", "text_with_files", "data",
    "format", "utils", "security", "xxd", "strings", "case_conv",
}

# Default execution settings
DEFAULT_TIMEOUT = 30
DEFAULT_WORKERS = 4
MANIFEST_FILENAME = "manifest.json"
MANIFEST_VERSION = 1

# Normalizer names available in YAML registry
NORMALIZER_NAMES = {"normalize_newlines", "strip_trailing_whitespace", "strip_path", "strip_temp_dir"}


def project_root() -> Path:
    """Navigate from tools/golden/ up to project root."""
    return Path(__file__).resolve().parent.parent.parent.parent.parent


def default_golden_dir() -> Path:
    """Default directory for golden master files."""
    return Path(__file__).resolve().parent.parent.parent / "golden_masters"


def default_binary() -> str:
    """Default path to the omni binary."""
    import os
    root = project_root()
    binary = root / "bin" / "omni"
    if os.name == "nt":
        binary = binary.with_suffix(".exe")
    return str(binary)


def default_registry() -> Path:
    """Path to the YAML test registry."""
    return Path(__file__).resolve().parent.parent.parent / "golden_tests.yaml"
