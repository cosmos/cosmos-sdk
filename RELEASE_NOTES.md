# Cosmos SDK v0.46.16 Release Notes

This patch release introduces one dependency bump in the v0.46.x line of the Cosmos SDK.

Ensure you have the following replaces in the `go.mod` of your application:

```go
// use cometbft
replace github.com/tendermint/tendermint => github.com/cometbft/cometbft v0.34.29
// replace broken goleveldb
replace github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
```

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/v0.46.16/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.46.15...v0.46.16

## End-of-Life Notice

`v0.46.16` is the last release of the `v0.46.x` line. Per this version, the v0.46.x line reached its end-of-life.
The SDK team maintains the [latest two major versions of the SDK](https://github.com/cosmos/cosmos-sdk/blob/main/RELEASE_PROCESS.md#major-release-maintenance). This means no features, improvements or bug fixes will be backported to the `v0.46.x` line. Per our policy, the `v0.46.x` line will receive security patches only.

We encourage all chains to upgrade to Cosmos SDK Eden (`v0.50.0`), or the `v0.47.x` line.
