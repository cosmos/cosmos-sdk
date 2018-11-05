# Simple Governance Module

## Module initialization 

First, let us go into the module's folder and create a folder for our module.

```bash
cd x/
mkdir simple_governance
cd simple_governance
mkdir -p client/cli client/rest
touch client/cli/simple_governance.go client/rest/simple_governance.go errors.go handler.go handler_test.go keeper_keys.go keeper_test.go keeper.go test_common.go test_types.go types.go codec.go
```

Let us start by adding the files we will need. Your module's folder should look something like that:

```
x
└─── simple_governance
      ├─── client
      │     ├───  cli
      │     │     └─── simple_governance.go
      │     └─── rest
      │           └─── simple_governance.go
      ├─── errors.go
      ├─── handler.go
      ├─── keeper_keys.go
      ├─── keeper.go
      ├─── types.go
      └─── codec.go
```

Let us go into the detail of each of these files.

## Types 

**File: [`x/simple_governance/types.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/types.go)**

In this file, we define the custom types for our module. This includes the types from the [State](app-design.md#State) section and the custom message types defined in the [Messages](app-design#Messages) section.

For each new type that is not a message, it is possible to add methods that make sense in the context of the application. In our case, we will implement an `updateTally` function to easily update the tally of a given proposal as vote messages come in.

Messages are a bit different. They implement the `Message` interface defined in the SDK's `types` folder. Here are the methods you need to implement when you define a custom message type:

- `Type()`: This function returns the name of our module's route. When messages are processed by the application, they are routed using the string returned by the `Type()` method.
- `GetSignBytes()`: Returns the byte representation of the message. It is used to sign the message.
- `GetSigners()`: Returns address(es) of the signer(s).
- `ValidateBasic()`: This function is used to discard obviously invalid messages. It is called at the beginning of `runTx()` in the baseapp file. If `ValidateBasic()` does not return `nil`, the app stops running the transaction.
- `Get()`: A basic getter, returns some property of the message.
- `String()`: Returns a human-readable version of the message

For our simple governance messages, this means:

- `Type()` will return `"simpleGov"`
- For `SubmitProposalMsg`, we need to make sure that the attributes are not empty and that the deposit is both valid and positive. Note that this is only basic validation, we will therefore not check in this method that the sender has sufficient funds to pay for the deposit
- For `VoteMsg`, we check that the address and option are valid and that the proposalID is not negative.
- As for other methods, less customization is required. You can check the code to see a standard way of implementing these.

## Keeper

**File: [`x/simple_governance/keeper.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/keeper.go)**

### Short intro to keepers

`Keepers` are a module abstraction that handle reading/writing to the module store. This is a practical implementation of the **Object Capability Model** for Cosmos. 


As module developers, we have to define keepers to interact with our module's store(s) not only from within our module, but also from other modules. When another module wants to access one of our module's store(s), a keeper for this store has to be passed to it at the application level. In practice, it will look like that:

```go
// in app.go

// instanciate keepers
keeperA = moduleA.newKeeper(app.moduleAStoreKey)
keeperB = moduleB.newKeeper(app.moduleBStoreKey)

// pass instance of keeperA to handler of module B
app.Router().
        AddRoute("moduleA", moduleA.NewHandler(keeperA)).
        AddRoute("moduleB", moduleB.NewHandler(keeperB, keeperA))   // Here module B can access one of module A's store via the keeperA instance
```

`KeeperA` grants a set of capabilities to the handler of module B. When developing a module, it is good practice to think about the sensitivity of the different capabilities that can be granted through keepers. For example, some module may need to read and write to module A's main store, while others only need to read it. If a module has multiple stores, then some keepers could grant access to all of them, while others would only grant access to specific sub-stores. It is the job of the module developer to make sure it is easy for  application developers to instanciate a keeper with the right capabilities. Of course, the handler of a module will most likely get an unrestricted instance of that module's keeper.

### Store for our app

Before we delve into the keeper itself, let us see what objects we need to store in our governance sub-store, and how to index them.

- `Proposals` will be indexed by `'proposals'|<proposalID>`.
- `Votes` (`Yes`, `No`, `Abstain`) will be indexed by `'proposals'|<proposalID>|'votes'|<voterAddress>`.

Notice the quote mark on `'proposals'` and `'votes'`. They indicate that these are constant keywords. So, for example, the option casted by voter with address `0x01` on proposal `0101` will be stored at index `'proposals'|0101|'votes'|0x01`.

