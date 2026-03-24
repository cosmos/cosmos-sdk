"""
Spec loader and chain scanner.

Reads YAML migration specs and provides detection/verification logic
that all MCP tools share.
"""

from __future__ import annotations

import os
import re
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import yaml


# ── Paths ────────────────────────────────────────────────────────────────────

MCP_DIR = Path(__file__).parent
MIGRATION_DIR = MCP_DIR.parent
REPO_ROOT = MIGRATION_DIR.parent.parent
SPEC_DIR = MIGRATION_DIR / "migration-spec" / "v50-to-v54"
UPGRADING_FILE = REPO_ROOT / "UPGRADING.md"
CHANGELOG_FILE = REPO_ROOT / "CHANGELOG.md"
RELEASE_NOTES_FILE = REPO_ROOT / "RELEASE_NOTES.md"


# ── Spec loading ──────────────────────────────────────────────────────────────


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
    detection_files: list[str] = field(default_factory=list)
    detection_go_mod: list[str] = field(default_factory=list)

    # Changes
    changes: dict[str, Any] = field(default_factory=dict)

    # Manual steps
    manual_steps: list[dict[str, str]] = field(default_factory=list)

    # Verification
    verification: dict[str, Any] = field(default_factory=dict)

    @property
    def has_fatal_warnings(self) -> bool:
        warnings = self.changes.get("imports", {}).get("warnings", [])
        return any(w.get("fatal", False) for w in warnings)

    @property
    def fatal_message(self) -> str:
        warnings = self.changes.get("imports", {}).get("warnings", [])
        for warning in warnings:
            if warning.get("fatal", False):
                return warning.get("message", "")
        return ""


def _normalize_manual_steps(raw_steps: Any) -> list[dict[str, str]]:
    """Accept both string and object manual-step forms."""
    normalized: list[dict[str, str]] = []
    for index, step in enumerate(raw_steps or [], start=1):
        if isinstance(step, str):
            description = step.strip()
            if description:
                normalized.append(
                    {
                        "id": f"manual-step-{index}",
                        "description": description,
                    }
                )
            continue

        if not isinstance(step, dict):
            continue

        description = str(step.get("description", "")).strip()
        if not description:
            continue

        normalized.append(
            {
                "id": str(step.get("id") or f"manual-step-{index}"),
                "description": description,
            }
        )

    return normalized


def load_specs(spec_dir: Path | None = None) -> list[Spec]:
    """Load all YAML specs from the spec directory."""
    directory = spec_dir or SPEC_DIR
    specs = []
    for spec_file in sorted(directory.glob("*.yaml")):
        with spec_file.open() as handle:
            raw = yaml.safe_load(handle)

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
                detection_files=detection.get("files", []),
                detection_go_mod=detection.get("go_mod", []),
                changes=raw.get("changes", {}),
                manual_steps=_normalize_manual_steps(raw.get("manual_steps", [])),
                verification=raw.get("verification", {}),
            )
        )
    return specs


# ── Spec ordering ─────────────────────────────────────────────────────────────

SPEC_ORDER = [
    "group-enterprise-migration",     # fatal check first
    "core-sdk-migration",             # broad dependency/import rewrites first
    "store-v2-migration",
    "bank-endblock-order",
    "crisis-removal",
    "circuit-contrib-migration",
    "nft-contrib-migration",
    "gov-keeper-migration",
    "gov-hooks-proposer-arg",
    "epochs-keeper-pointer",
    "epochs-app-module-pointer",
    "ante-handler-simplification",
    "app-structure-cleanup",
]


def order_specs(specs: list[Spec]) -> list[Spec]:
    """Sort specs into the correct application order."""
    order_map = {spec_id: index for index, spec_id in enumerate(SPEC_ORDER)}
    return sorted(specs, key=lambda spec: order_map.get(spec.id, 999))


# ── File discovery ────────────────────────────────────────────────────────────


def _iter_files(
    directory: str,
    *,
    suffixes: tuple[str, ...] | None = None,
    include_names: set[str] | None = None,
) -> list[str]:
    """List files under a directory tree with optional filtering."""
    results: list[str] = []
    include_names = include_names or set()
    for root, _, files in os.walk(directory):
        for name in files:
            if suffixes and not name.endswith(suffixes):
                if name not in include_names:
                    continue
            results.append(os.path.join(root, name))
    return sorted(results)


def list_go_files(directory: str) -> list[str]:
    """List Go source files under a directory."""
    return _iter_files(directory, suffixes=(".go",))


def list_go_mod_files(directory: str) -> list[str]:
    """List go.mod files under a directory."""
    return _iter_files(directory, include_names={"go.mod"})


def _read_file(path: str) -> str:
    return Path(path).read_text(errors="replace")


def _search_files(
    paths: list[str],
    pattern: str,
    *,
    literal: bool = False,
    max_results: int = 50,
) -> list[str]:
    """Search a list of files for a pattern."""
    if not pattern:
        return []

    regex = re.compile(re.escape(pattern) if literal else pattern, re.MULTILINE)
    matches: list[str] = []

    for path in paths:
        try:
            content = _read_file(path)
        except OSError:
            continue

        if regex.search(content):
            matches.append(path)
            if len(matches) >= max_results:
                break

    return matches


