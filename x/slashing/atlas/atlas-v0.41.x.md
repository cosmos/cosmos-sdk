
# x/slashing

Slashing is responsible for punishing malicious actors.

## Usage

1. Import the module.

  ```go
    import (
      "github.com/cosmos/cosmos-sdk/x/slashing"
      slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
      slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
    )
  ```

2. Add AppModuleBasic to your ModuleBasics.

  ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        slashing.AppModuleBasic{},
      }
    )
  ```

3. Add the slashing keeper to your apps struct.

  ```go
    type app struct {
      // ...
      SlashingKeeper   slashingkeeper.Keeper
      // ...
    }
  ```
4. Add the slashing store key to the group of store keys.
 
  ```go
   func NewApp(...) *App {
     // ...
      keys := sdk.NewKVStoreKeys(
       slashingtypes.StoreKey,
      )
     // ...
   }
  ```

5. Create the keeper.

  ```go
   func NewApp(...) *App {
      // ...
    	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, keys[slashingtypes.StoreKey], &stakingKeeper, app.GetSubspace(slashingtypes.ModuleName),
	)
   }
  ```

7. Add the slashing module to the app's ModuleManager.

  ```go
   func NewApp(...) *App {
     // ...
   app.mm = module.NewManager(
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
	)
   }
  ```
9. Set the slashing module begin blocker order.

  ```go
    func NewApp(...) *App {
     // ...
      app.mm.SetOrderBeginBlockers(
      // ...
      slashingtypes.ModuleName,
      //...
      )
    }
  ```


10.  Set the slashing module genesis order.

  ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(slashingtypes.ModuleName, ...)
   }
  ``` 


11. Add the gov module to the simulation manager (if you have one set).

  ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
       // ...
     )
   }
  ```

## Genesis

```go
// GenesisState defines the slashing module's genesis state.
type GenesisState struct {
	// params defines all the paramaters of related to deposit.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// signing_infos represents a map between validator addresses and their
	// signing infos.
	SigningInfos []SigningInfo `protobuf:"bytes,2,rep,name=signing_infos,json=signingInfos,proto3" json:"signing_infos" yaml:"signing_infos"`
	// signing_infos represents a map between validator addresses and their
	// missed blocks.
	MissedBlocks []ValidatorMissedBlocks `protobuf:"bytes,3,rep,name=missed_blocks,json=missedBlocks,proto3" json:"missed_blocks" yaml:"missed_blocks"`
}
```

## Messages

Slashing has both queries and transactions. They can be used in the command line interface, REST and gRPC APIs.

### CLI

#### Queries

```sh
app q slashing
Querying commands for the slashing module

Usage:
  app query slashing [flags]
  app query slashing [command]

Available Commands:
  params        Query the current slashing parameters
  signing-info  Query a validator's signing information
  signing-infos Query signing information of all validators

Flags:
  -h, --help   help for slashing

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app query slashing [command] --help" for more information about a command.
```

#### Transactions

```sh
app tx slashing                                                                                   
Slashing transaction subcommands

Usage:
  app tx slashing [flags]
  app tx slashing [command]

Available Commands:
  unjail      unjail validator previously jailed for downtime

Flags:
  -h, --help   help for slashing

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app tx slashing [command] --help" for more information about a command.
```


### REST

The REST api for slashing can be found here https://cosmos.network/rpc/master under the slashing section. 


### gRPC

gRPC supports both queries and transactions for the slashing module. 

#### Query

[gRPC query](https://docs.cosmos.network/master/core/proto-docs.html#cosmos/slashing/v1beta1/query.proto)

#### Tx

[gRPC Tx](https://docs.cosmos.network/master/core/proto-docs.html#cosmos-slashing-v1beta1-tx-proto)

View supported messages at [docs.cosmos.network/v0.41/modules/slashing](https://docs.cosmos.network/v0.41/modules/slashing/03_messages.html)
