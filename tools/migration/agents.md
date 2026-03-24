# Migration Agent Orchestration Guide

This file tells an AI agent how to migrate a Cosmos SDK chain repo from one SDK
version to another. Read it in full before touching any files.

---

## What you are doing

You are migrating a Cosmos SDK chain from **v0.50.x–v0.53.x → v0.54.x**. The
specs apply to any chain on v0.50 or later — the module wiring, import paths, and
keeper patterns are structurally the same across v50–v53. Chains on older versions
(v0.47 or earlier) may need additional work beyond what these specs cover.

The goal is a chain repo that builds cleanly against the v54 SDK with no
references to removed or deprecated modules.

The migration is **spec-driven**: each component with breaking changes has a YAML
spec in `tools/migration/migration-spec/v50-to-v54/`. Your job is to:

1. Scan the target repo to detect which specs apply.
2. Apply each relevant spec in dependency order.
3. Verify success.
4. Report any changes that required manual intervention.

---

## Where to look

```
tools/migration/
  agents.md                      ← this file
  migration-spec/v50-to-v54/        ← one YAML file per migration concern
    core.yaml                    ← always apply first
    store-v2.yaml
    bank-endblock.yaml
    crisis.yaml
    circuit.yaml
    nft.yaml
    group.yaml                   ← fatal if detected; stop immediately
    gov.yaml
    gov-hooks.yaml
    epochs.yaml
    epochs-app-module.yaml
    ante.yaml
    app-structure.yaml           ← apply last
```

The target chain repo (the one you are migrating) is a separate directory,
**not** this tools/ tree. Never modify files in tools/migration/ itself.

---

## How to detect the current version

```bash
# 1. Find go.mod files
find . -name go.mod | head -20

# 2. Check the SDK version
grep 'github.com/cosmos/cosmos-sdk' go.mod

# 3. Look for v53-specific import patterns
grep -r 'cosmossdk.io/x/' . --include='*.go' | head -20
grep -r 'x/crisis' . --include='*.go' | head -10
grep -r 'x/group' . --include='*.go' | head -10
```

If the SDK version in go.mod is `v0.50.x`, `v0.51.x`, `v0.52.x`, or `v0.53.x`,
proceed. The migration specs are valid for all of these — the breaking changes
between v50 and v53 are minimal and do not affect the patterns matched by these
specs. If the version is already `v0.54.x`, the migration may be partially done —
scan for leftovers before assuming it is complete.

---

## How to select specs

Run detection for each spec in order. A spec **applies** if any of its
`detection.imports`, `detection.patterns`, `detection.files`, or
`detection.go_mod` rules match.

```bash
# Check if a spec applies — example for crisis
grep -r 'cosmos/cosmos-sdk/x/crisis\|x/crisis' . --include='*.go' -l
grep -r 'app\.CrisisKeeper\|crisistypes\.StoreKey' . --include='*.go' -l
```

**Always apply in this order** (later specs may depend on earlier ones):

| Order | Spec             | Reason                                              |
|-------|------------------|-----------------------------------------------------|
| 1     | group.yaml       | Fatal check — halt immediately if group is found    |
| 2     | core.yaml        | Import rewrites must happen before other specs      |
| 3     | store-v2.yaml    | Store import and BaseApp helper removals            |
| 4     | bank-endblock.yaml | Put x/bank first in SetOrderEndBlockers          |
| 5     | crisis.yaml      | Keeper/module removal                               |
| 6     | circuit.yaml     | Import rewrite + ante cleanup                       |
| 7     | nft.yaml         | Import rewrite                                      |
| 8     | gov.yaml         | Keeper signature surgery                            |
| 9     | gov-hooks.yaml   | AfterProposalSubmission signature update            |
| 10    | epochs.yaml      | Keeper pointer change                               |
| 11    | epochs-app-module.yaml | Remove stale keeper dereferences             |
| 12    | ante.yaml        | File deletion + call-site rewrite                   |
| 13    | app-structure.yaml | DI file deletion + misc cleanup                  |

---

## How to reason about the migration

Before making any edits, build a mental model of the target app:

1. **What modules does it use?** Scan imports in `app/app.go` (or equivalent).
2. **Does it have custom wiring?** Look for anything beyond standard simapp patterns
   — extra keepers, custom ante handlers, non-standard module ordering.
