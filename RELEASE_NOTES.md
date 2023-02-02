# Cosmos SDK v0.45.13 Release Notes

This release introduces a one bug fix, namely [#14798](https://github.com/cosmos/cosmos-sdk/pull/14798).

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.45.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.45.12...v0.45.13

**NOTE:** The changes mentioned in `v0.45.9` are **no longer required**. The following replace directive can be removed from the chains.

```go
# Can be deleted from go.mod
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```

Instead, `github.com/confio/ics23/go` must be **bumped to `v0.9.0`**.
