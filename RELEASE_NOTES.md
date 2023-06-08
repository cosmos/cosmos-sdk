# Cosmos SDK v0.46.13 Release Notes

This release includes few improvements and bug fixes.
Notably, the [barberry security fix](https://forum.cosmos.network/t/cosmos-sdk-security-advisory-barberry/10825). All chains using Cosmos SDK v0.46.0 and above must upgrade to `v0.46.13` **immediately**. A chain is safe as soon as **33%+1** of the voting power has upgraded. Coordinate with your validators to upgrade as soon as possible.

Additionally, it includes new commands for snapshots management and bootstrapping from a local snapshot (add `snapshot.Cmd(appCreator)` to the chain root command for using it).

Did you know Cosmos SDK Twilight (a.k.a v0.47) has been released? Upgrade easily by reading the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md#v047x).

Ensure you have the following replaces in the `go.mod` of your application:

```go
// use cometbft
replace github.com/tendermint/tendermint => github.com/cometbft/cometbft v0.34.28
// replace broken goleveldb
replace github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
```

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.46.12...v0.46.13
