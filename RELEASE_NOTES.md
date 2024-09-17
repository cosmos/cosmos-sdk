# Cosmos SDK v0.50.10 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/58)

## 🚀 Highlights

For this month patch release of the v0.50.x line, some bugs were fixed.

Notably, we fixed the following:

* Add the root command `module-hash-by-height` to query and retrieve module hashes at a specific height
* `PreBlock` events (mainly `x/upgrade`) are now emitted (this time, for real)
* A fix in runtime baseapp option ordering, giving issue when other modules were having options

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/CHANGELOG.md) for an exhaustive list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/v0.50.9...v0.50.10) from the last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) when migrating from `v0.47.x` to `v0.50.1`.
Note, that the next SDK release, v0.52, will not include `x/params` migration, when migrating from < v0.47, v0.50.x **or** v0.47.x, is a mandatory migration.
