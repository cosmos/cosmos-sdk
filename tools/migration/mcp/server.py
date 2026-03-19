"""
Cosmos SDK Migration MCP Server.

Exposes migration tooling as MCP tools, resources, and prompts so that
AI agents (Claude, etc.) can migrate chain repos from v50+ to v54.

Usage:
    # stdio transport (for Claude Desktop, Claude Code, etc.)
    uv run cosmos-migration-mcp

    # Or directly:
    python -m server
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import shutil
from dataclasses import asdict
from pathlib import Path
from typing import Any

import yaml
from mcp.server.fastmcp import FastMCP

from specs import (
    SPEC_DIR,
    Spec,
    load_specs,
    order_specs,
    scan_chain,
    verify_spec,
    VerifyResult,
)

mcp = FastMCP(
    "cosmos-migration",
    version="0.1.0",
    description="Cosmos SDK chain migration server (v50+ → v54). "
    "Provides tools for scanning chains, planning migrations, applying "
    "specs, and verifying results.",
)


# ═══════════════════════════════════════════════════════════════════════════════
# TOOLS
# ═══════════════════════════════════════════════════════════════════════════════


@mcp.tool()
def scan_chain_tool(chain_dir: str) -> dict[str, Any]:
    """
    Scan a chain repository and detect which migration specs apply.

    Analyzes go.mod for the SDK version, then checks every spec's detection
    patterns against the codebase. Returns: SDK version, applicable specs
    (in correct application order), expected warnings, and any fatal blocks
    that would halt migration.

    Args:
        chain_dir: Absolute path to the chain repository root.
    """
    result = scan_chain(chain_dir)
    return asdict(result)


@mcp.tool()
def get_migration_plan(chain_dir: str) -> dict[str, Any]:
    """
    Generate a structured migration plan for a chain repository.

    Scans the chain, selects applicable specs, and for each spec describes:
    - What changes will be made (imports, removals, text replacements, etc.)
    - Which files will be affected
    - Expected warnings
    - Manual steps required

    Does NOT modify any files. This is a preview/dry-run.

    Args:
        chain_dir: Absolute path to the chain repository root.
    """
    specs = load_specs()
    scan = scan_chain(chain_dir, specs)

    if scan.fatal_blocks:
        return {
            "status": "blocked",
            "sdk_version": scan.sdk_version,
            "fatal_blocks": scan.fatal_blocks,
            "message": "Migration cannot proceed. Resolve fatal blocks first.",
        }

    plan_specs = []
    spec_map = {s.id: s for s in specs}

    for spec_id in scan.applicable_specs:
        spec = spec_map.get(spec_id)
        if not spec:
            continue

        changes = spec.changes
        affected_files = _estimate_affected_files(chain_dir, spec)

        plan_entry = {
            "spec_id": spec.id,
            "title": spec.title,
            "description": spec.description.strip(),
            "changes_summary": _summarize_changes(changes),
            "affected_files": affected_files,
            "manual_steps": [
                {"id": s["id"], "description": s["description"].strip()}
                for s in spec.manual_steps
            ],
            "has_warnings": any(
                not w.get("fatal", False)
                for w in changes.get("imports", {}).get("warnings", [])
            ),
        }
        plan_specs.append(plan_entry)

    return {
        "status": "ready",
        "sdk_version": scan.sdk_version,
        "specs_to_apply": len(plan_specs),
        "plan": plan_specs,
        "warnings": scan.warnings,
        "application_order": [s["spec_id"] for s in plan_specs],
    }


@mcp.tool()
def apply_spec(chain_dir: str, spec_id: str, dry_run: bool = False) -> dict[str, Any]:
    """
    Apply a single migration spec to a chain repository.

    Executes the changes defined in the spec: file removals, import rewrites
    (as text replacements), statement removals (as text replacements),
    and text replacements. Returns a summary of what was changed.

    For AST-level changes (like govkeeper.NewKeeper surgery), the tool
    returns the manual_steps with instructions — the agent should apply
    those edits directly.

    Args:
        chain_dir: Absolute path to the chain repository root.
        spec_id: The ID of the spec to apply (e.g., "crisis-removal").
        dry_run: If True, report what would change without modifying files.
    """
    specs = load_specs()
    spec = next((s for s in specs if s.id == spec_id), None)
    if not spec:
        return {"error": f"Spec '{spec_id}' not found"}

    if spec.has_fatal_warnings:
        return {
            "error": "This spec has fatal warnings and cannot be applied automatically.",
            "message": spec.fatal_message,
        }

    results: dict[str, Any] = {
        "spec_id": spec_id,
        "dry_run": dry_run,
        "file_removals": [],
        "text_replacements": [],
        "manual_steps_required": [],
        "warnings": [],
    }

    changes = spec.changes

    # ── File removals ─────────────────────────────────────────────────────
    for removal in changes.get("file_removals", []):
        fname = removal["file_name"]
        must_match = removal.get("contains_must_match", "")
        matches = _find_files(chain_dir, fname)
        for fpath in matches:
            if must_match:
                content = Path(fpath).read_text(errors="replace")
                if must_match not in content:
                    continue
            if dry_run:
                results["file_removals"].append(
                    {"file": fpath, "action": "would_delete"}
                )
            else:
                os.remove(fpath)
                results["file_removals"].append(
                    {"file": fpath, "action": "deleted"}
                )

    # ── Text replacements ─────────────────────────────────────────────────
    for repl in changes.get("text_replacements", []):
        old = repl.get("old", "")
        new = repl.get("new", "")
        file_match = repl.get("file_match", "")
        requires = repl.get("requires_contains", [])

        if not old:
            continue

        # Find candidate files
        go_files = _find_go_files(chain_dir)
        for fpath in go_files:
            basename = os.path.basename(fpath)
            if file_match and not basename.endswith(file_match):
                continue

            content = Path(fpath).read_text(errors="replace")

            # Check requires_contains
            if requires and not all(r in content for r in requires):
                continue

            if old in content:
                if dry_run:
                    results["text_replacements"].append(
                        {
                            "file": os.path.relpath(fpath, chain_dir),
                            "pattern": old[:80] + ("..." if len(old) > 80 else ""),
                            "action": "would_replace",
                        }
                    )
                else:
                    content = content.replace(old, new)
                    Path(fpath).write_text(content)
                    results["text_replacements"].append(
                        {
                            "file": os.path.relpath(fpath, chain_dir),
                            "pattern": old[:80] + ("..." if len(old) > 80 else ""),
                            "action": "replaced",
                        }
                    )

    # ── Import rewrites (as text replacement) ─────────────────────────────
    for rewrite in changes.get("imports", {}).get("rewrites", []):
        old_path = rewrite["old"]
        new_path = rewrite["new"]
        all_packages = rewrite.get("all_packages", False)

        go_files = _find_go_files(chain_dir)
        for fpath in go_files:
            content = Path(fpath).read_text(errors="replace")
            if old_path not in content:
                continue
            # Don't replace if new path already present (idempotent)
            if new_path in content and old_path not in content.replace(new_path, ""):
                continue

            if all_packages:
                updated = content.replace(old_path, new_path)
            else:
                # Exact match only (with quotes)
                updated = content.replace(f'"{old_path}"', f'"{new_path}"')

            if updated != content:
                if dry_run:
                    results["text_replacements"].append(
                        {
                            "file": os.path.relpath(fpath, chain_dir),
                            "pattern": f"import {old_path} → {new_path}",
                            "action": "would_rewrite",
                        }
                    )
                else:
                    Path(fpath).write_text(updated)
                    results["text_replacements"].append(
                        {
                            "file": os.path.relpath(fpath, chain_dir),
                            "pattern": f"import {old_path} → {new_path}",
                            "action": "rewritten",
                        }
                    )

    # ── Collect warnings ──────────────────────────────────────────────────
    for w in changes.get("imports", {}).get("warnings", []):
        if not w.get("fatal", False):
            results["warnings"].append(w.get("message", ""))

    # ── Manual steps ──────────────────────────────────────────────────────
    for step in spec.manual_steps:
        results["manual_steps_required"].append(
            {
                "id": step["id"],
                "description": step["description"].strip(),
            }
        )

    return results


@mcp.tool()
def verify_spec_tool(chain_dir: str, spec_id: str) -> dict[str, Any]:
    """
    Run verification checks for a specific spec against a chain directory.

    Checks must_not_import, must_not_contain, and must_contain rules
    defined in the spec's verification section.

    Args:
        chain_dir: Absolute path to the chain repository root.
        spec_id: The ID of the spec to verify.
    """
    specs = load_specs()
    spec = next((s for s in specs if s.id == spec_id), None)
    if not spec:
        return {"error": f"Spec '{spec_id}' not found"}

    result = verify_spec(spec, chain_dir)
    return asdict(result)


@mcp.tool()
def verify_all_specs(chain_dir: str) -> dict[str, Any]:
    """
    Run verification checks for ALL applicable specs against a chain directory.

    First scans the chain to detect applicable specs, then runs verification
    for each one. Returns per-spec pass/fail and an overall summary.

    Args:
        chain_dir: Absolute path to the chain repository root.
    """
    specs = load_specs()
    scan = scan_chain(chain_dir, specs)
    spec_map = {s.id: s for s in specs}

    results = []
    all_passed = True
    for spec_id in scan.applicable_specs:
        spec = spec_map.get(spec_id)
        if not spec:
            continue
        r = verify_spec(spec, chain_dir)
        results.append(asdict(r))
        if not r.passed:
            all_passed = False

    return {
        "chain_dir": chain_dir,
        "all_passed": all_passed,
        "specs_checked": len(results),
        "results": results,
    }


@mcp.tool()
def verify_build(chain_dir: str, timeout: int = 300) -> dict[str, Any]:
    """
    Run `go build ./...` in the chain directory and return structured results.

    Args:
        chain_dir: Absolute path to the chain directory containing go.mod.
        timeout: Build timeout in seconds (default 300).
    """
    try:
        result = subprocess.run(
            ["go", "build", "./..."],
            capture_output=True,
            text=True,
            cwd=chain_dir,
            timeout=timeout,
        )
        if result.returncode == 0:
            return {"passed": True, "output": result.stdout[:2000]}
        else:
            # Parse errors for structured output
            errors = _parse_go_build_errors(result.stderr)
            return {
                "passed": False,
                "error_count": len(errors),
                "errors": errors[:20],  # cap at 20
                "raw_stderr": result.stderr[:3000],
            }
    except subprocess.TimeoutExpired:
        return {"passed": False, "error": f"Build timed out after {timeout}s"}
    except FileNotFoundError:
        return {"passed": False, "error": "go binary not found in PATH"}


@mcp.tool()
def list_specs() -> list[dict[str, Any]]:
    """
    List all available migration specs with their metadata.

    Returns spec ID, title, version, description summary, and whether
    the spec has fatal warnings.
    """
    specs = load_specs()
    return [
        {
            "id": s.id,
            "title": s.title,
            "version": s.version,
            "description": s.description.strip()[:200],
            "has_fatal_warnings": s.has_fatal_warnings,
            "has_manual_steps": len(s.manual_steps) > 0,
        }
        for s in order_specs(specs)
    ]


@mcp.tool()
def get_spec(spec_id: str) -> dict[str, Any]:
    """
    Get the full content of a specific migration spec.

    Returns the complete YAML spec as a structured dict, including
    all detection rules, changes, manual steps, and verification checks.

    Args:
        spec_id: The ID of the spec to retrieve.
    """
    specs = load_specs()
    spec = next((s for s in specs if s.id == spec_id), None)
    if not spec:
        return {"error": f"Spec '{spec_id}' not found"}
    return spec.raw


@mcp.tool()
def check_warnings(chain_dir: str) -> dict[str, Any]:
    """
    Check a chain directory for all migration warnings (fatal and non-fatal).

    Scans imports and returns any warnings that would be emitted during
    migration, grouped by severity.

    Args:
        chain_dir: Absolute path to the chain repository root.
    """
    scan = scan_chain(chain_dir)
    return {
        "chain_dir": chain_dir,
        "fatal_blocks": scan.fatal_blocks,
        "warnings": scan.warnings,
        "has_fatal": len(scan.fatal_blocks) > 0,
        "has_warnings": len(scan.warnings) > 0,
    }


# ═══════════════════════════════════════════════════════════════════════════════
# RESOURCES
# ═══════════════════════════════════════════════════════════════════════════════


@mcp.resource("specs://v50-to-v54/{spec_id}")
def get_spec_resource(spec_id: str) -> str:
    """Read a migration spec as YAML text."""
    spec_file = SPEC_DIR / f"{spec_id}.yaml"
    if not spec_file.exists():
        # Try matching by ID instead of filename
        for f in SPEC_DIR.glob("*.yaml"):
            with open(f) as fh:
                data = yaml.safe_load(fh)
            if data.get("id") == spec_id:
                return f.read_text()
        return f"Spec '{spec_id}' not found"
    return spec_file.read_text()


@mcp.resource("agents://orchestration")
def get_orchestration_guide() -> str:
    """Read the agents.md orchestration guide."""
    agents_file = Path(__file__).parent.parent / "agents.md"
    if agents_file.exists():
        return agents_file.read_text()
    return "agents.md not found"


@mcp.resource("specs://v50-to-v54/index")
def get_spec_index() -> str:
    """List all available specs with IDs and titles."""
    specs = load_specs()
    lines = ["# Available Migration Specs (v50+ → v54)\n"]
    for s in order_specs(specs):
        fatal = " [FATAL]" if s.has_fatal_warnings else ""
        manual = " [MANUAL STEPS]" if s.manual_steps else ""
        lines.append(f"- **{s.id}**: {s.title}{fatal}{manual}")
    return "\n".join(lines)


# ═══════════════════════════════════════════════════════════════════════════════
# PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════


@mcp.prompt()
def migrate_chain(chain_dir: str) -> str:
    """
    Full migration workflow prompt: scan → plan → apply → verify.

    Orchestrates the complete migration of a chain from v50+ to v54.
    """
    return f"""You are migrating the Cosmos SDK chain at: {chain_dir}

