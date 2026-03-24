"""
Cosmos SDK Migration MCP Server.

Exposes migration tooling as MCP tools, resources, and prompts so that
AI agents can migrate Cosmos SDK chains from v50+ to v54.
"""

from __future__ import annotations

import os
import re
import subprocess
from dataclasses import asdict
from pathlib import Path
from typing import Any, Callable

import yaml
from mcp.server.fastmcp import FastMCP

from specs import (
    CHANGELOG_FILE,
    RELEASE_NOTES_FILE,
    SPEC_DIR,
    UPGRADING_FILE,
    Spec,
    list_go_files,
    list_go_mod_files,
    load_specs,
    order_specs,
    scan_chain,
    verify_spec,
)

mcp = FastMCP(
    "cosmos-migration",
    version="0.2.0",
    description="Cosmos SDK chain migration server (v50+ → v54). "
    "Provides tools for scanning chains, planning migrations, applying "
    "specs, and verifying results against the repository upgrade docs.",
)


# ═══════════════════════════════════════════════════════════════════════════════
# TOOLS
# ═══════════════════════════════════════════════════════════════════════════════


@mcp.tool()
def scan_chain_tool(chain_dir: str) -> dict[str, Any]:
    """
    Scan a chain repository and detect which migration specs apply.

    Analyzes go.mod for the SDK version, then checks every spec's detection
    rules against the codebase. Returns: SDK version, applicable specs
    (in correct application order), expected warnings, and any fatal blocks
    that would halt migration.

    Args:
        chain_dir: Absolute path to the chain repository root.
    """
    return asdict(scan_chain(chain_dir))


