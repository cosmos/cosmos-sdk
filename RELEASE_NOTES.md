# Cosmos SDK v0.50.8 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/58)

## 🚀 Highlights

For this month patch release of the v0.50.x line, a few improvements were added to the SDK and some bugs were fixed.

Notably, we added and fixed the following:

* Allow to import private key from mnemonic file using `<appd> keys add testing --recover --source ./mnemonic.txt`
* Fixed the json parsing in `simd q wait-tx`

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.50.8/CHANGELOG.md) for an exhaustive list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/v0.50.7...v0.50.8) from the last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) when migrating from `v0.47.x` to `v0.50.1`.
Note, that the next SDK release, v0.51, will not include `x/params` migration, when migrating from < v0.47, v0.50.x **or** v0.47.x, is a mandatory migration.
