
# x/gov

Gov governs the protocol. There are multiple 

## Usage

1. Import the module.

  ```go
    import (
      "github.com/cosmos/cosmos-sdk/x/gov"
      govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
      govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
    )
  ```

2. Add AppModuleBasic to your ModuleBasics.

  ```go
    var (
      ModuleBasics = module.NewBasicManager(
        // ...
        gov.NewAppModuleBasic(
          paramsclient.ProposalHandler, distrclient.ProposalHandler, upgradeclient.ProposalHandler,upgradeclient.CancelProposalHandler,
        ),
      }
    )
  ```

3. Give gov module account permissions.


  ```go
      // module account permissions
      var maccPerms = map[string][]string{
        govtypes.ModuleName:            {authtypes.Burner},
      }
  ```

4. Add the <module_name> keeper to your apps struct.

  ```go
    type app struct {
      // ...
      GovKeeper        govkeeper.Keeper
      // ...
    }
  ```
5. Add the <module_name> store key to the group of store keys.
 
  ```go
   func NewApp(...) *App {
     // ...
      keys := sdk.NewKVStoreKeys(
       govtypes.StoreKey,
      )
     // ...
   }
  ```
6. Create the gov router. This is the way governance can change parameters of other modules. 

  ```go
  	govRouter := govtypes.NewRouter()
	  govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
    AddRoute()
    // todo add modules that use governance to change parameters here
		// Example: AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
  ```
7. Create the keeper. 

  ```go
   func NewApp(...) *App {
      // ...
    app.GovKeeper = govkeeper.NewKeeper(
      appCodec, keys[govtypes.StoreKey], app.GetSubspace(govtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
      &stakingKeeper, govRouter,
    )
   }
  ```

8. Add the gov module to the app's ModuleManager.

  ```go
   func NewApp(...) *App {
     // ...
     app.mm = module.NewManager(
       // ...
       gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
       // ...
     )
   }
  ```
9. Set the gov module end blocker order.

  ```go
    func NewApp(...) *App {
     // ...
      app.mm.SetOrderEndBlockers(govtypes.ModuleName)
    }
  ```


10.  Set the gov module genesis order.

  ```go
   func NewApp(...) *App {
     // ...
     app.mm.SetOrderInitGenesis(govtypes.ModuleName,, ...)
   }
  ``` 


11. Add the gov module to the simulation manager (if you have one set).

  ```go
   func NewApp(...) *App {
     // ...
     app.sm = module.NewSimulationManager(
       // ...
       distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
       // ...
     )
   }
  ```

## Genesis

```go
type GenesisState struct {
	// starting_proposal_id is the ID of the starting proposal.
	StartingProposalId uint64 `protobuf:"varint,1,opt,name=starting_proposal_id,json=startingProposalId,proto3" json:"starting_proposal_id,omitempty" yaml:"starting_proposal_id"`
	// deposits defines all the deposits present at genesis.
	Deposits Deposits `protobuf:"bytes,2,rep,name=deposits,proto3,castrepeated=Deposits" json:"deposits"`
	// votes defines all the votes present at genesis.
	Votes Votes `protobuf:"bytes,3,rep,name=votes,proto3,castrepeated=Votes" json:"votes"`
	// proposals defines all the proposals present at genesis.
	Proposals Proposals `protobuf:"bytes,4,rep,name=proposals,proto3,castrepeated=Proposals" json:"proposals"`
	// params defines all the paramaters of related to deposit.
	DepositParams DepositParams `protobuf:"bytes,5,opt,name=deposit_params,json=depositParams,proto3" json:"deposit_params" yaml:"deposit_params"`
	// params defines all the paramaters of related to voting.
	VotingParams VotingParams `protobuf:"bytes,6,opt,name=voting_params,json=votingParams,proto3" json:"voting_params" yaml:"voting_params"`
	// params defines all the paramaters of related to tally.
	TallyParams TallyParams `protobuf:"bytes,7,opt,name=tally_params,json=tallyParams,proto3" json:"tally_params" yaml:"tally_params"`
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