Follow these steps exactly:

1. **Scan**: Call `scan_chain_tool(chain_dir="{chain_dir}")` to detect the SDK
   version and which specs apply.

2. **Check for fatal blocks**: If the scan returns `fatal_blocks`, STOP
   immediately and report the issue. Do not modify any files.

3. **Plan**: Call `get_migration_plan(chain_dir="{chain_dir}")` to see exactly
   what will change. Review the plan before proceeding.

4. **Apply specs in order**: For each spec in the plan's `application_order`,
   call `apply_spec(chain_dir="{chain_dir}", spec_id="<id>")`.
   After each spec, review the result for any manual_steps_required.

5. **Handle manual steps**: For specs with manual_steps (like gov.yaml's
   NewKeeper surgery), read the step description and apply the edit directly
   using your code editing capabilities.

6. **Run go mod tidy**: After all specs are applied, run `go mod tidy` in the
   chain directory to clean up dependencies.

7. **Verify**: Call `verify_all_specs(chain_dir="{chain_dir}")` to confirm all
   verification checks pass.

8. **Build**: Call `verify_build(chain_dir="{chain_dir}")` to confirm the chain
   compiles. If it fails, use the `debug_build_failure` prompt.

9. **Report**: Summarize what was done, what warnings were emitted, and any
   remaining manual steps.