@mcp.tool()
def get_migration_plan(chain_dir: str) -> dict[str, Any]:
    """
    Generate a structured migration plan for a chain repository.

    Scans the chain, selects applicable specs, and for each spec describes:
    - What changes will be made (imports, go.mod, removals, arg edits, etc.)
    - Which files will likely be affected
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

    spec_map = {spec.id: spec for spec in specs}
    plan: list[dict[str, Any]] = []

    for spec_id in scan.applicable_specs:
        spec = spec_map.get(spec_id)
        if not spec:
            continue

        plan.append(
            {
                "spec_id": spec.id,
                "title": spec.title,
                "description": spec.description.strip(),
                "changes_summary": _summarize_changes(spec.changes),
                "affected_files": _estimate_affected_files(chain_dir, spec),
                "manual_steps": spec.manual_steps,
                "upstream_sources": spec.raw.get("upstream_sources", []),
                "has_warnings": any(
                    not warning.get("fatal", False)
                    for warning in spec.changes.get("imports", {}).get("warnings", [])
                ),
            }
        )

    return {
        "status": "ready",
        "sdk_version": scan.sdk_version,
        "specs_to_apply": len(plan),
        "plan": plan,
        "warnings": scan.warnings,
        "application_order": [entry["spec_id"] for entry in plan],
    }


@mcp.tool()
def apply_spec(chain_dir: str, spec_id: str, dry_run: bool = False) -> dict[str, Any]:
    """
    Apply a single migration spec to a chain repository.

    Executes go.mod updates, import rewrites, statement removals, map entry
    removals, call-argument edits, special-case structural rewrites, and
    text replacements defined in the spec.

    For remaining non-mechanical changes, the tool returns manual steps the
    agent should handle directly.

    Args:
        chain_dir: Absolute path to the chain repository root.
        spec_id: The ID of the spec to apply.
        dry_run: If True, report what would change without modifying files.
    """
    spec = _load_spec(spec_id)
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
        "go_mod_changes": [],
        "file_removals": [],
        "import_rewrites": [],
        "statement_removals": [],
        "map_entry_removals": [],
        "call_arg_edits": [],
        "special_case_rewrites": [],
        "text_replacements": [],
        "manual_steps_required": list(spec.manual_steps),
        "warnings": [],
    }

    changes = spec.changes

    _apply_file_removals(chain_dir, changes.get("file_removals", []), dry_run, results)
    _apply_go_mod_changes(chain_dir, changes.get("go_mod", {}), dry_run, results)
    _apply_import_rewrites(chain_dir, changes.get("imports", {}).get("rewrites", []), dry_run, results)
    _apply_statement_removals(chain_dir, changes.get("statement_removals", []), dry_run, results)
    _apply_map_entry_removals(chain_dir, changes.get("map_entry_removals", []), dry_run, results)
    _apply_call_arg_edits(chain_dir, changes.get("call_arg_edits", []), dry_run, results)
    _apply_special_cases(chain_dir, changes.get("special_cases", []), dry_run, results)
    _apply_text_replacements(chain_dir, changes.get("text_replacements", []), dry_run, results)

    for warning in changes.get("imports", {}).get("warnings", []):
        if not warning.get("fatal", False):
            results["warnings"].append(warning.get("message", ""))

    return results


@mcp.tool()
def verify_spec_tool(chain_dir: str, spec_id: str) -> dict[str, Any]:
    """
    Run verification checks for a specific spec against a chain directory.

    Checks must_not_import, must_not_contain, and must_contain rules
    defined in the spec's verification section.
    """
    spec = _load_spec(spec_id)
    if not spec:
        return {"error": f"Spec '{spec_id}' not found"}
    return asdict(verify_spec(spec, chain_dir))


@mcp.tool()
def verify_all_specs(chain_dir: str, spec_ids: list[str] | None = None) -> dict[str, Any]:
    """
    Run verification checks for multiple specs against a chain directory.

    If spec_ids is omitted, verification uses the currently detectable specs.
    For post-migration verification, prefer passing the application_order from
    get_migration_plan so removed specs are still checked.
    """
    specs = load_specs()
    spec_map = {spec.id: spec for spec in specs}
    selected_ids = spec_ids or scan_chain(chain_dir, specs).applicable_specs

    results = []
    all_passed = True
    for spec_id in selected_ids:
        spec = spec_map.get(spec_id)
        if not spec:
            results.append(
                {
                    "spec_id": spec_id,
                    "passed": False,
                    "failures": [f"Spec '{spec_id}' not found"],
                }
            )
            all_passed = False
            continue

        result = verify_spec(spec, chain_dir)
        results.append(asdict(result))
        if not result.passed:
            all_passed = False

    return {
        "chain_dir": chain_dir,
        "all_passed": all_passed,
        "specs_checked": len(results),
        "checked_spec_ids": selected_ids,
        "results": results,
    }


@mcp.tool()
def verify_build(chain_dir: str, timeout: int = 300) -> dict[str, Any]:
    """
    Run `go build ./...` in the chain directory and return structured results.
    """
    try:
        result = subprocess.run(
            ["go", "build", "./..."],
            capture_output=True,
            text=True,
            cwd=chain_dir,
            timeout=timeout,
        )
    except subprocess.TimeoutExpired:
        return {"passed": False, "error": f"Build timed out after {timeout}s"}
    except FileNotFoundError:
        return {"passed": False, "error": "go binary not found in PATH"}

    if result.returncode == 0:
        return {"passed": True, "output": result.stdout[:2000]}

    errors = _parse_go_build_errors(result.stderr)
    return {
        "passed": False,
        "error_count": len(errors),
        "errors": errors[:20],
        "raw_stderr": result.stderr[:3000],
    }


@mcp.tool()
def list_specs() -> list[dict[str, Any]]:
    """List all available migration specs with metadata."""
    specs = load_specs()
    return [
        {
            "id": spec.id,
            "title": spec.title,
            "version": spec.version,
            "description": spec.description.strip()[:200],
            "has_fatal_warnings": spec.has_fatal_warnings,
            "has_manual_steps": bool(spec.manual_steps),
            "upstream_sources": spec.raw.get("upstream_sources", []),
        }
        for spec in order_specs(specs)
    ]


@mcp.tool()
def get_spec(spec_id: str) -> dict[str, Any]:
    """Get the full content of a specific migration spec."""
    spec = _load_spec(spec_id)
    if not spec:
        return {"error": f"Spec '{spec_id}' not found"}
    return spec.raw


@mcp.tool()
def check_warnings(chain_dir: str) -> dict[str, Any]:
    """Check a chain directory for fatal and non-fatal migration warnings."""
    scan = scan_chain(chain_dir)
    return {
        "chain_dir": chain_dir,
        "fatal_blocks": scan.fatal_blocks,
        "warnings": scan.warnings,
        "has_fatal": bool(scan.fatal_blocks),
        "has_warnings": bool(scan.warnings),
    }


# ═══════════════════════════════════════════════════════════════════════════════
# RESOURCES
# ═══════════════════════════════════════════════════════════════════════════════


@mcp.resource("specs://v50-to-v54/{spec_id}")
def get_spec_resource(spec_id: str) -> str:
    """Read a migration spec as YAML text."""
    spec_file = SPEC_DIR / f"{spec_id}.yaml"
    if spec_file.exists():
        return spec_file.read_text()

    for candidate in SPEC_DIR.glob("*.yaml"):
        with candidate.open() as handle:
            data = yaml.safe_load(handle)
        if data.get("id") == spec_id:
            return candidate.read_text()

    return f"Spec '{spec_id}' not found"


@mcp.resource("specs://v50-to-v54/index")
def get_spec_index() -> str:
    """List all available specs with IDs and titles."""
    lines = ["# Available Migration Specs (v50+ → v54)\n"]
    for spec in order_specs(load_specs()):
        fatal = " [FATAL]" if spec.has_fatal_warnings else ""
        manual = " [MANUAL STEPS]" if spec.manual_steps else ""
        lines.append(f"- **{spec.id}**: {spec.title}{fatal}{manual}")
    return "\n".join(lines)


@mcp.resource("agents://orchestration")
def get_orchestration_guide() -> str:
    """Read the agents.md orchestration guide."""
    agents_file = Path(__file__).parent.parent / "agents.md"
    return agents_file.read_text() if agents_file.exists() else "agents.md not found"


@mcp.resource("docs://upgrading/v0.54")
def get_upgrading_guide() -> str:
    """Read the repository v0.54 upgrade guide."""
    return UPGRADING_FILE.read_text()


@mcp.resource("docs://upgrade-checklist/v0.54")
def get_upgrade_checklist() -> str:
    """Read the v0.54 upgrade checklist section."""
    return _read_markdown_section(UPGRADING_FILE, "## Upgrade Checklist")


@mcp.resource("docs://changelog/v0.54-breaking")
def get_changelog_breaking_changes() -> str:
    """Read the v0.54 breaking-changes section from the changelog."""
    return _read_markdown_section(CHANGELOG_FILE, "### Breaking Changes")


@mcp.resource("docs://release-notes/v0.54")
def get_release_notes() -> str:
    """Read the v0.54 release notes."""
    return RELEASE_NOTES_FILE.read_text()


# ═══════════════════════════════════════════════════════════════════════════════
# PROMPTS
# ═══════════════════════════════════════════════════════════════════════════════


@mcp.prompt()
def migrate_chain(chain_dir: str) -> str:
    """
    Full migration workflow prompt: scan → plan → apply → verify.
    """
    return f"""You are migrating the Cosmos SDK chain at: {chain_dir}

