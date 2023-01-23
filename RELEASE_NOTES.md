# Cosmos SDK v0.45.12 Release Notes

This release introduces a number of bug fixes and improvements. Notably with an update to State Streaming APIs.

Moreover, this release contains a store fix. The changes have been tested against the Cosmos Hub and Juno mainnet with no issues. However, there is a low probability of an edge case happening. Hence, it is recommended to do a **coordinated upgrade** to avoid any issues.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.45.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.45.11...v0.45.12

**NOTE:** The changes mentioned in `v0.45.9` are **no longer required**. The following replace directive can be removed from the chains.

```go
# Can be deleted from go.mod
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```

Instead, `github.com/confio/ics23/go` must be **bumped to `v0.9.0`**.
