# Cosmos SDK v0.46.8 Release Notes

This release introduces bug fixes and improvements. Notably, the SDK have now switched to Informal Systems' Tendermint fork.
Their fork has no changes compared to the upstream Tendermint, but it is now [maintained by Informal Systems](https://twitter.com/informalinc/status/1613580954383040512). Chains are invited to do the same.

Moreover, this release contains a store fix. The changes have been tested against a v0.46.x chain mainnet with no issues. However, there is a low probability of an edge case happening. Hence, it is recommended to do a **coordinated upgrade** to avoid any issues.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.7...v0.46.8

**NOTE**: The changes mentioned in `v0.46.3` are no longer required. The following replace directive can be removed from the chains.

```go
# Can be deleted from go.mod
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```

Instead, `github.com/confio/ics23/go` must be **bumped to `v0.9.0`**.
