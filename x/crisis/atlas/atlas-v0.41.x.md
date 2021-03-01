# x/crisis

The `x/crisis` module is responsible for handling invariants that may be broken. An invariant can be total supply of the chain, if the number does not add up, this invariant will be broken. 

## Usage

1. Import the module.

   ```go
   import (
    "github.com/cosmos/cosmos-sdk/x/crisis"
    crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
    crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
   )
   ```

2. Add `AppModuleBasic` to your `ModuleBasics`.

    ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        crisis.AppModuleBasic{},
      }
    )
    ```

3. Add the crisis keeper to your apps struct.

    ```go
      type app struct {
        // ...
        CrisisKeeper     crisiskeeper.Keeper
        // ...
      }
    ```

4. Create the keeper. Note the capability keeper requires a memory key

   ```go
   func NewApp(...) *App {
      // ...
      // create capability keeper with router
    app.CrisisKeeper = crisiskeeper.NewKeeper(
      app.GetSubspace(crisistypes.ModuleName), invCheckPeriod, app.BankKeeper, authtypes.FeeCollectorName,
	)
   }
   ```

5. Add the `x/crisis` module to the app's `ModuleManager`.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm = module.NewManager(
       // ...
       crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
       // ...
     )
   }
   ```
7. Set the `x/crisis` module begin blocker order.

    ```go
    func NewApp(...) *App {
     // ...
      app.mm.SetOrderEndBlockers(crisistypes.ModuleName)
    }
    ```

6.  Set the `x/crisis` module genesis order.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(crisistypes.ModuleName,, ...)
   }
   ```

7. Register Invariants.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm.RegisterInvariants(&app.CrisisKeeper)
   }
   ```

12. Add the `x/crisis` module to the simulation manager (if you have one set).

   ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       capability.NewAppModule(appCodec, *app.CapabilityKeeper),
       // ...
     )
   }

## Genesis

The `x/crisis` module defines its genesis state as follows:

```proto
// GenesisState defines the crisis module's genesis state.
type GenesisState struct {
	// constant_fee is the fee used to verify the invariant in the crisis
	// module.
	ConstantFee types.Coin `protobuf:"bytes,3,opt,name=constant_fee,json=constantFee,proto3" json:"constant_fee" yaml:"constant_fee"`
}
```

## Messages

The crisis module provides client interactions only through the command line. 

### CLI

The crisis module allows the user to submit proofs of an broken invariant. 

```sh
app tx crisis  
Crisis transactions subcommands

Usage:
  simd tx crisis [flags]
  simd tx crisis [command]

Available Commands:
  invariant-broken Submit proof that an invariant broken to halt the chain

Flags:
  -h, --help   help for crisis

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "simd tx crisis [command] --help" for more information about a command.
```
