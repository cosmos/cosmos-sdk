# Cosmos SDK v0.46.13 Release Notes

This release includes a few improvements and bug fixes.
Notably, a bump to [CometBFT v0.34.28](https://github.com/cometbft/cometbft/blob/v0.34.28/CHANGELOG.md#v03428).
Additionally, it includes new commands for snapshots management and bootstrapping from a local snapshot.

Did you know Cosmos SDK Twilight (a.k.a v0.47) has been released? Upgrade easily by reading the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md#v047x).

Note, from `v0.46.11`+, the following replace is *mandatory* in the `go.mod` of your application:

```go
// use cometbft
replace github.com/tendermint/tendermint => github.com/cometbft/cometbft v0.34.28
```

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.46.12...v0.46.13
