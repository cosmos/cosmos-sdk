# Cosmos SDK v0.50.4 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/58)

## 🚀 Highlights

Some months ago Cosmos SDK Eden was released. Missed the announcement? Read it [here](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.1).
For this month patch release of the v0.50.x line, a few features and improvements were added to the SDK.

Notably, we added and fixed the following:

* Adds in-place testnet CLI command for creating testnets from local state (kudos to @czarcas7ic)
* Multiple fixes in baseapp, with fixes in `DefaultProposalHandler` and vote extensions
* Add a missed check in `x/auth/vesting`: [GHSA-4j93-fm92-rp4m](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-4j93-fm92-rp4m)

We recommended to upgrade to this patch release as soon as possible.  
When upgrading from <= v0.50.3, please ensure that 2/3 of the validator power upgrade to v0.50.4.

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.50.4/CHANGELOG.md) for an exhaustive list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/release/v0.50.3...v0.50.4) from the last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) when migrating from `v0.47.x` to `v0.50.1`.
Note, that the next SDK release, v0.51.0, will not include `x/params` migration, when migrating from < v0.47, v0.50.x **or** v0.47.x, is a mandatory migration.
