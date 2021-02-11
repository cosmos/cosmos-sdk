# x/evidence

The `x/evidence` module is responsible for handling multi-asset coin transfers between
accounts and tracking special-case pseudo-transfers which must work differently
with particular kinds of accounts.

## Usage

1. Import the module.

   ```go
   import (
       "github.com/cosmos/cosmos-sdk/x/evidence"
   )
   ```

2. Add `AppModuleBasic` to your `ModuleBasics`.

    ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        evidence.AppModuleBasic{},
      }
    )
    ```

3. Create the module's parameter subspace in your application constructor.

   ```go
   func NewApp(...) *App {
     // ...
     app.subspaces[evidence.ModuleName] = app.ParamsKeeper.Subspace(evidence.DefaultParamspace)
   }
   ```

4. Create the keeper. Note, the `x/evidence` module depends on the `x/auth` module
   and a list of blacklisted account addresses which funds are not allowed to be
   sent to. Your application will need to define this method based your needs.

   ```go
   func NewApp(...) *App {
     // ...
     app.BankKeeper = bank.NewBaseKeeper(
       app.AccountKeeper, app.subspaces[evidence.ModuleName], app.BlacklistedAccAddrs(),
     )
   }
   ```

5. Add the `x/evidence` module to the app's `ModuleManager`.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm = module.NewManager(
       // ...
       evidence.NewAppModule(app.EvidenceKeeper, app.AccountKeeper),
       // ...
     )
   }
   ```

6. Set the `x/evidence` module genesis order.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(..., evidence.ModuleName, ...)
   }
   ```

7. Add the `x/evidence` module to the simulation manager (if you have one set).

   ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       evidence.NewAppModule(app.EvidenceKeeper, app.AccountKeeper),
       // ...
     )
   }

## Genesis

The `x/evidence` module defines its genesis state as follows:

```proto
type GenesisState struct {
 // evidence defines all the evidence at genesis.
 Evidence []*types.Any `protobuf:"bytes,1,rep,name=evidence,proto3" json:"evidence,omitempty"`
}
```

## Messages

## Client
