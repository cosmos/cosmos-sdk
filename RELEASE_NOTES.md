# Cosmos SDK v0.50.2 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/58)

## 🚀 Highlights

A month ago, Cosmos SDK Eden was released. Missed the announcement? Read it [here](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.1).
For second patch release of the v0.50.x line, a few features and improvements were added to the SDK.

Notably, we added and fixed the following:

* Allow to import base64 encoded pubkeys in the keyring using `<appd> keys add <name> --pubkey-base64 <base64-pubkey>` 
* A bug when migrating from v0.45/v0.46 directly to v0.50 due to missing `ConsensusParams` 
* An issue when simulating gas for transactions when using a multisig

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.50.2/CHANGELOG.md) for an exhaustive list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/release/v0.50.1...v0.50.2) from the last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) when migrating from `v0.47.x` to `v0.50.1`.
Note, that the next SDK release, v0.51.0, will not include `x/params` migration, when migrating from < v0.47, v0.50.x **or** v0.47.x, is a mandatory migration.
