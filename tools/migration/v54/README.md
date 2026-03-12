# v54 Migration Tool

Automated migration tool for upgrading a base Cosmos SDK chain from **v0.53 to v0.54**.

> **Scope:** This tool handles base SDK migrations for chains that do not use IBC or EVM.
> IBC and EVM migration support will be added separately.

## What it migrates

- **Go module updates** — bumps SDK and related dependency versions in `go.mod`
- **Import path rewrites** — migrates `cosmossdk.io/x/*` vanity URLs to `github.com/cosmos/cosmos-sdk/x/*`
- **Removed modules** — drops `go.mod` entries for modules folded back into the SDK monorepo
- **Type renames** — updates any renamed types/structs between v53 and v54
- **Function signature changes** — handles argument list changes in SDK APIs
- **Complex rewrites** — multi-statement replacements for deprecated helper functions

## Installation

```shell
cd tools/migration/v54
go install .
```

## Usage

Run in the root of your chain's Go module:

```shell
v54 .
```

Or specify a directory:

```shell
v54 path/to/your/chain
```

After running, finalize with:

```shell
go mod tidy
```

## Status

🚧 **Work in progress** — migration rules are being populated. The framework is ready; version-specific rules need to be filled in based on the final v54 breaking changes.
