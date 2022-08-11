# ADR 60: ABCI++ (Phase I)

## Changelog

* 2022-08-10: Initial Draft (@alexanderbez, @marbar3778)

## Status

Proposed

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
transactions, i.e. passing the AnteHandler, to it's own mempool data structure.
In order to provide a flexible approach to meet the varying needs of application
developers, we will define both a mempool interface and a data structure utilizing
Golang generics that will allow developers to only need to focus on transaction
ordering. Developers requiring absolute full control can implement their own
custom mempool implementation.

We define the general mempool interface as follows (subject to change):

```go
// MempoolTx we define a mempool transaction interface that is as minimal as
// possible, only requiring applications to define the size of the transaction
// to be used when reaping and getting the transaction itself. Interface type
// casting can be used in the actual mempool implementation.
type MempoolTx interface {
	// Size returns the size of the transaction in bytes.
	Size(codec.Codec) int
	Tx() sdk.Tx
}

type Mempool[T MempoolTx] interface {
	// Insert attempts to insert a MempoolTx into the mempool returning an error
	// upon failure.
	Insert(T) error
	// Reap returns the next available transaction in the mempool, returning an error
	// upon failure. The notion of 'available' or 'next' is defined by the application's
	// mempool implementation.
	Reap() (T, error)
	// ReapMaxBytes performs the the same function as Reap, except it returns
	// multiple transactions limited by their total size.
	ReapMaxBytes(n int) ([]T, error)
	// NumTxs returns the number of transactions currently in the mempool.
	NumTxs() int
}
```

We will define an implementation of `Mempool[T MempoolTx]` that will cover a
majority of application use cases. Namely, it will prioritize transactions by
priority and transaction sender, allowing for multiple prioritized transactions
from the same sender. The mempool will be defined as a wrapper around a simple
priority queue using a max binary heap, along with additional indexes/metadata
to store senders and their nonces.

Transaction reaping will essentially happen via a two-phase approach:

1. Reap one or more transactions from the priority queue and collect them into a 
   buffer.
2. For each transaction in the buffer, ensure it does not violate any nonce predicates.
3. Continue reaping until the desired number of valid transactions have been selected.
4. Re-insert any transactions remaining in the buffer.

```go
type PriorityMempool[T MempoolTx] struct {
	queue   *PriorityQueue[MempoolTx]
	senders map[string][]int64
	// ...
}
```

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
block, we must iterate over the list of transactions and ...

We will only populate the `Status` field of the `ProcessProposalResponse` with
`ACCEPT` if ALL the transactions were accepted as valid, otherwise we will
populate with `REJECT`.

<!-- 
## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR. -->

## References

* https://github.com/tendermint/tendermint/blob/master/spec/abci%2B%2B/README.md
* [1] https://github.com/tendermint/tendermint/issues/7750#issuecomment-1076806155
* [2] https://github.com/tendermint/tendermint/issues/7750#issuecomment-1075717151


<!-- ----------------------------------------------------------------------- -->
<!-- ----------------------------------------------------------------------- -->
<!-- ----------------------------------------------------------------------- -->

<!-- 
### Prepare Proposal

There are two designs being considered for `Prepare Proposal`:

#### App-defined Mempool

An app defined mempool is a mempool that lives within the application. Currently Tendermint houses all the ordering and block creation logic. While this abstractions simplifies applications it prohibits many usecases from being achieved. 

The appside mempool is populated via `Checktx`. Tendermint will continue calling `Checktx` for the application to verify the validity of a transaction. When the application receives the transaction, it will add it to its own mempool. While the app is continuing the app will have the option to modify or order transactions within any preference it would like. The ordering of transactions is optional for full nodes. When a block is being proposed, Tendermint will send an array of bytes to the application and ask if the bytes are to be included in the block or replaced with transactions the app would like. At this time the application can replace, modify or delete transcations from the array of bytes. 

##### Pros

- Apps have more fine control over complete tx ordering and inclusion of transactions 
- Remove reliance on the Tendermint's mempool to do ordering

##### Cons

- Added implemention complexity
- Duplicating txs in the mempool 

#### Extra Bytes

The extra bytes approach is where the application/node operator could tell tendermint to send a multiple of the max amount of bytes to be included in a block. When `Prepare` proposal is called Tendermint will provide the app with more transactions that can fit in the block, the Application will say which of the transaction to include and which to discard. While this design removes the complexity of defining a mempool in the application, the application does not have complete control over all the transactions in the mempool. 

##### Pros

- App complexity is low
- Not duplicating mempool txs in the application 

##### Cons

- App has control over a subset of all transactions in the mempool


### Process Proposal

At the process proposal phase Tendermint will ask validators to verify the contents of a block. To do this, Tendermint will send the contents of the proposed block to the application. At this stage the application will verify the content of the block in a stateless manner. If the content fails the stateless checks, the application will have the option to tell Tendermint to vote against the block. 

## Decision

> This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..."
> {decision body}

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

* {reference link} -->
