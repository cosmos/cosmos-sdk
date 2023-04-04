# Cosmos SDK v0.46.12 Release Notes

This release introduces a number of improvements and bug fixes, notably a new query for the `x/group` module, for querying all groups on a chain.

Note, from `v0.46.11`+, the following replace is *mandatory* in the `go.mod` of your application:

```go
// use cometbft
replace github.com/tendermint/tendermint => github.com/cometbft/cometbft v0.34.27
```

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.46.11...v0.46.12
