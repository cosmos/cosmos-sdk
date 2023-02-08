# Cosmos SDK v0.45.11 Release Notes

This release introduces a number of bug fixes and improvements.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.45.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.45.10...v0.45.11

**NOTE**: The changes mentioned in `v0.45.9` are **still** required:

```go
# Chains must add the following to their go.mod for the application:
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```
