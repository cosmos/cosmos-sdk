# ADR 60: ABCI++ (Phase I)

## Changelog

* 2022-08-10: Initial Draft (@alexanderbez, @marbar3778)

## Status

PROPOSED

## Abstract

This ADR describes the initial adoption of [ABCI++](https://github.com/tendermint/tendermint/blob/master/spec/abci%2B%2B/README.md),
the next evolution of ABCI, within the Cosmos SDK. ABCI++ aims to provide
application developers with more flexibility and control over application and
consensus semantics, e.g. in-application mempools, in-process oracles, and
order-book style matching engines.

## Context

Tendermint will release ABCI++ in phases. Notably, at the time of this writing,
Tendermint is releasing v0.37.0 which will include `PrepareProposal` and `ProcessProposal`.

The `PrepareProposal` ABCI method is concerned with a block proposer requesting
the application to evaluate a series of transactions to be included in the next
block, defined as a slice of `TxRecord` objects. The application can either
accept, reject, or completely ignore some or all of these transactions. This is
an important consideration to make as the application can essentially define and
control its own mempool allowing it to define sophisticated transaction priority
and filtering mechanisms, by completely ignoring the `TxRecords` Tendermint
sends it, favoring its own transactions. This essentially means that the Tendermint
mempool acts more like a gossip data structure.

The second ABCI method, `ProcessProposal`, is used to process the block proposer's
proposal as defined by `PrepareProposal`. This ABCI method requests that the
application evaluate the entire proposed block for validity.

It is important to note that in this phase of ABCI++ integration, the application
is NOT responsible for locking semantics -- Tendermint will still be responsible
for that. In the future, however, the application will be responsible for locking,
which allows for parallel execution possibilities.

## Decision

We will integrate the first phase of ABCI++, which will be introduced in Tendermint
v0.37.0, in the next major release of the Cosmos SDK. We will integrate the two
aforementioned ABCI++ methods on the `BaseApp` type. We describe the implementations
of the two methods individually below.

Prior to describing the implementation of the two new methods, it is important to
note that the existing ABCI methods, `CheckTx`, `DeliverTx`, etc, still exist and
serve the same functions as they do now.

### `PrepareProposal`

Prior to evaluating the decision for how to implement `PrepareProposal`, it is
important to note that `CheckTx` will still be executed and will be responsible
for evaluating transaction validity as it does now, with one very important
_additive_ distinction.

When executing transactions in `CheckTx`, the application will now add valid
transactions, i.e. passing the AnteHandler, to its own mempool data structure.
In order to provide a flexible approach to meet the varying needs of application
developers, we will define both a mempool interface and a data structure utilizing
Golang generics, allowing developers to focus only on transaction
ordering. Developers requiring absolute full control can implement their own
custom mempool implementation.

> Note, there will most likely be a need to modify the AnteHandler method signature
> to now accept a series of TxRecords.

We define the general mempool interface as follows (subject to change):

```go
// MempoolTx we define an app-side mempool transaction interface that is as
// minimal as possible, only requiring applications to define the size of the
// transaction to be used when reaping and getting the transaction itself.
// Interface type casting can be used in the actual app-side mempool implementation.
type MempoolTx interface {
	// Size returns the size of the transaction in bytes.
	Size(codec.Codec) int
	Tx() sdk.Tx
}

// PrepareTxRecord defines a wrapper around a MempoolTx that is returned from
// PrepareProposal which includes an Action to inform Tendermint what to do with
// the transaction.
type PrepareTxRecord[T MempoolTx] struct {
	Tx T
	Action abci.TxAction
}

type Mempool[T MempoolTx] interface {
	// Insert attempts to insert a MempoolTx into the app-side mempool returning
	// an error upon failure.
	Insert(T) error
	// ReapMaxBytes returns the next set of available transactions from the app-side
	// mempool, up to maxBytes or until the mempool is empty. The application can
	// decide to return transactions from its own mempool or from the incoming
	// TxRecords or some combination of both. The notion of 'available' or 'next'
	// is defined by the application's mempool implementation.
	ReapMaxBytes(txRecords abci.TxRecords, maxBytes int) ([]PrepareTxRecord[T], error)
	// NumTxs returns the number of transactions currently in the mempool.
	NumTxs() int
}
```

We will define an implementation of `Mempool[T MempoolTx]` that will cover a
majority of application use cases. Namely, it will prioritize transactions by
priority and transaction sender, allowing for multiple prioritized transactions
from the same sender. The app-side mempool will be defined as a wrapper around a
simple priority queue using a max binary heap, along with additional indexes/metadata
to store senders and their nonces, allowing for simple multi-dimensional
prioritization (2-ary).

Transaction reaping will essentially happen via a two-phase approach:

1. Reap one or more transactions from the priority queue and collect them into 
   one of two buffers -- _valid_ and _invalid_.
2. For transactions that DO NOT violate the nonce validation, they are included
   in the _valid_ buffer.
3. For transactions that DO violate the nonce validation, they are included in
   the _invalid_ buffer.
4. Continue this process until the desired number of valid transactions are
   reaped or until the mempool is empty.
5. Provide Tendermint the list of all transactions from the _valid_ buffer.
6. Re-insert all transactions, from both buffers, back into app-side mempool.
   This is to ensure we do not discard transactions from the app-side mempool in
	 case `ProcessProposal` fails or in case that the proposal, while passing
	 `ProcessProposal` is not the one decided for that height, i.e. the height took
	 more than one round.

```go
type PriorityMempool[T MempoolTx] struct {
	queue   *PriorityQueue[MempoolTx]
	senders map[string][]int64
	// ...
}
```

> The `PriorityMempool[T MempoolTx]` implementation will support Options such as
> limiting the mempool size by a fixed number of bytes.

Previous discussions<sup>1</sup> have come to the agreement that Tendermint will
perform a request to the application, via `PrepareProposalRequest`, with a certain
amount of transactions reaped from Tendermint's local mempool. The exact amount
of transactions reaped will be determined by a local operator configuration.
This is referred to as the "one-shot approach" seen in discussions.

When Tendermint reaps transactions from the local mempool and sends them to the
application via `PrepareProposalRequest`, the application will have to evaluate
the transactions. Specifically, it will need to inform Tendermint if it should
reject and or include each transaction. Note, the application can even _replace_
transactions entirely with other transactions.

When evaluating transactions from `PrepareProposalRequest`, the application will
ignore _all_ transactions sent to it in the request and instead reap up to
`PrepareProposalRequest.max_tx_bytes` from it's own mempool. There is no need to
execute the transactions for validity as they have already passed CheckTx.

### `ProcessProposal`

The `ProcessProposal` ABCI method is relatively straightforward. It is responsible
for ensuring validity of the proposed block containing transactions that were
selected from the `PrepareProposal` step. In order to check validity of the proposed
block, we must iterate over the list of transactions and execute them using the
same mode/execution strategy as `CheckTx`. In other words, we execute each transaction
using the AnteHandler only -- no messages are executed. However, we cannot just
execute `CheckTx` again, because `BaseApp` already has a modified `checkState` at
this point. So when executing `ProcessProposal`, we create a similar branched
state, `processProposalState`, off of `deliverState`. Using `processProposalState`
we execute the AnteHandler for each transaction. Note, the `processProposalState`
is never committed and is completely discarded after `ProcessProposal` completes.

We will only populate the `Status` field of the `ProcessProposalResponse` with
`ACCEPT` if ALL the transactions were accepted as valid, otherwise we will
populate with `REJECT`.

### `DeliverTx`

Since transactions are not truly removed from the app-side mempool during
`PrepareProposal`, since `ProcessProposal` can fail and we do not want to lose
transactions, we need to finally remove the transaction from the app-side mempool
during `DeliverTx` since during this phase, the transactions are being included
in the proposed block.

Alternatively, we can keep the transactions as truly being removed during the
reaping phase in `PrepareProposal` and add them back to the app-side mempool in
case `ProcessProposal` fails.

## Consequences

### Backwards Compatibility

ABCI++ is naturally not backwards compatible with prior versions of the Cosmos SDK
and Tendermint. For example, an application that requests `PrepareProposalRequest`
to the same application that does not speak ABCI++ will naturally fail.

However, in the first phase of the integration, the existing ABCI methods as we
know them today will still exist and function as they currently do.

### Positive

* Applications now have full control over transaction ordering and priority.
* Lays the groundwork for the full integration of ABCI++, which will unlock more
  app-side use cases around block construction and integration with the Tendermint
  consensus engine.

### Negative

* Requires that the "mempool", as a general data structure that collects and stores
  uncommitted transactions will be duplicated between both Tendermint and the
  Cosmos SDK.
* Additional requests between Tendermint and the Cosmos SDK in the context of
  block execution. Albeit, the overhead should be negligible.

### Neutral

## Further Discussions

It is possible to design the app-side implementation of the `Mempool[T MempoolTx]`
in many different ways using different data structures and implementations. All
of which have different tradeoffs. The proposed solution keeps things simple
and covers cases that would be required for most basic applications. There are
tradeoffs that can be made to improve performance of reaping and inserting into
the provided mempool implementation.

## References

* https://github.com/tendermint/tendermint/blob/master/spec/abci%2B%2B/README.md
* [1] https://github.com/tendermint/tendermint/issues/7750#issuecomment-1076806155
* [2] https://github.com/tendermint/tendermint/issues/7750#issuecomment-1075717151
