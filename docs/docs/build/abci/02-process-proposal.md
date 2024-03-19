# Process Proposal

`ProcessProposal` handles the validation of a proposal from `PrepareProposal`,
which also includes a block header. Meaning, that after a block has been proposed
the other validators have the right to vote on a block. The validator in the
default implementation of `PrepareProposal` runs basic validity checks on each
transaction.

Note, `ProcessProposal` MAY NOT be non-deterministic, i.e. it must be deterministic.
This means if `ProcessProposal` panics or fails and we reject, all honest validator
processes will prevote nil and the CometBFT round will proceed again until a valid
proposal is proposed.

Here is the implementation of the default implementation:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.50.0-alpha.0/baseapp/abci_utils.go#L153-L159
```

Like `PrepareProposal` this implementation is the default and can be modified by
the application developer in [`app.go`](./01-app-go-v2.md). If you decide to implement
your own `ProcessProposal` handler, you must be sure to ensure that the transactions
provided in the proposal DO NOT exceed the maximum block gas and `maxtxbytes` (if set).

```go
processOpt := func(app *baseapp.BaseApp) {
abciPropHandler := baseapp.NewDefaultProposalHandler(mempool, app)
app.SetProcessProposal(abciPropHandler.ProcessProposalHandler())
}

baseAppOptions = append(baseAppOptions, processOpt)
```
