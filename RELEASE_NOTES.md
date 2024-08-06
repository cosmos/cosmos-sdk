# Cosmos SDK v0.50.9 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/58)

## 🚀 Highlights

For this month patch release of the v0.50.x line, some bugs were fixed.

Notably, we fixed the following:

* `PreBlock` events (mainly `x/upgrade`) are now emitted
* Improve compatibility of depinject v1.0.0 with `app.yaml` / `app.json`

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.50.9/CHANGELOG.md) for an exhaustive list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/v0.50.8...v0.50.9) from the last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) when migrating from `v0.47.x` to `v0.50.1`.
Note, that the next SDK release, v0.52, will not include `x/params` migration, when migrating from < v0.47, v0.50.x **or** v0.47.x, is a mandatory migration.
