# Cosmos SDK v0.46.11 Release Notes

This release includes the migration to CometBFT. This migration should be not breaking.
From `v0.46.11`+, the following replace is *mandatory* in the `go.mod` of your application:

```go
// use cometbft
replace github.com/tendermint/tendermint => github.com/cometbft/cometbft v0.34.27
```

A more complete migration is happening in Cosmos SDK v0.47.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.46.10...v0.46.11