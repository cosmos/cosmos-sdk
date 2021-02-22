
# x/upgrade

Upgrade handles live upgrades of your application.

## Usage

1. Import the module.

  ```go
    import (
      upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
      upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
      upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
    )
  ```

2. Add AppModuleBasic to your ModuleBasics.

  ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        upgrade.AppModuleBasic{},
      }
    )
  ```

3. Add the upgrade keeper to your apps struct.

  ```go
    type app struct {
      // ...
      UpgradeKeeper    upgradekeeper.Keeper
      // ...
    }
  ```
5. Add the upgrading store key to the group of store keys.
 
  ```go
   func NewApp(...) *App {
     // ...
      keys := sdk.NewKVStoreKeys(
       upgradetypes.StoreKey,
      )
     // ...
   }
  ```

6. Create the keeper.

  ```go
   func NewApp(...) *App {
      // ...
      app.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath)
   }
  ```

7. Add the upgrade module to the app's ModuleManager.

  ```go
  func NewApp(...) *App {
     // ...
    app.mm = module.NewManager(
      upgrade.NewAppModule(app.UpgradeKeeper),
    )
  }
  ```

8. Set the upgrade module begin blocker order.

  ```go
    func NewApp(...) *App {
     // ...
      app.mm.SetOrderBeginBlockers(
      // ...
      upgradetypes.ModuleName,
      //...
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
