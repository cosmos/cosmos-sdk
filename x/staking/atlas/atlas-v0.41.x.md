
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

Staking has multiple queries and transactions that one can use for interacting with the module.

### CLI

#### Queries

```sh
app q staking           
Querying commands for the staking module

Usage:
  app query staking [flags]
  app query staking [command]

Available Commands:
  delegation                 Query a delegation based on address and validator address
  delegations                Query all delegations made by one delegator
  delegations-to             Query all delegations made to one validator
  historical-info            Query historical info at given height
  params                     Query the current staking parameters information
  pool                       Query the current staking pool values
  redelegation               Query a redelegation record based on delegator and a source and destination validator address
  redelegations              Query all redelegations records for one delegator
  redelegations-from         Query all outgoing redelegatations from a validator
  unbonding-delegation       Query an unbonding-delegation record based on delegator and validator address
  unbonding-delegations      Query all unbonding-delegations records for one delegator
  unbonding-delegations-from Query all unbonding delegatations from a validator
  validator                  Query a validator
  validators                 Query for all validators

Flags:
  -h, --help   help for staking

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app query staking [command] --help" for more information about a command.
```

#### Transactions

```sh
app tx staking                                                                                    
Staking transaction subcommands

Usage:
  app tx staking [flags]
  app tx staking [command]

Available Commands:
  create-validator create new validator initialized with a self-delegation to it
  delegate         Delegate liquid tokens to a validator
  edit-validator   edit an existing validator account
  redelegate       Redelegate illiquid tokens from one validator to another
  unbond           Unbond shares from a validator

Flags:
  -h, --help   help for staking

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app tx staking [command] --help" for more information about a command.
```


### REST

The REST api for slashing can be found here https://cosmos.network/rpc/master under query section. 

### gRPC

gRPC supports both queries and transactions for the staking module. 

#### Query

[gRPC query](https://docs.cosmos.network/master/core/proto-docs.html#cosmos-upgrade-v1beta1-query-proto)
