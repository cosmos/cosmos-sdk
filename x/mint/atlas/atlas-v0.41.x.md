
# x/mint

Mint handles the minting of new tokens. This can be associated with a inflation rate.

## Usage

1. Import the module.

  ```go
    import (
      "github.com/cosmos/cosmos-sdk/x/mint"
	    mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	    minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
    )
  ```

2. Add AppModuleBasic to your ModuleBasics.

  ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        mint.AppModuleBasic{},
      }
    )
  ```

3. Give the mint module account permissions.


  ```go
      // module account permissions
      var maccPerms = map[string][]string{
        minttypes.ModuleName:           {authtypes.Minter},
      }
  ```

4. Add the mint keeper to your apps struct.

  ```go
    type app struct {
      // ...
      MintKeeper       mintkeeper.Keeper
      // ...
    }
  ```
5. Add the mint store key to the group of store keys.
 
  ```go
   func NewApp(...) *App {
     // ...
      keys := sdk.NewKVStoreKeys(
       minttypes.StoreKey
      )
     // ...
   }
  ```

6. Create the keeper. 

  ```go
   func NewApp(...) *App {
      // ...
    app.MintKeeper = mintkeeper.NewKeeper(
		  appCodec, keys[minttypes.StoreKey], app.GetSubspace(minttypes.ModuleName), &stakingKeeper,
		  app.AccountKeeper, app.BankKeeper, authtypes.FeeCollectorName,
	  )
   }
  ```

8. Add the mint module to the app's ModuleManager.

  ```go
   func NewApp(...) *App {
     // ...
   app.mm = module.NewManager(
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper),
	)
   }
  ```
9. Set the mint module begin blocker order.

  ```go
    func NewApp(...) *App {
     // ...
      app.mm.SetOrderBeginBlockers(
      // ...
      minttypes.ModuleName,
      //...
      )
    }
  ```


10.  Set the mint module genesis order.

  ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(minttypes.ModuleName, ...)
   }
  ``` 


11. Add the mint module to the simulation manager (if you have one set).

  ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper),
       // ...
     )
   }
  ```

## Genesis

```go
type GenesisState struct {
	// minter is a space for holding current inflation information.
	Minter Minter `protobuf:"bytes,1,opt,name=minter,proto3" json:"minter"`
	// params defines all the paramaters of the module.
	Params Params `protobuf:"bytes,2,opt,name=params,proto3" json:"params"`
}
```

## Messages

The mint module supports a command line interface, REST and gRPC APIs. There are no transaction APIs only query APIs.

### CLI

Mint up

#### Queries

```sh
app q mint
Querying commands for the minting module

Usage:
  app query mint [flags]
  app query mint [command]

Available Commands:
  annual-provisions Query the current minting annual provisions value
  inflation         Query the current minting inflation value
  params            Query the current minting parameters

Flags:
  -h, --help   help for mint

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app query mint [command] --help" for more information about a command.
```


### REST

The rest api endpoints can be found here https://cosmos.network/rpc/master under the mint section.

### gRPC

mint supports both queries and transactions for gRPC. 

#### Query

[gRPC query](https://docs.cosmos.network/master/core/proto-docs.html#cosmos-mint-v1beta1-query-proto)
