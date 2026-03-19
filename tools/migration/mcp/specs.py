"""
Spec loader and chain scanner.

Reads YAML migration specs and provides detection/verification logic
that all MCP tools share.
"""

from __future__ import annotations

import os
import re
import subprocess
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import yaml


# ── Spec loading ──────────────────────────────────────────────────────────────

SPEC_DIR = Path(__file__).parent.parent / "migration-spec" / "v50-to-v54"


@dataclass
class Spec:
    """A parsed migration spec."""

    id: str
    title: str
    version: str
    description: str
    raw: dict[str, Any]

    # Detection
    detection_imports: list[str] = field(default_factory=list)
    detection_patterns: list[str] = field(default_factory=list)

    # Changes
    changes: dict[str, Any] = field(default_factory=dict)

    # Manual steps
    manual_steps: list[dict[str, Any]] = field(default_factory=list)

    # Verification
    verification: dict[str, Any] = field(default_factory=dict)

    @property
    def has_fatal_warnings(self) -> bool:
        warnings = self.changes.get("imports", {}).get("warnings", [])
        return any(w.get("fatal", False) for w in warnings)

    @property
    def fatal_message(self) -> str:
        warnings = self.changes.get("imports", {}).get("warnings", [])
        for w in warnings:
            if w.get("fatal", False):
                return w.get("message", "")
        return ""


def load_specs(spec_dir: Path | None = None) -> list[Spec]:
    """Load all YAML specs from the spec directory."""
    d = spec_dir or SPEC_DIR
    specs = []
    for f in sorted(d.glob("*.yaml")):
        with open(f) as fh:
            raw = yaml.safe_load(fh)
        detection = raw.get("detection", {})
        specs.append(
            Spec(
                id=raw["id"],
                title=raw["title"],
                version=raw.get("version", ""),
                description=raw.get("description", ""),
                raw=raw,
                detection_imports=detection.get("imports", []),
                detection_patterns=detection.get("patterns", []),
                changes=raw.get("changes", {}),
                manual_steps=raw.get("manual_steps", []),
                verification=raw.get("verification", {}),
            )
        )
    return specs


# ── Spec ordering ─────────────────────────────────────────────────────────────

SPEC_ORDER = [
    "group-enterprise-migration",       # fatal check first
    "core-sdk-migration",               # import rewrites before module-specific specs
    "crisis-removal",
    "circuit-contrib-migration",
    "nft-contrib-migration",
    "gov-keeper-migration",
    "epochs-keeper-pointer",
    "ante-handler-simplification",
    "app-structure-cleanup",            # cleanup last
]


def order_specs(specs: list[Spec]) -> list[Spec]:
    """Sort specs into the correct application order."""
    order_map = {sid: i for i, sid in enumerate(SPEC_ORDER)}
    return sorted(specs, key=lambda s: order_map.get(s.id, 999))


# ── Chain scanning ────────────────────────────────────────────────────────────

@dataclass
class ScanResult:
    """Result of scanning a chain directory."""

    chain_dir: str
    sdk_version: str
    applicable_specs: list[str]
    warnings: list[dict[str, str]]
    fatal_blocks: list[dict[str, str]]
    detection_details: dict[str, dict[str, list[str]]]


def _grep_dir(pattern: str, directory: str, max_results: int = 50) -> list[str]:
    """Search for a pattern in .go files under directory."""
    try:
        result = subprocess.run(
            ["grep", "-rl", "--include=*.go", pattern, directory],
            capture_output=True,
            text=True,
            timeout=30,
        )
        files = [f for f in result.stdout.strip().split("\n") if f]
        return files[:max_results]
    except Exception:
        return []


