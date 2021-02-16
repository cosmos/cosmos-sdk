# x/capability

The `x/capability` module is responsible for handling permissions for cross modules requests.

## Usage

1. Import the module.

   ```go
   import (
    "github.com/cosmos/cosmos-sdk/x/capability"
    capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
    capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
   )
   ```

2. Add `AppModuleBasic` to your `ModuleBasics`.

    ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        capability.AppModuleBasic{},
      }
    )
    ```

3. Add the capability keeper to your apps struct.

    ```go
      type app struct {
        // ...
        CapabilityKeeper *capabilitykeeper.Keeper
        // ...
      }
    ```

4. Add the capability store key to the group of store keys.

   ```go
   func NewApp(...) *App {
     // ...
      keys := sdk.NewKVStoreKeys(
      capabilitytypes.StoreKey,
      )
     // ...
   }
   ```

6. Create New memory store keys. 

   ```go
   func NewApp(...) *App {
     // ...
    memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
     // ...
   }
   ```

7. Create the keeper. Note the capability keeper requires a memory key

   ```go
   func NewApp(...) *App {
      // ...
      // create capability keeper with router
      app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
   }
   ```

8. Add the `x/capability` module to the app's `ModuleManager`.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm = module.NewManager(
       // ...
       capability.NewAppModule(appCodec, *app.CapabilityKeeper),
       // ...
     )
   }
   ```

9.  Set the `x/capability` module genesis order.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(capabilitytypes.ModuleName, ...)
   }
   ```

10. Add the `x/capability` module to the simulation manager (if you have one set).

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

The `x/capability` module defines its genesis state as follows:

```proto
// GenesisOwners defines the capability owners with their corresponding index.
type GenesisOwners struct {
	// index is the index of the capability owner.
	Index uint64 `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	// index_owners are the owners at the given index.
	IndexOwners CapabilityOwners `protobuf:"bytes,2,opt,name=index_owners,json=indexOwners,proto3" json:"index_owners" yaml:"index_owners"`
}
```

## Messages

Capability module does not support client interactions.
