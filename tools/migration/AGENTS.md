# Migration Agent Orchestration Guide

This file tells an AI agent how to migrate a Cosmos SDK chain repo from
`v0.50.x` through `v0.53.x` to `v0.54.x` using the structured specs in this
repository.

## What you are doing

Your job is to:

1. Scan the target chain repo.
2. Detect the current SDK version and enabled migration concerns.
3. Select the relevant specs in dependency order.
4. Produce and execute a migration plan in small, reviewable steps.
5. Run verification until the repo is in a clean v54 state.

The source of truth is the YAML spec set in
`tools/migration/migration-spec/v50-to-v54/`. The `migrate-to-v54` CLI is a
scanner, planner, and verifier. It does not replace agent judgment.

## Where to look

Start here:

- `tools/migration/agents.md`
- `tools/migration/README.md`
- `tools/migration/migration-spec/v50-to-v54/*.yaml`
- `tools/migration/cmd/migrate-to-v54`

Then inspect the target chain repo:

- `go.mod` and any nested `go.mod` files
- `app/app.go`, `app.go`, `app_config.go`, `app_di.go`, `root_di.go`
- custom `ante.go` files
- module wiring, keeper construction, and imports

## How to reason about the migration

Work spec-by-spec, not file-by-file.

Apply this order:

1. `group.yaml`
2. `core.yaml`
3. `crisis.yaml`
4. `circuit.yaml`
5. `nft.yaml`
6. `gov.yaml`
7. `epochs.yaml`
8. `ante.yaml`
9. `app-structure.yaml`

Key rules:

- `group.yaml` is blocking. If group usage is detected, stop and report it.
- `core.yaml` should run first because it sets the baseline dependency and
  import state.
- Chain-specific modules stay in place unless a spec explicitly says otherwise.
- If a spec describes a manual step, make the smallest edit that satisfies the
  spec and preserves existing chain wiring.

## What commands to run

From this repository:

```bash
cd tools/migration
go run ./cmd/migrate-to-v54 --repo /path/to/chain scan
go run ./cmd/migrate-to-v54 --repo /path/to/chain plan
go run ./cmd/migrate-to-v54 --repo /path/to/chain verify
```

Optional deeper verification after edits:

```bash
cd tools/migration
go run ./cmd/migrate-to-v54 --repo /path/to/chain verify --go-mod-tidy --go-build --go-test
```

Useful direct inspection commands in the target repo:

```bash
find . -name go.mod
rg "github.com/cosmos/cosmos-sdk" -g'go.mod'
rg "x/crisis|x/group|x/circuit|x/nft|x/gov|x/epochs" -g'*.go'
rg "NewKeeper|NewAnteHandler|app_di.go|root_di.go" -g'*.go'
```

## How to stage edits

Take one concern at a time.

For each selected spec:

1. Confirm the detection signals in the target repo.
2. Read the spec description and manual steps.
3. Apply the minimum set of edits required by that spec.
4. Re-run `plan` or `verify` before moving to the next spec if the change is
   structurally important.

Do not mix unrelated cleanup into the migration.

## What success looks like

The migration is complete when:

- the repo is on `github.com/cosmos/cosmos-sdk v0.54.x`
- blocking specs are either resolved or explicitly escalated
- selected specs no longer fail their verification checks
- `go mod tidy`, `go build ./...`, and `go test ./...` pass when run
- no leftover references remain to removed modules or pre-v54 wiring that the
  selected specs are responsible for

## How to report status

When you finish, report:

- detected SDK version(s)
- selected specs in execution order
- edits applied for each spec
- manual interventions taken
- verification results
- any remaining blockers or follow-up work
