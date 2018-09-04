# ABCI

The Application BlockChain Interface, or ABCI, is a powerfully
delineated boundary between the Cosmos-SDK and Tendermint.
It separates the logical state transition machine of your application from
its secure replication across many physical machines.

By providing a clear, language agnostic boundary between applications and consensus,
ABCI provides tremendous developer flexibility and [support in many
languages](https://tendermint.com/ecosystem). That said, it is still quite a low-level protocol, and
requires frameworks to be built to abstract over that low-level componentry.
The Cosmos-SDK is one such framework.

While we've already seen `DeliverTx`, the workhorse of any ABCI application,
here we will introduce the other ABCI requests sent by Tendermint, and
how we can use them to build more advanced applications. For a more complete
depiction of the ABCI and how its used, see
[the
specification](https://github.com/tendermint/tendermint/blob/master/docs/app-dev/abci-spec.md)

## InitChain

In our previous apps, we built out all the core logic, but we never specified
how the store should be initialized. For that, we use the `app.InitChain` method,
which is called once by Tendermint the very first time the application boots up.

The InitChain request contains a variety of Tendermint information, like the consensus
parameters and an initial validator set, but it also contains an opaque blob of
application specific bytes - typically JSON encoded.
Apps can decide what to do with all of this information by calling the
`app.SetInitChainer` method.

For instance, let's introduce a `GenesisAccount` struct that can be JSON encoded
and part of a genesis file. Then we can populate the store with such accounts
during InitChain:

```go
TODO
```

If we include a correctly formatted `GenesisAccount` in our Tendermint
genesis.json file, the store will be initialized with those accounts and they'll
be able to send transactions!

## BeginBlock

BeginBlock is called at the beginning of each block, before processing any
transactions with DeliverTx.
It contains information on what validators have signed.

## EndBlock

EndBlock is called at the end of each block, after processing all transactions
with DeliverTx.
It allows the application to return updates to the validator set.

## Commit

Commit is called after EndBlock. It persists the application state and returns
the Merkle root hash to be included in the next Tendermint block. The root hash
can be in Query for Merkle proofs of the state.

## Query

Query allows queries into the application store according to a path.

## CheckTx

CheckTx is used for the mempool. It only runs the AnteHandler. This is so
potentially expensive message handling doesn't begin until the transaction has
actually been committed in a block. The AnteHandler authenticates the sender and
ensures they have enough to pay the fee for the transaction. If the transaction
later fails, the sender still pays the fee.
