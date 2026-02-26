# Process Proposal

`ProcessProposal` handles the validation of a proposal from `PrepareProposal`,
which also includes a block header. After a block has been proposed,
the other validators have the right to accept or reject that block. The validator in the
default implementation of `PrepareProposal` runs basic validity checks on each
transaction.

Note, `ProcessProposal` MUST be deterministic. Non-deterministic behaviors will cause apphash mismatches.
This means if `ProcessProposal` panics or fails and we reject, all honest validator
processes should reject (i.e., prevote nil). If so, CometBFT will start a new round with a new block proposal and the same cycle will happen with `PrepareProposal`
and `ProcessProposal` for the new proposal.

Here is the implementation of the default implementation:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/baseapp/abci_utils.go#L219-L226
```

Like `PrepareProposal`, this implementation is the default and can be modified by
the application developer in [`app_di.go`](../building-apps/01-app-go-di.md). If you decide to implement
your own `ProcessProposal` handler, you must ensure that the transactions
provided in the proposal DO NOT exceed the maximum block gas and `maxtxbytes` (if set).

```go
processOpt := func(app *baseapp.BaseApp) {
    abciPropHandler := baseapp.NewDefaultProposalHandler(mempool, app)
    app.SetProcessProposal(abciPropHandler.ProcessProposalHandler())
}

baseAppOptions = append(baseAppOptions, processOpt)
```
