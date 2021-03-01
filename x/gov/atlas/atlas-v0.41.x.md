
# x/gov

Gov governs the protocol. There are multiple proposals one can make. 

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

4. Add the gov keeper to your apps struct.

  ```go
    type app struct {
      // ...
      GovKeeper        govkeeper.Keeper
      // ...
    }
  ```
5. Add the gov store key to the group of store keys.
 
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

Gov supports cli, REST and gRPC for all queries and messages

### CLI

CLI support for the gov module is for both transactions and queries. Below you will see a print out of the commands. 

#### Queries

```sh
app q gov      
Querying commands for the governance module

Usage:
  app query gov [flags]
  app query gov [command]

Available Commands:
  deposit     Query details of a deposit
  deposits    Query deposits on a proposal
  param       Query the parameters (voting|tallying|deposit) of the governance process
  params      Query the parameters of the governance process
  proposal    Query details of a single proposal
  proposals   Query proposals with optional filters
  proposer    Query the proposer of a governance proposal
  tally       Get the tally of a proposal vote
  vote        Query details of a single vote
  votes       Query votes on a proposal

Flags:
  -h, --help   help for gov

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app query gov [command] --help" for more information about a command.
```

#### Transactions

```sh
app tx gov
Governance transactions subcommands

Usage:
  app tx gov [flags]
  app tx gov [command]

Available Commands:
  deposit         Deposit tokens for an active proposal
  submit-proposal Submit a proposal along with an initial deposit
  vote            Vote for an active proposal, options: yes/no/no_with_veto/abstain
  weighted-vote   Vote for an active proposal, options: yes/no/no_with_veto/abstain

Flags:
  -h, --help   help for gov

Global Flags:
      --chain-id string     The network chain ID
      --home string         directory for config and data (default "/Users/markobaricevic/.simapp")
      --log_format string   The logging format (json|plain) (default "plain")
      --log_level string    The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
      --trace               print out full stack trace on errors

Use "app tx gov [command] --help" for more information about a command.
```


### REST

The rest api endpoints can be found here https://cosmos.network/rpc/master under the governance section.

### gRPC

Gov supports both queries and transactions for gRPC. 

#### Query

[gRPC query](https://docs.cosmos.network/master/core/proto-docs.html#cosmos-gov-v1beta1-query-proto)

#### Tx

[gRPC Tx](https://docs.cosmos.network/master/core/proto-docs.html#cosmos-gov-v1beta1-tx-proto)


View supported messages at [docs.cosmos.network/v0.41/modules/gov](https://docs.cosmos.network/v0.41/modules/gov/03_messages.html)
