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