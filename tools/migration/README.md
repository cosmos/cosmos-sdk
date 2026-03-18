# Migration Helper

This directory contains the new migration-helper layout for Cosmos SDK `v0.54`
upgrades.

It is intentionally split into three parts:

1. `migration-spec/`
   Structured YAML rules for each migration concern and version jump.
2. `AGENTS.md`
   An orchestration guide for AI-assisted migrations.
3. `cmd/migrate-to-v54`
   A fresh CLI that scans a chain repo, selects the relevant specs, generates a
   migration plan, and runs verification.

## Layout

```text
tools/migration/
  README.md
  AGENTS.md
  go.mod
  migration-spec/
    v50-to-v54/
      *.yaml
  cmd/
    migrate-to-v54/
      main.go
  internal/
    migrationtool/
      load.go
      scan.go
      verify.go
```

The migration-specific agent guide lives at `tools/migration/agents.md`.

## Usage

```bash
cd tools/migration
go run ./cmd/migrate-to-v54 --repo /path/to/chain scan
go run ./cmd/migrate-to-v54 --repo /path/to/chain plan
go run ./cmd/migrate-to-v54 --repo /path/to/chain verify
```

Optional command execution during verification:

```bash
go run ./cmd/migrate-to-v54 --repo /path/to/chain verify --go-mod-tidy --go-build --go-test
```

## Output model

- `scan` reports detected SDK versions and repo signals.
- `plan` reports the ordered spec list and what each spec requires.
- `verify` runs each selected spec's verification checks and can optionally run
  `go mod tidy`, `go build ./...`, and `go test ./...` inside the target repo.