These keywords are used to faciliate range queries. Range queries (TODO: Link to formal spec) allow developer to query a subspace of the store, and return an iterator. They are made possible by the nice properties of the [IAVL+ tree](https://github.com/tendermint/iavl) that is used in the background. In practice, this means that it is possible to store and query a Key-Value pair in O(1), while still being able to iterate over a given subspace of Key-Value pairs. For example, we can query all the addresses that voted on a given proposal, along with their votes, by calling `rangeQuery(SimpleGovStore, <proposalID|'addresses'>)`.

### Keepers for our app

In our case, we only have one store to access, the `SimpleGov` store. We will need to set and get values inside this store via our keeper. However, these two actions do not have the same impact in terms of security. While there should no problem in granting read access to our store to other modules, write access is way more sensitive. So ideally application developers should be able to create either a governance mapper that can only get values from the store, or one that can both get and set values. To this end, we will introduce two keepers: `Keeper` and `KeeperRead`. When application developers create their application, they will be able to decide which of our module's keeper to use.

Now, let us try to think about which keeper from **external** modules our module's keepers need access to.
Each proposal requires a deposit. This means our module needs to be able to both read and write to the module that handles tokens, which is the `bank` module. We also need to be able to determine the voting power of each voter based on their stake. To this end, we need read access to the store of the `staking` module. However, we don't need write access to this store. We should therefore indicate that in our module, and the application developer should be careful to only pass a read-only keeper of the `staking` module to our module's handler.

With all that in mind, we can define the structure of our `Keeper`:

```go
    type Keeper struct {
        SimpleGov    sdk.StoreKey        // Key to our module's store
        cdc                 *codec.Codec         // Codec to encore/decode structs
        ck                  bank.Keeper         // Needed to handle deposits. This module onlyl requires read/writes to Atom balance
        sm                  stake.Keeper        // Needed to compute voting power. This module only needs read access to the staking store.
        codespace           sdk.CodespaceType   // Reserves space for error codes
    }
```

And the structure of our `KeeperRead`:

```go
type KeeperRead struct {
    Keeper
}
```

`KeeperRead` will inherit all methods from `Keeper`, except those that we override. These will be the methods that perform writes to the store.

### Functions and Methods

The first function we have to create is the constructor.

```go
func NewKeeper(SimpleGov sdk.StoreKey, ck bank.Keeper, sm stake.Keeper, codespace sdk.CodespaceType) Keeper
```

This function is called from the main [`app.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/app/app.go) file to instanciate a new `Keeper`. A similar function exits for `KeeperRead`.

```go
func NewKeeperRead(SimpleGov sdk.StoreKey, ck bank.Keeper, sm stake.Keeper, codespace sdk.CodespaceType) KeeperRead
```

Depending on the needs of the application and its modules, either `Keeper`, `KeeperRead`, or both, will be instanciated at application level.

*Note: Both the `Keeper` type name and `NewKeeper()` function's name are standard names used in every module. It is no requirement to follow this standard, but doing so can facilitate the life of application developers*

Now, let us describe the methods we need for our module's `Keeper`. For the full implementation, please refer to `keeper.go`.

- `GetProposal`: Get a `Proposal` given a `proposalID`. Proposals need to be decoded from `byte` before they can be read.
- `SetProposal`: Set a `Proposal` at index `'proposals'|<proposalID>`. Proposals need to be encoded to `byte` before they can be stored.
- `NewProposalID`: A function to generate a new unique `proposalID`.
- `GetVote`: Get a vote `Option` given a `proposalID` and a `voterAddress`.
- `SetVote`: Set a vote `Option` given a `proposalID` and a `voterAddress`.
- Proposal Queue methods: These methods implement a standard proposal queue to store `Proposals` on a First-In First-Out basis. It is used to tally the votes at the end of the voting period.

The last thing that needs to be done is to override certain methods for the `KeeperRead` type. `KeeperRead` should not have write access to the stores. Therefore, we will override the methods `SetProposal()`, `SetVote()` and `NewProposalID()`, as well as `setProposalQueue()` from the Proposal Queue's methods. For `KeeperRead`, these methods will just throw an error.

*Note: If you look at the code, you'll notice that the context `ctx` is a parameter of many of the methods. The context `ctx` provides useful information on the current state such as the current block height and allows the keeper `k` to access the `KVStore`. You can check all the methods of `ctx` [here](https://github.com/cosmos/cosmos-sdk/blob/develop/types/context.go#L144-L168)*.

## Handler 

**File: [`x/simple_governance/handler.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/handler.go)**

### Constructor and core handlers

Handlers implement the core logic of the state-machine. When a transaction is routed from the app to the module, it is run by the `handler` function.

In practice, one `handler` will be implemented for each message of the module. In our case, we have two message types. We will therefore need two `handler` functions. We will also need a constructor function to route the message to the correct `handler`:

```go
func NewHandler(k Keeper) sdk.Handler {
    return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
        switch msg := msg.(type) {
        case SubmitProposalMsg:
            return handleSubmitProposalMsg(ctx, k, msg)
        case VoteMsg:
            return handleVoteMsg(ctx, k, msg)
        default:
            errMsg := "Unrecognized gov Msg type: " + reflect.TypeOf(msg).Name()
            return sdk.ErrUnknownRequest(errMsg).Result()
        }
    }
}
```

The messages are routed to the appropriate `handler` depending on their type. For our simple governance module, we only have two `handlers`, that correspond to our two message types. They have similar signatures:

```go
func handleSubmitProposalMsg(ctx sdk.Context, k Keeper, msg SubmitProposalMsg) sdk.Result
```

Let us take a look at the parameters of this function:

- The context `ctx` to access the stores.
- The keeper `k` allows the handler to read and write from the different stores, including the module's store (`SimpleGovernance` in our case) and all the stores from other modules that the keeper `k` has been granted an access to (`stake` and `bank` in our case).
- The message `msg` that holds all the information provided by the sender of the transaction.

The function returns a `Result` that is returned to the application. It contains several useful information such as the amount of `Gas` for this transaction and wether the message was succesfully processed or not. At this point, we exit the boundaries of our simple governance module and go back to root application level. The `Result` will differ from application to application. You can check the `sdk.Result` type directly [here](https://github.com/cosmos/cosmos-sdk/blob/develop/types/result.go) for more info.

### BeginBlocker and EndBlocker

In contrast to most smart-contracts platform, it is possible to perform automatic (i.e. not triggered by a transaction sent by an end-user) execution of logic in Cosmos-SDK applications.

This automatic execution of code takes place in the `BeginBlock` and `EndBlock` functions that are called at the beginning and at the end of every block. They are powerful tools, but it is important for application developers to be careful with them. For example, it is crutial that developers control the amount of computing that happens in these functions, as expensive computation could delay the block time, and never-ending loop freeze the chain altogether.

`BeginBlock` and `EndBlock` are composable functions, meaning that each module can implement its own `BeginBlock` and `EndBlock` logic. When needed, `BeginBlock` and `EndBlock` logic is implemented in the module's `handler`. Here is the standard way to proceed for `EndBlock` (`BeginBlock` follows the exact same pattern):

```go
func NewEndBlocker(k Keeper) sdk.EndBlocker {
    return func(ctx sdk.Context, req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
        err := checkProposal(ctx, k)
        if err != nil {
            panic(err)
        }
        return
    }
}
```

Do not forget that each module need to declare its `BeginBlock` and `EndBlock` constructors at application level. See the [Application - Bridging it all together](app-structure.md).

For the purpose of our simple governance application, we will use `EndBlock` to automatically tally the results of the vote. Here are the different steps that will be performed:

1. Get the oldest proposal from the `ProposalProcessingQueue`
2. Check if the `CurrentBlock` is the block at which the voting period for this proposal ends. If Yes, go to 3.. If no, exit.
3. Check if proposal is accepted or rejected. Update the proposal status.
4. Pop the proposal from the `ProposalProcessingQueue` and go back to 1.

Let us perform a quick safety analysis on this process.
- The loop will not run forever because the number of proposals in `ProposalProcessingQueue` is finite
- The computation should not be too expensive because tallying of individual proposals is not expensive and the number of proposals is expected be relatively low. That is because proposals require a `Deposit` to be accepted. `MinDeposit` should be high enough so that we don't have too many `Proposals` in the queue.
- In the eventuality that the application becomes so successful that the `ProposalProcessingQueue` ends up containing so many proposals that the blockchain starts slowing down, the module should be modified to mitigate the situation. One clever way of doing it is to cap the number of iteration per individual `EndBlock` at `MaxIteration`. This way, tallying will be spread over many blocks if the number of proposals is too important and block time should remain stable. This would require to modify the current check `if (CurrentBlock == Proposal.SubmitBlock + VotingPeriod)` to `if (CurrentBlock > Proposal.SubmitBlock + VotingPeriod) AND (Proposal.Status == ProposalStatusActive)`.

## Codec

**File: [`x/simple_governance/codec.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/codec.go)**

The `codec.go` file allows developers to register the concrete message types of their module into the codec. In our case, we have two messages to declare:

```go
func RegisterCodec(cdc *codec.Codec) {
    cdc.RegisterConcrete(SubmitProposalMsg{}, "simple_governance/SubmitProposalMsg", nil)
    cdc.RegisterConcrete(VoteMsg{}, "simple_governance/VoteMsg", nil)
}
```
Don't forget to call this function in `app.go` (see [Application - Bridging it all together](app-structure.md)) for more).

## Errors 

**File: [`x/simple_governance/errors.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/errors.go)**

The `error.go` file allows us to define custom error messages for our module.  Declaring errors should be relatively similar in all modules. You can look in the `error.go` file directly for a concrete example. The code is self-explanatory.

Note that the errors of our module inherit from the `sdk.Error` interface and therefore possess the method `Result()`. This method is useful when there is an error in the `handler` and an error has to be returned in place of an actual result.

## Command-Line Interface

**File: [`x/simple_governance/client/cli/simple_governance.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/client/cli/simple_governance.go)**

Go in the `cli` folder and create a `simple_governance.go` file. This is where we will define the commands for our module.

The CLI builds on top of [Cobra](https://github.com/spf13/cobra). Here is the schema to build a command on top of Cobra:

```go
    // Declare flags
    const(
        Flag = "flag"
        ...
    )

    // Main command function. One function for each command.
    func Command(codec *codec.Codec) *cobra.Command {
        // Create the command to return
        command := &cobra.Command{
            Use: "actual command",
            Short: "Short description",
            Run: func(cmd *cobra.Command, args []string) error {
                // Actual function to run when command is used
            },
        }

        // Add flags to the command
        command.Flags().<Type>(FlagNameConstant, <example_value>, "<Description>")

        return command
    }
```

## Rest API

**File: [`x/simple_governance/client/rest/simple_governance.goo`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/client/rest/simple_governance.go)**

The Rest Server, also called [Light-Client Daemon (LCD)](https://github.com/cosmos/cosmos-sdk/tree/master/client/lcd), provides support for **HTTP queries**.

________________________________________________________

USER INTERFACE <=======> REST SERVER <=======> FULL-NODE

________________________________________________________

It allows end-users that do not want to run full-nodes themselves to interract with the chain. The LCD can be configured to perform **Light-Client verification** via the flag `--trust-node`, which can be set to `true` or `false`.

- If *light-client verification* is enabled, the Rest Server acts as a light-client and needs to be run on the end-user's machine. It allows them to interract with the chain in a trustless way without having to store the whole chain locally.

- If *light-client verification* is disabled, the Rest Server acts as a simple relayer for HTTP calls. In this setting, the Rest server needs not be run on the end-user's machine. Instead, it will probably be run by the same entity that operates the full-node the server connects to. This mode is useful if end-users trust the full-node operator and do not want to store anything locally.

Now, let us define endpoints that will be available for users to query through HTTP requests. These endpoints will be defined in a `simple_governance.go` file stored in the `rest` folder.

| Method | URL                             | Description                                                 |
|--------|---------------------------------|-------------------------------------------------------------|
| GET    | /proposals                      | Range query to get all submitted proposals                  |
| POST   | /proposals                      | Submit a new proposal                                       |
| GET    | /proposals/{id}                 | Returns a proposal given its ID                             |
| GET    | /proposals/{id}/votes           | Range query to get all the votes casted on a given proposal |
| POST   | /proposals/{id}/votes           | Cast a vote on a given proposal                             |
| GET    | /proposals/{id}/votes/{address} | Returns the vote of a given address on a given proposal     |

It is the job of module developers to provide sensible endpoints so that front-end developers and service providers can properly interact with it.

Additionaly, here is a [link](https://hackernoon.com/restful-api-designing-guidelines-the-best-practices-60e1d954e7c9) for REST APIs best practices.
