#!/usr/bin/env python3
"""CI gate: validate the SLSA Provenance v1 predicate that `omni attest` emits.

Generates a local-builder attestation with the committed attest fixtures, decodes
the DSSE payload, extracts the predicate, and checks:

  1. the predicate matches testing/schemas/slsa-provenance-v1.schema.json
     (via the `jsonschema` package when installed; otherwise a pure-stdlib
     structural check of the required fields), and
  2. the emitted builder.id is exactly an ADR-0009-allowlisted value.

This runs in CI only (never inside the omni binary). It dogfoods the real
`omni attest` command, so any drift of the emitted predicate from the SLSA v1.0
shape — or any builder.id outside the ADR-0009 allowlist — fails the build.

Exit 0 = valid; non-zero = drift/overclaim (prints the reason).
"""
from __future__ import annotations

import base64
import json
import os
import subprocess
import sys
from pathlib import Path

# ADR-0009 builder.id allowlist (must match pkg/attest constants).
BUILDER_ID_RELEASE = "https://github.com/inovacc/omni/.github/workflows/release.yml@refs/heads/main"
BUILDER_ID_LOCAL = "https://github.com/inovacc/omni/attest/local@v1"
ALLOWED_BUILDER_IDS = {BUILDER_ID_RELEASE, BUILDER_ID_LOCAL}

REPO = Path(__file__).resolve().parents[2]
FIXTURES = REPO / "testing" / "golden" / "fixtures" / "attest"
SCHEMA = REPO / "testing" / "schemas" / "slsa-provenance-v1.schema.json"
PASSPHRASE = "golden-fixture-passphrase"


def fail(msg: str) -> "NoReturn":  # type: ignore[name-defined]
    print(f"FAIL: {msg}", file=sys.stderr)
    sys.exit(1)


def omni_bin() -> str:
    for cand in (os.environ.get("OMNI_BIN"), str(REPO / "bin" / "omni"), str(REPO / "bin" / "omni.exe")):
        if cand and Path(cand).exists():
            return cand
    fail("omni binary not found (set OMNI_BIN or run `task build`)")


def generate_predicate() -> dict:
    env = dict(os.environ, OMNI_SIGN_PASSPHRASE=PASSPHRASE)
    proc = subprocess.run(
        [omni_bin(), "attest", "--key", str(FIXTURES / "test.key"),
         "--artifact", str(FIXTURES / "artifact.bin"), "--predicate-type", "slsa-provenance"],
        capture_output=True, text=True, env=env,
    )
    if proc.returncode != 0:
        fail(f"omni attest failed (exit {proc.returncode}): {proc.stderr.strip()}")
    envelope = json.loads(proc.stdout)
    payload = base64.b64decode(envelope["payload"])
    statement = json.loads(payload)
    if statement.get("predicateType") != "https://slsa.dev/provenance/v1":
        fail(f"unexpected predicateType: {statement.get('predicateType')!r}")
    predicate = statement.get("predicate")
    if not isinstance(predicate, dict):
        fail("statement has no object predicate")
    return predicate


def validate_structure(predicate: dict) -> None:
    """Validate against the schema with jsonschema if available, else stdlib."""
    schema = json.loads(SCHEMA.read_text(encoding="utf-8"))
    try:
        import jsonschema  # type: ignore
        jsonschema.validate(instance=predicate, schema=schema)
        print("  schema: validated with jsonschema")
        return
    except ImportError:
        pass
    # Pure-stdlib fallback: enforce the schema's required fields.
    bd = predicate.get("buildDefinition")
    rd = predicate.get("runDetails")
    if not isinstance(bd, dict):
        fail("predicate.buildDefinition missing/!object")
    if not isinstance(rd, dict):
        fail("predicate.runDetails missing/!object")
    if not bd.get("buildType"):
        fail("buildDefinition.buildType missing/empty")
    if not isinstance(bd.get("externalParameters"), dict):
        fail("buildDefinition.externalParameters missing/!object")
    builder = rd.get("builder")
    if not isinstance(builder, dict) or not builder.get("id"):
        fail("runDetails.builder.id missing/empty")
    print("  schema: validated with stdlib structural check (jsonschema not installed)")


def main() -> None:
    predicate = generate_predicate()
    validate_structure(predicate)
    builder_id = predicate["runDetails"]["builder"]["id"]
    if builder_id not in ALLOWED_BUILDER_IDS:
        fail(f"builder.id {builder_id!r} is not ADR-0009-allowlisted (no SLSA overclaim allowed)")
    print(f"OK: SLSA Provenance v1 predicate is schema-valid; builder.id = {builder_id}")


if __name__ == "__main__":
    main()
