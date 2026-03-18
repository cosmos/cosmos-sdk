# Migration Framework

A spec-driven, AST-based framework for migrating Cosmos SDK chains across breaking version changes.

## Overview

Migrating a chain to a new SDK version involves dozens of coordinated changes: import rewrites, keeper signature updates, module removals, file deletions, and go.mod bumps. Encoding all of this as a single monolithic script breaks down quickly when chains diverge from simapp (custom aliases, extra modules, chain-specific wiring).

This framework takes a three-pronged approach:

1. **`migration-spec/`** — YAML files that are the authoritative source of truth for each migration concern. One file per component/version jump. Human-readable, diffable, and reusable across chains.

2. **`agents.md`** — An orchestration guide for AI agents. Explains where to look, how to reason about the migration, what commands to run, how to stage edits, and what success looks like. An agent that reads this file can migrate a chain without any compiled tooling.

3. **Go AST engine** (`migrate.go` + supporting files) — A compiled Go library that executes the structured rules in the migration specs programmatically. Powers the `v54/` binary for automated, verifiable migrations.

---

## Directory structure

```
tools/migration/
  agents.md                      ← AI agent orchestration guide
  migration-spec/
    v50-to-v54/                     ← specs for the v50+ → v54 upgrade
      core.yaml                  ← SDK version bumps, vanity URL rewrites (always apply first)
      crisis.yaml                ← Remove x/crisis entirely
      circuit.yaml               ← x/circuit moved to contrib (keep, warn)
      nft.yaml                   ← x/nft moved to contrib (keep, warn)
      group.yaml                 ← x/group → enterprise (fatal, halt migration)
      gov.yaml                   ← govkeeper.NewKeeper signature change
      epochs.yaml                ← EpochsKeeper value → pointer
      ante.yaml                  ← Remove custom ante.go wrapper
      app-structure.yaml         ← DI files, misc ordering cleanup
  migrate.go                     ← Core migration engine
  v54/                           ← Compiled v53→v54 migration binary
    main.go
    calls.go
    imports.go
    removals.go
```

---

## How to use

### For AI-assisted migration (recommended for real chains)

Point an AI agent (Claude, etc.) at `agents.md` and the target chain repo:

> "Read tools/migration/agents.md and then migrate the chain at <path> to v54."

The agent reads the YAML specs, scans the repo to detect which modules are in
use, applies changes using standard code editing (AST rewrites, text
replacements, file deletions), and verifies the result builds. This is the
recommended path for chains with custom wiring because the agent can reason about
edge cases that a static binary cannot handle — different import aliases,
non-standard file layouts, chain-specific modules like slinky, etc.

### For simapp / standard chains via the Go binary

```bash
cd tools/migration/v54
go run . --dir /path/to/chain-repo
```

The compiled Go binary executes the same rules described in the YAML specs, but
as hardcoded Go structs. It works well for standard simapp-shaped repos. For
chains with significant customisation, use the agent-assisted path above.

### For manual migration

Read each relevant YAML spec in `migration-spec/v50-to-v54/` for the components
your chain uses. Each spec has a `description`, `changes`, `manual_steps`, and
`verification` section that together describe exactly what to do and how to
confirm it worked.

### E2E regression testing

The `e2e/` directory contains a test runner that clones real chain repos from
GitHub, applies the migration via the Go binary, and verifies the result:

```bash
cd tools/migration/e2e
./run.sh              # all non-skipped chains
./run.sh simapp-v53   # a specific chain
./run.sh dydx-v4      # test detection on a complex chain
```

See `e2e/README.md` for details on adding chains and interpreting results.

---

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
4. Optionally: create a `v55/` compiled binary that executes the specs programmatically.

The YAML specs are the source of truth. The Go binary is an optimisation — not a requirement.

---

## Go AST engine capabilities

The core engine (`migrate.go`) supports:

- **Import path rewriting** — with wildcard sub-package matching and alias-aware Except handling
- **Import warnings** — emit warnings (or fatal errors) for imports that require manual attention
- **Statement removal** — remove assignment and expression statements by target pattern, with support for removing preceding/following associated statements
- **Map entry removal** — remove key-value pairs from map literals by key pattern
- **Call argument edits** — remove or insert arguments in specific function calls, matched by function pattern or method name
- **AST argument surgery** — complex argument transformations via Go callbacks (for cases where the new argument must be synthesised from existing ones, and the package alias is dynamic)
- **Type and struct field changes** — rename types, remove or modify struct fields
- **Text replacements** — post-AST find-and-replace for patterns too complex for AST manipulation (multi-line, deeply nested, or file-scoped)
- **File removal** — delete files with a content safety check before deletion
- **go.mod management** — update, add, remove, and strip replace directives
