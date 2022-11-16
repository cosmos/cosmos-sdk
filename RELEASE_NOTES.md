# Cosmos SDK v0.46.5 Release Notes

This release introduces a number of bug fixes and improvements. Notably, an upgrade to Tendermint [v0.34.23](https://github.com/tendermint/tendermint/releases/tag/v0.34.23).

If your chain's state has coin metadata, an issue has been discovered in the bank module coin metadata migration. This issue is fixed in `v0.46.5`.  

* If you are planning to migrate to v0.46, please use `v0.46.5`. All releases prior to `v0.46.5`, **must NOT be used**. All previous version of `v0.46` are retracted.
* If your chain is already on v0.46 using `<= v0.46.4` and has coin metadata, a **coordinated upgrade** to `v0.46.5` is required.
* If your chain is already on v0.46 using `<= v0.46.4` but has no coin metadata, this release is **non-breaking**.

Moreover, an issue have been found in the group module. This issue is fixed in `v0.46.5`.

* If you use the group module, upgrade to `v0.46.5` **immediately**. A **coordinated upgrade** to `v0.46.5` is required.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.4...v0.46.5

**NOTE**: The changes mentioned in `v0.46.3` are **still** required:

```go
# Chains must add the following to their go.mod for the application:
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```
