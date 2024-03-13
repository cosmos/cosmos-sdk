# Cosmos SDK v0.50.5 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/58)

## 🚀 Highlights

This is time for another patch release of Cosmos SDK Eden.
This release includes a few notable fixes:

* Fix a bypass delegator slashing: [GHSA-86h5-xcpx-cfqc](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-86h5-xcpx-cfqc)
* Fix an issue in `baseapp.ValidateVoteExtensions` helper: [GHSA-95rx-m9m5-m94v](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-95rx-m9m5-m94v)
* Allow to provide custom signers for `x/auth/tx` using depinject

We recommended to upgrade to this patch release as soon as possible.  
When upgrading from <= v0.50.4, please ensure that 2/3 of the validator power upgrade to v0.50.5.

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.50.5/CHANGELOG.md) for an exhaustive list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/release/v0.50.4...v0.50.5) from the last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) when migrating from `v0.47.x` to `v0.50.1`.
Note, that the next SDK release, v0.51, will not include `x/params` migration, when migrating from < v0.47, v0.50.x **or** v0.47.x, is a mandatory migration.
