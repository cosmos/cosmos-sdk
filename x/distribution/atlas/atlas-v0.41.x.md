# x/distribution

The `x/distribution` module distributes fees and staking rewards to users.

## Usage

1. Import the module.

   ```go
   import (
    distr "github.com/cosmos/cosmos-sdk/x/distribution"
    distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
    distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
    distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
   )
   ```

2. Add `AppModuleBasic` to your `ModuleBasics`.

    ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        distr.AppModuleBasic{},
      }
    )
    ```

3. Give distribution module account permissions.

    ```go
  	// module account permissions
    var maccPerms = map[string][]string{
      distrtypes.ModuleName:          nil,
    }
    ```

4. Allow the distribution module to receive funds.

    ```go
      allowedReceivingModAcc = map[string]bool{
        distrtypes.ModuleName: true,
      }
    ```

5. Add the distribution keeper to your apps struct.

    ```go
      type app struct {
        // ...
        DistrKeeper      distrkeeper.Keeper
        // ...
      }
    ```

6. Add the distribution store key to the group of store keys.

   ```go
   func NewApp(...) *App {
     // ...
      keys := sdk.NewKVStoreKeys(
       distrtypes.StoreKey,
      )
     // ...
   }
   ```

7. Create the keeper. Note the distribution keeper requires an account, bank and staking keeper. This is in order to distribute rewards. 

   ```go
   func NewApp(...) *App {
      // ...
      // create capability keeper with router
      app.DistrKeeper = distrkeeper.NewKeeper(
		    appCodec, keys[distrtypes.StoreKey], app.GetSubspace(distrtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		    &stakingKeeper, authtypes.FeeCollectorName, app.ModuleAccountAddrs(),
	)
   }
   ```

8. Add the `x/distribution` module to the app's `ModuleManager`.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm = module.NewManager(
       // ...
       distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
       // ...
     )
   }
   ```

9. Set the `x/distribution` module begin blocker order.

    ```go
    func NewApp(...) *App {
     // ...
      app.mm.SetOrderBeginBlockers(
        // ...
        distrtypes.ModuleName,
        // ...
      )
    }
    ```

10.  Set the `x/distribution` module genesis order.

   ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(distrtypes.ModuleName,, ...)
   }
   ```

11. Add the `x/distribution` module to the simulation manager (if you have one set).

   ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
       // ...
     )
   }

## Genesis

The `x/distribution` module defines its genesis state as follows:

```proto
// GenesisState defines the distribution module's genesis state.
type GenesisState struct {
	// params defines all the paramaters of the module.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params" yaml:"params"`
	// fee_pool defines the fee pool at genesis.
	FeePool FeePool `protobuf:"bytes,2,opt,name=fee_pool,json=feePool,proto3" json:"fee_pool" yaml:"fee_pool"`
	// fee_pool defines the delegator withdraw infos at genesis.
	DelegatorWithdrawInfos []DelegatorWithdrawInfo `protobuf:"bytes,3,rep,name=delegator_withdraw_infos,json=delegatorWithdrawInfos,proto3" json:"delegator_withdraw_infos" yaml:"delegator_withdraw_infos"`
	// fee_pool defines the previous proposer at genesis.
	PreviousProposer string `protobuf:"bytes,4,opt,name=previous_proposer,json=previousProposer,proto3" json:"previous_proposer,omitempty" yaml:"previous_proposer"`
	// fee_pool defines the outstanding rewards of all validators at genesis.
	OutstandingRewards []ValidatorOutstandingRewardsRecord `protobuf:"bytes,5,rep,name=outstanding_rewards,json=outstandingRewards,proto3" json:"outstanding_rewards" yaml:"outstanding_rewards"`
	// fee_pool defines the accumulated commisions of all validators at genesis.
	ValidatorAccumulatedCommissions []ValidatorAccumulatedCommissionRecord `protobuf:"bytes,6,rep,name=validator_accumulated_commissions,json=validatorAccumulatedCommissions,proto3" json:"validator_accumulated_commissions" yaml:"validator_accumulated_commissions"`
	// fee_pool defines the historical rewards of all validators at genesis.
	ValidatorHistoricalRewards []ValidatorHistoricalRewardsRecord `protobuf:"bytes,7,rep,name=validator_historical_rewards,json=validatorHistoricalRewards,proto3" json:"validator_historical_rewards" yaml:"validator_historical_rewards"`
	// fee_pool defines the current rewards of all validators at genesis.
	ValidatorCurrentRewards []ValidatorCurrentRewardsRecord `protobuf:"bytes,8,rep,name=validator_current_rewards,json=validatorCurrentRewards,proto3" json:"validator_current_rewards" yaml:"validator_current_rewards"`
	// fee_pool defines the delegator starting infos at genesis.
	DelegatorStartingInfos []DelegatorStartingInfoRecord `protobuf:"bytes,9,rep,name=delegator_starting_infos,json=delegatorStartingInfos,proto3" json:"delegator_starting_infos" yaml:"delegator_starting_infos"`
	// fee_pool defines the validator slash events at genesis.
	ValidatorSlashEvents []ValidatorSlashEventRecord `protobuf:"bytes,10,rep,name=validator_slash_events,json=validatorSlashEvents,proto3" json:"validator_slash_events" yaml:"validator_slash_events"`
}
```

## Messages

View supported messages at [docs.cosmos.network/v0.40/modules/Distribution](https://docs.cosmos.network/v0.41/modules/distribution/04_messages.html)

## Client

### CLI

The distribution module supports the blow command to query information located in the modules store.

```sh
app q distribution                   
Querying commands for the distribution module

Usage:
  app query distribution [flags]
  app query distribution [command]

Available Commands:
  commission                    Query distribution validator commission
  community-pool                Query the amount of coins in the community pool
  params                        Query distribution params
  rewards                       Query all distribution delegator rewards or rewards from a particular validator
  slashes                       Query distribution validator slashes
  validator-outstanding-rewards Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations

Flags:
  -h, --help   help for distribution

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

```

### REST

Evidence REST API supports only queries of evidence. To submit evidence please use gRPC or the cli.

### gRPC

Distribution supports both querying and submitting transactions via gRPC

#### Query

[gRPC query](https://docs.cosmos.network/master/core/proto-docs.html#cosmos/evidence/v1beta1/query.proto)

#### Tx

[gRPC Tx](https://docs.cosmos.network/master/core/proto-docs.html#cosmos-distribution-v1beta1-tx-proto)

View supported messages at [docs.cosmos.network/v0.40/modules/distribution](https://docs.cosmos.network/v0.40/modules/distribution/03_messages.html)

### gRPC

Distribution supports both querying and submitting transactions via gRPC

#### Query

[gRPC query](https://docs.cosmos.network/master/core/proto-docs.html#cosmos/distribution/v1beta1/query.proto)

#### Tx

[gRPC Tx](https://docs.cosmos.network/master/core/proto-docs.html#cosmos-distribution-v1beta1-tx-proto)

View supported messages at [docs.cosmos.network/v0.40/modules/distribution](https://docs.cosmos.network/master/modules/distribution/04_messages.html)
