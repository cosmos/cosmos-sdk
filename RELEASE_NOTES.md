# Cosmos SDK v0.42.0 "Stargate" Release Notes

This release contains a single, but important security fix for all non "Cosmos Hub" chains (e.g. any chain that does not use the default `cosmos` bech32 prefix). The fix addresses a bug in evidence handling on the Cosmos SDK that rendered the `v0.41.x` and `v0.40.x` release series unsafe for most chains. Please see the PR below for more details.

## Bug Fixes

- [#8461](https://github.com/cosmos/cosmos-sdk/pull/8461) Fix bech32 prefix in evidence validator address conversion