def _grep_dir(pattern: str, directory: str, max_results: int = 50) -> list[str]:
    """Search for a regex pattern in .go files under directory."""
    return _search_files(
        list_go_files(directory),
        pattern,
        literal=False,
        max_results=max_results,
    )


def _find_named_files(directory: str, names: list[str], max_results: int = 50) -> list[str]:
    """Find files by exact filename match."""
    targets = set(names)
    matches: list[str] = []
    for path in _iter_files(directory):
        if os.path.basename(path) in targets:
            matches.append(path)
            if len(matches) >= max_results:
                break
    return matches


def _detect_sdk_version(chain_dir: str) -> str:
    """Parse go.mod to find the SDK version."""
    go_mod_files = list_go_mod_files(chain_dir)
    if not go_mod_files:
        return "unknown"

    content = _read_file(go_mod_files[0])
    for line in content.splitlines():
        stripped = line.strip()
        if "github.com/cosmos/cosmos-sdk" not in stripped or stripped.startswith("//"):
            continue
        for part in stripped.split():
            if part.startswith("v0."):
                return part
    return "unknown"


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


def scan_chain(chain_dir: str, specs: list[Spec] | None = None) -> ScanResult:
    """
    Scan a chain directory and determine which specs apply.

    Returns a structured result with SDK version, applicable specs,
    warnings, and fatal blocks.
    """
    specs = specs or load_specs()

    sdk_version = _detect_sdk_version(chain_dir)
    applicable: list[str] = []
    warnings: list[dict[str, str]] = []
    fatal_blocks: list[dict[str, str]] = []
    details: dict[str, dict[str, list[str]]] = {}

    go_files = list_go_files(chain_dir)
    go_mod_files = list_go_mod_files(chain_dir)

    for spec in order_specs(specs):
        matched_imports = _search_files(
            go_files,
            "|".join(re.escape(imp) for imp in spec.detection_imports),
            max_results=50,
        ) if spec.detection_imports else []

        matched_patterns = _search_files(
            go_files,
            "|".join(re.escape(pattern) for pattern in spec.detection_patterns),
            max_results=50,
        ) if spec.detection_patterns else []

        matched_files = _find_named_files(chain_dir, spec.detection_files) if spec.detection_files else []

        matched_go_mod = []
        if spec.detection_go_mod:
            matched_go_mod = _search_files(
                go_mod_files,
                "|".join(re.escape(entry) for entry in spec.detection_go_mod),
                max_results=50,
            )

        if matched_imports or matched_patterns or matched_files or matched_go_mod:
            applicable.append(spec.id)
            details[spec.id] = {
                "matched_imports": [os.path.relpath(path, chain_dir) for path in matched_imports],
                "matched_patterns": [os.path.relpath(path, chain_dir) for path in matched_patterns],
                "matched_files": [os.path.relpath(path, chain_dir) for path in matched_files],
                "matched_go_mod": [os.path.relpath(path, chain_dir) for path in matched_go_mod],
            }

            if spec.has_fatal_warnings:
                fatal_blocks.append(
                    {"spec_id": spec.id, "message": spec.fatal_message}
                )

            for warning in spec.changes.get("imports", {}).get("warnings", []):
                if not warning.get("fatal", False):
                    warnings.append(
                        {"spec_id": spec.id, "message": warning.get("message", "")}
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
    failures: list[str] = []
    verification = spec.verification
    go_files = list_go_files(chain_dir)

    for import_path in verification.get("must_not_import", []):
        files = _search_files(go_files, f'"{import_path}"', literal=True)
        if files:
            short = [os.path.relpath(path, chain_dir) for path in files[:3]]
            failures.append(
                f"must_not_import '{import_path}' found in: {', '.join(short)}"
            )

    for entry in verification.get("must_not_contain", []):
        pattern = entry if isinstance(entry, str) else entry.get("pattern", "")
        file_match = entry.get("file_match", "") if isinstance(entry, dict) else ""
        if not pattern:
            continue

        files = _search_files(go_files, pattern, literal=True)
        if file_match:
            files = [path for path in files if os.path.relpath(path, chain_dir).endswith(file_match)]
        if files:
            short = [os.path.relpath(path, chain_dir) for path in files[:3]]
            failures.append(
                f"must_not_contain '{pattern}' found in: {', '.join(short)}"
            )

    for entry in verification.get("must_contain", []):
        pattern = entry if isinstance(entry, str) else entry.get("pattern", "")
        file_match = entry.get("file_match", "") if isinstance(entry, dict) else ""
        if not pattern:
            continue

        files = _search_files(go_files, pattern, literal=True)
        if file_match:
            files = [path for path in files if os.path.relpath(path, chain_dir).endswith(file_match)]
        if not files:
            failures.append(f"must_contain '{pattern}' not found anywhere")

    return VerifyResult(
        spec_id=spec.id,
        passed=not failures,
        failures=failures,
    )