def _detect_sdk_version(chain_dir: str) -> str:
    """Parse go.mod to find the SDK version."""
    go_mod = Path(chain_dir) / "go.mod"
    if not go_mod.exists():
        # Search subdirectories
        for sub in Path(chain_dir).rglob("go.mod"):
            go_mod = sub
            break
    if not go_mod.exists():
        return "unknown"

    content = go_mod.read_text()
    for line in content.splitlines():
        line = line.strip()
        # Match: github.com/cosmos/cosmos-sdk v0.50.x
        if "github.com/cosmos/cosmos-sdk" in line and not line.startswith("//"):
            parts = line.split()
            for part in parts:
                if part.startswith("v0."):
                    return part
    return "unknown"


def scan_chain(chain_dir: str, specs: list[Spec] | None = None) -> ScanResult:
    """
    Scan a chain directory and determine which specs apply.

    Returns a structured result with SDK version, applicable specs,
    warnings, and fatal blocks.
    """
    if specs is None:
        specs = load_specs()

    sdk_version = _detect_sdk_version(chain_dir)
    applicable = []
    warnings = []
    fatal_blocks = []
    details: dict[str, dict[str, list[str]]] = {}

    for spec in order_specs(specs):
        matched_imports: list[str] = []
        matched_patterns: list[str] = []

        # Check detection imports
        for imp in spec.detection_imports:
            files = _grep_dir(imp, chain_dir)
            if files:
                matched_imports.append(imp)

        # Check detection patterns
        for pat in spec.detection_patterns:
            files = _grep_dir(re.escape(pat), chain_dir)
            if files:
                matched_patterns.append(pat)

        if matched_imports or matched_patterns:
            applicable.append(spec.id)
            details[spec.id] = {
                "matched_imports": matched_imports,
                "matched_patterns": matched_patterns,
            }

            # Check for fatal warnings
            if spec.has_fatal_warnings:
                fatal_blocks.append(
                    {"spec_id": spec.id, "message": spec.fatal_message}
                )

            # Check for non-fatal warnings
            import_warnings = spec.changes.get("imports", {}).get("warnings", [])
            for w in import_warnings:
                if not w.get("fatal", False):
                    warnings.append(
                        {"spec_id": spec.id, "message": w.get("message", "")}
                    )

    return ScanResult(
        chain_dir=chain_dir,
        sdk_version=sdk_version,
        applicable_specs=applicable,
        warnings=warnings,
        fatal_blocks=fatal_blocks,
        detection_details=details,
    )


# ── Verification ──────────────────────────────────────────────────────────────

@dataclass
class VerifyResult:
    """Result of verifying a spec against a directory."""

    spec_id: str
    passed: bool
    failures: list[str]


def verify_spec(spec: Spec, chain_dir: str) -> VerifyResult:
    """Run verification checks for a single spec."""
    failures = []
    v = spec.verification

    # must_not_import
    for imp in v.get("must_not_import", []):
        files = _grep_dir(f'"{imp}', chain_dir)
        if files:
            short = [os.path.relpath(f, chain_dir) for f in files[:3]]
            failures.append(f"must_not_import '{imp}' found in: {', '.join(short)}")

    # must_not_contain
    for entry in v.get("must_not_contain", []):
        pattern = entry if isinstance(entry, str) else entry.get("pattern", "")
        file_match = entry.get("file_match", "") if isinstance(entry, dict) else ""
        if not pattern:
            continue
        files = _grep_dir(pattern, chain_dir)
        if file_match:
            files = [f for f in files if f.endswith(file_match)]
        if files:
            short = [os.path.relpath(f, chain_dir) for f in files[:3]]
            failures.append(
                f"must_not_contain '{pattern}' found in: {', '.join(short)}"
            )

    # must_contain
    for entry in v.get("must_contain", []):
        pattern = entry if isinstance(entry, str) else entry.get("pattern", "")
        if not pattern:
            continue
        files = _grep_dir(pattern, chain_dir)
        if not files:
            failures.append(f"must_contain '{pattern}' not found anywhere")

    return VerifyResult(
        spec_id=spec.id, passed=len(failures) == 0, failures=failures
    )
