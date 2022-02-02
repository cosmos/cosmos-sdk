# Cosmos SDK v0.44.6 Release Notes

This release introduces bug fixes and improvements on the Cosmos SDK v0.44 series:

- Populate `ctx.ConsensusParams` for begin/end blockers.
- Significantly speedup iterator creation after delete heavy workloads, which significantly improves IBC migration times.
- Ensure that `LegacyAminoPubKey` struct correctly unmarshals from JSON.
- Add evidence to std/codec to be able to decode evidence in client interactions.

See the [Cosmos SDK v0.44.6 Changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.44.6/CHANGELOG.md) for the exhaustive list of all changes.

**Full Changelog**: https://github.com/cosmos/cosmos-sdk/compare/v0.44.5...v0.44.6
