# Cosmos SDK v0.55.1 Release Notes

## Highlights

Cosmos SDK v0.55.1 is a patch release with three changes: a fix for fee allowance expiry queue entries that the keeper cannot decode on chains that upgraded from v0.53.x or earlier, IAVL v1.2.9, and patched Go and dependency versions.

Upgrading from v0.55.0 **requires a coordinated chain upgrade**. The `x/feegrant` consensus version goes from 2 to 3, and its store migration runs through `RunMigrations` in the upgrade handler.

### x/feegrant migration

Chains that upgraded from v0.53.x or earlier can store fee allowance expiry queue entries with an empty value. The keeper cannot decode them, so the feegrant `EndBlocker` returns an error at the first block after such an entry expires. The v2 to v3 migration rewrites those entries.

A chain is affected if it upgraded from v0.53.x or earlier and still has unexpired fee grants with an expiration that were created before that upgrade. Affected chains must run the migration before the earliest of those grants expires. Chains that started on v0.54 or later are not affected.

See the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/v0.55.1/UPGRADING.md#xfeegrant-v0551) for how to check a chain.

### Other changes

* IAVL v1.2.9: snapshot import returns an error on malformed input instead of panicking.
* Go 1.26.11 and patched dependencies (klauspost/compress, grpc, golang.org/x, pion, otel).

## Changelog

See the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.55.1/CHANGELOG.md) for the full list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/v0.55.0...v0.55.1) from the last release.