3. **Are there chain-specific files?** Things like `app_di.go`, `root_di.go`,
   custom `ante.go`, or `app_config.go` with extra modules.
4. **Does it have its own Go modules?** Some chains split keepers into sub-modules
   with their own `go.mod`. Each sub-module needs its own `go.mod` update.

For each spec you are applying, read the spec's `description` field. It tells you
**why** the change is being made — essential context for handling edge cases.

---

## How to stage edits

Work **one spec at a time**. For each spec:

### Step 1 — Check detection
Confirm the spec applies using the patterns in `detection`.

### Step 2 — Handle fatal specs first
If `group.yaml` detection matches, **stop everything**. Print:

```
FATAL: x/group usage detected in <files>.
The group module is not supported by the automated v54 migration tool.
It requires a manual move to enterprise/group.
Contact Cosmos Labs before proceeding: https://cosmos.network/enterprise
```

Do not modify any files. Return control to the user.

### Step 3 — Apply structured changes in order

**Important**: Text replacements are match-based, so applying a spec twice is
generally safe — the `old` pattern simply won't be found the second time. For
structural rewrites such as the gov keeper signature change, check whether the
new v0.54 form is already present before editing.

For each spec, apply changes in this sub-order:

1. **file_removals** — delete files first so AST doesn't process dead code
2. **go_mod changes** — update versions, remove modules, strip local replaces
3. **import rewrites** — rewrite import paths (AST-level)
4. **statement_removals** — remove keeper init and wiring statements
5. **map_entry_removals** — remove map/slice entries
6. **call_arg_edits** — remove/add arguments to specific function calls
7. **special_cases** — apply targeted structural rewrites described by the spec
8. **manual_steps** — apply whatever still cannot be handled safely
9. **text_replacements** — post-structural cleanup

### Step 4 — Emit warnings

After applying a spec, emit any `imports.warnings` messages in the output so
the user can see what requires attention (e.g., contrib modules that won't be
maintained).

### Step 5 — Verify
Run the checks in `verification` to confirm the spec was applied cleanly.

---

## What commands to run

After all specs are applied:

```bash
# 1. Update go.mod dependencies
go mod tidy

# 2. Verify the build
go build ./...

# 3. Run tests
go test ./...

# 4. Run the repo's lint command if one exists

# 5. Check for leftover references to removed modules
grep -r 'cosmossdk.io/x/circuit\|cosmossdk.io/x/nft' . --include='*.go'
grep -r 'x/crisis' . --include='*.go'
grep -r 'x/group' . --include='*.go'
grep -r 'NewUncachedContext\|SimWriteState\|SetStoreMetrics' . --include='*.go'
```

If `go mod tidy` fails, common causes:

- **Version constraint conflict** — two modules require incompatible SDK versions.
  Pin the SDK version explicitly in go.mod and re-run tidy.
- **Missing module** — a module was removed from go.mod but is still imported.
  Find the dangling import and either remove it or add the module back.
- **Replace directive** — a local-path replace (e.g., `replace foo => ../foo`) was
  not stripped. Run with `strip_local_replaces: true` or remove manually.
- **Forked replace directive** — a non-local replace (e.g., `replace github.com/cosmos/cosmos-sdk => github.com/yourorg/cosmos-sdk ...`)
  was preserved. Do not delete it blindly; audit whether the fork has a compatible
  v54 target or whether the chain must migrate off the fork first.

If `go build ./...` fails, read the errors carefully. Common causes:

- **Missing import** — a text replacement removed a line that was needed elsewhere.
  Check if the pattern was too broad (`file_match` can scope it).
- **Dangling reference** — a symbol was used but its import was stripped. Find
  where the symbol is still referenced and either remove the reference or re-add
  a scoped import.
- **Type mismatch** — a keeper type changed (e.g., value → pointer). Check
  epochs.yaml and the struct definition.
- **Custom ante wrapper** — if the chain has a project-local `HandlerOptions`
  or additional ante decorators, do not force the simapp rewrite. Keep the local
  wrapper and migrate it manually.
- **Wrong argument count** — the gov NewKeeper surgery may not have applied if
  the call site had fewer than 9 args. Check manually.

---

## How to handle chain-specific edge cases

### Custom module aliases
Some chains alias packages differently from simapp. For example, the gov keeper
package might be imported as `keeper` not `govkeeper`. The migration tool resolves
aliases from the AST, but if you're applying changes manually:

