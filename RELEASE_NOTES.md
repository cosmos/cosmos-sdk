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

## Maintenance Policy

Cosmos SDK Olympus (v0.52) final release is approaching really soon. That means the Eden line (v0.50.x) will soon only be supported for bug fixes only, as per our release policy. Earlier versions are not maintained.  

Note, that the next SDK release, v0.52, does not include `x/params` migration, when migrating from < v0.47, v0.50.x **or** v0.47.x, is a mandatory migration.

Start integrating with [Cosmos SDK Eden (v0.52)](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md#v052x) and enjoy and the new features and performance improvements.
