
# x/staking

Staking is proof of stake. It handles the logic around validators and delegators. The spec is located [here](https://docs.cosmos.network/v0.41/modules/staking/)

## Usage

1. Import the module.

  ```go
    import (
      "github.com/cosmos/cosmos-sdk/x/staking"
      stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
      stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
    )
  ```

2. Add AppModuleBasic to your ModuleBasics.

  ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        staking.AppModuleBasic{},
      }
    )
  ```

3. Give the staking module account permissions.


  ```go
      // module account permissions
      var maccPerms = map[string][]string{
    		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		    stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
      }
  ```

4. Add the staking keeper to your apps struct.

  ```go
    type app struct {
      // ...
      StakingKeeper    stakingkeeper.Keeper
      // ...
    }
  ```
5. Add the staking store key to the group of store keys.
 
  ```go
   func NewApp(...) *App {
     // ...
      keys := sdk.NewKVStoreKeys(
       stakingtypes.StoreKey
      )
     // ...
   }
  ```

6. Create the keeper.

  ```go
   func NewApp(...) *App {
      // ...
    stakingKeeper := stakingkeeper.NewKeeper(
      appCodec, keys[stakingtypes.StoreKey], app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName),
    )
   }
  ```

7. Set staking hooks. These are used to notify the distribution and slashing modules of any needed information. 

  ```go
   func NewApp(...) *App {
      // ...
    // register the staking hooks
    // NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
    app.StakingKeeper = *stakingKeeper.SetHooks(
      stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
    )
   }
  ```

9. Add the staking module to the app's ModuleManager.

  ```go
  func NewApp(...) *App {
     // ...
    app.mm = module.NewManager(
      staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
    )
  }
  ```

10. Set the staking module begin blocker order.

  ```go
    func NewApp(...) *App {
     // ...
      app.mm.SetOrderBeginBlockers(
      // ...
      stakingtypes.ModuleName,
      //...
      )
    }
  ```


10.  Set the slashing module genesis order.

  ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(stakingtypes.ModuleName, ...)
   }
  ``` 


11. Add the gov module to the simulation manager (if you have one set).

  ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
       // ...
     )
   }
  ```

## Genesis

```go
// GenesisState defines the staking module's genesis state.
type GenesisState struct {
	// params defines all the paramaters of related to deposit.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// last_total_power tracks the total amounts of bonded tokens recorded during
	// the previous end block.
	LastTotalPower github_com_cosmos_cosmos_sdk_types.Int `protobuf:"bytes,2,opt,name=last_total_power,json=lastTotalPower,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"last_total_power" yaml:"last_total_power"`
	// last_validator_powers is a special index that provides a historical list
	// of the last-block's bonded validators.
	LastValidatorPowers []LastValidatorPower `protobuf:"bytes,3,rep,name=last_validator_powers,json=lastValidatorPowers,proto3" json:"last_validator_powers" yaml:"last_validator_powers"`
	// delegations defines the validator set at genesis.
	Validators []Validator `protobuf:"bytes,4,rep,name=validators,proto3" json:"validators"`
	// delegations defines the delegations active at genesis.
	Delegations []Delegation `protobuf:"bytes,5,rep,name=delegations,proto3" json:"delegations"`
	// unbonding_delegations defines the unbonding delegations active at genesis.
	UnbondingDelegations []UnbondingDelegation `protobuf:"bytes,6,rep,name=unbonding_delegations,json=unbondingDelegations,proto3" json:"unbonding_delegations" yaml:"unbonding_delegations"`
	// redelegations defines the redelegations active at genesis.
	Redelegations []Redelegation `protobuf:"bytes,7,rep,name=redelegations,proto3" json:"redelegations"`
	Exported      bool           `protobuf:"varint,8,opt,name=exported,proto3" json:"exported,omitempty"`
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
