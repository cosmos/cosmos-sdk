# Cosmos SDK v0.52.0-alpha.1 Release Notes

There are no release notes for pre-releases.

A beta release will be cut in the coming days after [audits](https://github.com/cosmos/cosmos-sdk/issues/21176).
Cosmos SDK modules won't be tagged before RC, when integrating with alphas and betas, use the latest pseudo version from the release branch for SDK modules or main for other packages. Refer to the [version matrix](https://github.com/cosmos/cosmos-sdk?tab=readme-ov-file#version-matrix) to understand what it means.
On the other hand, `cosmossdk.io/core` v1 beta will be cut [soon](https://github.com/cosmos/cosmos-sdk/issues/21176), to allow you to upgrade your modules easily.
Lastly, this alpha contains some dependencies annoyance that will be resolved before the RC (i.e. the SDK is temporarily importing `auth`, `bank` and `staking`).

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.52.x/CHANGELOG.md) for an exhaustive list of changes.  
Refer to the [UPGRADING.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.52.x/UPGRADING.md) for upgrading your application.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/release/v0.50.x...release/v0.52.x
