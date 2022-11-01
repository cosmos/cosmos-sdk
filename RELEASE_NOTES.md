# Cosmos SDK v0.46.4 Release Notes

This release introduces a number of bug fixes, features and improvements.  
Notably, a new query for accessing module accounts info by module name (thanks @gsk967).

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.3...v0.46.4

**NOTE**: The changes mentioned in `v0.46.3` are **still** required:

> Chains must add the following to their go.mod for the application:
>
> ```go
> replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
> ```
