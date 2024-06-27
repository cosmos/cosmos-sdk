# Cosmos SDK v0.50.7 Release Notes

💬 [**Release Discussion**](https://github.com/orgs/cosmos/discussions/58)

## 🚀 Highlights

For this month patch release of the v0.50.x line, a few improvements were added to the SDK and some bugs were fixed.

Notably, we added and fixed the following:

* Add extra checks in x/consensus `MsgUpdateParams` to prevent footguns when updating the consensus parameters.
    * Forgetting a field in a x/consensus parameter change gov proposal could lead to a chain halt.
    * The fix is in theory consensus breaking, but in practice, it is only a footgun prevention (the path only triggers if the proposal was executed and was invalid). Please ensure that all validators are on v0.50.7 before the execution of a `x/consensus` params update proposal.
* Remove txs from the mempool when they fail in RecheckTX

## 📝 Changelog

Check out the [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.50.7/CHANGELOG.md) for an exhaustive list of changes, or [compare changes](https://github.com/cosmos/cosmos-sdk/compare/release/v0.50.6...v0.50.7) from the last release.

Refer to the [upgrading guide](https://github.com/cosmos/cosmos-sdk/blob/release/v0.50.x/UPGRADING.md) when migrating from `v0.47.x` to `v0.50.1`.
Note, that the next SDK release, v0.51, will not include `x/params` migration, when migrating from < v0.47, v0.50.x **or** v0.47.x, is a mandatory migration.
