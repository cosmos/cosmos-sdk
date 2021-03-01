
# x/params

Params handles the modification of module parameters. If used in conjunction with with Gov, Gov will be change the parameters.

## Usage

1. Import the module.

  ```go
    import (
    "github.com/cosmos/cosmos-sdk/x/params"
    paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
    paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
    paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
    paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
    )
  ```

2. Add AppModuleBasic to your ModuleBasics.

  ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        params.AppModuleBasic{},
      }
    )
  ```

3. Add the params keeper to your apps struct.

  ```go
    type app struct {
      // ...
      ParamsKeeper     paramskeeper.Keeper
      // ...
    }
  ```
4. Add the params store key to the group of store keys.
 
  ```go
   func NewApp(...) *App {
     // ...
      keys := sdk.NewKVStoreKeys(
       paramstypes.StoreKey,
      )
     // ...
   }
  ```

5. Create the keeper. 

  ```go
  func NewApp(...) *App {
      // ...
    bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()))
  }
  ```

6. Create and set a params store, this is used for consensus related parameters. 

  ```go
   func NewApp(...) *App {
      // ...
    app.ParamsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])
   }
  ```

7. Add the params module to the app's ModuleManager.

  ```go
   func NewApp(...) *App {
     // ...
   app.mm = module.NewManager(
		params.NewAppModule(app.ParamsKeeper),
	)
   }
  ```

8.  Add the param module to the simulation manager (if you have one set).

  ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       params.NewAppModule(app.ParamsKeeper),
       // ...
     )
   }
  ```

## Messages

Params supports changing parameters of the system through either a governance proposal, if used in conjunction with the gov module. 

### CLI


#### Queries

```sh
 app q params
Querying commands for the params module

Usage:
  app query params [flags]
  app query params [command]

Available Commands:
  subspace    Query for raw parameters by subspace and key

Flags:
  -h, --help   help for params

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app query params [command] --help" for more information about a command.
```

#### Transactions

Transactions are submitted through another module. In the case of the cosmos hub this is gov. 

### REST

Queries are supported via rest. They can be found here https://cosmos.network/rpc/master. Transactions are submitting through governance. 

### gRPC

Params support both queries and transactions for gRPC. The transactions need to be administered via a different module. If used in conjunction with gov they would be modified via governance proposals. 

#### Query

[gRPC query](https://docs.cosmos.network/master/core/proto-docs.html#cosmos-params-v1beta1-query-proto)
