# Cosmos SDK v0.46.14 Release Notes

This patch release introduces a few bug fixes and improvements to the v0.46.x line of the Cosmos SDK. Notably, an improvement to `<appd> prune` UX and improving the error handling when there is a snapshot creation failure.

Ensure you have the following replaces in the `go.mod` of your application:

```go
// use cometbft
replace github.com/tendermint/tendermint => github.com/cometbft/cometbft v0.34.29
// replace broken goleveldb
replace github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
```

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.46.13...v0.46.14
