# Cosmos SDK v0.44.5 Release Notes

This release introduces bug fixes and improvements on the Cosmos SDK v0.44 series:

- Emit ante handler events for failed transactions: ant events can cause blockchain change (eg tx fees) and related events should be emitted.
- (fix) Upgrade IAVL to 0.17.3 to solve race condition bug in IAVL.

See the [Cosmos SDK v0.44.5 Changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.44.5/CHANGELOG.md) for the exhaustive list of all changes.

**Full Changelog**: https://github.com/cosmos/cosmos-sdk/compare/v0.44.4...v0.44.5
