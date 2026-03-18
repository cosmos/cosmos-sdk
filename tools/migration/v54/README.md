# v50+ → v54 Migration Tool

Automated migration tool for upgrading Cosmos SDK applications from v0.50.x through v0.53.x to v0.54.

## Usage

```bash
cd tools/migration/v54
go run . /path/to/your/app
goimports -w /path/to/your/app
cd /path/to/your/app && go mod tidy
```

## What It Does

All changes are fully automated — no manual intervention required.

### Version Handling
- `v0.50.x` to `v0.52.x`: the tool first performs a minimal bridge by updating the main `github.com/cosmos/cosmos-sdk` requirement to `v0.53.6`
- `v0.53.x`: the tool applies the full v53 -> v54 rewrite set directly
- `v0.54+`: rejected as already migrated

### go.mod Changes
- Bumps `github.com/cosmos/cosmos-sdk` to v0.54 pseudo version
- Bumps companion modules (api, client/v2, core, depinject, store, math)
- Bumps CometBFT to v0.39.0-beta.3
- Replaces `cosmossdk.io/log` with `cosmossdk.io/log/v2`
- Removes `cosmossdk.io/x/*` vanity URL modules (folded into SDK monorepo)

### Import Rewrites
- `cosmossdk.io/log` → `cosmossdk.io/log/v2`
- `cosmossdk.io/x/{feegrant,evidence,upgrade,tx}` → `github.com/cosmos/cosmos-sdk/x/*`
- `cosmossdk.io/x/{circuit,nft}` → `github.com/cosmos/cosmos-sdk/contrib/x/*`
- **Warning** emitted for `x/group` (moved to enterprise with commercial license)

### Module Removals (circuit, nft, group)
- Removes struct fields from SimApp
- Removes keeper initialization statements
- Removes AppModule registrations from `module.NewManager`
- Removes store keys from `NewKVStoreKeys`
- Removes entries from genesis ordering, begin/end blocker ordering
- Removes `nft.ModuleName` from `maccPerms`
- Removes `app.BaseApp.SetCircuitBreaker` call

### Function Signature Changes
- `govkeeper.NewKeeper`: removes StakingKeeper arg, adds `NewDefaultCalculateVoteResultsAndVotingPower` wrapper
- `nodeservice.RegisterNodeService`: adds 4th argument (earliest version closure)
- `SetOrderEndBlockers`: adds `banktypes.ModuleName` at front

### Struct Changes
- `EpochsKeeper` field changed from value to pointer type
- EpochsKeeper init rewritten to use local variable + pointer assignment

### Ante Handler
- Deletes custom `ante.go` wrapper (circuit ante decorator removed)
- Rewrites `setAnteHandler` to use `ante.NewAnteHandler(ante.HandlerOptions{...})` directly

### Other Changes
- `app.BaseApp.GRPCQueryRouter()` → `app.GRPCQueryRouter()`
- `app.BaseApp.Simulate` → `app.Simulate`
- `SetModuleVersionMap` error now handled
- Duplicate `auth/tx` import consolidated to `authtx` alias

## Post-Migration

Run `goimports -w . && go mod tidy` to clean up unused imports and resolve dependencies.
