# x/distribution

The `x/distribution` module distributes fees and staking rewards to users.

## Usage

1. Import the module.

   ```go
   import (
    distr "github.com/cosmos/cosmos-sdk/x/distribution"
    distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
    distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
    distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
   )
   ```

2. Add `AppModuleBasic` to your `ModuleBasics`.

    ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        distr.AppModuleBasic{},
      }
    )
    ```

3. Give distribution module account permissions.

    ```go
  	// module account permissions
    var maccPerms = map[string][]string{
      distrtypes.ModuleName:          nil,
    }
    ```

4. Allow the distribution module to receive funds.

    ```go
      allowedReceivingModAcc = map[string]bool{
        distrtypes.ModuleName: true,
      }
    ```

5. Add the distribution keeper to your apps struct.

    ```go
      type app struct {
        // ...
        DistrKeeper      distrkeeper.Keeper
        // ...
      }
    ```

6. Add the distribution store key to the group of store keys.

   ```go
   func NewApp(...) *App {
     // ...
      keys := sdk.NewKVStoreKeys(
       distrtypes.StoreKey,
      )
     // ...
   }
   ```

7. Create the keeper. Note the distribution keeper requires an account, bank and staking keeper. This is in order to distribute rewards. 

   ```go
   func NewApp(...) *App {
      // ...
      // create capability keeper with router
      app.DistrKeeper = distrkeeper.NewKeeper(
		    appCodec, keys[distrtypes.StoreKey], app.GetSubspace(distrtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		    &stakingKeeper, authtypes.FeeCollectorName, app.ModuleAccountAddrs(),
	)
   }
   ```

8. Add the `x/distribution` module to the app's `ModuleManager`.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm = module.NewManager(
       // ...
       distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
       // ...
     )
   }
   ```

9. Set the `x/distribution` module begin blocker order.

    ```go
    func NewApp(...) *App {
     // ...
      app.mm.SetOrderBeginBlockers(
        // ...
        distrtypes.ModuleName,
        // ...
      )
    }
    ```

10.  Set the `x/distribution` module genesis order.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(distrtypes.ModuleName,, ...)
   }
   ```

11. Add the `x/distribution` module to the simulation manager (if you have one set).

   ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
       // ...
     )
   }

## Genesis

The `x/distribution` module defines its genesis state as follows:

```proto
// DelegatorWithdrawInfo is the address for where distributions rewards are
// withdrawn to by default this struct is only used at genesis to feed in
// default withdraw addresses.
type DelegatorWithdrawInfo struct {
	// delegator_address is the address of the delegator.
	DelegatorAddress string `protobuf:"bytes,1,opt,name=delegator_address,json=delegatorAddress,proto3" json:"delegator_address,omitempty" yaml:"delegator_address"`
	// withdraw_address is the address to withdraw the delegation rewards to.
	WithdrawAddress string `protobuf:"bytes,2,opt,name=withdraw_address,json=withdrawAddress,proto3" json:"withdraw_address,omitempty" yaml:"withdraw_address"`
}
```

## Messages

The distribution module provides interaction via command line, REST and gRPC.

### CLI

#### Queries

#### Transactions


### REST

<!-- todo: -->

### gRPC

Distribution supports both querying and submitting transactions via gRPC

#### Query

[gRPC query](https://docs.cosmos.network/master/core/proto-docs.html#cosmos/distribution/v1beta1/query.proto)

#### Tx

[gRPC Tx](https://docs.cosmos.network/master/core/proto-docs.html#cosmos-distribution-v1beta1-tx-proto)

View supported messages at [docs.cosmos.network/v0.40/modules/distribution](https://docs.cosmos.network/master/modules/distribution/04_messages.html)
