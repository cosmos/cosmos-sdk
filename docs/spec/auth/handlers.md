## Handlers

The auth module presently has no transaction handlers of its own, but does expose
the special `AnteHandler`, used for performing basic validity checks on a transaction,
such that it could be thrown out of the mempool. Note that the ante handler is called on
`CheckTx`, but *also* on `DeliverTx`, as Tendermint proposers presently have the ability
to include in their proposed block transactions which fail `CheckTx`.

### Ante Handler
