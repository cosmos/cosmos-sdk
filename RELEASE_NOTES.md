# Cosmos SDK v0.46.5 Release Notes

This release introduces a number of bug fixes and improvements.  
Notably, an upgrade to Tendermint [v0.34.23](https://github.com/tendermint/tendermint/releases/tag/v0.34.23).

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.4...v0.46.5

**NOTE**: The changes mentioned in `v0.46.3` are **still** required:

> Chains must add the following to their go.mod for the application:
>
> ```go
> replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
> ```
