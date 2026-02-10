"""Record golden master baselines."""

from __future__ import annotations

import json
from pathlib import Path

from golden.manifest import (
    ManifestEntry, Manifest, compute_sha256,
    get_git_commit, now_iso, load_manifest, save_manifest,
)
from golden.types import RunResult


def record_golden(
    results: list[RunResult],
    golden_dir: Path,
    verbose: bool = False,
) -> Manifest:
    """Write golden master files and update manifest."""
    golden_dir.mkdir(parents=True, exist_ok=True)
    manifest = load_manifest(golden_dir)
    manifest.omni_commit = get_git_commit()
    manifest.recorded_at = now_iso()

    for result in results:
        tc = result.test_case
        cat_dir = golden_dir / tc.golden_subdir
        cat_dir.mkdir(parents=True, exist_ok=True)

        # Write .stdout sidecar
        stdout_path = cat_dir / f"{tc.golden_filename}.stdout"
        with open(stdout_path, "w", encoding="utf-8", newline="\n") as f:
            f.write(result.stdout)

        # Write JSON metadata
        meta = {
            "exit_code": result.exit_code,
            "stdout_file": f"{tc.golden_filename}.stdout",
            "stderr": result.stderr,
        }
        json_path = cat_dir / f"{tc.golden_filename}.json"
        with open(json_path, "w", encoding="utf-8", newline="\n") as f:
            json.dump(meta, f, indent=2)
            f.write("\n")

        # Update manifest entry
        entry = ManifestEntry(
            category=tc.category,
            name=tc.name,
            golden_path=f"{tc.golden_subdir}/{tc.golden_filename}.stdout",
            sha256=compute_sha256(result.stdout),
            exit_code=result.exit_code,
            normalizations=tc.normalizations,
        )
        manifest.upsert_entry(entry)

        if verbose:
            print(f"  recorded {tc.golden_key} (exit={result.exit_code})")

    save_manifest(manifest, golden_dir)
    return manifest
