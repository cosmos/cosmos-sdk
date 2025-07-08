# Cosmos SDK v0.53.3 Release Notes

## 🚀 Highlights

This patch release fixes [GHSA-p22h-3m2v-cmgh](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-p22h-3m2v-cmgh).
It resolves a `x/distribution` module issue that can halt chains when the historical rewards pool overflows.
Chains using the `x/distribution` module are affected by this issue.

We recommended upgrading to this patch release as soon as possible.

This patch is state-breaking; chains must perform a coordinated upgrade. This patch cannot be applied in a rolling upgrade.

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.53.3/CHANGELOG.md) for an exhaustive list of changes or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/v0.53.2...v0.53.3) from the last release.