# x/auth

The `x/auth` module is responsible for specifying the base transaction and
account types for an application, as well as AnteHandler and authentication logic.

## Usage

1. Import the module.

   ```go
   import (
       "github.com/cosmos/cosmos-sdk/x/auth"
   )
   ```

2. Add `AppModuleBasic` to your `ModuleBasics`.

    ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        auth.AppModuleBasic{},
      }
    )
    ```

3. Create the module's parameter subspace in your application constructor.

   ```go
   func NewApp(...) *App {
     // ...
     app.subspaces[auth.ModuleName] = app.ParamsKeeper.Subspace(auth.DefaultParamspace)
   }
   ```

4. Create the keeper.

   ```go
   func NewApp(...) *App {
      // ...
      app.AccountKeeper = auth.NewAccountKeeper(
       app.cdc, keys[auth.StoreKey], app.subspaces[auth.ModuleName], auth.ProtoBaseAccount,
      )
   }
   ```

5. Add the `x/auth` module to the app's `ModuleManager`.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm = module.NewManager(
       // ...
       auth.NewAppModule(app.AccountKeeper),
       // ...
     )
   }
   ```

6. Set the `x/auth` module genesis order.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(..., auth.ModuleName, ...)
   }
   ```

7. Add the `x/auth` module to the simulation manager (if you have one set).

   ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       auth.NewAppModule(app.AccountKeeper),
       // ...
     )
   }

8. Set the `AnteHandler` if you're using the default provided by `x/auth`. Note,
the default `AnteHandler` provided by the `x/auth` module depends on the `x/supply`
module.

   ```go
   func NewApp(...) *App {
     app.SetAnteHandler(ante.NewAnteHandler(
       app.AccountKeeper,
       app.SupplyKeeper, 
       auth.DefaultSigVerificationGasConsumer,
     ))
   }
   ```

## Client

### CLI

### REST
