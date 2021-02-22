
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

<!-- Todo: add a short description about client interactions -->

### CLI
<!-- Todo: add a short description about client interactions -->

#### Queries
<!-- Todo: add a short description about cli query interactions -->

#### Transactions
<!-- Todo: add a short description about cli transaction interactions -->


### REST
<!-- Todo: add a short description about REST interactions -->

#### Query
<!-- Todo: add a short description about REST query interactions -->

#### Tx
<!-- Todo: add a short description about REST transaction interactions -->

### gRPC
<!-- Todo: add a short description about gRPC interactions -->

#### Query
<!-- Todo: add a short description about gRPC query interactions -->

#### Tx
<!-- Todo: add a short description about gRPC transactions interactions -->
