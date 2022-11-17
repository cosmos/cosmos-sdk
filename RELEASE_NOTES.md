# Cosmos SDK v0.46.5 Release Notes

This release introduces a number of serious bug fixes and improvements. Notably, an upgrade to Tendermint [v0.34.23](https://github.com/tendermint/tendermint/releases/tag/v0.34.23).

If you are planning to migrate to v0.46, please use `v0.46.5`. All releases prior to `v0.46.5` are [retracted](https://go.dev/ref/mod#go-mod-file-retract) and **must NOT be used** (`go get` directly upgrades the SDK version to `>= v0.46.5` thanks to the retraction, current builds are not affected).

If your chain's state has coin metadata, an issue has been discovered in the bank module coin metadata migration. This issue is fixed in `v0.46.5`.  

* If your chain is already on v0.46 using `<= v0.46.4` and has coin metadata, a **coordinated upgrade** to `v0.46.5` is required.
    * Use the helper function `Migrate_V0464_To_V0465` for migrating a chain **already on v0.46 with versions <=v0.46.4** to the latest v0.46.5 correct state.
* If your chain is already on v0.46 using `<= v0.46.4` but has no coin metadata, this release is **non-breaking**.

Moreover, serious issues have been found in the group module. These issues are fixed in `v0.46.5`.

* If you use the group module, upgrade to `v0.46.5` **immediately**. A **coordinated upgrade** to `v0.46.5` is required.

When a chain is already using `<= v0.46.4`, but has no coin metadata and no group module, this release is **non-breaking**.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.4...v0.46.5

**NOTE**: The changes mentioned in `v0.46.3` are **still** required:

```go
# Chains must add the following to their go.mod for the application:
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```
