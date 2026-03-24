# Migration Framework

A spec-driven migration guide for upgrading Cosmos SDK chains across breaking version changes.

## Overview

Migrating a chain to a new SDK version involves dozens of coordinated changes: import rewrites, keeper signature updates, module removals, file deletions, and go.mod bumps.

This framework takes a three-pronged approach:

1. **`migration-spec/`** — YAML files that are the authoritative source of truth for each migration concern. One file per component/version jump. Human-readable, diffable, and reusable across chains.

2. **`agents.md`** — An orchestration guide for AI agents. Explains where to look, how to reason about the migration, what commands to run, how to stage edits, and what success looks like. An agent that reads this file can migrate a chain without any compiled tooling.

3. **`mcp/`** — An executable MCP server that applies the spec model directly and exposes repository docs as MCP resources for agents that support tool use.

---

## Directory structure

```
tools/migration/
  agents.md                      ← AI agent orchestration guide
  migration-spec/
    v50-to-v54/                  ← specs for the v50+ → v54 upgrade
      core.yaml                  ← SDK version bumps, vanity URL rewrites (always apply first)
      store-v2.yaml              ← store/v2 imports + removed BaseApp store helpers
      bank-endblock.yaml         ← move x/bank to the front of SetOrderEndBlockers
      crisis.yaml                ← Remove x/crisis entirely
      circuit.yaml               ← x/circuit moved to contrib (keep, warn)
      nft.yaml                   ← x/nft moved to contrib (keep, warn)
      group.yaml                 ← x/group → enterprise (fatal, halt migration)
      gov.yaml                   ← govkeeper.NewKeeper signature change
      gov-hooks.yaml             ← AfterProposalSubmission adds proposer address
      epochs.yaml                ← EpochsKeeper value → pointer
      epochs-app-module.yaml     ← x/epochs NewAppModule should receive a pointer
      ante.yaml                  ← Remove custom ante.go wrapper
      app-structure.yaml         ← DI files, misc ordering cleanup
  mcp/                           ← MCP server for AI agent integration
    server.py                  ← Tools, resources, prompts
    specs.py                   ← Spec loader, scanner, verifier
```

---

## How to use

### Via MCP server (recommended)

Connect the `cosmos-migration` MCP server to Claude Desktop, Claude Code, or
any MCP-compatible agent:

```bash
# Claude Code
claude mcp add cosmos-migration -- uv --directory tools/migration/mcp run server.py
```

Then ask: *"Migrate the chain at ~/code/mychain to v54."*

The agent uses `scan_chain_tool` → `get_migration_plan` → `apply_spec` (per
spec) → `verify_all_specs` → `verify_build`. See `mcp/README.md` for full setup.
The spec set remains focused on the `v0.54` breaking changes; the `v0.50+`
support claim means those `v0.54` rules also fit older app layouts when the
same features existed.

### For AI-assisted migration (without MCP)

Point an AI agent at `agents.md` and the target chain repo:

> "Read tools/migration/agents.md and then migrate the chain at <path> to v54."

The agent reads the YAML specs, scans the repo, applies changes using standard
code editing, and verifies the result builds. This path works without the MCP
server but requires the agent to parse specs and run grep/sed manually.

Important: fatal checks such as `x/group` are preflight gates, not cleanup
steps. A migration that stops on a fatal spec should do so before deleting files
or rewriting `go.mod`.

### For manual migration

Read each relevant YAML spec in `migration-spec/v50-to-v54/` for the components
your chain uses. Each spec has a `description`, `changes`, `manual_steps`, and
`verification` section that together describe exactly what to do and how to
confirm it worked.

## YAML spec format

Each spec file follows this structure:

```yaml
id: <unique-identifier>
title: <human-readable title>
version: v53 -> v54
description: |
  What this migration does and why.

upstream_sources:
  - docs://upgrading/v0.54

detection:
  imports: [list of import prefixes that indicate this spec applies]
  patterns: [code patterns to search for]
  files: [filenames that indicate this spec applies]
  go_mod: [go.mod entries that indicate this spec applies]

changes:
  go_mod:        { remove, update, add, strip_local_replaces }
  imports:       { rewrites: [...], warnings: [...] }
  statement_removals:  [...]
  map_entry_removals:  [...]
  call_arg_edits:      [...]
  special_cases:       [...]   # targeted structural rewrites in the MCP server
  text_replacements:   [...]
  file_removals:       [...]

manual_steps:
  - id: <stable-id>
    description: <manual follow-up instructions>

verification:
  must_not_import:  [...]
  must_not_contain: [...]
  must_contain:     [...]
```