```bash
# Find the actual alias used
grep 'cosmos-sdk/x/gov/keeper' . -r --include='*.go'
```

Use whatever alias the file declares — do NOT hardcode `govkeeper`.

### Extra modules in the app
If the chain has modules beyond the standard simapp set (e.g., dydx-specific
modules like x/vault, x/clob, x/listing, x/slinky), those are **out of scope**
for these specs. Do not remove or modify them. Apply only the changes described
in the specs and leave chain-specific wiring untouched.

### Multiple go.mod files
Cosmos chains often have sub-modules (e.g., `proto/`, `api/`, `testutil/`).
Each needs its SDK dependency updated independently:

```bash
find . -name go.mod | xargs grep 'cosmos/cosmos-sdk'
```

Apply `go_mod.update` to each file that references the SDK.

### Chains on v0.50 (not v0.53)
Chains like dydx are on SDK v0.50.x, not v0.53.x. The breaking changes between
v50 and v53 are minimal for the patterns these specs match — the same vanity URLs,
the same module wiring, the same keeper signatures. Apply the same specs. The only
differences you may encounter:

- Some cosmossdk.io/x/* packages may not have existed in v50 (they were split out
  later). If an import rewrite has no match, it's a no-op — safe to skip.
- v50 chains may have slightly different `go.mod` module versions for non-SDK
  dependencies. The `go_mod.update` entries are SDK-only; run `go mod tidy` after
  to resolve transitive deps.
- v50 chains may still use `github.com/cosmos/cosmos-sdk/x/crisis` directly
  (never migrated to the vanity URL). The detection and removal patterns match
  both paths.

### slinky / oracle modules
These may reference types from cosmossdk.io packages that have also moved.
Check each import individually and apply the relevant vanity-URL rewrite from
core.yaml. If a module is not in the core.yaml rewrites list, it may need a
separate spec or manual handling.

### The circuitante import in non-standard locations
Some chains include circuitante in files other than ante.go. The text replacement
in circuit.yaml strips the import line but leaves the usage. Search for all
occurrences:

```bash
grep -r 'circuitante' . --include='*.go'
```

Remove each circuitante usage manually and replace with the direct SDK ante
pattern if needed.

---

## What success looks like

A successful migration means:

- [ ] `go build ./...` passes with no errors
- [ ] `go test ./...` passes (or only pre-existing failures remain)
- [ ] No imports of `cosmossdk.io/x/circuit`, `cosmossdk.io/x/nft`, removed vanity URLs
- [ ] No references to `app.CrisisKeeper`, `crisistypes.StoreKey`, `x/crisis`
- [ ] `govkeeper.NewDefaultCalculateVoteResultsAndVotingPower` is present in gov keeper init
- [ ] `AfterProposalSubmission` declarations include the proposer address arg
- [ ] `app.EpochsKeeper = &epochsKeeper` is present (if epochs was used)
- [ ] No `NewUncachedContext`, `SimWriteState`, or `SetStoreMetrics` calls remain
- [ ] `ante.go` is deleted (if it contained circuitante)
- [ ] End-of-run warnings emitted for circuit and nft (expected — not failures)
- [ ] `go.mod` versions updated and `go mod tidy` clean

---

## Reporting

At the end of the migration, produce a summary:

```
Migration complete: v53 → v54

Specs applied:
  ✓ core
  ✓ crisis
  ✓ circuit   ⚠ warning: contrib module
  ✓ nft       ⚠ warning: contrib module
  ✓ gov
  ✓ epochs
  ✓ ante
  ✓ app-structure

Warnings (action required):
  - circuit: this module has been moved to contrib and will not be maintained
    by the Cosmos SDK team. Files affected: app/app.go, x/circuit/...
  - nft: this module has been moved to contrib and will not be maintained
    by the Cosmos SDK team. Files affected: app/app.go, x/nft/...

Manual steps remaining:
  - None

Build status: PASS
Test status: PASS
```

If group was detected and migration was halted:

```
Migration HALTED: x/group detected

The following files reference x/group and must be resolved before migration
can proceed:

  app/app.go:42   github.com/cosmos/cosmos-sdk/x/group
  app/app.go:1337 app.GroupKeeper = groupkeeper.NewKeeper(...)

Action required: Contact Cosmos Labs to obtain an enterprise/group license,
then follow the enterprise migration guide before re-running this tool.
```
