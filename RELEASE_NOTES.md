# Cosmos SDK v0.50.12 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/58)

## 🚀 Highlights

This patch release fixes [GHSA-x5vx-95h7-rv4p](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-x5vx-95h7-rv4p).
It resolves a `x/group` module issue that can halt chain when handling a malicious proposal.
Only users of the `x/group` module are affected by this issue.

We recommended to upgrade to this patch release as soon as possible.
When upgrading from <= v0.50.11, please use a chain upgrade to ensure that 2/3 of the validator power upgrade to v0.50.12.

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.50.12/CHANGELOG.md) for an exhaustive list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/v0.50.11...v0.50.12) from the last release.
