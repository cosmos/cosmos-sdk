# Upgrade Reference

This document provides a quick reference for the upgrades from `v0.53.x` to `v0.54.x` of Cosmos SDK.

Note, always read the **App Wiring Changes** section for more information on application wiring updates.

### TLDR

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.54.x/CHANGELOG.md).

## Authority Parameter Removal

Authority management has been centralized to the `x/consensus` module. Individual module keepers no longer accept an `authority` parameter in their constructors.

### Affected Modules

The following modules have removed the `authority` parameter from their keeper constructors:

* `x/auth`
* `x/bank`
* `x/distribution`
* `x/gov`
* `x/mint`
* `x/protocolpool`
* `x/slashing`
* `x/staking`
* `x/upgrade`

### Migration

**Before:**

```go
app.BankKeeper = bankkeeper.NewBaseKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[banktypes.StoreKey]),
    app.AccountKeeper,
    blockedAddresses,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(), // authority parameter - REMOVE
    logger,
)
```

**After:**

```go
app.BankKeeper = bankkeeper.NewBaseKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[banktypes.StoreKey]),
    app.AccountKeeper,
    blockedAddresses,
    logger, // authority parameter removed
)
```

Apply this pattern to all affected modules listed above.

### How It Works

Modules now retrieve the authority from `ConsensusParams` via the context at runtime. The authority value is stored in consensus params and can be updated via governance proposals.

**Example - Authority Validation:**

```go
func (k msgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)

    // Retrieve and validate authority from consensus params
    if sdkCtx.ConsensusParams().Authority.Authority != req.Authority {
        return nil, errors.Wrapf(sdkerrors.ErrUnauthorized,
            "invalid authority: expected %s, got %s",
            req.Authority, sdkCtx.ConsensusParams().Authority.Authority)
    }

    // ... rest of the handler
}
```