"""


@mcp.prompt()
def assess_chain(chain_dir: str) -> str:
    """
    Scan-only prompt: detect version, list applicable specs, estimate effort.

    Does not modify any files.
    """
    return f"""Assess the migration readiness of the chain at: {chain_dir}

1. Call `scan_chain_tool(chain_dir="{chain_dir}")` to detect the SDK version
   and applicable specs.

2. Call `get_migration_plan(chain_dir="{chain_dir}")` to see the full plan.

3. Summarize:
   - Current SDK version
   - Number of specs that apply
   - Any fatal blocks (e.g., x/group)
   - Expected warnings (e.g., circuit/nft contrib moves)
   - Number of files that will be affected
   - Any manual steps required
   - Estimated complexity: simple (standard simapp), moderate (some custom
     wiring), or complex (heavily forked, many custom modules)

Do NOT modify any files. This is assessment only.
"""


@mcp.prompt()
def debug_build_failure(chain_dir: str, error_output: str) -> str:
    """
    Diagnose a post-migration build failure.

    Given the error output from `go build`, identify which spec caused the
    issue and suggest a fix.
    """
    return f"""The chain at {chain_dir} failed to build after migration.

Build error output:
```
{error_output}
```

Diagnose this failure:

1. **Parse the error**: Identify the file, line, and error type.

