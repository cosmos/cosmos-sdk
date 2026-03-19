# Cosmos SDK Migration MCP Server

An MCP (Model Context Protocol) server that lets AI agents migrate Cosmos SDK
chains from v0.50+ to v0.54. Teams point it at their chain repo and the agent
handles detection, planning, application, and verification.

## Setup

```bash
cd tools/migration/mcp

# Install dependencies
uv pip install -e .

# Or with pip
pip install -e .
```

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "cosmos-migration": {
      "command": "uv",
      "args": ["--directory", "/path/to/cosmos-sdk/tools/migration/mcp", "run", "server.py"]
    }
  }
}
```

### Claude Code

```bash
claude mcp add cosmos-migration -- uv --directory /path/to/tools/migration/mcp run server.py
```

## Tools

| Tool | Description |
|---|---|
| `scan_chain_tool` | Detect SDK version and which specs apply to a chain |
| `get_migration_plan` | Preview all changes without modifying files |
| `apply_spec` | Apply a single spec (with optional dry_run) |
| `verify_spec_tool` | Run verification checks for one spec |
| `verify_all_specs` | Run verification for all applicable specs |
| `verify_build` | Run `go build ./...` and return structured results |
| `list_specs` | List all available migration specs |
| `get_spec` | Get full content of a specific spec |
| `check_warnings` | Check for fatal blocks and warnings |

## Resources

| URI | Description |
|---|---|
| `specs://v50-to-v54/{id}` | Read a spec as YAML text |
| `specs://v50-to-v54/index` | List all specs with metadata |
| `agents://orchestration` | Read the agents.md guide |

## Prompts

| Prompt | Description |
|---|---|
| `migrate_chain` | Full workflow: scan → plan → apply → verify → report |
| `assess_chain` | Scan-only: detect version, estimate effort, no changes |
| `debug_build_failure` | Diagnose post-migration build errors |

## Example conversation

> **User**: Migrate my chain at ~/code/mychain to v54
>
> **Agent** (using migrate_chain prompt):
> 1. Calls `scan_chain_tool("~/code/mychain")`
>    → SDK v0.50.6, 7 specs apply, no fatal blocks
> 2. Calls `get_migration_plan("~/code/mychain")`
>    → 43 files affected, 2 warnings (circuit, nft)
> 3. Applies each spec in order via `apply_spec`
> 4. Handles manual steps (govkeeper surgery) directly
> 5. Calls `verify_all_specs` → all pass
> 6. Calls `verify_build` → pass
> 7. Reports summary with warnings

## Architecture

```
mcp/
  server.py          ← MCP server (tools, resources, prompts)
  specs.py           ← Spec loader, scanner, verifier
  pyproject.toml     ← Package config

  Uses:
  ../migration-spec/v50-to-v54/*.yaml   ← Spec definitions
  ../agents.md                           ← Orchestration guide
```

The MCP server is a thin layer over the spec loader in `specs.py`. All
migration logic reads from the same YAML specs that the Go binary and
agents.md reference. The three execution paths (MCP tools, Go binary,
manual following agents.md) all use the same source of truth.
