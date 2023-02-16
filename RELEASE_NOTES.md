# Cosmos SDK v0.45.14 Release Notes

This release fixes a possible way to DoS a node.

**NOTE**: Add or update the following replace in the `go.mod` of your application:

```go
// use informal system fork of tendermint
replace github.com/tendermint/tendermint => github.com/informalsystems/tendermint v0.34.26
```

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.45.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.45.13...v0.45.14
