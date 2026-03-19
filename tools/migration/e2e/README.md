# E2E Migration Tests

Regression tests for the v54 migration using the compiled Go binary against
real chain repos from GitHub.

> **Note on execution paths**: This E2E suite uses the Go binary (`v54/`) as
> the migration executor. This is the **regression testing** path — it verifies
> that the compiled rules produce correct output for known chains. The
> **production migration** path is agent-driven: an AI agent reads `agents.md`
> and the YAML specs, then applies changes directly. The two paths should
> produce equivalent results for standard patterns; the agent path handles
> chain-specific edge cases that the Go binary cannot.

## Quick start

```bash
# Run all non-skipped chains
./run.sh

# Run a specific chain
./run.sh simapp-v53

# Run multiple chains
./run.sh simapp-v53 dydx-v4

# Keep work directory for debugging
E2E_KEEP=1 ./run.sh simapp-v53

# Skip the go build step (just test detection + migration + verification)
E2E_SKIP_BUILD=1 ./run.sh dydx-v4

# Use a pre-built migration binary
MIGRATE_BIN=./migrate-v54 ./run.sh
```

## Requirements

- Go 1.23+
- git
- python3 with pyyaml (`pip install pyyaml`)

## Adding a new test chain

Edit `chains.yaml` and add an entry:

```yaml
- id: my-chain
  repo: https://github.com/org/chain.git
  ref: v1.0.0                    # tag, branch, or commit
  sdk_version: v0.50.x
  app_dir: .                     # where go.mod lives
  migrate_dir: .                 # what to pass to --dir
  expected:
    specs:                       # which spec IDs should be detected
      - core-sdk-migration
      - crisis-removal
    warnings:                    # which specs emit warnings
      - circuit-contrib-migration
    fatal: null                  # set to a spec ID if migration should halt
    build: true                  # should `go build ./...` pass?
  skip: false
  notes: >
    Description of this chain's quirks.
```

Set `build: false` for chains with heavy custom modules where `go build` isn't
expected to pass from migration alone. The verification checks will still run.

Set `skip: true` for chains you want in the registry but aren't ready to test
yet. Explicitly naming a skipped chain on the command line will still run it:
`./run.sh osmosis`.

## What it tests

For each chain, the runner:

1. Clones the repo at the specified ref into a temp directory
2. Builds the migration tool from `../v54/`
3. Runs the migration binary against the chain's `migrate_dir`
4. Checks for expected fatal halts (group module)
5. Runs `go mod tidy`
6. Runs verification checks from all specs in `migration-spec/v50-to-v54/`:
   - `must_not_import` — these import paths must not appear in any `.go` file
   - `must_not_contain` — these patterns must not appear
   - `must_contain` — these patterns must appear somewhere
7. Runs `go build ./...` (if `build: true`)
8. Reports pass/fail per chain with a summary

## Interpreting results

```
[e2e] ━━━ simapp-v53 (v0.53.0) ━━━
[e2e]   Cloning ...
  ✓ Cloned
  ✓ Migration completed
  ⚠ circuit: this module has been moved to contrib...
  ⚠ nft: this module has been moved to contrib...
  PASS [core-sdk-migration]
  PASS [crisis-removal]
  ...
  ✓ Build succeeded

[e2e] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[e2e] Results: 2/2 passed, 0 failed
```

- **✓** = step passed
- **⚠** = warning (expected, not a failure)
- **✗** = failure (verification or build)
- **PASS [spec-id]** = all verification checks for that spec passed
- **FAIL [spec-id]** = a verification check failed (details follow)
