# Cosmos SDK v0.46.11 Release Notes

The SDK sets its minimum version to Go 1.19. This is not because the SDK uses new Go 1.19 functionalities, but to signal that we recommend chains to upgrade to Go 1.19 — Go 1.18 is not supported by the Go Team anymore.
Note, that SDK recommends chains to use the same Go version across all of their network.
We recommend, as well, chains to perform a **coordinated upgrade** when migrating from Go 1.18 to Go 1.19.

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
