# x/feegrant

The `x/feegrant` module distributes fees and staking rewards to users.

## Usage

1. Import the module.

   ```go
   import (
    feegrant "github.com/cosmos/cosmos-sdk/x/feegrant"
    feegrantante "github.com/cosmos/cosmos-sdk/x/feegrant/ante"
    feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
    feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant/types"
   )
   ```

2. Add `AppModuleBasic` to your `ModuleBasics`.

    ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        feegrant.AppModuleBasic{},
      }
    )
    ```

3. Add the feegrant keeper to your apps struct.

    ```go
      type app struct {
        // ...
        FeeGrantKeeper   feegrantkeeper.Keeper
        // ...
      }
    ```

4. Add the feegrant store key to the group of store keys.

   ```go
   func NewApp(...) *App {
     // ...
      keys := sdk.NewKVStoreKeys(
       feegranttypes.StoreKey,
      )
     // ...
   }
   ```

5. Create the keeper. Note the feegrant keeper requires an account keeper.

   ```go
   func NewApp(...) *App {
      // ...
      // create capability keeper with router
      app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, keys[feegranttypes.StoreKey], app.AccountKeeper)
	)
   }
   ```

8. Add the `x/feegrant` module to the app's `ModuleManager`.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm = module.NewManager(
       // ...
       feegrant.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
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

```sh
app q distribution
Querying commands for the distribution module

Usage:
  app query distribution [flags]
  app query distribution [command]

Available Commands:
  commission                    Query distribution validator commission
  community-pool                Query the amount of coins in the community pool
  params                        Query distribution params
  rewards                       Query all distribution delegator rewards or rewards from a particular validator
  slashes                       Query distribution validator slashes
  validator-outstanding-rewards Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations

Flags:
  -h, --help   help for distribution

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app query distribution [command] --help" for more information about a command.
```

#### Transactions

```sh
app tx distribution
Distribution transactions subcommands

Usage:
  app tx distribution [flags]
  app tx distribution [command]

Available Commands:
  fund-community-pool  Funds the community pool with the specified amount
  set-withdraw-addr    change the default withdraw address for rewards associated with an address
  withdraw-all-rewards withdraw all delegations rewards for a delegator
  withdraw-rewards     Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator

Flags:
  -h, --help   help for distribution

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app tx distribution [command] --help" for more information about a command.
```


### REST

Evidence REST API supports only queries of evidence. To submit evidence please use gRPC or the cli.

### gRPC

Evidence supports both querying and submitting transactions via gRPC

#### Query

[gRPC query](https://docs.cosmos.network/master/core/proto-docs.html#cosmos/distribution/v1beta1/query.proto)

#### Tx

[gRPC Tx](https://docs.cosmos.network/master/core/proto-docs.html#cosmos-distribution-v1beta1-tx-proto)

View supported messages at [docs.cosmos.network/v0.40/modules/distribution](https://docs.cosmos.network/master/modules/distribution/04_messages.html)
