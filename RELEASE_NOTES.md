# Cosmos SDK v0.45.15 Release Notes

This release includes the migration to [CometBFT v0.34.27](https://github.com/cometbft/cometbft/blob/v0.34.27/CHANGELOG.md#v03427).
This migration should be minimally breaking for chains.
From `v0.45.15`+, the following replace is *mandatory* in the `go.mod` of your application:

```go
// use cometbft
replace github.com/tendermint/tendermint => github.com/cometbft/cometbft v0.34.27
```

Additionally, the SDK sets its minimum version to Go 1.19. This is not because the SDK uses new Go 1.19 functionalities, but to signal that we recommend chains to upgrade to Go 1.19 — Go 1.18 is not supported by the Go Team anymore.
Note, that SDK recommends chains to use the same Go version across all of their network.
We recommend, as well, chains to perform a **coordinated upgrade** when migrating from Go 1.18 to Go 1.19.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.45.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.45.14...v0.45.15

## End-of-Life Notice

`v0.45.15` is the last release of the `v0.45.x` line. Per this version, the v0.45.x line reached its end-of-life.
The SDK team maintains the two latest major versions of the SDK. This means no features, improvements or bug fixes will be backported to the `v0.45.x` line. Per our policy, the `v0.45.x` line will receive security patches only.

We encourage all chains to upgrade to the latest release of the SDK, or the `v0.46.x` line.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md) for how to upgrade a chain to the latest release.

## FAQ Migration to CometBFT v0.34.27

### I use `tm-db` but I get an import error with `cometbft-db`

For preventing API breaking changes, the SDK team has kept using `tm-db` for `v0.45.x` and `v0.46.x`.
However, the CometBFT team kept using `cometbft-db` for their `v0.34.x` line.
This means if your app directly interact with CometBFT (e.g. for a force pruning command), you will need to use `cometbft-db` there.
When not interacting with CometBFT directly, you can use `tm-db` as usual.

### I get import errors with `btcd`

If you are using an old version of `btcd`, you will need to upgrade to the latest version.
The previous versions had vulnerabilities so the SDK and CometBFT have upgraded to the latest version.
In the latest version `btcsuite/btcd` and `btcsuite/btcd/btcec` are two separate go modules.

### I encounter state sync issues

Please ensure you have built the binary with the same Go version as the network.
You can easily verify that by querying `/cosmos/base/tendermint/v1beta1/node_info` of a node in the network, and checking the `go_version` field.
