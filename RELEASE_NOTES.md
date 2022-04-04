# Cosmos SDK v0.44.7 Release Notes

This release introduces bug fixes and improvements on the Cosmos SDK v0.44 series:

- The `everything` prune strategy now keeps the latest 2 heights (instead of only the latest one). If this flag is set, state sync is automatically disabled.
- The bank keeper has a new `WithMintCoinsRestriction` method to allow custom module-specific restrictions on minting (e.g. only minting a certain denom). This function is not on the `bank.Keeper` interface, so it's not API-breaking, but only additive on the `Keeper` implementation; please use casting to access this method.
- Fix data race in store trace component.

See the [Cosmos SDK v0.44.7 Changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.44.7/CHANGELOG.md) for the exhaustive list of all changes.

**Full Changelog**: https://github.com/cosmos/cosmos-sdk/compare/v0.44.6...v0.44.7
