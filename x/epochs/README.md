---
sidebar_position: 1
---

# `x/epochs`

## Abstract

Often in the SDK, we would like to run certain code every so often. The
purpose of `epochs` module is to allow other modules to set that they
would like to be signaled once every period. So another module can
specify it wants to execute code once a week, starting at UTC-time = x.
`epochs` creates a generalized epoch interface to other modules so that
they can easily be signaled upon such events.

## Contents

1. **[Concept](#concepts)**
2. **[State](#state)**
3. **[Events](#events)**
4. **[Keeper](#keepers)**
5. **[Hooks](#hooks)**
6. **[Queries](#queries)**

## Concepts

The epochs module defines on-chain timers that execute at fixed time intervals.
Other SDK modules can then register logic to be executed at the timer ticks.
We refer to the period in between two timer ticks as an "epoch".

Every timer has a unique identifier.
Every epoch will have a start time, and an end time, where `end time = start time + timer interval`.
On mainnet, we only utilize one identifier, with a time interval of `one day`.

The timer will tick at the first block whose block time is greater than the timer end time,
and set the start as the prior timer end time. (Notably, it's not set to the block time!)
This means that if the chain has been down for a while, you will get one timer tick per block,
until the timer has caught up.

## State

The Epochs module keeps a single `EpochInfo` per identifier.
This contains the current state of the timer with the corresponding identifier.
Its fields are modified at every timer tick.
EpochInfos are initialized as part of genesis initialization or upgrade logic,
and are only modified on begin blockers.

## Events

The `epochs` module emits the following events:

### BeginBlocker

| Type        | Attribute Key | Attribute Value |
| ----------- | ------------- | --------------- |
| epoch_start | epoch_number  | {epoch_number}  |
| epoch_start | start_time    | {start_time}    |

### EndBlocker

| Type      | Attribute Key | Attribute Value |
| --------- | ------------- | --------------- |
| epoch_end | epoch_number  | {epoch_number}  |

## Keepers

### Keeper Functions

The epochs keeper provides the following functions to manage epochs:

```go
// GetEpochInfo returns epoch info by identifier.
func (k *Keeper) GetEpochInfo(ctx sdk.Context, identifier string) (types.EpochInfo, error)

// AddEpochInfo adds a new epoch info. Will return an error if the epoch fails validation,
// or re-uses an existing identifier. This method also sets the start time if left unset,
// and sets the epoch start height.
func (k *Keeper) AddEpochInfo(ctx sdk.Context, epoch types.EpochInfo) error

// AllEpochInfos iterate through epochs to return all epochs info.
func (k *Keeper) AllEpochInfos(ctx sdk.Context) ([]types.EpochInfo, error)

// NumBlocksSinceEpochStart returns the number of blocks since the epoch started.
// If the epoch started on block N, then calling this during block N (after BeforeEpochStart)
// would return 0. Calling it any point in block N+1 (assuming the epoch doesn't increment)
// would return 1.
func (k *Keeper) NumBlocksSinceEpochStart(ctx sdk.Context, identifier string) (int64, error)
```

## Hooks

```go
  // the first block whose timestamp is after the duration is counted as the end of the epoch
  AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error
  // new epoch is next block of epoch end block
  BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error
```

### How modules receive hooks

On hook receiver function of other modules, they need to filter
`epochIdentifier` and only do executions for only specific
epochIdentifier. Filtering epochIdentifier could be in `Params` of other
modules so that they can be modified by governance.

This is the standard dev UX of this:

```golang
func (k MyModuleKeeper) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
    params := k.GetParams(ctx)
    if epochIdentifier == params.DistrEpochIdentifier {
    // my logic
  }
  return nil
}
```

### Wiring Hooks

**Manual Wiring:**

Import the following:

```go
import (
    // ...
    "github.com/cosmos/cosmos-sdk/x/epochs"
    epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
    epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
)
```

Add the epochs keeper to your application struct:

```go
EpochsKeeper *epochskeeper.Keeper
```

Add the store key:

```go
keys := storetypes.NewKVStoreKeys(
    // ...
    epochstypes.StoreKey,
)
```

Instantiate the keeper:

```go
epochsKeeper := epochskeeper.NewKeeper(
    runtime.NewKVStoreService(keys[epochstypes.StoreKey]),
    appCodec,
)

app.EpochsKeeper = &epochsKeeper
```

Set up hooks for the epochs keeper:

```go
app.EpochsKeeper.SetHooks(
    epochstypes.NewMultiEpochHooks(
        // insert epoch hooks receivers here
        app.SomeOtherModule
    ),
)
```

Add the epochs module to the module manager:

```go
app.ModuleManager = module.NewManager(
    // ...
    epochs.NewAppModule(appCodec, app.EpochsKeeper),
)
```

Add entries for SetOrderBeginBlockers and SetOrderInitGenesis:

```go
app.ModuleManager.SetOrderBeginBlockers(
    // ...
    epochstypes.ModuleName,
)
```

```go
app.ModuleManager.SetOrderInitGenesis(
    // ...
    epochstypes.ModuleName,
)
```

**DI Wiring:**

First, set up the keeper for the application.

Import the epochs keeper:

```go
epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
```

Add the keeper to your application struct:

```go
EpochsKeeper *epochskeeper.Keeper
```

Add the keeper to the depinject system:

```go
depinject.Inject(
    appConfig,
    &appBuilder,
    &app.appCodec,
    &app.legacyAmino,
    &app.txConfig,
    &app.interfaceRegistry,
    // ... other modules
    &app.EpochsKeeper, // NEW MODULE!
)
```

Next, set up configuration for the module.

Import the following:

```go
import (
    epochsmodulev1 "cosmossdk.io/api/cosmos/epochs/module/v1"
    
    _ "github.com/cosmos/cosmos-sdk/x/epochs" // import for side-effects
    epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
)
```

Add an entry for BeginBlockers and InitGenesis:

```go
BeginBlockers: []string{
    // ...
    epochstypes.ModuleName,
},
```

```go
InitGenesis: []string{
    // ...
    epochstypes.ModuleName,
},
```

Add an entry for epochs in the ModuleConfig:

```go
{
    Name:   epochstypes.ModuleName,
    Config: appconfig.WrapAny(&epochsmodulev1.Module{}),
},
```

depinject can automatically add your hooks to the epochs `Keeper`. For it do so, specify an output of your module with the type `epochtypes.EpochHooksWrapper`, ie:

```go
type TestInputs struct {
	depinject.In
}

type TestOutputs struct {
	depinject.Out

	Hooks types.EpochHooksWrapper
}

func DummyProvider(in TestInputs) TestOutputs {
	return TestOutputs{
		Hooks: types.EpochHooksWrapper{
			EpochHooks: testEpochHooks{},
		},
	}
}
```

for an example see [`depinject_test.go`](https://github.com/cosmos/cosmos-sdk/tree/main/x/epochs/depinject_test.go)

### Panic isolation

If a given epoch hook panics, its state update is reverted, but we keep
proceeding through the remaining hooks. This allows more advanced epoch
logic to be used, without concern over state machine halting, or halting
subsequent modules.

This does mean that if there is behavior you expect from a prior epoch
hook, and that epoch hook reverted, your hook may also have an issue. So
do keep in mind "what if a prior hook didn't get executed" in the safety
checks you consider for a new epoch hook.

## Queries

The Epochs module provides the following queries to check the module's state.

```protobuf
service Query {
  // EpochInfos provide running epochInfos
  rpc EpochInfos(QueryEpochsInfoRequest) returns (QueryEpochsInfoResponse) {}
  // CurrentEpoch provide current epoch of specified identifier
  rpc CurrentEpoch(QueryCurrentEpochRequest) returns (QueryCurrentEpochResponse) {}
}
```

### Epoch Infos

Query the currently running epochInfos

```sh
<appd> query epochs epoch-infos
```

:::details Example

An example output:

```sh
epochs:
- current_epoch: "183"
  current_epoch_start_height: "2438409"
  current_epoch_start_time: "2021-12-18T17:16:09.898160996Z"
  duration: 86400s
  epoch_counting_started: true
  identifier: day
  start_time: "2021-06-18T17:00:00Z"
- current_epoch: "26"
  current_epoch_start_height: "2424854"
  current_epoch_start_time: "2021-12-17T17:02:07.229632445Z"
  duration: 604800s
  epoch_counting_started: true
  identifier: week
  start_time: "2021-06-18T17:00:00Z"
```

:::

### Current Epoch

Query the current epoch by the specified identifier

```sh
<appd> query epochs current-epoch [identifier]
```

:::details Example

Query the current `day` epoch:

```sh
<appd> query epochs current-epoch day
```

Which in this example outputs:

```sh
current_epoch: "183"
```

:::