Use the migration specs together with the repository docs resources when a
spec includes manual steps or when a chain deviates from simapp.

1. **Scan**: Call `scan_chain_tool(chain_dir="{chain_dir}")` to detect the SDK
   version and applicable specs.

2. **Check for fatal blocks**: If the scan returns `fatal_blocks`, STOP
   immediately and report the issue. Do not modify any files.

3. **Plan**: Call `get_migration_plan(chain_dir="{chain_dir}")` and keep the
   returned `application_order`.

4. **Apply specs in order**: For each spec in `application_order`, call
   `apply_spec(chain_dir="{chain_dir}", spec_id="<id>")`.

5. **Handle remaining manual steps**: For specs that still return manual steps,
   consult `docs://upgrading/v0.54` and the spec's `upstream_sources`, then
   apply the remaining edits directly.

6. **Run go mod tidy**: After all specs are applied, run `go mod tidy` in the
   chain directory to clean up dependencies.

7. **Verify**: Call
   `verify_all_specs(chain_dir="{chain_dir}", spec_ids=<application_order>)`.

8. **Build**: Call `verify_build(chain_dir="{chain_dir}")`. If it fails, use
   the `debug_build_failure` prompt.

9. **Report**: Summarize what was changed, which warnings remain, and which
   manual steps, if any, still require operator attention.
"""


@mcp.prompt()
def assess_chain(chain_dir: str) -> str:
    """
    Scan-only prompt: detect version, list applicable specs, estimate effort.
    """
    return f"""Assess the migration readiness of the chain at: {chain_dir}

1. Call `scan_chain_tool(chain_dir="{chain_dir}")`.
2. Call `get_migration_plan(chain_dir="{chain_dir}")`.
3. Summarize:
   - Current SDK version
   - Number of specs that apply
   - Any fatal blocks (for example `x/group`)
   - Expected warnings (for example contrib module moves)
   - Approximate files affected
   - Manual steps required
   - Estimated complexity: simple, moderate, or complex

Do NOT modify any files.
"""


@mcp.prompt()
def debug_build_failure(chain_dir: str, error_output: str) -> str:
    """
    Diagnose a post-migration build failure.
    """
    return f"""The chain at {chain_dir} failed to build after migration.

Build error output:
```
{error_output}
```

Diagnose the failure:

1. Identify the file, line, and error type.
2. Determine whether this maps to:
   - a missed spec application,
   - a manual step from a spec,
   - a chain-specific customization that the specs intentionally leave alone.
3. Use `docs://upgrading/v0.54` and `docs://changelog/v0.54-breaking` for
   ambiguous API migrations.
4. Suggest or apply a focused fix, then re-run
   `verify_build(chain_dir="{chain_dir}")`.
