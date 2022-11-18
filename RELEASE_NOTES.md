# Cosmos SDK v0.46.6 Release Notes

This release introduces small bug fixes and improvements.

Please read the release notes of [v0.46.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.46.5) if you are upgrading from `<=0.46.4`.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

Full Commit History: https://github.com/cosmos/cosmos-sdk/compare/v0.46.5...v0.46.6

**NOTE**: The changes mentioned in `v0.46.3` are **still** required:

```go
# Chains must add the following to their go.mod for the application:
replace github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
```
