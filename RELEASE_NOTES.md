# Cosmos SDK v0.47.3 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/categories/announcements)

## 🚀 Highlights

Missed the v0.47.0 announcement? Read it [here](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.0).
For this third patch release of the `v0.47.x` line, some of the notable changes include:

* A command to be able to bootstrap comet from a local snapshot with [`<app> comet bootstrap-state`](https://docs.cosmos.network/v0.47/run-node/run-node#local-state-sync).
* The default logger is now `cosmossdk.io/log`, which supports coloring 🟥🟩🟪🟦 and filtering again.
* A bug fix in `x/group` migration. Chains migrating from v0.46.x to v0.47.x must use at least v0.47.**3**.

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.47.3/CHANGELOG.md) for an exhaustive list of changes or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/release/v0.47.2...v0.47.3) from last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md) when migrating from `v0.46.x` to `v0.47.0`.
