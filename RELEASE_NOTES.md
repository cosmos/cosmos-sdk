# Cosmos SDK v0.46.11 Release Notes

This release includes the migration to [CometBFT v0.34.27](https://github.com/cometbft/cometbft/blob/v0.34.27/CHANGELOG.md#v03427).
This migration should be not be breaking for chains.
From `v0.46.11`+, the following replace is *mandatory* in the `go.mod` of your application:

```go
// use cometbft
replace github.com/tendermint/tendermint => github.com/cometbft/cometbft v0.34.27
```

Additionally, the SDK sets its minimum version to Go 1.19. This is not because the SDK uses new Go 1.19 functionalities, but to signal that we recommend chains to upgrade to Go 1.19 — Go 1.18 is not supported by the Go Team anymore.
Note, that SDK recommends chains to use the same Go version across all of their network.
We recommend, as well, chains to perform a **coordinated upgrade** when migrating from Go 1.18 to Go 1.19.

Please see the [CHANGELOG](https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/CHANGELOG.md) for an exhaustive list of changes.

**Full Commit History**: https://github.com/cosmos/cosmos-sdk/compare/v0.46.10...v0.46.11
