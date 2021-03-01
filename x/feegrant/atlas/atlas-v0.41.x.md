# x/feegrant

The `x/feegrant` module allows accounts to grant fee allowances and to use fees from their accounts.

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

9. Set the `x/feegrant` module begin blocker order.

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

10.  Set the `x/feegrant` module genesis order.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(distrtypes.ModuleName,, ...)
   }
   ```

11. Add the `x/feegrant` module to the simulation manager (if you have one set).

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

The `x/feegrant` module defines its genesis state as follows:

```proto
// GenesisState contains a set of fee allowances, persisted from the store
type GenesisState struct {
	FeeAllowances []FeeAllowanceGrant `protobuf:"bytes,1,rep,name=fee_allowances,json=feeAllowances,proto3" json:"fee_allowances"`
}
```

## Messages

The feegrant module provides interaction via command line, REST and gRPC.

### CLI

#### Queries

```sh
app q feegrant          
Querying commands for the feegrant module

Usage:
  app query feegrant [flags]
  app query feegrant [command]

Available Commands:
  grant       Query details of a single grant
  grants      Query all grants of a grantee

Flags:
  -h, --help   help for feegrant

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app query feegrant [command] --help" for more information about a command.
```

#### Transactions

```sh
app tx feegrant
Grant and revoke fee allowance for a grantee by a granter

Usage:
  app tx feegrant [flags]
  app tx feegrant [command]

Available Commands:
  grant       Grant Fee allowance to an address
  revoke      revoke fee-grant

Flags:
  -h, --help   help for feegrant

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app tx feegrant [command] --help" for more information about a command.
```


### REST

Feegrant REST API supports queries and transactions. 

### gRPC

Feegrant supports both querying and submitting transactions via gRPC

#### Query

[gRPC query](https://docs.cosmos.network/master/core/proto-docs.html#cosmos/feegrant/v1beta1/query.proto)

#### Tx

[gRPC Tx](https://docs.cosmos.network/master/core/proto-docs.html#cosmos-feegrant-v1beta1-tx-proto)
