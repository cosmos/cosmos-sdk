# x/evidence

The `x/evidence` module is responsible for handling multi-asset coin transfers between
accounts and tracking special-case pseudo-transfers which must work differently
with particular kinds of accounts.

## Usage

1. Import the module.

   ```go
   import (
      "github.com/cosmos/cosmos-sdk/x/evidence"
      evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
      evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
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

3. Add the evidence keeper to your apps struct.

    ```go
      type app struct {
        // ...
        EvidenceKeeper   evidencekeeper.Keeper
        // ...
      }
    ```

4. Add the evidence store key to the group of store keys.

   ```go
   func NewApp(...) *App {
     // ...
      keys := sdk.NewKVStoreKeys(
      evidencetypes.StoreKey, 
      )
     // ...
   }
   ```

5. Create the keeper. Note, the `x/evidence` module depends on the `x/staking` and `x/slashing` modules.

   ```go
   func NewApp(...) *App {
      // ...
      // create evidence keeper with router
      evidenceKeeper := evidencekeeper.NewKeeper(
        appCodec, keys[evidencetypes.StoreKey], &app.StakingKeeper, app.SlashingKeeper,
      )
   }
   ```

6. Add the `x/evidence` module to the app's `ModuleManager`.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm = module.NewManager(
       // ...
       evidence.NewAppModule(app.EvidenceKeeper),
       // ...
     )
   }
   ```

7. Set the `x/evidence` module begin blocker order.

    ```go
    func NewApp(...) *App {
     // ...
      app.mm.SetOrderBeginBlockers(
        // ...
        evidencetypes.ModuleName, 
        // ...
      )
    }
    ```

8. Set the `x/evidence` module genesis order.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(..., evidencetypes.ModuleName, ...)
   }
   ```

9. Add the `x/evidence` module to the simulation manager (if you have one set).

   ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       evidence.NewAppModule(app.EvidenceKeeper),
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
<!-- todo: change to v0.41 when its available -->

View supported messages at [docs.cosmos.network/v0.40/modules/evidence](https://docs.cosmos.network/v0.40/modules/evidence/03_messages.html)

## Client

### CLI

The evidence module supports the blow command to query evidence.

```sh
Usage:
  app query evidence [flags]

Flags:
      --count-total       count total number of records in evidence to query for
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for evidence
      --limit uint        pagination limit of evidence to query for (default 100)
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint       pagination offset of evidence to query for
  -o, --output string     Output format (text|json) (default "text")
      --page uint         pagination page of evidence to query for. This sets offset to a multiple of limit (default 1)
      --page-key string   pagination page-key of evidence to query for
```

### REST

### gRPC

View supported messages at [docs.cosmos.network/v0.40/modules/evidence](https://docs.cosmos.network/v0.40/modules/evidence/03_messages.html)