2. **Identify the cause**: Common post-migration errors:
   - "undefined: X" — a symbol was removed by a spec but still referenced.
     Check which spec removed it and whether a reference was missed.
   - "too many arguments" / "not enough arguments" — a function signature
     change was not applied. Check gov.yaml or the call_arg_edits.
   - "cannot use X as type Y" — a type changed (e.g., value → pointer).
     Check epochs.yaml.
   - "imported and not used" — an import was rewritten but the usage was
     not updated. Run goimports to fix.

3. **Suggest a fix**: Either:
   - A specific code edit to make
   - A spec that needs to be re-applied
   - A manual step that was missed

4. **Apply the fix** if you can do so confidently, then re-run
   `verify_build(chain_dir="{chain_dir}")`.
"""


# ═══════════════════════════════════════════════════════════════════════════════
# HELPERS
# ═══════════════════════════════════════════════════════════════════════════════


def _find_go_files(directory: str) -> list[str]:
    """Find all .go files in a directory tree."""
    result = []
    for root, _, files in os.walk(directory):
        for f in files:
            if f.endswith(".go"):
                result.append(os.path.join(root, f))
    return result


def _find_files(directory: str, filename: str) -> list[str]:
    """Find all files with a given name in a directory tree."""
    result = []
    for root, _, files in os.walk(directory):
        if filename in files:
            result.append(os.path.join(root, filename))
    return result


def _estimate_affected_files(chain_dir: str, spec: Spec) -> list[str]:
    """Estimate which files a spec would touch."""
    affected = set()
    changes = spec.changes

    # Text replacements
    for repl in changes.get("text_replacements", []):
        old = repl.get("old", "")
        file_match = repl.get("file_match", "")
        if old:
            from specs import _grep_dir

            files = _grep_dir(re.escape(old[:40]), chain_dir)
            for f in files:
                if file_match and not f.endswith(file_match):
                    continue
                affected.add(os.path.relpath(f, chain_dir))

    # Import rewrites
    for rewrite in changes.get("imports", {}).get("rewrites", []):
        from specs import _grep_dir

        files = _grep_dir(rewrite["old"], chain_dir)
        for f in files:
            affected.add(os.path.relpath(f, chain_dir))

    # File removals
    for removal in changes.get("file_removals", []):
        matches = _find_files(chain_dir, removal["file_name"])
        for f in matches:
            affected.add(os.path.relpath(f, chain_dir))

    return sorted(affected)[:50]


def _summarize_changes(changes: dict[str, Any]) -> dict[str, int]:
    """Count the changes in each category."""
    return {
        "import_rewrites": len(changes.get("imports", {}).get("rewrites", [])),
        "import_warnings": len(changes.get("imports", {}).get("warnings", [])),
        "text_replacements": len(changes.get("text_replacements", [])),
        "file_removals": len(changes.get("file_removals", [])),
        "statement_removals": len(changes.get("statement_removals", [])),
        "map_entry_removals": len(changes.get("map_entry_removals", [])),
        "call_arg_edits": len(changes.get("call_arg_edits", [])),
    }


def _parse_go_build_errors(stderr: str) -> list[dict[str, str]]:
    """Parse go build stderr into structured errors."""
    errors = []
    for line in stderr.strip().splitlines():
        # Pattern: ./file.go:line:col: error message
        match = re.match(r"^(.+\.go):(\d+):(\d+):\s*(.+)$", line)
        if match:
            errors.append(
                {
                    "file": match.group(1),
                    "line": match.group(2),
                    "col": match.group(3),
                    "message": match.group(4),
                }
            )
        elif line.strip() and not line.startswith("#"):
            errors.append({"message": line.strip()})
    return errors


# ═══════════════════════════════════════════════════════════════════════════════
# ENTRYPOINT
# ═══════════════════════════════════════════════════════════════════════════════


def main():
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
