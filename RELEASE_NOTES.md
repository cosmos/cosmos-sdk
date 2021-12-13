# Cosmos SDK v0.45.0 Release Notes

Cosmos SDK v0.45.0 is a logical continuation of the v0.44.\* series, but brings a couple of state- and API-breaking changes requested by the community.

**State-Breaking Changes**:

- We now charge gas in two new places: on `.Seek()` even if there are no entries, and for the key length (on top of the value length).

**API-Breaking Changes**:

- The `BankKeeper` interface has a new `HasSupply` method to ensure that input denom actually exists on chain.
- The `CommitMultiStore` interface contains a new `SetIAVLCacheSize` method for a configurable IAVL cache size.

See our [CHANGELOG](./CHANGELOG.md) for the exhaustive list of all changes, or a full [commit diff](https://github.com/cosmos/cosmos-sdk/compare/v0.44.5...v0.45.0).
