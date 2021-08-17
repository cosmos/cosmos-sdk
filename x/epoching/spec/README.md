<!--
order: 20
title: Epoching Overview
parent:
  title: "epoching"
-->

# `x/epoching`

## Abstract

The epoching module allows modules to queue messages for execution at a certain block height. Each module will have its own instance of the epoching module, this allows each module to have its own message queue and own duration for epochs.

## Example

In this example, we are creating an epochkeeper for a module that will be used by the module to queue messages to be executed at a later point in time.

```go
type Keeper struct {
  storeKey           sdk.StoreKey
  cdc                codec.BinaryMarshaler
  epochKeeper        epochkeeper.Keeper
}

// NewKeeper creates a new staking Keeper instance
func NewKeeper(cdc codec.BinaryMarshaler, key sdk.StoreKey) Keeper {
 return Keeper{
  storeKey:           key,
  cdc:                cdc,
  epochKeeper:        epochkeeper.NewKeeper(cdc, key),
 }
}
```

### Contents

1. **[State](01_state.md)**
