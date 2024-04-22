# Cosmos SDK v0.50.6 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/58)

## 🚀 Highlights

For this month patch release of the v0.50.x line, a few features and improvements were added to the SDK.

Notably, we added and fixed the following:

* Add start customizability to start command options. Customize how an application starts with the new `StartCommandHandler` field in `server.StartCmdOptions` struct.
* Fixing [GHSA-4j93-fm92-rp4m](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-4j93-fm92-rp4m) in `x/feegrant` and `x/authz` modules. The upgrade instuctions were provided in the [v0.50.4 release notes](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.4).

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.50.6/CHANGELOG.md) for an exhaustive list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/release/v0.50.5...v0.50.6) from the last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) when migrating from `v0.47.x` to `v0.50.1`.
Note, that the next SDK release, v0.51, will not include `x/params` migration, when migrating from < v0.47, v0.50.x **or** v0.47.x, is a mandatory migration.
