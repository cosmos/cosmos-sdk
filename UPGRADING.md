# Upgrade Reference

This document provides a quick reference for the upgrades from `v0.53.x` to `v0.54.x` of Cosmos SDK.

Note, always read the **App Wiring Changes** section for more information on application wiring updates.

### TLDR

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.54.x/CHANGELOG.md).

## x/gov

### Keeper Initialization

The `x/gov` module has been decoupled from `x/staking`. The `keeper.NewKeeper` constructor now requires a `CalculateVoteResultsAndVotingPowerFn` parameter instead of a `StakingKeeper`.

**Before:**
```go
govKeeper := keeper.NewKeeper(
    cdc,
    storeService,
    authKeeper,
    bankKeeper,
    stakingKeeper,  // StakingKeeper parameter
    distrKeeper,
    router,
    config,
    authority,
)
```

**After:**
```go
govKeeper := keeper.NewKeeper(
    cdc,
    storeService,
    authKeeper,
    bankKeeper,
    keeper.NewDefaultCalculateVoteResultsAndVotingPower(stakingKeeper),  // Function parameter
    distrKeeper,
    router,
    config,
    authority,
)
```

For applications using depinject, the governance module now accepts an optional `CalculateVoteResultsAndVotingPowerFn`. If not provided, it will use the `StakingKeeper` (also optional) to create the default function.