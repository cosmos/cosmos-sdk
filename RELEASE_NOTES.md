# Cosmos SDK v0.46.7 Release Notes

This release introduces bug fixes and improvements. Notably, the upgrade to Tendermint [v0.34.24](https://github.com/tendermint/tendermint/releases/tag/v0.34.24).

Please read the release notes of [v0.46.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.5) if you are upgrading from `<=0.46.4`.

A critical vulnerability has been fixed in the group module. For safety, `v0.46.5` and `v0.46.6` are retracted, even though chains not using the group module are not affected. When using the group module, please upgrade immediately to `v0.46.7`.

An issue has been discovered in the gov module's votes migration. It does not impact proposals and votes tallying, but the gRPC queries on votes are incorrect. This issue is fixed in `v0.46.7`, however:
- if your chain is already on v0.46 using `<= v0.46.6`, a **coordinated upgrade** to v0.46.7 is required. You can use the helper function `govv046.Migrate_V046_6_To_V046_7` for migrating a chain **already on v0.46 with versions <=v0.46.6** to the latest v0.46.7 correct state.
- if your chain is on a previous version <= v0.45, then simply use v0.46.7 when upgrading to v0.46.

**NOTE**: The changes mentioned in `v0.46.3` are no longer required. The following replace directive can be removed from the chains.

```go
# Can be deleted from go.mod
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```

Instead, `github.com/confio/ics23/go` must be **bumped to `v0.9.0`**.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.6...v0.46.7
