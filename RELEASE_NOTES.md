# Cosmos SDK v0.47.16 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/6)

## 🚀 Highlights

This patch release fixes [GHSA-x5vx-95h7-rv4p](https://github.com/cosmos/cosmos-sdk/security/advisories/GHSA-x5vx-95h7-rv4p).
It resolves a `x/group` module issue that can halt chain when handling a malicious proposal.
Only users of the `x/group` module are affected by this issue.

We recommended to upgrade to this patch release as soon as possible.
When upgrading from <= v0.47.15, please use a chain upgrade to ensure that 2/3 of the validator power upgrade to v0.47.16.

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.47.16/CHANGELOG.md) for an exhaustive list of changes or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/v0.47.15...v0.47.16) from last release.
