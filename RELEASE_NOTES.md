# Cosmos SDK v0.50.11 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/58)

## 🚀 Highlights

We are back on schedule for our monthly v0.50.x patch releases.
The last two months, next to ramping up on v0.52 and v2, we added a few bug fixes and (UX) improvements.

Notable changes:

* New Linux-only backend that adds Linux kernel's `keyctl` support
* Fix fallback genesis path in server
* Skip sims test when running dry on validators

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.50.11/CHANGELOG.md) for an exhaustive list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/v0.50.10...v0.50.11) from the last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) when migrating from `v0.47.x` to `v0.50.1`.
Note, that the next SDK release, v0.52, will not include `x/params` migration, when migrating from < v0.47, v0.50.x **or** v0.47.x, is a mandatory migration.
