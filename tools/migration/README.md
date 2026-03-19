# Migration Framework

A spec-driven migration guide for upgrading Cosmos SDK chains across breaking version changes.

## Overview

Migrating a chain to a new SDK version involves dozens of coordinated changes:
import rewrites, keeper signature updates, module removals, file deletions, and
`go.mod` bumps. Encoding all of this as a single monolithic script breaks down
quickly when chains diverge from simapp with custom aliases, extra modules, or
chain-specific wiring.

This directory keeps only the pieces that matter for the current approach:

1. **`migration-spec/`** — YAML files that are the source of truth for each
   migration concern. One file per component/version jump. Human-readable,
   diffable, and reusable across chains.

2. **`agents.md`** — An orchestration guide for AI agents. It explains where to
   look, how to reason about the migration, what commands to run, and what
   success looks like.

---

## Directory structure

```
tools/migration/
  agents.md                      ← AI agent orchestration guide
  migration-spec/
    v50-to-v54/                  ← specs for the v50+ → v54 upgrade
      core.yaml                  ← SDK version bumps, vanity URL rewrites (always apply first)
      crisis.yaml                ← Remove x/crisis entirely
      circuit.yaml               ← x/circuit moved to contrib (keep, warn)
      nft.yaml                   ← x/nft moved to contrib (keep, warn)
      group.yaml                 ← x/group → enterprise (fatal, halt migration)
      gov.yaml                   ← govkeeper.NewKeeper signature change
      epochs.yaml                ← EpochsKeeper value → pointer
      ante.yaml                  ← Remove custom ante.go wrapper
      app-structure.yaml         ← DI files, misc ordering cleanup
```

---

## How to use

### For AI-assisted migration

Point an AI agent (Claude, etc.) at `agents.md` and the target chain repo:

> "Read tools/migration/agents.md and then migrate the chain at <path> to v54."

The agent reads the YAML specs, scans the repo to detect which modules are in
use, applies changes using standard code editing (AST rewrites, text
replacements, file deletions), and verifies the result builds. This is the
recommended path for chains with custom wiring because the agent can reason about
edge cases that purely mechanical migration steps cannot handle, such as
different import aliases, non-standard file layouts, forked replace
directives, historical upgrade files, and custom ante handlers.

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

detection:
  imports: [list of import prefixes that indicate this spec applies]
  patterns: [code patterns to search for]

changes:
  go_mod:        { remove, update, add, strip_local_replaces }
  imports:       { rewrites: [...], warnings: [...] }
  statement_removals:  [...]
  map_entry_removals:  [...]
  call_arg_edits:      [...]
  text_replacements:   [...]
  file_removals:       [...]
  manual_steps:        [...]   # for changes that can't be automated

verification:
  must_not_import:  [...]
  must_not_contain: [...]
  must_contain:     [...]
```

---

## Adding a migration for a new version

1. Create `migration-spec/v54-v55/` (or whatever the version jump is).
2. Write one YAML file per concern following the format above.
3. Update `agents.md` with any new ordering rules or special cases for that version.

The YAML specs are the source of truth.

## Current v54 caveats

- Fatal specs must abort before any target mutations. A repo that hits
  `group.yaml` should remain clean after the failed run.
- Simapp-shaped apps may still require manual comparison against the current v54
  reference layout after the mechanical rewrites are applied.
- Published release candidates and the live v54 monorepo can drift. Validate
  against the actual target dependency graph and reference app layout instead of
  assuming a release-candidate module graph is a perfect proxy.