"""


# ═══════════════════════════════════════════════════════════════════════════════
# APPLY HELPERS
# ═══════════════════════════════════════════════════════════════════════════════


def _load_spec(spec_id: str) -> Spec | None:
    return next((spec for spec in load_specs() if spec.id == spec_id), None)


def _apply_file_removals(
    chain_dir: str,
    removals: list[dict[str, Any]],
    dry_run: bool,
    results: dict[str, Any],
) -> None:
    for removal in removals:
        file_name = removal["file_name"]
        required_text = removal.get("contains_must_match", "")
        for path in _find_files(chain_dir, file_name):
            if required_text and required_text not in _read_file(path):
                continue
            action = "would_delete" if dry_run else "deleted"
            results["file_removals"].append(
                {"file": os.path.relpath(path, chain_dir), "action": action}
            )
            if not dry_run:
                os.remove(path)


def _apply_go_mod_changes(
    chain_dir: str,
    go_mod_changes: dict[str, Any],
    dry_run: bool,
    results: dict[str, Any],
) -> None:
    if not go_mod_changes:
        return

    for path in list_go_mod_files(chain_dir):
        original = _read_file(path)
        updated = original
        changed_details: list[str] = []

        if go_mod_changes.get("strip_local_replaces"):
            updated, changed = _strip_local_replaces(updated)
            if changed:
                changed_details.append("strip_local_replaces")

        updated, removed = _remove_go_mod_modules(updated, go_mod_changes.get("remove", []))
        if removed:
            changed_details.extend(f"remove:{module}" for module in removed)

        updated, updated_modules = _update_go_mod_modules(updated, go_mod_changes.get("update", {}))
        if updated_modules:
            changed_details.extend(f"update:{module}" for module in updated_modules)

        updated, added = _add_go_mod_modules(updated, go_mod_changes.get("add", {}))
        if added:
            changed_details.extend(f"add:{module}" for module in added)

        if updated == original:
            continue

        relpath = os.path.relpath(path, chain_dir)
        results["go_mod_changes"].append(
            {
                "file": relpath,
                "action": "would_update" if dry_run else "updated",
                "details": changed_details,
            }
        )
        if not dry_run:
            Path(path).write_text(updated)


def _apply_import_rewrites(
    chain_dir: str,
    rewrites: list[dict[str, Any]],
    dry_run: bool,
    results: dict[str, Any],
) -> None:
    go_files = list_go_files(chain_dir)
    for rewrite in rewrites:
        old_path = rewrite["old"]
        new_path = rewrite["new"]
        all_packages = rewrite.get("all_packages", False)

        for path in go_files:
            content = _read_file(path)
            if old_path not in content:
                continue

            updated = (
                content.replace(old_path, new_path)
                if all_packages
                else content.replace(f'"{old_path}"', f'"{new_path}"')
            )
            if updated == content:
                continue

            relpath = os.path.relpath(path, chain_dir)
            results["import_rewrites"].append(
                {
                    "file": relpath,
                    "action": "would_rewrite" if dry_run else "rewritten",
                    "pattern": f"{old_path} → {new_path}",
                }
            )
            if not dry_run:
                Path(path).write_text(updated)


def _apply_statement_removals(
    chain_dir: str,
    removals: list[dict[str, Any]],
    dry_run: bool,
    results: dict[str, Any],
) -> None:
    if not removals:
        return

    for path in list_go_files(chain_dir):
        original = _read_file(path)
        updated = original
        removed_items: list[str] = []

        for removal in removals:
            anchor = removal.get("assign_target") or removal.get("call_pattern")
            if not anchor:
                continue
            updated, count = _remove_statement_occurrences(updated, anchor)
            if count:
                removed_items.extend([anchor] * count)

        if updated == original:
            continue

        relpath = os.path.relpath(path, chain_dir)
        results["statement_removals"].append(
            {
                "file": relpath,
                "action": "would_remove" if dry_run else "removed",
                "patterns": removed_items,
            }
        )
        if not dry_run:
            Path(path).write_text(updated)


def _apply_map_entry_removals(
    chain_dir: str,
    removals: list[dict[str, Any]],
    dry_run: bool,
    results: dict[str, Any],
) -> None:
    if not removals:
        return

    for path in list_go_files(chain_dir):
        original = _read_file(path)
        updated = original
        removed_keys: list[str] = []

        for removal in removals:
            map_var = removal.get("map_var", "")
            if map_var and map_var not in updated:
                continue
            for key in removal.get("keys", []):
                pattern = re.compile(rf"^[ \t]*{re.escape(key)}[ \t]*,?[ \t]*\n?", re.MULTILINE)
                updated, count = pattern.subn("", updated)
                if count:
                    removed_keys.extend([key] * count)

        if updated == original:
            continue

        relpath = os.path.relpath(path, chain_dir)
        results["map_entry_removals"].append(
            {
                "file": relpath,
                "action": "would_remove" if dry_run else "removed",
                "keys": removed_keys,
            }
        )
        if not dry_run:
            Path(path).write_text(updated)


def _apply_call_arg_edits(
    chain_dir: str,
    edits: list[dict[str, Any]],
    dry_run: bool,
    results: dict[str, Any],
) -> None:
    if not edits:
        return

    for path in list_go_files(chain_dir):
        original = _read_file(path)
        updated = original
        changed_calls: list[str] = []

        for edit in edits:
            updated, count = _apply_call_arg_edit(updated, edit)
            if count:
                call_name = edit.get("func_pattern") or edit.get("method_name") or "call"
                changed_calls.extend([call_name] * count)

        if updated == original:
            continue

        relpath = os.path.relpath(path, chain_dir)
        results["call_arg_edits"].append(
            {
                "file": relpath,
                "action": "would_edit" if dry_run else "edited",
                "calls": changed_calls,
            }
        )
        if not dry_run:
            Path(path).write_text(updated)


def _apply_special_cases(
    chain_dir: str,
    special_cases: list[str],
    dry_run: bool,
    results: dict[str, Any],
) -> None:
    if not special_cases:
        return

    handlers: dict[str, Callable[[str], tuple[str, int]]] = {
        "gov_new_keeper": _rewrite_gov_new_keeper_calls,
        "uncached_context": _rewrite_uncached_context_calls,
        "gov_hooks_after_proposal_submission": _rewrite_gov_hook_signatures,
    }

    for case_name in special_cases:
        handler = handlers.get(case_name)
        if not handler:
            continue

        for path in list_go_files(chain_dir):
            original = _read_file(path)
            updated, count = handler(original)
            if not count or updated == original:
                continue

            relpath = os.path.relpath(path, chain_dir)
            results["special_case_rewrites"].append(
                {
                    "file": relpath,
                    "action": "would_rewrite" if dry_run else "rewritten",
                    "case": case_name,
                    "count": count,
                }
            )
            if not dry_run:
                Path(path).write_text(updated)


def _apply_text_replacements(
    chain_dir: str,
    replacements: list[dict[str, Any]],
    dry_run: bool,
    results: dict[str, Any],
) -> None:
    if not replacements:
        return

    for replacement in replacements:
        old = replacement.get("old", "")
        new = replacement.get("new", "")
        file_match = replacement.get("file_match", "")
        required = replacement.get("requires_contains", [])
        if not old:
            continue

        for path in _iter_text_files(chain_dir):
            relpath = os.path.relpath(path, chain_dir)
            if file_match and not _matches_file(relpath, file_match):
                continue

            content = _read_file(path)
            if required and not all(entry in content for entry in required):
                continue
            if old not in content:
                continue

            updated = content.replace(old, new)
            if updated == content:
                continue

            results["text_replacements"].append(
                {
                    "file": relpath,
                    "action": "would_replace" if dry_run else "replaced",
                    "pattern": old[:80] + ("..." if len(old) > 80 else ""),
                }
            )
            if not dry_run:
                Path(path).write_text(updated)


# ═══════════════════════════════════════════════════════════════════════════════
# STRUCTURAL REWRITE HELPERS
# ═══════════════════════════════════════════════════════════════════════════════


def _remove_statement_occurrences(content: str, anchor: str) -> tuple[str, int]:
    count = 0
    search_from = 0

    while True:
        index = content.find(anchor, search_from)
        if index == -1:
            return content, count

        statement_start = content.rfind("\n", 0, index) + 1
        statement_end = _find_statement_end(content, statement_start)
        content = content[:statement_start] + content[statement_end:]
        count += 1
        search_from = statement_start


def _find_statement_end(content: str, start: int) -> int:
    depth = 0
    in_string: str | None = None
    escape = False
    index = start

    while index < len(content):
        char = content[index]

        if in_string:
            if escape:
                escape = False
            elif char == "\\" and in_string != "`":
                escape = True
            elif char == in_string:
                in_string = None
            index += 1
            continue

        if content.startswith("//", index):
            newline = content.find("\n", index)
            if newline == -1:
                return len(content)
            index = newline + 1
            if depth == 0:
                return index
            continue

        if content.startswith("/*", index):
            close = content.find("*/", index + 2)
            if close == -1:
                return len(content)
            index = close + 2
            continue

        if char in {'"', "'", "`"}:
            in_string = char
        elif char in "([{":
            depth += 1
        elif char in ")]}":
            depth = max(depth - 1, 0)
        elif char == "\n" and depth == 0:
            return index + 1

        index += 1

    return len(content)


def _apply_call_arg_edit(content: str, edit: dict[str, Any]) -> tuple[str, int]:
    if "func_pattern" in edit:
        token = edit["func_pattern"] + "("
    elif "method_name" in edit:
        token = "." + edit["method_name"] + "("
    else:
        return content, 0

    def transform(args: list[str]) -> list[str] | None:
        new_args = list(args)
        original = list(args)

        for pattern in edit.get("remove", []):
            new_args = [arg for arg in new_args if not _arg_matches(arg, pattern)]

        for addition in sorted(edit.get("add", []), key=lambda entry: entry.get("position", len(new_args))):
            expr = addition["expr"]
            if any(_normalize_arg(arg) == _normalize_arg(expr) for arg in new_args):
                continue
            position = max(0, min(addition.get("position", len(new_args)), len(new_args)))
            new_args.insert(position, expr)

        return new_args if _normalize_args_list(original) != _normalize_args_list(new_args) else None

    return _rewrite_call_args_by_token(content, token, transform)


def _rewrite_call_args_by_token(
    content: str,
    token: str,
    transform: Callable[[list[str]], list[str] | None],
) -> tuple[str, int]:
    replacements: list[tuple[int, int, str]] = []
    count = 0
    search_from = 0

    while True:
        token_index = content.find(token, search_from)
        if token_index == -1:
            break

        open_paren = token_index + len(token) - 1
        close_paren = _find_matching_paren(content, open_paren)
        if close_paren == -1:
            break

        args_source = content[open_paren + 1 : close_paren]
        new_args = transform(_split_top_level_args(args_source))
        if new_args is not None:
            base_indent = _line_indent(content, token_index)
            replacements.append(
                (
                    open_paren + 1,
                    close_paren,
                    _render_args(args_source, new_args, base_indent),
                )
            )
            count += 1

        search_from = close_paren + 1

    for start, end, replacement in reversed(replacements):
        content = content[:start] + replacement + content[end:]

    return content, count


def _rewrite_gov_new_keeper_calls(content: str) -> tuple[str, int]:
    aliases = _find_import_aliases(
        content,
        "github.com/cosmos/cosmos-sdk/x/gov/keeper",
        default_alias="keeper",
    )
    total = 0

    for alias in aliases:
        token = f"{alias}.NewKeeper("

        def transform(args: list[str], *, keeper_alias: str = alias) -> list[str] | None:
            if any("NewDefaultCalculateVoteResultsAndVotingPower" in arg for arg in args):
                return None
            if len(args) < 9:
                return None

            staking_keeper = args[4].strip().rstrip(",")
            new_args = list(args[:4]) + list(args[5:9])
            new_args.append(
                f"{keeper_alias}.NewDefaultCalculateVoteResultsAndVotingPower({staking_keeper})"
            )
            return new_args

        content, count = _rewrite_call_args_by_token(content, token, transform)
        total += count

    return content, total


def _rewrite_uncached_context_calls(content: str) -> tuple[str, int]:
    token = ".NewUncachedContext("
    replacements: list[tuple[int, int, str]] = []
    count = 0
    search_from = 0

    while True:
        dot_index = content.find(token, search_from)
        if dot_index == -1:
            break

        receiver_start = _receiver_start(content, dot_index)
        open_paren = dot_index + len(token) - 1
        close_paren = _find_matching_paren(content, open_paren)
        if close_paren == -1:
            break

        receiver = content[receiver_start:dot_index]
        args = _split_top_level_args(content[open_paren + 1 : close_paren])
        replacement = None

        if len(args) >= 2 and _normalize_arg(args[0]) == "false":
            replacement = f"{receiver}.NewNextBlockContext({args[1].strip().rstrip(',')})"
        elif args and _normalize_arg(args[0]) == "true":
            replacement = f"{receiver}.NewContext(true)"

        if replacement:
            replacements.append((receiver_start, close_paren + 1, replacement))
            count += 1

        search_from = close_paren + 1

    for start, end, replacement in reversed(replacements):
        content = content[:start] + replacement + content[end:]

    return content, count


def _rewrite_gov_hook_signatures(content: str) -> tuple[str, int]:
    aliases = _find_import_aliases(
        content,
        "github.com/cosmos/cosmos-sdk/types",
        default_alias="types",
    )
    if not aliases:
        return content, 0

    addr_type = f"{aliases[0]}.AccAddress"
    token = "AfterProposalSubmission("
    replacements: list[tuple[int, int, str]] = []
    count = 0
    search_from = 0

    while True:
        token_index = content.find(token, search_from)
        if token_index == -1:
            break

        open_paren = token_index + len(token) - 1
        close_paren = _find_matching_paren(content, open_paren)
        if close_paren == -1:
            break

        tail = content[close_paren + 1 : close_paren + 24]
        if "error" not in tail:
            search_from = close_paren + 1
            continue

        args_source = content[open_paren + 1 : close_paren]
        args = _split_top_level_args(args_source)
        if any("AccAddress" in arg for arg in args) or len(args) != 2:
            search_from = close_paren + 1
            continue

        args.append(f"proposerAddr {addr_type}")
        replacements.append(
            (
                open_paren + 1,
                close_paren,
                _render_args(args_source, args, _line_indent(content, token_index)),
            )
        )
        count += 1
        search_from = close_paren + 1

    for start, end, replacement in reversed(replacements):
        content = content[:start] + replacement + content[end:]

    return content, count


def _find_import_aliases(content: str, import_path: str, default_alias: str) -> list[str]:
    pattern = re.compile(
        rf'^\s*(?:(?P<alias>[A-Za-z_][A-Za-z0-9_]*)\s+)?"{re.escape(import_path)}"\s*$',
        re.MULTILINE,
    )
    aliases = []
    for match in pattern.finditer(content):
        aliases.append(match.group("alias") or default_alias)
    return aliases


def _split_top_level_args(args_source: str) -> list[str]:
    args: list[str] = []
    current: list[str] = []
    depth = 0
    in_string: str | None = None
    escape = False
    index = 0

    while index < len(args_source):
        char = args_source[index]

        if in_string:
            current.append(char)
            if escape:
                escape = False
            elif char == "\\" and in_string != "`":
                escape = True
            elif char == in_string:
                in_string = None
            index += 1
            continue

        if args_source.startswith("//", index):
            newline = args_source.find("\n", index)
            if newline == -1:
                current.append(args_source[index:])
                break
            current.append(args_source[index:newline + 1])
            index = newline + 1
            continue

        if args_source.startswith("/*", index):
            close = args_source.find("*/", index + 2)
            if close == -1:
                current.append(args_source[index:])
                break
            current.append(args_source[index:close + 2])
            index = close + 2
            continue

        if char in {'"', "'", "`"}:
            in_string = char
            current.append(char)
        elif char in "([{":
            depth += 1
            current.append(char)
        elif char in ")]}":
            depth -= 1
            current.append(char)
        elif char == "," and depth == 0:
            arg = "".join(current).strip()
            if arg:
                args.append(arg)
            current = []
        else:
            current.append(char)

        index += 1

    tail = "".join(current).strip()
    if tail:
        args.append(tail)

    return args


def _find_matching_paren(content: str, open_paren: int) -> int:
    depth = 0
    in_string: str | None = None
    escape = False
    index = open_paren

    while index < len(content):
        char = content[index]

        if in_string:
            if escape:
                escape = False
            elif char == "\\" and in_string != "`":
                escape = True
            elif char == in_string:
                in_string = None
            index += 1
            continue

        if content.startswith("//", index):
            newline = content.find("\n", index)
            if newline == -1:
                return -1
            index = newline + 1
            continue

        if content.startswith("/*", index):
            close = content.find("*/", index + 2)
            if close == -1:
                return -1
            index = close + 2
            continue

        if char in {'"', "'", "`"}:
            in_string = char
        elif char == "(":
            depth += 1
        elif char == ")":
            depth -= 1
            if depth == 0:
                return index

        index += 1

    return -1


def _render_args(original_args: str, args: list[str], base_indent: str) -> str:
    if "\n" in original_args or any("\n" in arg for arg in args):
        item_indent = _argument_indent(original_args, base_indent + "\t")
        rendered = [f"{item_indent}{arg.strip().rstrip(',')}," for arg in args if arg.strip()]
        if not rendered:
            return ""
        return "\n" + "\n".join(rendered) + "\n" + base_indent

    return ", ".join(arg.strip().rstrip(",") for arg in args if arg.strip())


def _argument_indent(args_source: str, fallback: str) -> str:
    for line in args_source.splitlines():
        stripped = line.strip()
        if not stripped:
            continue
        return re.match(r"[ \t]*", line).group(0)
    return fallback


def _line_indent(content: str, index: int) -> str:
    line_start = content.rfind("\n", 0, index) + 1
    line = content[line_start:index]
    return re.match(r"[ \t]*", line).group(0)


def _receiver_start(content: str, dot_index: int) -> int:
    index = dot_index - 1
    while index >= 0 and re.match(r"[A-Za-z0-9_.]", content[index]):
        index -= 1
    return index + 1


def _arg_matches(arg: str, pattern: str) -> bool:
    normalized_arg = _normalize_arg(arg)
    if "..." in pattern:
        prefix = _normalize_arg(pattern.split("...", 1)[0])
        return normalized_arg.startswith(prefix)
    return normalized_arg == _normalize_arg(pattern)


def _normalize_arg(value: str) -> str:
    return re.sub(r"\s+", "", value).rstrip(",")


def _normalize_args_list(args: list[str]) -> list[str]:
    return [_normalize_arg(arg) for arg in args if arg.strip()]


# ═══════════════════════════════════════════════════════════════════════════════
# GO.MOD HELPERS
# ═══════════════════════════════════════════════════════════════════════════════


def _strip_local_replaces(content: str) -> tuple[str, bool]:
    changed = False
    lines = content.splitlines(keepends=True)
    rebuilt: list[str] = []
    in_replace_block = False

    for line in lines:
        stripped = line.strip()
        if stripped.startswith("replace ("):
            in_replace_block = True
            rebuilt.append(line)
            continue

        if in_replace_block and stripped == ")":
            in_replace_block = False
            rebuilt.append(line)
            continue

        if "=>" not in stripped:
            rebuilt.append(line)
            continue

        target = stripped.split("=>", 1)[1].strip().split()[0]
        is_local = target.startswith("./") or target.startswith("../") or target.startswith("/")
        if is_local:
            changed = True
            continue

        rebuilt.append(line)

    return _remove_empty_go_mod_blocks("".join(rebuilt)), changed


def _remove_go_mod_modules(content: str, modules: list[str]) -> tuple[str, list[str]]:
    removed: list[str] = []
    lines = content.splitlines(keepends=True)
    rebuilt: list[str] = []

    for line in lines:
        stripped = line.strip()
        module = _go_mod_line_module(stripped)
        if module and module in modules:
            removed.append(module)
            continue
        rebuilt.append(line)

    return _remove_empty_go_mod_blocks("".join(rebuilt)), sorted(set(removed))


def _update_go_mod_modules(content: str, updates: dict[str, str]) -> tuple[str, list[str]]:
    updated_modules: list[str] = []
    lines = content.splitlines(keepends=True)

    for index, line in enumerate(lines):
        stripped = line.strip()
        for module, version in updates.items():
            if _line_starts_with_module(stripped, module):
                lines[index] = _replace_go_mod_version(line, module, version)
                updated_modules.append(module)
                break

    return "".join(lines), sorted(set(updated_modules))


def _add_go_mod_modules(content: str, additions: dict[str, str]) -> tuple[str, list[str]]:
    if not additions:
        return content, []

    present = {module for module in additions if re.search(rf"(?m)^\s*{re.escape(module)}\s+v", content)}
    missing = [module for module in additions if module not in present]
    if not missing:
        return content, []

    addition_lines = "".join(f"require {module} {additions[module]}\n" for module in missing)
    if not content.endswith("\n"):
        content += "\n"
    return content + addition_lines, missing


def _go_mod_line_module(stripped: str) -> str | None:
    if not stripped or stripped.startswith("//") or stripped in {"require (", "replace ("} or stripped == ")":
        return None
    if stripped.startswith("require "):
        parts = stripped.split()
        return parts[1] if len(parts) >= 3 else None
    if stripped.startswith("replace "):
        parts = stripped.split()
        return parts[1] if len(parts) >= 4 else None
    parts = stripped.split()
    if len(parts) >= 2 and parts[1].startswith("v"):
        return parts[0]
    return None


def _line_starts_with_module(stripped: str, module: str) -> bool:
    if stripped.startswith("require "):
        parts = stripped.split()
        return len(parts) >= 3 and parts[1] == module and parts[2].startswith("v")
    parts = stripped.split()
    return len(parts) >= 2 and parts[0] == module and parts[1].startswith("v")


def _replace_go_mod_version(line: str, module: str, version: str) -> str:
    newline = "\n" if line.endswith("\n") else ""
    stripped = line.strip()

    if stripped.startswith("require "):
        parts = stripped.split()
        comment = ""
        if "//" in line:
            comment = " //" + line.split("//", 1)[1].rstrip("\n")
        return f"require {module} {version}{comment}{newline}"

    indent = re.match(r"[ \t]*", line).group(0)
    comment = ""
    if "//" in line:
        comment = " //" + line.split("//", 1)[1].rstrip("\n")
    return f"{indent}{module} {version}{comment}{newline}"


def _remove_empty_go_mod_blocks(content: str) -> str:
    content = re.sub(r"(?ms)^require\s*\(\s*\)\n?", "", content)
    content = re.sub(r"(?ms)^replace\s*\(\s*\)\n?", "", content)
    return content


# ═══════════════════════════════════════════════════════════════════════════════
# DOC / REPORT HELPERS
# ═══════════════════════════════════════════════════════════════════════════════


def _find_files(directory: str, filename: str) -> list[str]:
    result = []
    for root, _, files in os.walk(directory):
        if filename in files:
            result.append(os.path.join(root, filename))
    return sorted(result)


def _read_file(path: str) -> str:
    return Path(path).read_text(errors="replace")


def _iter_text_files(directory: str) -> list[str]:
    files = []
    for root, _, names in os.walk(directory):
        for name in names:
            if name.endswith(".go") or name == "go.mod":
                files.append(os.path.join(root, name))
    return sorted(files)


def _matches_file(relpath: str, file_match: str) -> bool:
    return relpath.endswith(file_match) or os.path.basename(relpath).endswith(file_match)


def _estimate_affected_files(chain_dir: str, spec: Spec) -> list[str]:
    affected: set[str] = set()
    go_files = list_go_files(chain_dir)
    go_mod_files = list_go_mod_files(chain_dir)
    all_text_files = _iter_text_files(chain_dir)
    changes = spec.changes

    if changes.get("go_mod"):
        for path in go_mod_files:
            affected.add(os.path.relpath(path, chain_dir))

    for replacement in changes.get("text_replacements", []):
        old = replacement.get("old", "")
        file_match = replacement.get("file_match", "")
        for path in all_text_files:
            relpath = os.path.relpath(path, chain_dir)
            if file_match and not _matches_file(relpath, file_match):
                continue
            if old and old in _read_file(path):
                affected.add(relpath)

    for rewrite in changes.get("imports", {}).get("rewrites", []):
        for path in go_files:
            if rewrite["old"] in _read_file(path):
                affected.add(os.path.relpath(path, chain_dir))

    for removal in changes.get("file_removals", []):
        for path in _find_files(chain_dir, removal["file_name"]):
            affected.add(os.path.relpath(path, chain_dir))

    for removal in changes.get("statement_removals", []):
        anchor = removal.get("assign_target") or removal.get("call_pattern", "")
        for path in go_files:
            if anchor and anchor in _read_file(path):
                affected.add(os.path.relpath(path, chain_dir))

    for removal in changes.get("map_entry_removals", []):
        for key in removal.get("keys", []):
            for path in go_files:
                if key in _read_file(path):
                    affected.add(os.path.relpath(path, chain_dir))

    for edit in changes.get("call_arg_edits", []):
        anchor = edit.get("func_pattern") or edit.get("method_name", "")
        for path in go_files:
            if anchor and anchor in _read_file(path):
                affected.add(os.path.relpath(path, chain_dir))

    for special_case in changes.get("special_cases", []):
        anchors = {
            "gov_new_keeper": "NewKeeper(",
            "uncached_context": "NewUncachedContext(",
            "gov_hooks_after_proposal_submission": "AfterProposalSubmission(",
        }
        anchor = anchors.get(special_case, "")
        for path in go_files:
            if anchor and anchor in _read_file(path):
                affected.add(os.path.relpath(path, chain_dir))

    return sorted(affected)[:80]


def _summarize_changes(changes: dict[str, Any]) -> dict[str, int]:
    return {
        "go_mod": 1 if changes.get("go_mod") else 0,
        "import_rewrites": len(changes.get("imports", {}).get("rewrites", [])),
        "import_warnings": len(changes.get("imports", {}).get("warnings", [])),
        "file_removals": len(changes.get("file_removals", [])),
        "statement_removals": len(changes.get("statement_removals", [])),
        "map_entry_removals": len(changes.get("map_entry_removals", [])),
        "call_arg_edits": len(changes.get("call_arg_edits", [])),
        "special_cases": len(changes.get("special_cases", [])),
        "text_replacements": len(changes.get("text_replacements", [])),
    }


def _read_markdown_section(path: Path, header: str) -> str:
    content = path.read_text()
    start = content.find(header)
    if start == -1:
        return content

    next_header = re.search(r"(?m)^## ", content[start + len(header) :])
    if not next_header:
        return content[start:].strip()
    end = start + len(header) + next_header.start()
    return content[start:end].strip()


def _parse_go_build_errors(stderr: str) -> list[dict[str, str]]:
    errors = []
    for line in stderr.strip().splitlines():
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


def main() -> None:
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
