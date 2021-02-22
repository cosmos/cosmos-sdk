
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
