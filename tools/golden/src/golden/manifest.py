"""SHA-256 manifest for golden master tracking."""

from __future__ import annotations

import hashlib
import json
import subprocess
from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from pathlib import Path

from golden.config import MANIFEST_FILENAME, MANIFEST_VERSION


@dataclass
class ManifestEntry:
    """A single entry in the manifest."""
    category: str
    name: str
    golden_path: str
    sha256: str
    exit_code: int
    normalizations: list[str] = field(default_factory=list)


@dataclass
class Manifest:
    """Golden master manifest with SHA-256 hashes for fast comparison."""
    version: int = MANIFEST_VERSION
    omni_commit: str = ""
    recorded_at: str = ""
    entries: list[ManifestEntry] = field(default_factory=list)

    def find_entry(self, category: str, name: str) -> ManifestEntry | None:
        for e in self.entries:
            if e.category == category and e.name == name:
                return e
        return None

    def upsert_entry(self, entry: ManifestEntry) -> None:
        for i, e in enumerate(self.entries):
            if e.category == entry.category and e.name == entry.name:
                self.entries[i] = entry
                return
        self.entries.append(entry)

    def remove_entry(self, category: str, name: str) -> None:
        self.entries = [e for e in self.entries if not (e.category == category and e.name == name)]


def compute_sha256(content: str) -> str:
    """SHA-256 hash of normalized content."""
    return hashlib.sha256(content.encode("utf-8")).hexdigest()


def get_git_commit() -> str:
    """Get current git short commit hash."""
    try:
        result = subprocess.run(
            ["git", "rev-parse", "--short", "HEAD"],
            capture_output=True, text=True, timeout=5,
        )
        return result.stdout.strip() if result.returncode == 0 else "unknown"
    except Exception:
        return "unknown"


def now_iso() -> str:
    """UTC ISO timestamp."""
    return datetime.now(timezone.utc).isoformat()


def load_manifest(golden_dir: Path) -> Manifest:
    """Load manifest from disk."""
    manifest_path = golden_dir / MANIFEST_FILENAME
    if not manifest_path.exists():
        return Manifest()
    with open(manifest_path, "r", encoding="utf-8") as f:
        data = json.load(f)
    entries = [ManifestEntry(**e) for e in data.get("entries", [])]
    return Manifest(
        version=data.get("version", MANIFEST_VERSION),
        omni_commit=data.get("omni_commit", ""),
        recorded_at=data.get("recorded_at", ""),
        entries=entries,
    )


def save_manifest(manifest: Manifest, golden_dir: Path) -> None:
    """Write manifest to disk."""
    golden_dir.mkdir(parents=True, exist_ok=True)
    manifest_path = golden_dir / MANIFEST_FILENAME
    data = {
        "version": manifest.version,
        "omni_commit": manifest.omni_commit,
        "recorded_at": manifest.recorded_at,
        "entries": [asdict(e) for e in manifest.entries],
    }
    with open(manifest_path, "w", encoding="utf-8", newline="\n") as f:
        json.dump(data, f, indent=2)
        f.write("\n")
